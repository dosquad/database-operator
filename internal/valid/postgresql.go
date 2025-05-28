package valid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)
	nameRegex      = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

	ErrInvalidName = errors.New("invalid name")
)

const (
	PostgreSQLNameDataLen = 64
	PostgreSQLNameDataMin = 3
)

// PGIdentifier a PostgreSQL identifier or name. Identifiers can be composed of
// multiple parts such as ["schema", "table"] or ["table", "column"].
//
// based on pgx.Identifier from `jackc/pgx`.
type PGIdentifier string

// String returns a string safe for SQL and URIs.
func (ident PGIdentifier) String() string {
	s := strings.ReplaceAll(string(ident), string([]byte{0}), "")
	s = nameRegex.ReplaceAllString(s, "")
	return strings.ReplaceAll(s, `"`, `""`)
}

// Sanitize returns a sanitized string safe for SQL interpolation.
func (ident PGIdentifier) Sanitize() string {
	return `"` + ident.String() + `"`
}

func (ident PGIdentifier) Validate() (string, error) {
	name := nameRegex.ReplaceAllString(string(ident), "")

	if !validNameRegex.MatchString(name) {
		return name, fmt.Errorf("%w: invalid characters", ErrInvalidName)
	}

	switch strings.ToLower(name) {
	case "postgres", "psql", "root":
		return name, ErrInvalidName
	}

	if len(name) > PostgreSQLNameDataLen-1 {
		return name, fmt.Errorf("%w: name too long", ErrInvalidName)
	}

	if len(name) <= PostgreSQLNameDataMin {
		return name, fmt.Errorf("%w: name too short", ErrInvalidName)
	}

	// if !strings.HasPrefix(name, helper.DatabaseResourcePrefix) {
	// 	return name, fmt.Errorf("%w: name does not start with resource prefix", ErrInvalidName)
	// }

	return name, nil
}

// PGValue a PostgreSQL value.
type PGValue string

// Sanitize returns a sanitized string safe for SQL interpolation.
func (ident PGValue) Sanitize() string {
	s := strings.ReplaceAll(string(ident), string([]byte{0}), "")
	s = strings.ReplaceAll(s, `'`, `''`)
	return "'" + s + "'"
}

// String returns a string safe for SQL and URI use.
func (ident PGValue) String() string {
	s := strings.ReplaceAll(string(ident), string([]byte{0}), "")
	s = strings.ReplaceAll(s, `'`, `''`)
	s = strings.ReplaceAll(s, `@`, ``)
	return s
}
