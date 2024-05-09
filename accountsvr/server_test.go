//nolint:gocognit // ignore complexity in test function.
package accountsvr_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dosquad/database-operator/accountsvr"
	accountsvrtest "github.com/dosquad/database-operator/accountsvr/test"
	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/testhelp"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testNewMockDB(t *testing.T) (
	time.Time,
	*accountsvrtest.MockDB,
	accountsvr.Server,
	context.Context,
	context.CancelFunc,
) {
	t.Helper()

	strt := time.Now()
	ctx, cancel := context.WithCancel(context.TODO())
	dsn := dbov1.PostgreSQLDSN("postgresql://localhost:53357/testdb")
	mDB := accountsvrtest.NewMockDB(t, nil, dsn)

	// conn, err := pgx.Connect(ctx, dsn.String())
	svr, err := accountsvr.NewDatabaseServerWithMock(ctx, dsn, mDB)
	if err != nil {
		testhelp.Errorf(t, strt, "accountsvr.NewDatabaseServer(): error, got '%s', want 'nil'", err)
	}

	return strt, mDB, svr, ctx, cancel
}

func TestAccountSvr_CheckInvalidName(t *testing.T) {
	t.Parallel()
	start, _, svr, _, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	tests := []struct {
		name                string
		databaseName        string
		expectDatabaseName  string
		expectError         error
		expectErrorContains string
	}{
		{"ExpectPass_SimpleName", "testname", "testname", nil, ""},
		{"ExpectPass_ComplexName", "k8s_01hdfk6ss05fm11pzs7k846eax", "k8s_01hdfk6ss05fm11pzs7k846eax", nil, ""},
		{"ExpectFail_ReservedName_psql", "psql", "psql", accountsvr.ErrInvalidName, ""},
		{"ExpectFail_ReservedName_root", "root", "root", accountsvr.ErrInvalidName, ""},
		{"ExpectFail_ReservedName_postgres", "postgres", "postgres", accountsvr.ErrInvalidName, ""},
		{"ExpectFail_ReservedName_CaseInsensitive", "POSTGRES", "POSTGRES", accountsvr.ErrInvalidName, ""},
		{"ExpectFail_StartWithDigit", "9name", "9name", accountsvr.ErrInvalidName, "invalid characters"},
		{"ExpectSuccess_ContainsDigit", "x9name", "x9name", nil, ""},
		{"ExpectSuccess_StripInvalidChar_Hash", "foo#bar", "foobar", nil, ""},
		{"ExpectSuccess_StripInvalidChar_Dash", "foo-bar", "foobar", nil, ""},
		{"ExpectSuccess_Length_62", strings.Repeat("x", 62), strings.Repeat("x", 62), nil, ""},
		{"ExpectSuccess_Length_63", strings.Repeat("x", 63), strings.Repeat("x", 63), nil, ""},
		{
			"ExpectFail_Length_64",
			strings.Repeat("x", 64), strings.Repeat("x", 64),
			accountsvr.ErrInvalidName, "name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := svr.CheckInvalidName(tt.databaseName)
			if !errors.Is(err, tt.expectError) {
				testhelp.Errorf(t, start, "svr.CheckInvalidName(): error, got '%+v', want '%+v'", err, tt.expectError)
			}

			if err != nil && tt.expectErrorContains != "" && !strings.Contains(err.Error(), tt.expectErrorContains) {
				testhelp.Errorf(t, start,
					"svr.CheckInvalidName(): error expected to contain string, got '%+v', want '%+v'",
					err, tt.expectErrorContains,
				)
			}

			if name != tt.expectDatabaseName {
				testhelp.Errorf(t, start, "svr.CheckInvalidName(): name, got '%s', want '%s'", name, tt.expectDatabaseName)
			}
		})
	}
}

func TestAccountSvr_Connect(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	isClosed := false
	mDB.OnIsClosed = func() bool {
		return isClosed
	}

	// Connection is not nil and is connected, do nothing return nil.
	if err := svr.Connect(ctx); err != nil {
		testhelp.Errorf(t, start, "accountsvr.Connect(ctx): error, got '%s', want 'nil'", err)
	}

	expectCalledFunc := map[string]int{
		"IsClosed": 1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.Connect(ctx): called functions -got +want:\n%s", diff)
	}

	mDB.CallCountReset()
	isClosed = true

	// Connection is not nil and not connected, dsn will fail as there is no real DB server.
	if err := svr.Connect(ctx); err == nil {
		testhelp.Errorf(t, start, "accountsvr.Connect(ctx): error, got 'nil', want 'error'")
	}

	expectCalledFunc = map[string]int{
		"IsClosed": 1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.Connect(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_Close(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	if err := svr.Close(ctx); err != nil {
		testhelp.Errorf(t, start, "accountsvr.Close(ctx): error, got '%s', want 'nil'", err)
	}

	expectCalledFunc := map[string]int{
		"Close": 1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.Close(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_ListUsers(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	expectUsers := []string{
		"foo",
		"bar",
	}
	users := svr.ListUsers(ctx)
	if diff := cmp.Diff(users, expectUsers); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): -got +want:\n%s", diff)
	}

	expectCalledFunc := map[string]int{
		"IsClosed": 1,
		"Query":    1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_ListUsers_QueryError(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	errQuery := errors.New("failed to query")
	mDB.OnQuery = func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
		return accountsvrtest.NewMockRows(mDB, nil, []string{"foo", "bar"}), errQuery
	}

	expectUsers := []string{}
	users := svr.ListUsers(ctx)
	if diff := cmp.Diff(users, expectUsers); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): -got +want:\n%s", diff)
	}

	expectCalledFunc := map[string]int{
		"IsClosed": 1,
		"Query":    1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_ListUsers_RowError(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	errQuery := errors.New("failed to query")
	mDB.OnQuery = func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
		mr := accountsvrtest.NewMockRows(mDB, nil, []string{"foo", "bar"})
		mr.OnErr = func() error {
			return errQuery
		}

		return mr, nil
	}

	expectUsers := []string{}
	users := svr.ListUsers(ctx)
	if diff := cmp.Diff(users, expectUsers); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): -got +want:\n%s", diff)
	}

	expectCalledFunc := map[string]int{
		"IsClosed": 1,
		"Query":    1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_ListUsers_RowValueError(t *testing.T) {
	t.Parallel()
	start, mDB, svr, ctx, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	errQuery := errors.New("failed to query")
	mDB.OnQuery = func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
		mr := accountsvrtest.NewMockRows(mDB, nil, []string{"foo", "bar"})
		mr.OnValues = func() ([]any, error) {
			return nil, errQuery
		}

		return mr, nil
	}

	expectUsers := []string{}
	users := svr.ListUsers(ctx)
	if diff := cmp.Diff(users, expectUsers); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): -got +want:\n%s", diff)
	}

	expectCalledFunc := map[string]int{
		"IsClosed": 1,
		"Query":    1,
	}

	if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.ListUsers(ctx): called functions -got +want:\n%s", diff)
	}
}

func TestAccountSvr_IsRole(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name          string
		roleName      string
		roleExists    bool
		expectedError error
	}{
		{"ExpectSuccess_RoleExists", "thunderball", true, nil},
		{"ExpectFail_RoleDoesNotExist", "goldfinger", false, nil},
		{"ExpectFail_ServerError", "internal-server-error", false, internalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			mDB.OnQuery = func(_ context.Context, _ string, a ...any) (pgx.Rows, error) {
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "username queried: %s", v)
						switch strings.ToLower(v) {
						case "thunderball": // valid role
							return accountsvrtest.NewMockRows(mDB, nil, []string{"thunderball"}), nil
						case "internal-server-error": // internal server error
							return nil, internalServerError
						}
					}
				}

				// no role found
				return accountsvrtest.NewMockRows(mDB, nil, []string{}), nil
			}

			isRole, err := svr.IsRole(ctx, tt.roleName)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start, "accountsvr.IsRole(ctx): error, got '%v', want '%v'", err, tt.expectedError)
			}

			if isRole != tt.roleExists {
				testhelp.Errorf(t, start, "accountsvr.IsRole(ctx): role exists, got '%t', want '%t'", isRole, tt.roleExists)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
				"Query":    1,
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.IsRole(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_IsDatabase(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name                 string
		databaseName         string
		expectDatabaseName   string
		expectDatabaseExists bool
		expectedError        error
	}{
		{"ExpectSuccess_DatabaseExists", "roly", "roly", true, nil},
		{"ExpectFail_DatabaseDoesNotExist", "poly", "poly", false, nil},
		{"ExpectFail_ServerError", "internal_server_error", "internal_server_error", false, internalServerError},
		{"ExpectSuccess_CorrectedDatabaseName", "roly-poly", "rolypoly", true, nil},
		{"ExpectFail_DatabaseNameLength", strings.Repeat("x", 64), strings.Repeat("x", 64), false, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			mDB.OnQuery = func(_ context.Context, _ string, a ...any) (pgx.Rows, error) {
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "database queried: %s", v)
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return accountsvrtest.NewMockRows(mDB, nil, []string{"roly"}), nil
						case "internal_server_error": // internal server error
							return nil, internalServerError
						}
					}
				}

				// no role found
				return accountsvrtest.NewMockRows(mDB, nil, []string{}), nil
			}

			dbName, dbExists, err := svr.IsDatabase(ctx, tt.databaseName)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start, "accountsvr.IsDatabase(ctx): error, got '%v', want '%v'", err, tt.expectedError)
			}

			if dbExists != tt.expectDatabaseExists {
				testhelp.Errorf(t, start,
					"accountsvr.IsDatabase(ctx): database exists, got '%t', want '%t'",
					dbExists, tt.expectDatabaseExists,
				)
			}

			if dbName != tt.expectDatabaseName {
				testhelp.Errorf(t, start,
					"accountsvr.IsDatabase(ctx): database name, got '%s', want '%s'", dbName, tt.expectDatabaseName,
				)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
				"Query":    1,
			}
			if errors.Is(err, accountsvr.ErrInvalidName) {
				delete(expectCalledFunc, "Query")
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.IsDatabase(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_CreateRole(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name           string
		rolename       string
		expectRolename string
		expectedExec   bool
		expectedError  error
	}{
		{"ExpectSuccess", "newroly", "newroly", true, nil},
		{"ExpectSuccess_CorrectedDatabaseName", "new-roly", "newroly", true, nil},
		{"ExpectFail_DatabaseExists", "roly", "", false, accountsvr.ErrRoleExists},
		{"ExpectFail_ServerError", "internal_server_error", "", false, internalServerError},
		{"ExpectFail_ServerErrorExec", "exec_internal_server_error", "", true, internalServerError},
		{"ExpectFail_DatabaseNameLength", strings.Repeat("x", 64), "", false, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			generatedPassword := ""

			mDB.OnQuery = func(_ context.Context, _ string, a ...any) (pgx.Rows, error) {
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "database queried: %s", v)
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return accountsvrtest.NewMockRows(mDB, nil, []string{"roly"}), nil
						case "internal_server_error": // internal server error
							return nil, internalServerError
						}
					}
				}

				// no role found
				return accountsvrtest.NewMockRows(mDB, nil, []string{}), nil
			}
			mDB.OnExec = func(_ context.Context, _ string, a ...any) (pgconn.CommandTag, error) {
				// testhelp.Logf(t, start, "mDB.Exec(): stmt, got '%s'", s)
				if len(a) > 1 {
					if v, ok := a[1].(string); ok {
						testhelp.Logf(t, start, "password: %s", v)
						generatedPassword = v
					}
				}
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "database create: %s", v)
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return pgconn.NewCommandTag(""), nil
						case "exec_internal_server_error": // internal server error
							return pgconn.NewCommandTag(""), internalServerError
						}
					}
				}

				// no role found
				return pgconn.NewCommandTag(""), nil
			}

			roleName, pw, err := svr.CreateRole(ctx, tt.rolename)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start,
					"accountsvr.CreateRole(ctx): error, got '%v', want '%v'", err, tt.expectedError,
				)
			}

			if roleName != tt.expectRolename {
				testhelp.Errorf(t, start,
					"accountsvr.CreateRole(ctx): database name, got '%s', want '%s'", roleName, tt.expectRolename,
				)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 2,
				"Query":    1,
			}
			if tt.expectedExec {
				expectCalledFunc["Exec"] = 1

				if err == nil && pw != generatedPassword {
					testhelp.Errorf(t, start,
						"accountsvr.CreateRole(ctx): generated password, got '%s', want '%s'", pw, generatedPassword,
					)
				}
			}

			if err == nil {
				if !strings.ContainsAny(pw, password.Digits) ||
					!strings.ContainsAny(pw, password.LowerLetters) ||
					!strings.ContainsAny(pw, password.UpperLetters) ||
					!strings.ContainsAny(pw, password.Symbols) ||
					len(pw) < 16 {
					testhelp.Errorf(t, start, "accountsvr.CreateRole(ctx): password complexity not met, got '%s'", pw)
				}
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.CreateRole(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_UpdateRolePassword(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	dbNotFound := errors.New("database not found")
	tests := []struct {
		name           string
		rolename       string
		expectRolename string
		expectedExec   bool
		expectedError  error
	}{
		{"ExpectSuccess", "roly", "roly", true, nil},
		{"ExpectSuccess_CorrectedDatabaseName", "roly-poly", "rolypoly", true, nil},
		{"ExpectFail_DatabaseDoesNotExists", "poly", "", true, dbNotFound},
		{"ExpectFail_ServerErrorExec", "exec_internal_server_error", "", true, internalServerError},
		{"ExpectFail_DatabaseNameLength", strings.Repeat("x", 64), "", false, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			generatedPassword := ""

			mDB.OnExec = func(_ context.Context, _ string, a ...any) (pgconn.CommandTag, error) {
				// testhelp.Logf(t, start, "mDB.Exec(): stmt, got '%s'", s)
				if len(a) > 1 {
					if v, ok := a[1].(string); ok {
						testhelp.Logf(t, start, "password: %s", v)
						generatedPassword = v
					}
				}
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "database create: %s", v)
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return pgconn.NewCommandTag(""), nil
						case "exec_internal_server_error": // internal server error
							return pgconn.NewCommandTag(""), internalServerError
						}
					}
				}

				// no role found
				return pgconn.NewCommandTag(""), dbNotFound
			}

			roleName, pw, err := svr.UpdateRolePassword(ctx, tt.rolename)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start, "accountsvr.UpdateRolePassword(ctx): error, got '%v', want '%v'", err, tt.expectedError)
			}

			if roleName != tt.expectRolename {
				testhelp.Errorf(t, start,
					"accountsvr.UpdateRolePassword(ctx): database name, got '%s', want '%s'",
					roleName, tt.expectRolename,
				)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
			}
			if tt.expectedExec {
				expectCalledFunc["Exec"] = 1

				if err == nil && pw != generatedPassword {
					testhelp.Errorf(t, start,
						"accountsvr.UpdateRolePassword(ctx): generated password, got '%s', want '%s'",
						pw, generatedPassword,
					)
				}
			}

			if err == nil {
				if !strings.ContainsAny(pw, password.Digits) ||
					!strings.ContainsAny(pw, password.LowerLetters) ||
					!strings.ContainsAny(pw, password.UpperLetters) ||
					!strings.ContainsAny(pw, password.Symbols) ||
					len(pw) < 16 {
					testhelp.Errorf(t, start, "accountsvr.UpdateRolePassword(ctx): password complexity not met, got '%s'", pw)
				}
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.UpdateRolePassword(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_CreateDatabase(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name               string
		databaseName       string
		roleName           string
		expectDatabaseName string
		expectedExec       bool
		expectedError      error
	}{
		{"ExpectSuccess", "newroly", "newroly", "newroly", true, nil},
		{"ExpectSuccess_CorrectedDatabaseName", "new-roly", "rolename", "newroly", true, nil},
		{"ExpectFail_ServerErrorExec", "exec_internal_server_error", "rolename", "", true, internalServerError},
		{"ExpectFail_DatabaseNameLength", strings.Repeat("x", 64), "rolename", "", false, accountsvr.ErrInvalidName},
		{"ExpectFail_RoleNameLength", "dbname", strings.Repeat("x", 64), "", false, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			mDB.OnExec = func(_ context.Context, _ string, a ...any) (pgconn.CommandTag, error) {
				// testhelp.Logf(t, start, "mDB.Exec(): stmt, got '%s'", s)
				if len(a) > 1 {
					if v, ok := a[1].(string); ok {
						testhelp.Logf(t, start, "role name: %s", v)
						// generatedPassword = v
					}
				}
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "database name: %s", v)
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return pgconn.NewCommandTag(""), nil
						case "exec_internal_server_error": // internal server error
							return pgconn.NewCommandTag(""), internalServerError
						}
					}
				}

				// no role found
				return pgconn.NewCommandTag(""), nil
			}

			dbName, err := svr.CreateDatabase(ctx, tt.databaseName, tt.roleName)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start, "accountsvr.CreateDatabase(ctx): error, got '%v', want '%v'", err, tt.expectedError)
			}

			if dbName != tt.expectDatabaseName {
				testhelp.Errorf(t, start,
					"accountsvr.CreateDatabase(ctx): database name, got '%s', want '%s'",
					dbName, tt.expectDatabaseName,
				)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
			}
			if tt.expectedExec {
				expectCalledFunc["Exec"] = 1
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.CreateDatabase(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_CreateSchema(t *testing.T) {
	t.Parallel()

	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name             string
		schemaName       string
		roleName         string
		expectSchemaName string
		expectRoleName   string
		expectedExec     bool
		expectedError    error
	}{
		{"ExpectSuccess", "newroly", "newroly", "newroly", "newroly", true, nil},
		{"ExpectSuccess_CorrectedSchemaName", "new-roly", "rolename", "newroly", "rolename", true, nil},
		{"ExpectSuccess_CorrectedRoleName", "newroly", "role-name", "newroly", "rolename", true, nil},
		{"ExpectSuccess_CorrectedSchemaAndRoleName", "new-roly", "role-name", "newroly", "rolename", true, nil},
		{"ExpectFail_ServerErrorExec", "exec_internal_server_error", "rolename", "", "rolename", true, internalServerError},
		{"ExpectFail_SchemaNameLength", strings.Repeat("x", 64), "rolename", "", "", false, accountsvr.ErrInvalidName},
		{"ExpectFail_RoleNameLength", "dbname", strings.Repeat("x", 64), "", "", false, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			execSchemaName := ""
			execRoleName := ""

			mDB.OnExec = func(_ context.Context, _ string, a ...any) (pgconn.CommandTag, error) {
				// testhelp.Logf(t, start, "mDB.Exec(): stmt, got '%s'", s)
				if len(a) > 1 {
					if v, ok := a[1].(string); ok {
						testhelp.Logf(t, start, "role name: %s", v)
						execRoleName = v
					}
				}
				if len(a) > 0 {
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "schema name: %s", v)
						execSchemaName = v
						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return pgconn.NewCommandTag(""), nil
						case "exec_internal_server_error": // internal server error
							return pgconn.NewCommandTag(""), internalServerError
						}
					}
				}

				// no role found
				return pgconn.NewCommandTag(""), nil
			}

			err := svr.CreateSchema(ctx, tt.schemaName, tt.roleName)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start, "accountsvr.CreateSchema(ctx): error, got '%v', want '%v'", err, tt.expectedError)
			}

			if err == nil {
				if execRoleName != tt.expectRoleName {
					testhelp.Errorf(t, start,
						"accountsvr.CreateSchema(ctx): role name, got '%s', want '%s'",
						execRoleName, tt.expectRoleName,
					)
				}

				if execSchemaName != tt.expectSchemaName {
					testhelp.Errorf(t, start,
						"accountsvr.CreateSchema(ctx): schema name, got '%s', want '%s'",
						execSchemaName, tt.expectSchemaName,
					)
				}
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
			}
			if tt.expectedExec {
				expectCalledFunc["Exec"] = 1
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.CreateSchema(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_GetDatabaseHostConfig(t *testing.T) {
	start, mDB, svr, _, cancel := testNewMockDB(t)
	t.Cleanup(cancel)
	mDB.OnConfig = func() *pgx.ConnConfig {
		return &pgx.ConnConfig{
			Config: pgconn.Config{
				Host: "config-host-test",
			},
		}
	}

	conf := svr.GetDatabaseHostConfig()

	if conf != "config-host-test" {
		testhelp.Errorf(t, start, "svr.GetDatabaseHostConfig(): got '%s', want '%s'", conf, "config-host-test")
	}
}

func TestAccountSvr_GetDatabaseHost(t *testing.T) {
	t.Parallel()

	start, mDB, svr, _, cancel := testNewMockDB(t)
	t.Cleanup(cancel)
	mDB.OnConfig = func() *pgx.ConnConfig {
		return &pgx.ConnConfig{
			Config: pgconn.Config{
				Host: "config-host-test",
			},
		}
	}

	tests := []struct {
		name         string
		dbAccount    *dbov1.DatabaseAccount
		expectedHost string
	}{
		{
			"ExpectSuccess_AccountHost_Secret",
			&dbov1.DatabaseAccount{
				Spec: dbov1.DatabaseAccountSpec{CreateRelay: true, SecretName: "secret"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "objectname",
					Namespace: "objectnamespace",
				},
			},
			"secret",
		},
		{
			"ExpectSuccess_AccountHost_Name",
			&dbov1.DatabaseAccount{
				Spec: dbov1.DatabaseAccountSpec{CreateRelay: true},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "objectname",
					Namespace: "objectnamespace",
				},
			},
			"objectname",
		},
		{"ExpectSuccess_DatabaseHost",
			&dbov1.DatabaseAccount{
				Spec: dbov1.DatabaseAccountSpec{CreateRelay: false, SecretName: "secret"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "objectname",
					Namespace: "objectnamespace",
				},
			},
			"config-host-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conf := svr.GetDatabaseHost(tt.dbAccount)

			if conf != tt.expectedHost {
				testhelp.Errorf(t, start, "svr.GetDatabaseHostConfig(): got '%s', want '%s'", conf, tt.expectedHost)
			}
		})
	}
}

func TestAccountSvr_CopyInitConfigToSecret(t *testing.T) {
	t.Parallel()

	start, _, svr, _, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	dbAccount := &dbov1.DatabaseAccount{
		Spec: dbov1.DatabaseAccountSpec{CreateRelay: false, SecretName: "secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objectname",
			Namespace: "objectnamespace",
		},
	}

	secret := &corev1.Secret{}

	svr.CopyInitConfigToSecret(dbAccount, secret)

	expectData := map[string][]byte{
		accountsvr.DatabaseKeyHost: []byte(accountsvrtest.Hostname),
		accountsvr.DatabaseKeyPort: []byte("1337"),
	}

	if diff := cmp.Diff(secret.Data, expectData); diff != "" {
		testhelp.Errorf(t, start, "accountsvr.CopyInitConfigToSecret(): expected secret.Data -got +want:\n%s", diff)
	}
}

// func GetSecretKV(secret *corev1.Secret, key string) string {
// 	if secret.Data == nil {
// 		secret.Data = make(map[string][]byte)
// 	}
// 	if secret.StringData == nil {
// 		secret.StringData = make(map[string]string)
// 	}

// 	if v, ok := secret.Data[key]; ok {
// 		return string(v)
// 	}

// 	return ""
// }

func TestAccountSvr_GetSecretKV(t *testing.T) {
	t.Parallel()
	start, _, svr, _, cancel := testNewMockDB(t)
	t.Cleanup(cancel)

	dbAccount := &dbov1.DatabaseAccount{
		Spec: dbov1.DatabaseAccountSpec{CreateRelay: false, SecretName: "secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objectname",
			Namespace: "objectnamespace",
		},
	}
	secret := &corev1.Secret{}

	svr.CopyInitConfigToSecret(dbAccount, secret)

	if v := accountsvr.GetSecretKV(secret, accountsvr.DatabaseKeyHost); v != accountsvrtest.Hostname {
		testhelp.Errorf(t, start,
			"accountsvr.GetSecretKV(%s): got '%s', want '%s'", accountsvr.DatabaseKeyHost, v, accountsvrtest.Hostname,
		)
	}

	if v := accountsvr.GetSecretKV(secret, accountsvr.DatabaseKeyPort); v != strconv.FormatInt(accountsvrtest.Port, 10) {
		testhelp.Errorf(t, start,
			"accountsvr.GetSecretKV(%s): got '%s', want '%d'", accountsvr.DatabaseKeyPort, v, accountsvrtest.Port,
		)
	}
}

func TestAccountSvr_GenerateDSN(t *testing.T) {
	t.Parallel()

	start := time.Now()
	tests := []struct {
		name        string
		secret      *corev1.Secret
		expectedDSN string
	}{
		{
			"ExpectedSuccess_Empty",
			&corev1.Secret{},
			"postgres://:@",
		},
		{
			"ExpectedSuccess_SecretData",
			&corev1.Secret{
				Data: map[string][]byte{
					accountsvr.DatabaseKeyHost:     []byte("database-host"),
					accountsvr.DatabaseKeyPort:     []byte(strconv.FormatInt(accountsvrtest.Port, 10)),
					accountsvr.DatabaseKeyDatabase: []byte("database"),
					accountsvr.DatabaseKeySchema:   []byte("schema"),
					accountsvr.DatabaseKeyUsername: []byte("username"),
					accountsvr.DatabaseKeyPassword: []byte("password"),
				},
			},
			"postgres://username:password@database-host:1337/database",
		},
		{
			"ExpectedSuccess_SecretStringData",
			&corev1.Secret{
				StringData: map[string]string{
					accountsvr.DatabaseKeyHost:     "database-host",
					accountsvr.DatabaseKeyPort:     strconv.FormatInt(accountsvrtest.Port, 10),
					accountsvr.DatabaseKeyDatabase: "database",
					accountsvr.DatabaseKeySchema:   "schema",
					accountsvr.DatabaseKeyUsername: "username",
					accountsvr.DatabaseKeyPassword: "password",
				},
			},
			"postgres://username:password@database-host:1337/database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dsn := accountsvr.GenerateDSN(tt.secret)

			if diff := cmp.Diff(dsn, tt.expectedDSN); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.GenerateDSN(): dsn -got +want:\n%s", diff)
			}
		})
	}
}

func TestAccountSvr_Delete(t *testing.T) {
	t.Parallel()
	internalServerError := errors.New("internal-server-error")
	tests := []struct {
		name                 string
		targetName           string
		expectedDatabaseName string
		expectedRoleName     string
		expectedExec         int
		expectedError        error
	}{
		{"ExpectSuccess", "newroly", "newroly", "newroly", 2, nil},
		{"ExpectSuccess_CorrectedSchemaName", "new-roly", "newroly", "newroly", 2, nil},
		{
			"ExpectFail_ServerErrorExecDatabase",
			"exec_db_internal_server_error", "exec_db_internal_server_error", "",
			1, internalServerError,
		},
		{
			"ExpectFail_ServerErrorExecRole",
			"exec_role_internal_server_error", "exec_role_internal_server_error", "exec_role_internal_server_error",
			2, internalServerError,
		},
		{"ExpectFail_SchemaNameLength", strings.Repeat("x", 64), "", "", 0, accountsvr.ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, mDB, svr, ctx, cancel := testNewMockDB(t)
			t.Cleanup(cancel)

			execDatabaseName := ""
			execRoleName := ""

			mDB.OnExec = func(_ context.Context, s string, a ...any) (pgconn.CommandTag, error) {
				if len(a) > 0 { //nolint:nestif // testing function.
					if v, ok := a[0].(string); ok {
						testhelp.Logf(t, start, "exec[%s]: %s", s, v)
						switch {
						case strings.HasPrefix(s, "DROP DATABASE "):
							execDatabaseName = v
						case strings.HasPrefix(s, "DROP ROLE "):
							execRoleName = v
						}

						switch strings.ToLower(v) {
						case "roly", "rolypoly": // valid role
							return pgconn.NewCommandTag(""), nil
						case "exec_db_internal_server_error": // internal server error
							if strings.HasPrefix(s, "DROP DATABASE ") {
								return pgconn.NewCommandTag(""), internalServerError
							}
						case "exec_role_internal_server_error": // internal server error
							if strings.HasPrefix(s, "DROP ROLE ") {
								return pgconn.NewCommandTag(""), internalServerError
							}
						}
					}
				}

				// no role found
				return pgconn.NewCommandTag(""), nil
			}

			err := svr.Delete(ctx, tt.targetName)
			if !errors.Is(err, tt.expectedError) {
				testhelp.Errorf(t, start,
					"accountsvr.Delete(ctx): error, got '%v', want '%v'", err, tt.expectedError,
				)
			}

			if execDatabaseName != tt.expectedDatabaseName {
				testhelp.Errorf(
					t, start,
					"accountsvr.Delete(ctx): database name, got '%s', want '%s'",
					execDatabaseName, tt.expectedDatabaseName,
				)
			}

			if execRoleName != tt.expectedRoleName {
				testhelp.Errorf(t, start,
					"accountsvr.Delete(ctx): role name, got '%s', want '%s'", execRoleName, tt.expectedRoleName,
				)
			}

			expectCalledFunc := map[string]int{
				"IsClosed": 1,
			}
			if tt.expectedExec > 0 {
				expectCalledFunc["Exec"] = tt.expectedExec
			}

			if diff := cmp.Diff(mDB.CallCountMap(), expectCalledFunc); diff != "" {
				testhelp.Errorf(t, start, "accountsvr.Delete(ctx): called functions -got +want:\n%s", diff)
			}
		})
	}
}
