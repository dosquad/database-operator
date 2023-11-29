package test

import (
	"context"
	"net/url"

	dbov1 "github.com/dosquad/database-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	TestDSN = "postgres://adminuser:adminpass@databasehost:5432/k8s_01h97g9exfs6bw874x0k567jr7"
)

type MockServer struct {
	dsn        *url.URL
	calledFunc map[string]int

	OnCheckInvalidName       func(name string) (string, error)
	OnConnect                func(ctx context.Context) error
	OnClose                  func(ctx context.Context) error
	OnListUsers              func(ctx context.Context) []string
	OnIsRole                 func(ctx context.Context, roleName string) (bool, error)
	OnIsDatabase             func(ctx context.Context, dbName string) (string, bool, error)
	OnCreateRole             func(ctx context.Context, roleName string) (string, string, error)
	OnUpdateRolePassword     func(ctx context.Context, roleName string) (string, string, error)
	OnCreateDatabase         func(ctx context.Context, dbName, roleName string) (string, error)
	OnCreateSchema           func(ctx context.Context, schemaName, roleName string) error
	OnGetDatabaseHostConfig  func() string
	OnGetDatabaseHost        func(dbAccount *dbov1.DatabaseAccount) string
	OnCopyInitConfigToSecret func(dbAccount *dbov1.DatabaseAccount, secret *corev1.Secret)
	OnDelete                 func(ctx context.Context, name string) error
}

func NewMockServer(dsn string) *MockServer {
	u, _ := url.Parse(dsn)
	m := &MockServer{
		dsn: u,
	}
	m.CallCountReset()

	return m
}

func (m *MockServer) CallCountMap() map[string]int {
	return m.calledFunc
}

func (m *MockServer) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (m *MockServer) CallCountReset() {
	m.calledFunc = map[string]int{}
}

func (m *MockServer) CheckInvalidName(name string) (string, error) {
	m.calledFunc["CheckInvalidName"]++
	if m.OnCheckInvalidName != nil {
		return m.OnCheckInvalidName(name)
	}

	return "", nil
}

func (m *MockServer) Connect(ctx context.Context) error {
	m.calledFunc["Connect"]++
	if m.OnConnect != nil {
		return m.OnConnect(ctx)
	}

	return nil
}

func (m *MockServer) Close(ctx context.Context) error {
	m.calledFunc["Close"]++
	if m.OnClose != nil {
		return m.OnClose(ctx)
	}

	return nil
}

func (m *MockServer) ListUsers(ctx context.Context) []string {
	m.calledFunc["ListUsers"]++
	if m.OnListUsers != nil {
		return m.OnListUsers(ctx)
	}

	return nil
}

func (m *MockServer) IsRole(ctx context.Context, roleName string) (bool, error) {
	m.calledFunc["IsRole"]++
	if m.OnIsRole != nil {
		return m.OnIsRole(ctx, roleName)
	}

	return true, nil
}

func (m *MockServer) IsDatabase(ctx context.Context, dbName string) (string, bool, error) {
	m.calledFunc["IsDatabase"]++
	if m.OnIsDatabase != nil {
		return m.OnIsDatabase(ctx, dbName)
	}

	return dbName, true, nil
}

func (m *MockServer) CreateRole(ctx context.Context, roleName string) (string, string, error) {
	m.calledFunc["CreateRole"]++
	if m.OnCreateRole != nil {
		return m.OnCreateRole(ctx, roleName)
	}

	// return roleName, password nil
	return roleName, "mockpassword", nil
}

func (m *MockServer) UpdateRolePassword(ctx context.Context, roleName string) (string, string, error) {
	m.calledFunc["UpdateRolePassword"]++
	if m.OnUpdateRolePassword != nil {
		return m.OnUpdateRolePassword(ctx, roleName)
	}

	return roleName, "", nil
}

func (m *MockServer) CreateDatabase(ctx context.Context, dbName, roleName string) (string, error) {
	m.calledFunc["CreateDatabase"]++
	if m.OnCreateDatabase != nil {
		return m.OnCreateDatabase(ctx, dbName, roleName)
	}

	return dbName, nil
}

func (m *MockServer) CreateSchema(ctx context.Context, schemaName, roleName string) error {
	m.calledFunc["CreateSchema"]++
	if m.OnCreateSchema != nil {
		return m.OnCreateSchema(ctx, schemaName, roleName)
	}

	return nil
}

func (m *MockServer) GetDatabaseHostConfig() string {
	m.calledFunc["GetDatabaseHostConfig"]++
	if m.OnGetDatabaseHostConfig != nil {
		return m.OnGetDatabaseHostConfig()
	}

	return m.dsn.Hostname()
}

func (m *MockServer) GetDatabaseHost(dbAccount *dbov1.DatabaseAccount) string {
	m.calledFunc["GetDatabaseHost"]++
	if m.OnGetDatabaseHost != nil {
		return m.OnGetDatabaseHost(dbAccount)
	}

	if dbAccount.GetSpecCreateRelay() {
		return dbAccount.GetStatefulSetName().Name
	}

	return m.GetDatabaseHostConfig()
}

func (m *MockServer) CopyInitConfigToSecret(dbAccount *dbov1.DatabaseAccount, secret *corev1.Secret) {
	m.calledFunc["CopyInitConfigToSecret"]++
	if m.OnCopyInitConfigToSecret != nil {
		m.OnCopyInitConfigToSecret(dbAccount, secret)
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["host"] = []byte(m.GetDatabaseHost(dbAccount))
	secret.Data["port"] = []byte(m.dsn.Port())
}

func (m *MockServer) Delete(ctx context.Context, name string) error {
	m.calledFunc["Delete"]++
	if m.OnDelete != nil {
		return m.OnDelete(ctx, name)
	}

	return nil
}
