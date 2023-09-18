package accountsvr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)

	nameRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

	ErrInvalidName = errors.New("invalid name")

	ErrRoleExists = errors.New("role already exists")
)

const (
	PostgreSQLNameDataLen = 64

	passwordLength         = 28
	passwordComplexDigits  = 10
	passwordComplexSymbols = 1
)

type Server struct {
	connString dbov1.PostgreSQLDSN
	conn       *pgx.Conn
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

func NewServer(ctx context.Context, connString dbov1.PostgreSQLDSN) (*Server, error) {
	s := &Server{
		connString: connString,
	}

	return s, s.Connect(ctx)
	// defer conn.Close(context.Background())
}

func (s *Server) CheckInvalidName(name string) (string, error) {
	name = nameRegex.ReplaceAllString(name, "")

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

	return name, nil
}

func (s *Server) Connect(ctx context.Context) error {
	logger := log.FromContext(ctx)

	if s.conn != nil {
		if !s.conn.IsClosed() {
			return nil
		}
	}

	conn, err := pgx.Connect(ctx, s.connString.String())
	if err != nil {
		logger.Error(err, "unable to connect to the database")

		return err
	}

	s.conn = conn

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s *Server) ListUsers(ctx context.Context) []string {
	_ = s.Connect(ctx)

	rows, err := s.conn.Query(ctx, `select * from pg_catalog.pg_user`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	o := []string{}

	for rows.Next() {
		if rows.Err() != nil {
			return o
		}
		o = append(o, fmt.Sprintf("%#v", rows.RawValues()))
	}

	return o
}

// TODO actually generate password
func (s *Server) generatePassword(ctx context.Context) string {
	logger := log.FromContext(ctx)

	res, err := password.Generate(passwordLength, passwordComplexDigits, passwordComplexSymbols, false, false)
	if err != nil {
		logger.Error(err, "unable to generate password")
		panic(err)
	}
	return res
}

func (s *Server) IsRole(ctx context.Context, roleName string) (bool, error) {
	_ = s.Connect(ctx)

	rows, err := s.conn.Query(ctx, `select usename from pg_catalog.pg_user where usename=$1`, roleName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (s *Server) IsDatabase(ctx context.Context, dbName string) (string, bool, error) {
	_ = s.Connect(ctx)

	dbName, err := s.CheckInvalidName(dbName)
	if err != nil {
		return dbName, false, err
	}

	rows, err := s.conn.Query(ctx, `SELECT FROM pg_database WHERE datname = $1`, dbName)
	if err != nil {
		return dbName, false, err
	}
	defer rows.Close()

	return dbName, rows.Next(), nil
}

func (s *Server) CreateRole(ctx context.Context, roleName string) (string, string, error) {
	_ = s.Connect(ctx)
	logger := log.FromContext(ctx)

	if v, err := s.IsRole(ctx, roleName); err != nil || v {
		if v {
			return "", "", ErrRoleExists
		}

		return "", "", err
	}

	roleName, rollNameErr := s.CheckInvalidName(roleName)
	if rollNameErr != nil {
		return "", "", fmt.Errorf("role name[%s]: %w", roleName, rollNameErr)
	}

	password := s.generatePassword(ctx)
	stmt := fmt.Sprintf(`CREATE ROLE %s LOGIN PASSWORD '%s'`, roleName, password)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))

	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return "", "", err
	}

	return roleName, password, nil
}

func (s *Server) UpdateRolePassword(ctx context.Context, roleName string) (string, string, error) {
	_ = s.Connect(ctx)
	logger := log.FromContext(ctx)

	roleName, roleNameErr := s.CheckInvalidName(roleName)
	if roleNameErr != nil {
		return "", "", fmt.Errorf("role name[%s]: %w", roleName, roleNameErr)
	}

	password := s.generatePassword(ctx)
	stmt := fmt.Sprintf(`ALTER ROLE %s LOGIN PASSWORD '%s'`, roleName, password)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
	if _, err := s.conn.Exec(ctx, stmt); err != nil {
		return "", "", err
	}

	return roleName, password, nil
}

func (s *Server) CreateDatabase(ctx context.Context, dbName, roleName string) (string, error) {
	_ = s.Connect(ctx)
	logger := log.FromContext(ctx)
	var err error

	dbName, err = s.CheckInvalidName(dbName)
	if err != nil {
		return "", fmt.Errorf("database name[%s]: %w", dbName, err)
	}

	roleName, err = s.CheckInvalidName(roleName)
	if err != nil {
		return "", fmt.Errorf("role name[%s]: %w", roleName, err)
	}

	stmt := fmt.Sprintf(`CREATE DATABASE %s OWNER %s`, dbName, roleName)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
	if _, execErr := s.conn.Exec(ctx, stmt); execErr != nil {
		return "", execErr
	}

	return dbName, nil
}

func (s *Server) CreateSchema(ctx context.Context, schemaName, roleName string) error {
	_ = s.Connect(ctx)
	logger := log.FromContext(ctx)
	var err error

	schemaName, err = s.CheckInvalidName(schemaName)
	if err != nil {
		return fmt.Errorf("schema name[%s]: %w", schemaName, err)
	}

	roleName, err = s.CheckInvalidName(roleName)
	if err != nil {
		return fmt.Errorf("role name[%s]: %w", roleName, err)
	}

	stmt := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s`, schemaName, roleName)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
	if _, execErr := s.conn.Exec(ctx, stmt); execErr != nil {
		return execErr
	}

	return nil
}

func (s *Server) GetDatabaseHostConfig() string {
	return s.conn.Config().Host
}

func (s *Server) GetDatabaseHost(dbAccount *dbov1.DatabaseAccount) string {
	if dbAccount.GetSpecCreateRelay() {
		return dbAccount.GetStatefulSetName().Name
	}

	return s.GetDatabaseHostConfig()
}

func (s *Server) CopyInitConfigToSecret(
	dbAccount *dbov1.DatabaseAccount,
	secret *corev1.Secret,
) {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[DatabaseKeyHost] = []byte(s.GetDatabaseHost(dbAccount))
	secret.Data[DatabaseKeyPort] = []byte(fmt.Sprintf("%d", s.conn.Config().Port))
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

func (s *Server) Delete(ctx context.Context, name string) error {
	_ = s.Connect(ctx)
	logger := log.FromContext(ctx)

	name, err := s.CheckInvalidName(name)
	if err != nil {
		return fmt.Errorf("name[%s]: %w", name, err)
	}

	stmt := fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, name)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
	if _, execErr := s.conn.Exec(ctx, stmt); execErr != nil {
		if !strings.Contains(execErr.Error(), " not found") {
			return execErr
		}
	}

	stmt = fmt.Sprintf(`DROP ROLE IF EXISTS %s`, name)
	logger.V(1).Info(fmt.Sprintf("SQL: %s", stmt))
	if _, execErr := s.conn.Exec(ctx, stmt); execErr != nil {
		if !strings.Contains(execErr.Error(), " not found") {
			return execErr
		}
	}

	return nil
}
