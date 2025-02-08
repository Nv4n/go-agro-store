package server

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func StrToUUID(unparsed string) (pgtype.UUID, error) {
	var parsed pgtype.UUID
	err := parsed.Scan(unparsed)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return parsed, nil
}
