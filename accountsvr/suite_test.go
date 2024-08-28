package accountsvr_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/dosquad/database-operator/internal/testhelp"
)

var (
	createDatabaseRegex = regexp.MustCompile(`CREATE DATABASE "([^"]+)" OWNER "([^"]+)"`)
	createSchemaRegex   = regexp.MustCompile(`CREATE SCHEMA IF NOT EXISTS "([^"]+)" AUTHORIZATION "([^"]+)"`)
	createRoleRegex     = regexp.MustCompile(`CREATE ROLE "([^"]+)" LOGIN PASSWORD '([^']+)'`)
	alterRoleRegex      = regexp.MustCompile(`ALTER ROLE "([^"]+)" LOGIN PASSWORD '([^']+)'`)
	dropDatabaseRegex   = regexp.MustCompile(`DROP DATABASE IF EXISTS "([^"]+)" WITH \(FORCE\)`)
	dropRoleRegex       = regexp.MustCompile(`DROP ROLE IF EXISTS "([^"]+)"`)
)

func replaceArgs(t *testing.T, start time.Time, s string, a []any) []any {
	t.Helper()

	testhelp.Logf(t, start, "mDB.Exec(): stmt, got '%s'", s)

	if len(a) == 0 {
		for ridx, rx := range []*regexp.Regexp{
			createDatabaseRegex, createSchemaRegex, createRoleRegex,
			alterRoleRegex, dropDatabaseRegex, dropRoleRegex,
		} {
			matches := rx.FindStringSubmatch(s)
			if len(matches) >= 2 {
				matches = matches[1:]
				a = make([]any, len(matches))
				for idx := range matches {
					a[idx] = matches[idx]
				}
				testhelp.Logf(t, start, "mDB.Exec(): regexp[%d] matches, got '%+v'", ridx, matches)
				return a
			}
		}
	}

	return a
}
