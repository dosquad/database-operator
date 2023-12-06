package test

import (
	v1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type ReconcilePreModfunc func(obj *v1.DatabaseAccount)
type ReconcileModDBFunc func(want *v1.DatabaseAccount)

type ReconcileTarget interface {
	ReconcileModDBFunc | ReconcileModSecretFunc
}

func ReconcileWantStage(wantStage v1.DatabaseAccountCreateStage) ReconcileModDBFunc {
	return func(want *v1.DatabaseAccount) {
		want.Status.Stage = wantStage
	}
}

func ReconcileWantReady(ready bool) ReconcileModDBFunc {
	return func(want *v1.DatabaseAccount) {
		want.Status.Ready = ready
	}
}

func ReconcileWantDBName(name v1.PostgreSQLResourceName) ReconcileModDBFunc {
	return func(want *v1.DatabaseAccount) {
		want.Status.Name = name
	}
}

func ReconcileWantDBFinalizer(want *v1.DatabaseAccount) {
	want.ObjectMeta.Finalizers = []string{"dbo.dosquad.github.io/database-account-finalizer"}
}

func ReconcileStatusDBUser(want *v1.DatabaseAccount) {
	want.Status.Name = v1test.DBUser
}

func CompareAccountsIgnore() cmp.Option {
	return cmpopts.IgnoreFields(
		v1.DatabaseAccount{},
		"Status.Name",
	)
}

func CompareEventsIgnore() cmp.Option {
	return cmpopts.IgnoreFields(
		v1test.MockRecorderMessage{},
		"Message",
	)
}
