package v1

import "net/url"

// PostgreSQLDSN is a kubernetes validation for a PostgreSQL DSN.
// +optional
type PostgreSQLDSN string

func (d PostgreSQLDSN) String() string {
	return string(d)
}

func (d PostgreSQLDSN) Host() string {
	if u, err := url.Parse(string(d)); err == nil {
		return u.Host
	}

	return ""
}
