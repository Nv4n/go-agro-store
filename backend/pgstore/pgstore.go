package pgstore

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

// PGStore represents the configured session store.
type PGStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Pool    *pgxpool.Pool
}

// PGSession represents a session record in the database.
type PGSession struct {
	ID         int64
	Key        []byte
	Data       []byte
	CreatedOn  time.Time
	ModifiedOn time.Time
	ExpiresOn  time.Time
}

// NewPGStore creates a new PGStore instance with a pgxpool connection.
// It also creates the necessary schema in the database.
func NewPGStore(dbURL string, keyPairs ...[]byte) (*PGStore, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	return NewPGStoreFromPool(pool, keyPairs...)
}

// NewPGStoreFromPool creates a PGStore using an existing pgxpool.Pool and creates the schema.
func NewPGStoreFromPool(pool *pgxpool.Pool, keyPairs ...[]byte) (*PGStore, error) {
	store := &PGStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		Pool: pool,
	}

	// Create table if it doesn't exist.
	if err := store.createSessionsTable(); err != nil {
		return nil, err
	}

	return store, nil
}

// Close terminates the pgxpool connection.
func (store *PGStore) Close() {
	store.Pool.Close()
}

// Get fetches a session for a given name after it has been added to the registry.
func (store *PGStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(store, name)
}

// New returns a new session for the given name without adding it to the registry.
func (store *PGStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(store, name)
	if session == nil {
		return nil, nil
	}

	// Copy store options into the session.
	opts := *store.Options
	session.Options = &opts
	session.IsNew = true

	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		// Decode the session key from the cookie.
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, store.Codecs...)
		if err == nil {
			err = store.load(session)
			if err == nil {
				session.IsNew = false
			} else if errors.Is(err, pgx.ErrNoRows) {
				// No rows found; session remains new.
				err = nil
			}
		}
	}

	store.MaxAge(store.Options.MaxAge)
	return session, err
}

// Save writes the session into the database and sets the cookie.
func (store *PGStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// If MaxAge < 0, delete the session.
	if session.Options.MaxAge < 0 {
		if err := store.destroy(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		// Generate a random session ID key.
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			), "=")
	}

	if err := store.save(session); err != nil {
		return err
	}

	// Encode the session ID for cookie storage.
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.Codecs...)
	if err != nil {
		return err
	}

	fmt.Println("HERE MF")
	log.Println("HERE AGAIN")
	slog.Info(fmt.Sprintf("SAMESITE %v", session.Options.SameSite))
	slog.Info(fmt.Sprintf("OPTIONS %v", session.Options))
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxLength restricts the maximum length of new sessions.
func (store *PGStore) MaxLength(l int) {
	for _, codec := range store.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxLength(l)
		}
	}
}

// MaxAge sets the maximum age for the store and its cookies.
func (store *PGStore) MaxAge(age int) {
	store.Options.MaxAge = age
	for _, codec := range store.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// load fetches a session by its ID from the database.
func (store *PGStore) load(session *sessions.Session) error {
	var s PGSession
	ctx := context.Background()
	query := "SELECT id, key, data, created_on, modified_on, expires_on FROM http_sessions WHERE key = $1"
	err := store.Pool.QueryRow(ctx, query, session.ID).Scan(
		&s.ID, &s.Key, &s.Data, &s.CreatedOn, &s.ModifiedOn, &s.ExpiresOn,
	)
	if err != nil {
		return fmt.Errorf("unable to find session: %w", err)
	}

	return securecookie.DecodeMulti(session.Name(), string(s.Data), &session.Values, store.Codecs...)
}

// save writes encoded session values to the database.
func (store *PGStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, store.Codecs...)
	if err != nil {
		return err
	}

	// Determine creation and expiration timestamps.
	var createdOn time.Time
	if t, ok := session.Values["created_on"].(time.Time); ok {
		createdOn = t
	} else {
		createdOn = time.Now()
	}

	var expiresOn time.Time
	if ex, ok := session.Values["expires_on"].(time.Time); ok {
		expiresOn = ex
		// If the expiration is past the MaxAge, update it.
		if expiresOn.Before(time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))) {
			expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
		}
	} else {
		expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
	}

	psession := PGSession{
		Key:        []byte(session.ID),
		Data:       []byte(encoded),
		CreatedOn:  createdOn,
		ModifiedOn: time.Now(),
		ExpiresOn:  expiresOn,
	}

	if session.IsNew {
		return store.insert(&psession)
	}

	return store.update(&psession)
}

// destroy deletes the session record from the database.
func (store *PGStore) destroy(session *sessions.Session) error {
	ctx := context.Background()
	_, err := store.Pool.Exec(ctx, "DELETE FROM http_sessions WHERE key = $1", session.ID)
	return err
}

// createSessionsTable creates the required table and indexes if they do not exist.
// The schema uses TEXT columns for the key and data.
func (store *PGStore) createSessionsTable() error {
	stmt := `
DO $$
BEGIN
    CREATE TABLE IF NOT EXISTS http_sessions (
        id BIGSERIAL PRIMARY KEY,
        key TEXT,
        data TEXT,
        created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        modified_on TIMESTAMPTZ,
        expires_on TIMESTAMPTZ
    );
    CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
    CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
EXCEPTION WHEN insufficient_privilege THEN
    IF NOT EXISTS (
        SELECT FROM pg_catalog.pg_tables 
        WHERE schemaname = current_schema() AND tablename = 'http_sessions'
    ) THEN
        RAISE;
    END IF;
WHEN others THEN RAISE;
END;
$$;
`
	ctx := context.Background()
	_, err := store.Pool.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("unable to create http_sessions table: %w", err)
	}
	return nil
}

// insert writes a new session record to the database.
func (store *PGStore) insert(s *PGSession) error {
	ctx := context.Background()
	stmt := `INSERT INTO http_sessions (key, data, created_on, modified_on, expires_on)
             VALUES ($1, $2, $3, $4, $5)`
	_, err := store.Pool.Exec(ctx, stmt, s.Key, s.Data, s.CreatedOn, s.ModifiedOn, s.ExpiresOn)
	return err
}

// update modifies an existing session record.
func (store *PGStore) update(s *PGSession) error {
	ctx := context.Background()
	stmt := `UPDATE http_sessions SET data=$1, modified_on=$2, expires_on=$3 WHERE key=$4`
	_, err := store.Pool.Exec(ctx, stmt, s.Data, s.ModifiedOn, s.ExpiresOn, s.Key)
	return err
}
