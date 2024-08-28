package accountsvr

import (
	"context"

	dbov1 "github.com/dosquad/database-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type Server interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	ListUsers(ctx context.Context) []string
	IsRole(ctx context.Context, roleName string) (bool, error)
	IsDatabase(ctx context.Context, dbName string) (string, bool, error)
	CreateRole(ctx context.Context, roleName string) (string, string, error)
	UpdateRolePassword(ctx context.Context, roleName string) (string, string, error)
	CreateDatabase(ctx context.Context, dbName, roleName string) (string, error)
	CreateSchema(ctx context.Context, schemaName, roleName string) error
	GetDatabaseHostConfig() string
	GetDatabaseHost(dbAccount *dbov1.DatabaseAccount) string
	CopyInitConfigToSecret(dbAccount *dbov1.DatabaseAccount, secret *corev1.Secret)
	Delete(ctx context.Context, name string) error

	// connect(ctx context.Context) error
	// generatePassword(ctx context.Context) string
}
