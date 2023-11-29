package test

import (
	"fmt"
	"net/url"

	"github.com/dosquad/database-operator/accountsvr"
	accountsvrtest "github.com/dosquad/database-operator/accountsvr/test"
	v1 "github.com/dosquad/database-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReconcileModSecretFunc func(want *corev1.Secret)

func ReconcileWantSecretOwnerRefs(dbAccount v1.DatabaseAccount) ReconcileModSecretFunc {
	return func(want *corev1.Secret) {
		want.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			*dbAccount.GetReference(),
		}
	}
}

func ReconcileWantSecretFinalizer(want *corev1.Secret) {
	want.ObjectMeta.Finalizers = []string{"dbo.dosquad.github.io/database-account-finalizer"}
}

func ReconcileWantSecretNamePassword(want *corev1.Secret) {
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyUsername, NewDatabaseAccountName().String())(want)
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyPassword, "mockpassword")(want)
}

func ReconcileWantSecretInit(want *corev1.Secret) {
	u, _ := url.Parse(accountsvrtest.TestDSN)
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyHost, u.Hostname())(want)
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyPort, u.Port())(want)
}

func ReconcileWantSecretDatabaseDSN(want *corev1.Secret) {
	// dsn := fmt.Sprintf(
	// 	"postgres://k8s_01h97g9exfs6bw874x0k567jr7:mockpassword@k8s_01h97g9exfs6bw874x0k567jr7",
	// )
	dbName := NewDatabaseAccountName().String()
	u, _ := url.Parse(accountsvrtest.TestDSN)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", dbName, "mockpassword", u.Host, dbName)
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyDatabase, dbName)(want)
	ReconcileWantSecretDataValue(accountsvr.DatabaseKeyDSN, dsn)(want)
}

func ReconcileWantSecretDataValue(key, value string) ReconcileModSecretFunc {
	return func(want *corev1.Secret) {
		if want.StringData == nil {
			want.StringData = map[string]string{}
		}

		if want.Data == nil {
			want.Data = map[string][]byte{}
		}

		want.Data[key] = []byte(value)
	}
}
