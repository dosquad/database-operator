package accountsvr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/helper"
	"github.com/dosquad/database-operator/internal/valid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	ErrRoleExists = errors.New("role already exists")
)

type DatabaseServer struct {
	connString dbov1.PostgreSQLDSN
	conn       databaseConnection
}

const (
	DatabaseKeyDSN            = "dsn"
	DatabaseKeyUsername       = "username"
	DatabaseKeyPassword       = "password"
	DatabaseKeyHost           = "host"
	DatabaseKeyPort           = "port"
	DatabaseKeySchema         = "schema"
	DatabaseKeyDatabase       = "database"
	DatabaseKeyOnDelete       = "onDelete"
	DatabaseKeyPGBouncerConf  = "pgbouncer.ini"
	DatabaseKeyPGBouncerUsers = "userlist.txt"
)

func NewDatabaseServer(ctx context.Context, connString dbov1.PostgreSQLDSN) (*DatabaseServer, error) {
	s := &DatabaseServer{
		connString: connString,
	}

	return s, s.connect(ctx)
	// defer conn.Close(context.Background())
}

func NewDatabaseServerWithMock(
	_ context.Context,
	connString dbov1.PostgreSQLDSN,
	conn databaseConnection,
) (*DatabaseServer, error) {
	s := &DatabaseServer{
		connString: connString,
		conn:       conn,
	}

	return s, nil
}

// func (s *DatabaseServer) CheckInvalidName(name string) (string, error) {
// 	name = nameRegex.ReplaceAllString(name, "")

// 	if !validNameRegex.MatchString(name) {
// 		return name, fmt.Errorf("%w: invalid characters", ErrInvalidName)
// 	}

// 	switch strings.ToLower(name) {
// 	case "postgres", "psql", "root":
// 		return name, ErrInvalidName
// 	}

// 	if len(name) > PostgreSQLNameDataLen-1 {
// 		return name, fmt.Errorf("%w: name too long", ErrInvalidName)
// 	}

// 	if !strings.HasPrefix(name, helper.DatabaseResourcePrefix) {
// 		return name, fmt.Errorf("%w: name does not start with resource prefix", ErrInvalidName)
// 	}

// 	return name, nil
// }

func (s *DatabaseServer) Connect(ctx context.Context) error {
	if s.conn != nil {
		if !s.conn.IsClosed() {
			return nil
		}
	}

	return s.connect(ctx)
}

func (s *DatabaseServer) connect(ctx context.Context) error {
	logger := logr.FromContext(ctx)

	conn, err := pgx.Connect(ctx, s.connString.String())
	if err != nil {
		logger.Error(err, "unable to connect to the database")

		return err
	}

	s.conn = conn

	return nil
}

func (s *DatabaseServer) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s *DatabaseServer) ListUsers(ctx context.Context) []string {
	_ = s.Connect(ctx)

	var rows pgx.Rows
	{
		var err error
		rows, err = s.conn.Query(ctx, `select usename from pg_catalog.pg_user`)
		if err != nil {
			return []string{}
		}
	}
	defer rows.Close()

	o := []string{}

	for rows.Next() {
		if rows.Err() != nil {
			return o
		}

		v, err := rows.Values()
		if err != nil {
			return o
		}

		if len(v) > 0 {
			if user, ok := v[0].(string); ok {
				o = append(o, user)
			}
		}
	}

	return o
}

func (s *DatabaseServer) IsRole(ctx context.Context, roleName string) (bool, error) {
	_ = s.Connect(ctx)

	var rows pgx.Rows
	{
		var err error
		rows, err = s.conn.Query(ctx, `select usename from pg_catalog.pg_user where usename=$1`, roleName)
		if err != nil {
			return false, err
		}
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (s *DatabaseServer) IsDatabase(ctx context.Context, dbName string) (string, bool, error) {
	_ = s.Connect(ctx)

	{
		var err error
		dbName, err = valid.PGIdentifier(dbName).Validate()
		if err != nil {
			return dbName, false, err
		}
	}

	var rows pgx.Rows
	{
		var err error
		rows, err = s.conn.Query(ctx, `SELECT FROM pg_database WHERE datname = $1`, dbName)
		if err != nil {
			return dbName, false, err
		}
	}
	defer rows.Close()

	return dbName, rows.Next(), nil
}

func (s *DatabaseServer) CreateRole(ctx context.Context, roleName string) (string, string, error) {
	_ = s.Connect(ctx)
	// logger := logr.FromContext(ctx)

	if v, err := s.IsRole(ctx, roleName); err != nil || v {
		if v {
			return "", "", ErrRoleExists
		}
		return "", "", err
	}

	{
		var err error
		roleName, err = valid.PGIdentifier(roleName).Validate()
		if err != nil {
			return "", "", fmt.Errorf("role name[%s]: %w", roleName, err)
		}
	}

	password := helper.GeneratePassword(ctx)

	stmt := fmt.Sprintf(
		`CREATE ROLE %s LOGIN PASSWORD %s`,
		valid.PGIdentifier(roleName).Sanitize(),
		valid.PGValue(password).Sanitize(),
	)
	// stmt := `CREATE ROLE $1 LOGIN PASSWORD $2`
	// logger.V(1).Info(fmt.Sprintf("SQL: %s (%s, %s)", stmt, roleName, password))
	// if _, err := s.conn.Exec(ctx, stmt, roleName, password); err != nil {
	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return "", "", err
	}

	return roleName, password, nil
}

func (s *DatabaseServer) UpdateRolePassword(ctx context.Context, roleName string) (string, string, error) {
	_ = s.Connect(ctx)
	// logger := logr.FromContext(ctx)

	{
		var err error
		roleName, err = valid.PGIdentifier(roleName).Validate()
		if err != nil {
			return "", "", fmt.Errorf("role name[%s]: %w", roleName, err)
		}
	}

	password := helper.GeneratePassword(ctx)

	stmt := fmt.Sprintf(`ALTER ROLE %s LOGIN PASSWORD %s`,
		valid.PGIdentifier(roleName).Sanitize(),
		valid.PGValue(password).Sanitize(),
	)
	// stmt := `ALTER ROLE $1 LOGIN PASSWORD $2`
	// logger.V(1).Info(fmt.Sprintf("SQL: %s (%s, %s)", stmt, roleName, password))
	// if _, err := s.conn.Exec(ctx, `ALTER ROLE $1 LOGIN PASSWORD $2`, roleName, password); err != nil {
	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return "", "", err
	}

	return roleName, password, nil
}

func (s *DatabaseServer) CreateDatabase(ctx context.Context, dbName, roleName string) (string, error) {
	_ = s.Connect(ctx)
	// logger := logr.FromContext(ctx)

	{
		var err error
		dbName, err = valid.PGIdentifier(dbName).Validate()
		if err != nil {
			return "", fmt.Errorf("database name[%s]: %w", dbName, err)
		}
	}

	{
		var err error
		roleName, err = valid.PGIdentifier(roleName).Validate()
		if err != nil {
			return "", fmt.Errorf("role name[%s]: %w", roleName, err)
		}
	}

	stmt := fmt.Sprintf(`CREATE DATABASE %s OWNER %s`,
		valid.PGIdentifier(dbName).Sanitize(),
		valid.PGIdentifier(roleName).Sanitize(),
	)
	// stmt := `CREATE DATABASE $1 OWNER $2`
	// logger.V(1).Info(fmt.Sprintf("SQL: %s (%s, %s)", stmt, dbName, roleName))
	// if _, err := s.conn.Exec(ctx, `CREATE DATABASE $1 OWNER $2`, dbName, roleName); err != nil {
	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return "", err
	}

	return dbName, nil
}

func (s *DatabaseServer) CreateSchema(ctx context.Context, schemaName, roleName string) error {
	_ = s.Connect(ctx)
	// logger := logr.FromContext(ctx)

	{
		var err error
		schemaName, err = valid.PGIdentifier(schemaName).Validate()
		if err != nil {
			return fmt.Errorf("schema name[%s]: %w", schemaName, err)
		}
	}

	{
		var err error
		roleName, err = valid.PGIdentifier(roleName).Validate()
		if err != nil {
			return fmt.Errorf("role name[%s]: %w", roleName, err)
		}
	}

	stmt := fmt.Sprintf(
		`CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s`,
		valid.PGIdentifier(schemaName).Sanitize(),
		valid.PGIdentifier(roleName).Sanitize(),
	)
	// stmt := `CREATE SCHEMA IF NOT EXISTS $1 AUTHORIZATION $2`
	// logger.V(1).Info(fmt.Sprintf("SQL: %s (%s, %s)", stmt, schemaName, roleName))
	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return err
	}

	return nil
}

func (s *DatabaseServer) GetDatabaseHostConfig() string {
	return s.conn.Config().Host
}

func (s *DatabaseServer) GetDatabaseHost(dbAccount *dbov1.DatabaseAccount) string {
	if dbAccount.GetSpecCreateRelay() {
		return dbAccount.GetStatefulSetName().Name
	}

	return s.GetDatabaseHostConfig()
}

func (s *DatabaseServer) CopyInitConfigToSecret(
	dbAccount *dbov1.DatabaseAccount,
	secret *corev1.Secret,
) {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[DatabaseKeyHost] = []byte(s.GetDatabaseHost(dbAccount))
	secret.Data[DatabaseKeyPort] = []byte(strconv.FormatUint(uint64(s.conn.Config().Port), 10))
}

func GetSecretKV(secret *corev1.Secret, key string) string {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}

	if v, ok := secret.Data[key]; ok {
		return string(v)
	}

	if v, ok := secret.StringData[key]; ok {
		return v
	}

	return ""
}

func GenerateDSN(secret *corev1.Secret) string {
	var host string

	if len(GetSecretKV(secret, DatabaseKeyPort)) > 0 {
		host = fmt.Sprintf("%s:%s", GetSecretKV(secret, DatabaseKeyHost), GetSecretKV(secret, DatabaseKeyPort))
	} else {
		host = GetSecretKV(secret, DatabaseKeyHost)
	}

	u := &url.URL{
		User: url.UserPassword(
			GetSecretKV(secret, DatabaseKeyUsername),
			GetSecretKV(secret, DatabaseKeyPassword),
		),
		Host:   host,
		Scheme: "postgres",
		Path:   GetSecretKV(secret, DatabaseKeyDatabase),
	}

	return u.String()
}

func (s *DatabaseServer) Delete(ctx context.Context, name string) error {
	_ = s.Connect(ctx)
	// logger := logr.FromContext(ctx)

	{
		var err error
		name, err = valid.PGIdentifier(name).Validate()
		if err != nil {
			return fmt.Errorf("name[%s]: %w", name, err)
		}
	}

	var returnErr error

	name = valid.PGIdentifier(name).Sanitize()

	{
		stmt := fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, name)
		// logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
		// if _, err := s.conn.Exec(ctx, `DROP DATABASE IF EXISTS $1 WITH (FORCE)`, name); err != nil {
		if _, err := s.conn.Exec(ctx, stmt); err != nil {
			if !strings.Contains(err.Error(), " not found") {
				return err
			}
			returnErr = multierr.Append(returnErr, fmt.Errorf("database drop failed: %w", err))
		}
	}

	{
		stmt := `DROP ROLE IF EXISTS ` + name
		// logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
		// if _, err := s.conn.Exec(ctx, `DROP ROLE IF EXISTS $1`, name); err != nil {
		if _, err := s.conn.Exec(ctx, stmt); err != nil {
			if !strings.Contains(err.Error(), " not found") {
				return err
			}
			returnErr = multierr.Append(returnErr, fmt.Errorf("roll drop failed: %w", err))
		}
	}

	if returnErr != nil {
		return returnErr
	}

	return nil
}
