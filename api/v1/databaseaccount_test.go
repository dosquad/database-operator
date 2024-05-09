package v1_test

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
	"github.com/dosquad/database-operator/internal/testhelp"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDatabaseAccount_GetReference(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	ownerRef := dba.GetReference()

	if ownerRef.APIVersion != v1test.DBAPIVersion {
		t.Errorf("dba.GetReference().APIVersion expected '%v' received '%v'", v1test.DBAPIVersion, ownerRef.APIVersion)
	}
	if ownerRef.Kind != "DatabaseAccount" {
		t.Errorf("dba.GetReference().Kind expected '%v' received '%v'", "DatabaseAccount", ownerRef.Kind)
	}
	if ownerRef.Name != v1test.DBAccount {
		t.Errorf("dba.GetReference().Name expected '%v' received '%v'", v1test.DBAccount, ownerRef.Name)
	}
	if ownerRef.UID != v1test.DBUUID {
		t.Errorf("dba.GetReference().UID expected '%v' received '%v'", v1test.DBUUID, ownerRef.UID)
	}
}

func TestGetSecretName(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	if v := dba.GetSecretName(); v.String() != "default/testaccount" {
		t.Errorf("dba.GetSecretName() expected '%v' received '%v'", "default/testaccount", v.String())
	}
}

func TestGetSecretName_Custom(t *testing.T) {
	dba := v1test.NewDatabaseAccount()
	dba.Spec.SecretName = "custom-secret-name"

	if v := dba.GetSecretName(); v.String() != "default/custom-secret-name" {
		t.Errorf("dba.GetSecretName() expected '%v' received '%v'", "default/custom-secret-name", v.String())
	}
}

func TestGetStatefulSetName(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	if v := dba.GetStatefulSetName(); v.String() != "default/testaccount" {
		t.Errorf("dba.GetStatefulSetName() expected '%v' received '%v'", "default/testaccount", v.String())
	}
}

func TestGetStatefulSetName_Custom(t *testing.T) {
	dba := v1test.NewDatabaseAccount()
	dba.Spec.SecretName = "custom-secret-name"

	if v := dba.GetStatefulSetName(); v.String() != "default/custom-secret-name" {
		t.Errorf("dba.GetStatefulSetName() expected '%v' received '%v'", "default/custom-secret-name", v.String())
	}
}

func TestGetDatabaseName(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	v, err := dba.GetDatabaseName()
	if err != nil {
		t.Errorf("dba.GetDatabaseName() expected no error, received '%v'", err)
	}

	if v != v1test.DBUser {
		t.Errorf("dba.GetDatabaseName() expected '%v' received '%v'", v1test.DBUser, v)
	}
}

func TestGetDatabaseName_Error(t *testing.T) {
	dba := v1test.NewDatabaseAccount()
	dba.Status.Name = ""

	_, err := dba.GetDatabaseName()
	if !errors.Is(err, v1.ErrMissingDatabaseUsername) {
		t.Errorf("dba.GetDatabaseName() expected ErrMissingDatabaseUsername error, received '%v'", err)
	}
}

func TestGetSpecOnDelete(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	if v := dba.GetSpecOnDelete(); v != v1.OnDeleteDelete {
		t.Errorf("dba.GetSpecOnDelete() default expected 'delete', received '%v'", v)
	}
}

func TestGetSpecOnDelete_Specified(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

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
	dba := v1test.NewDatabaseAccount()
	dba.Spec.OnDelete = "foobar"

	if v := dba.GetSpecOnDelete(); v != v1.OnDeleteDelete {
		t.Errorf("dba.GetSpecOnDelete() default expected 'delete', received '%v'", v)
	}
}

func TestGetSpecCreateRelay_DefaultFalse(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	if v := dba.GetSpecCreateRelay(); v {
		t.Errorf("dba.GetSpecCreateRelay() default expected 'false', received '%t'", v)
	}
}

func TestGetSpecCreateRelay_Specified(t *testing.T) {
	dba := v1test.NewDatabaseAccount()

	for _, tt := range []bool{true, false} {
		dba.Spec.CreateRelay = tt

		if v := dba.GetSpecCreateRelay(); v != tt {
			t.Errorf("dba.GetSpecCreateRelay() default expected '%v', received '%t'", tt, v)
		}
	}
}

// func getClient(t *testing.T) client.Client {
// 	t.Helper()

// 	var sch = runtime.NewScheme()
// 	var ns = "default"

// 	testenv := &envtest.Environment{CRDDirectoryPaths: []string{"./testdata"}}
// 	cfg, cfgErr := testenv.Start()
// 	if cfgErr != nil {
// 		t.Fatalf("testenv.Start: error '%s'", cfgErr)
// 	}

// 	if err := rbacv1.AddToScheme(sch); err != nil {
// 		t.Fatalf("rbacv1.AddToScheme: error '%s'", err)
// 	}

// 	if err := appsv1.AddToScheme(sch); err != nil {
// 		t.Fatalf("appsv1.AddToScheme: error '%s'", err)
// 	}

// 	nonNamespacedClient, err := client.New(cfg, client.Options{Scheme: sch})
// 	if err != nil {
// 		t.Fatalf("client.New: error '%s'", err)
// 	}
// 	if nonNamespacedClient == nil {
// 		t.Fatalf("client.New: error nonNamespacedClient is nil")
// 	}
// 	return client.NewNamespacedClient(nonNamespacedClient, ns)
// }

func TestUpdateStatus(t *testing.T) {
	t.Parallel()
	start := time.Now()

	dba := v1test.NewDatabaseAccount()
	ctx := context.TODO()
	c := v1test.NewMockClient()

	var changeObject v1.DatabaseAccount

	c.TestStatusWriter.OnUpdate = func(
		_ context.Context, o client.Object,
		_ ...client.SubResourceUpdateOption,
	) error {
		switch v := o.(type) {
		case *v1.DatabaseAccount:
			if v != nil {
				changeObject = *v
			}
		default:
			testhelp.Errorf(t, start, "c.TestStatusWriter.OnUpdate: unknown object type: %+v", v)
		}

		return nil
	}

	if err := dba.UpdateStatus(ctx, c); err != nil {
		testhelp.Errorf(t, start, "dba.UpdateStatus(): error got '%v', want 'nil'", err)
	}

	expectedCallMap := map[string]int{
		"Update": 1,
	}

	if diff := cmp.Diff(c.TestStatusWriter.CallCountMap(), expectedCallMap); diff != "" {
		testhelp.Errorf(t, start, "dba.UpdateStatus(): called functions -got +want:\n%s", diff)
	}

	if diff := cmp.Diff(changeObject, dba); diff != "" {
		testhelp.Errorf(t, start, "dba.UpdateStatus(): change object -got +want:\n%s", diff)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	start := time.Now()

	dba := v1test.NewDatabaseAccount()
	ctx := context.TODO()
	c := v1test.NewMockClient()

	var changeObject v1.DatabaseAccount

	c.OnUpdate = func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
		switch v := obj.(type) {
		case *v1.DatabaseAccount:
			if v != nil {
				changeObject = *v
			}
		default:
			testhelp.Errorf(t, start, "c.TestStatusWriter.OnUpdate: unknown object type: %+v", v)
		}

		return nil
	}

	if err := dba.Update(ctx, c); err != nil {
		testhelp.Errorf(t, start, "dba.Update(): error got '%v', want 'nil'", err)
	}

	expectedCallMap := map[string]int{
		"MockClientWriter.Update": 1,
	}

	if diff := cmp.Diff(c.CallCountMap(), expectedCallMap); diff != "" {
		testhelp.Errorf(t, start, "dba.Update(): called functions -got +want:\n%s", diff)
	}

	if diff := cmp.Diff(changeObject, dba); diff != "" {
		testhelp.Errorf(t, start, "dba.Update(): change object -got +want:\n%s", diff)
	}
}

//nolint:gocognit // Table driven test.
func TestSetStage(t *testing.T) {
	t.Parallel()
	start := time.Now()
	ctx := context.TODO()

	tests := []struct {
		name                 string
		state                v1.DatabaseAccountCreateStage
		expectedUpdateCount  int
		applyExpectedChanges func(*v1.DatabaseAccount)
	}{
		{"", v1.UnknownStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.UnknownStage }},
		{"", v1.InitStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.InitStage }},
		{"", v1.UserCreateStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.UserCreateStage }},
		{"", v1.DatabaseCreateStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.DatabaseCreateStage }},
		{"", v1.RelayCreateStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.RelayCreateStage }},
		{"", v1.ErrorStage, 1, func(da *v1.DatabaseAccount) { da.Status.Stage = v1.ErrorStage }},
		{"", v1.ReadyStage, 0, func(da *v1.DatabaseAccount) { // If stage is not changing, Update should not be called.
			*da = v1.DatabaseAccount{}
		}},
		{"", v1.TerminatingStage, 1, func(da *v1.DatabaseAccount) {
			da.Status.Stage = v1.TerminatingStage
			da.Status.Ready = false
		}},
	}

	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.state.String()
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dba := v1test.NewDatabaseAccount()
			c := v1test.NewMockClient()

			var changeObject v1.DatabaseAccount

			c.TestStatusWriter.OnUpdate = func(
				_ context.Context, o client.Object,
				_ ...client.SubResourceUpdateOption,
			) error {
				switch v := o.(type) {
				case *v1.DatabaseAccount:
					if v != nil {
						changeObject = *v
					}
				default:
					testhelp.Errorf(t, start, "c.TestStatusWriter.OnUpdate: unknown object type: %+v", v)
				}

				return nil
			}

			if err := dba.SetStage(ctx, c, tt.state); err != nil {
				testhelp.Errorf(t, start, "dba.SetStage(): error, got '%v', want 'nil'", err)
			}

			expectedCallMap := map[string]int{
				"MockStatusClient.Status":                  1,
				"MockStatusClient.TestStatusWriter.Update": 1,
			}
			expectedStatusCallMap := map[string]int{
				"Update": tt.expectedUpdateCount,
			}
			if tt.expectedUpdateCount == 0 {
				expectedCallMap = map[string]int{}
				expectedStatusCallMap = map[string]int{}
			}

			if diff := cmp.Diff(c.CallCountMap(), expectedCallMap); diff != "" {
				testhelp.Errorf(t, start, "dba.SetStage(): client, called functions -got +want:\n%s", diff)
			}

			if diff := cmp.Diff(c.TestStatusWriter.CallCountMap(), expectedStatusCallMap); diff != "" {
				testhelp.Errorf(t, start, "dba.SetStage(): client.StatusWriter, called functions -got +want:\n%s", diff)
			}

			originaldba := v1test.NewDatabaseAccount()
			tt.applyExpectedChanges(&originaldba)

			if diff := cmp.Diff(changeObject, originaldba); diff != "" {
				testhelp.Errorf(t, start, "dba.SetStage(): object change -got +want:\n%s", diff)
			}
		})
	}
}
