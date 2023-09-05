package v1_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	v1 "github.com/dosquad/database-operator/api/v1"
)

const (
	dbUser       = `k8s_01h97g9exfs6bw874x0k567jr7`
	dbApp        = "testapp"
	dbAccount    = "testaccount"
	dbAPIVersion = "dbo.dosquad.github.io/v1"
	dbUUID       = "a6533e70-9716-4c68-91a1-851a6fef074f"
	jsonData     = `{
		"apiVersion": "` + dbAPIVersion + `",
		"kind": "DatabaseAccount",
		"metadata": {
			"annotations": {},
			"creationTimestamp": "2023-04-03T05:06:07Z",
			"generation": 1,
			"labels": {
				"app": "` + dbApp + `"
			},
			"name": "` + dbAccount + `",
			"namespace": "default",
			"resourceVersion": "10",
			"uid": "` + dbUUID + `"
		},
		"status": {
			"name": "` + dbUser + `",
			"ready": true,
			"stage": "Ready"
		}
	}`
)

func newDatabaseAccount() v1.DatabaseAccount {
	dbac := v1.DatabaseAccount{}

	if err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&dbac); err != nil {
		panic(err)
	}

	return dbac
}

func TestDatabaseAccount_GetReference(t *testing.T) {
	dba := newDatabaseAccount()

	ownerRef := dba.GetReference()

	if ownerRef.APIVersion != dbAPIVersion {
		t.Errorf("dba.GetReference().APIVersion expected '%v' received '%v'", dbAPIVersion, ownerRef.APIVersion)
	}
	if ownerRef.Kind != "DatabaseAccount" {
		t.Errorf("dba.GetReference().Kind expected '%v' received '%v'", "DatabaseAccount", ownerRef.Kind)
	}
	if ownerRef.Name != dbAccount {
		t.Errorf("dba.GetReference().Name expected '%v' received '%v'", dbAccount, ownerRef.Name)
	}
	if ownerRef.UID != dbUUID {
		t.Errorf("dba.GetReference().UID expected '%v' received '%v'", dbUUID, ownerRef.UID)
	}
}

func TestGetSecretName(t *testing.T) {
	dba := newDatabaseAccount()

	if v := dba.GetSecretName(); v.String() != "default/testaccount" {
		t.Errorf("dba.GetSecretName() expected '%v' received '%v'", "default/testaccount", v.String())
	}
}

func TestGetSecretName_Custom(t *testing.T) {
	dba := newDatabaseAccount()
	dba.Spec.SecretName = "custom-secret-name"

	if v := dba.GetSecretName(); v.String() != "default/custom-secret-name" {
		t.Errorf("dba.GetSecretName() expected '%v' received '%v'", "default/custom-secret-name", v.String())
	}
}

func TestGetStatefulSetName(t *testing.T) {
	dba := newDatabaseAccount()

	if v := dba.GetStatefulSetName(); v.String() != "default/testaccount" {
		t.Errorf("dba.GetStatefulSetName() expected '%v' received '%v'", "default/testaccount", v.String())
	}
}

func TestGetStatefulSetName_Custom(t *testing.T) {
	dba := newDatabaseAccount()
	dba.Spec.SecretName = "custom-secret-name"

	if v := dba.GetStatefulSetName(); v.String() != "default/custom-secret-name" {
		t.Errorf("dba.GetStatefulSetName() expected '%v' received '%v'", "default/custom-secret-name", v.String())
	}
}

func TestGetDatabaseName(t *testing.T) {
	dba := newDatabaseAccount()

	v, err := dba.GetDatabaseName()
	if err != nil {
		t.Errorf("dba.GetDatabaseName() expected no error, received '%v'", err)
	}

	if v != dbUser {
		t.Errorf("dba.GetDatabaseName() expected '%v' received '%v'", dbUser, v)
	}
}

func TestGetDatabaseName_Error(t *testing.T) {
	dba := newDatabaseAccount()
	dba.Status.Name = ""

	_, err := dba.GetDatabaseName()
	if !errors.Is(err, v1.ErrMissingDatabaseUsername) {
		t.Errorf("dba.GetDatabaseName() expected ErrMissingDatabaseUsername error, received '%v'", err)
	}
}

func TestGetSpecOnDelete(t *testing.T) {
	dba := newDatabaseAccount()

	if v := dba.GetSpecOnDelete(); v != v1.OnDeleteDelete {
		t.Errorf("dba.GetSpecOnDelete() default expected 'delete', received '%v'", v)
	}
}

func TestGetSpecOnDelete_Specified(t *testing.T) {
	dba := newDatabaseAccount()

	for _, tt := range []v1.DatabaseAccountOnDelete{
		v1.OnDeleteDelete,
		v1.OnDeleteRetain,
	} {
		dba.Spec.OnDelete = tt

		if v := dba.GetSpecOnDelete(); v != tt {
			t.Errorf("dba.GetSpecOnDelete() default expected '%v', received '%v'", tt, v)
		}
	}
}

func TestGetSpecOnDelete_InvalidReturnDefault(t *testing.T) {
	dba := newDatabaseAccount()
	dba.Spec.OnDelete = "foobar"

	if v := dba.GetSpecOnDelete(); v != v1.OnDeleteDelete {
		t.Errorf("dba.GetSpecOnDelete() default expected 'delete', received '%v'", v)
	}
}

func TestGetSpecCreateRelay_DefaultFalse(t *testing.T) {
	dba := newDatabaseAccount()

	if v := dba.GetSpecCreateRelay(); v {
		t.Errorf("dba.GetSpecCreateRelay() default expected 'false', received '%t'", v)
	}
}

func TestGetSpecCreateRelay_Specified(t *testing.T) {
	dba := newDatabaseAccount()

	for _, tt := range []bool{true, false} {
		dba.Spec.CreateRelay = tt

		if v := dba.GetSpecCreateRelay(); v != tt {
			t.Errorf("dba.GetSpecCreateRelay() default expected '%v', received '%t'", tt, v)
		}
	}
}
