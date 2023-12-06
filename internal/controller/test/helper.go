package test

import (
	"strings"

	v1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
)

// NewDatabaseAccountName returns a newly generated database/username.
func NewDatabaseAccountName() v1.PostgreSQLResourceName {
	return v1.PostgreSQLResourceName(strings.ToLower(v1test.DBUser))
}
