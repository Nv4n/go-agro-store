package server

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/jackc/pgx/v5/pgtype"
)

const csrfTokenKey = "csrf_token"

func StrToUUID(unparsed string) (pgtype.UUID, error) {
	var parsed pgtype.UUID
	err := parsed.Scan(unparsed)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return parsed, nil
}

// GenerateCSRFToken creates a new CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
