package controller_test

import (
	"context"
	"testing"

	v1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
	"github.com/dosquad/database-operator/internal/controller"
	controllertest "github.com/dosquad/database-operator/internal/controller/test"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile_Stage_Unknown_AddFinalizer(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{ // this is both loops of Reconcile (adding finalizer then UnknownStage)
			"MockClientReader.Get":    1,
			"MockClientWriter.Update": 1,
		},
		expectServerCallMap:   map[string]int{},
		expectRecorderCallMap: map[string]int{},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)":    1,
			"MockClientWriter.Update(*v1.DatabaseAccount)": 1,
		},
		expectNormalMessage:  []v1test.MockRecorderMessage{},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(t, v1.UnknownStage, nil, nil)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantDBFinalizer,
	}

	testReconcileResultsTestSet(ts, expect)
}

func TestReconcile_Stage_Unknown_WithFinalizer(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{ // this is both loops of Reconcile (adding finalizer then UnknownStage)
			"MockClientReader.Get":                     1,
			"MockStatusClient.Status":                  1,
			"MockStatusClient.TestStatusWriter.Update": 1,
		},
		expectServerCallMap:   map[string]int{},
		expectRecorderCallMap: map[string]int{},
		expectTestCallMap: map[string]int{ // status updates aren't included.
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
		},
		expectNormalMessage:  []v1test.MockRecorderMessage{},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(
		t, v1.UnknownStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBFinalizer,
		},
		nil,
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.InitStage),
		controllertest.ReconcileWantDBFinalizer,
	}
	// ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
	// 	controllertest.ReconcileWantSecretInit,
	// }

	testReconcileResultsTestSet(ts, expect)
}

func TestReconcile_Stage_Init(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{
			"MockClientReader.Get":                     2,
			"MockClientWriter.Create":                  1,
			"MockClientWriter.Update":                  1,
			"MockStatusClient.Status":                  1,
			"MockStatusClient.TestStatusWriter.Update": 1,
		},
		expectServerCallMap: map[string]int{
			"CopyInitConfigToSecret": 2,
			"GetDatabaseHost":        2,
			"GetDatabaseHostConfig":  2,
		},
		expectRecorderCallMap: map[string]int{
			"NormalEvent": 1,
		},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
			"MockClientReader.Get(*v1.Secret)":          1,
			"MockClientWriter.Create(*v1.Secret)":       1,
			"MockClientWriter.Update(*v1.Secret)":       1,
		},
		expectNormalMessage: []v1test.MockRecorderMessage{
			v1test.NewMockRecorderMessage(controller.ReasonQueued, ""),
		},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(
		t, v1.InitStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
			controllertest.ReconcileWantDBFinalizer,
		},
		nil,
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.UserCreateStage),
		controllertest.ReconcileWantDBFinalizer,
	}
	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
		controllertest.ReconcileWantSecretInit,
	}

	testReconcileResultsTestSet(ts, expect)
}

func TestReconcile_Stage_UserCreate(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{
			"MockClientReader.Get":                     2,
			"MockClientWriter.Update":                  1,
			"MockStatusClient.Status":                  1,
			"MockStatusClient.TestStatusWriter.Update": 1,
		},
		expectServerCallMap: map[string]int{
			"CreateRole": 1,
		},
		expectRecorderCallMap: map[string]int{
			"NormalEvent": 3,
		},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
			"MockClientReader.Get(*v1.Secret)":          1,
			"MockClientWriter.Update(*v1.Secret)":       1,
		},
		expectNormalMessage: []v1test.MockRecorderMessage{
			v1test.NewMockRecorderMessage(controller.ReasonUserCreate, ""),
			v1test.NewMockRecorderMessage(controller.ReasonUserCreate, ""),
			v1test.NewMockRecorderMessage(controller.ReasonDatabaseCreate, ""),
		},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(
		t, v1.UserCreateStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBFinalizer,
			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
		},
		[]controllertest.ReconcileModSecretFunc{
			controllertest.ReconcileWantSecretInit,
		},
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.DatabaseCreateStage),
		controllertest.ReconcileWantDBFinalizer,
	}
	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
		controllertest.ReconcileWantSecretNamePassword,
	}

	testReconcileResultsTestSet(ts, expect)
}

func TestReconcile_Stage_DatabaseCreate(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{
			"MockClientReader.Get":                     2,
			"MockClientWriter.Update":                  1,
			"MockStatusClient.Status":                  1,
			"MockStatusClient.TestStatusWriter.Update": 1,
		},
		expectServerCallMap: map[string]int{
			"CreateDatabase": 1,
			"IsDatabase":     1,
		},
		expectRecorderCallMap: map[string]int{
			"NormalEvent": 2,
		},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
			"MockClientReader.Get(*v1.Secret)":          1,
			"MockClientWriter.Update(*v1.Secret)":       1,
		},
		expectNormalMessage: []v1test.MockRecorderMessage{
			v1test.NewMockRecorderMessage(controller.ReasonDatabaseCreate, ""),
			v1test.NewMockRecorderMessage(controller.ReasonReady, ""),
		},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(
		t, v1.DatabaseCreateStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBFinalizer,
			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
		},
		[]controllertest.ReconcileModSecretFunc{
			controllertest.ReconcileWantSecretInit,
			controllertest.ReconcileWantSecretNamePassword,
		},
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.ReadyStage),
		controllertest.ReconcileWantReady(true),
		controllertest.ReconcileWantDBFinalizer,
	}
	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
		controllertest.ReconcileWantSecretDatabaseDSN,
		controllertest.ReconcileWantSecretNamePassword,
	}

	ts.svr.OnIsDatabase = func(_ context.Context, dbName string) (string, bool, error) {
		return dbName, false, nil
	}

	testReconcileResultsTestSet(ts, expect)
}

func TestReconcile_Stage_DatabaseCreate_DatabaseExists(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{
			"MockClientReader.Get":                     2,
			"MockClientWriter.Update":                  1,
			"MockStatusClient.Status":                  1,
			"MockStatusClient.TestStatusWriter.Update": 1,
		},
		expectServerCallMap: map[string]int{
			"IsDatabase": 1,
		},
		expectRecorderCallMap: map[string]int{
			"NormalEvent":  1,
			"WarningEvent": 1,
		},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
			"MockClientReader.Get(*v1.Secret)":          1,
			"MockClientWriter.Update(*v1.Secret)":       1,
		},
		expectNormalMessage: []v1test.MockRecorderMessage{
			v1test.NewMockRecorderMessage(controller.ReasonReady, ""),
		},
		expectWarningMessage: []v1test.MockRecorderMessage{
			v1test.NewMockRecorderMessage(controller.ReasonDatabaseCreate, "Database already exists"),
		},
	}
	ts := newTestSet(
		t, v1.DatabaseCreateStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBFinalizer,
			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
		},
		[]controllertest.ReconcileModSecretFunc{
			controllertest.ReconcileWantSecretInit,
			controllertest.ReconcileWantSecretNamePassword,
		},
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.ReadyStage),
		controllertest.ReconcileWantReady(true),
		controllertest.ReconcileWantDBFinalizer,
	}
	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
		controllertest.ReconcileWantSecretDatabaseDSN,
		controllertest.ReconcileWantSecretNamePassword,
	}

	testReconcileResultsTestSet(ts, expect)
}

// func TestReconcile_Stage_RelayCreate(t *testing.T) {
// 	t.Parallel()
// 	expect := expectSet{
// 		expectResult: reconcile.Result{},
// 		expectClientCallMap: map[string]int{
// 			"MockClientReader.Get":                     3,
// 			"MockClientWriter.Create":                  2,
// 			"MockStatusClient.Status":                  1,
// 			"MockStatusClient.TestStatusWriter.Update": 1,
// 		},
// 		expectServerCallMap: map[string]int{},
// 		expectRecorderCallMap: map[string]int{
// 			"NormalEvent": 3,
// 		},
// 		expectTestCallMap: map[string]int{
// 			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
// 			"MockClientReader.Get(*v1.Service)":         1,
// 			"MockClientReader.Get(*v1.StatefulSet)":     1,
// 			"MockClientWriter.Create(*v1.Service)":      1,
// 			"MockClientWriter.Create(*v1.StatefulSet)":  1,
// 		},
// 		expectNormalMessage: []v1test.MockRecorderMessage{
// 			v1test.NewMockRecorderMessage(controller.ReasonRelayCreate, "Created StatefulSet"),
// 			v1test.NewMockRecorderMessage(controller.ReasonRelayCreate, "Created Service"),
// 			v1test.NewMockRecorderMessage(controller.ReasonReady, "Ready to use"),
// 		},
// 		expectWarningMessage: []v1test.MockRecorderMessage{},
// 	}
// 	ts := newTestSet(
// 		t, v1.RelayCreateStage,
// 		[]controllertest.ReconcileModDBFunc{
// 			controllertest.ReconcileWantDBFinalizer,
// 			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName())},
// 		[]controllertest.ReconcileModSecretFunc{
// 			controllertest.ReconcileWantSecretInit,
// 			controllertest.ReconcileWantSecretNamePassword,
// 		},
// 	)
// 	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
// 		controllertest.ReconcileWantStage(v1.ReadyStage),
// 		controllertest.ReconcileWantReady(true),
// 		controllertest.ReconcileWantDBFinalizer,
// 	}
// 	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
// 		controllertest.ReconcileWantSecretNamePassword,
// 	}

// 	ts.svr.OnIsDatabase = func(ctx context.Context, dbName string) (string, bool, error) {
// 		return dbName, false, nil
// 	}

// 	testReconcileResultsTestSet(ts, expect)
// }

// func TestReconcile_Stage_RelayCreate_TwoRounds(t *testing.T) {
// 	t.Parallel()
// 	expect := expectSet{
// 		expectResult: reconcile.Result{},
// 		expectClientCallMap: map[string]int{
// 			"MockClientReader.Get":                     6,
// 			"MockClientWriter.Create":                  2,
// 			"MockStatusClient.Status":                  1,
// 			"MockStatusClient.Update":                  1,
// 			"MockStatusClient.TestStatusWriter.Update": 1,
// 		},
// 		expectServerCallMap: map[string]int{},
// 		expectRecorderCallMap: map[string]int{
// 			"NormalEvent": 3,
// 		},
// 		expectTestCallMap: map[string]int{
// 			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
// 			"MockClientReader.Get(*v1.Service)":         1,
// 			"MockClientReader.Get(*v1.StatefulSet)":     1,
// 			"MockClientWriter.Create(*v1.Service)":      1,
// 			"MockClientWriter.Create(*v1.StatefulSet)":  1,
// 		},
// 		expectNormalMessage: []v1test.MockRecorderMessage{
// 			v1test.NewMockRecorderMessage(controller.ReasonRelayCreate, "Created StatefulSet"),
// 			v1test.NewMockRecorderMessage(controller.ReasonRelayCreate, "Created Service"),
// 			v1test.NewMockRecorderMessage(controller.ReasonReady, "Ready to use"),
// 		},
// 		expectWarningMessage: []v1test.MockRecorderMessage{},
// 	}
// 	ts := newTestSet(
// 		t, v1.RelayCreateStage,
// 		[]controllertest.ReconcileModDBFunc{
// 			controllertest.ReconcileWantDBFinalizer,
// 			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName())},
// 		[]controllertest.ReconcileModSecretFunc{
// 			controllertest.ReconcileWantSecretInit,
// 			controllertest.ReconcileWantSecretNamePassword,
// 		},
// 	)
// 	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
// 		controllertest.ReconcileWantStage(v1.ReadyStage),
// 		controllertest.ReconcileWantReady(true),
// 		controllertest.ReconcileWantDBFinalizer,
// 	}
// 	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
// 		controllertest.ReconcileWantSecretNamePassword,
// 	}
// 	ts.reconcileLoop = 2

// 	ts.svr.OnIsDatabase = func(ctx context.Context, dbName string) (string, bool, error) {
// 		return dbName, false, nil
// 	}

// 	testReconcileResultsTestSet(ts, expect)
// }

func TestReconcile_Stage_Ready(t *testing.T) {
	t.Parallel()
	expect := expectSet{
		expectResult: reconcile.Result{},
		expectClientCallMap: map[string]int{
			"MockClientReader.Get":    3,
			"MockClientWriter.Update": 1,
		},
		expectServerCallMap:   map[string]int{},
		expectRecorderCallMap: map[string]int{},
		expectTestCallMap: map[string]int{
			"MockClientReader.Get(*v1.DatabaseAccount)": 1,
			"MockClientReader.Get(*v1.Secret)":          2,
			"MockClientWriter.Update(*v1.Secret)":       1,
		},
		expectNormalMessage:  []v1test.MockRecorderMessage{},
		expectWarningMessage: []v1test.MockRecorderMessage{},
	}
	ts := newTestSet(
		t, v1.ReadyStage,
		[]controllertest.ReconcileModDBFunc{
			controllertest.ReconcileWantDBFinalizer,
			controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
			controllertest.ReconcileWantReady(true),
		},
		[]controllertest.ReconcileModSecretFunc{
			controllertest.ReconcileWantSecretInit,
			controllertest.ReconcileWantSecretDatabaseDSN,
			controllertest.ReconcileWantSecretNamePassword,
		},
	)
	ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
		controllertest.ReconcileWantStage(v1.ReadyStage),
		controllertest.ReconcileWantDBFinalizer,
	}
	ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
		controllertest.ReconcileWantSecretOwnerRefs(ts.ctr.GetDatabaseAccount()),
	}

	// ts.svr.OnIsDatabase = func(ctx context.Context, dbName string) (string, bool, error) {
	// 	return dbName, false, nil
	// }

	testReconcileResultsTestSet(ts, expect)
}

// func TestReconcile_Stage_Ready(t *testing.T) {
// 	t.Parallel()
// 	expect := expectSet{
// 		expectResult: reconcile.Result{},
// 		expectClientCallMap: map[string]int{
// 			"MockClientReader.Get":    1,
// 			"MockClientWriter.Update": 1,
// 		},
// 		expectServerCallMap:   map[string]int{},
// 		expectRecorderCallMap: map[string]int{},
// 		expectTestCallMap:     map[string]int{},
// 		expectNormalMessage:   []v1test.MockRecorderMessage{},
// 		expectWarningMessage:  []v1test.MockRecorderMessage{},
// 	}
// 	for _, initFinalizerExists := range []bool{false, true} {
// 		initFinalizerExists := initFinalizerExists
// 		expect := expect
// 		ts := newTestSet(
// 			t, v1.ReadyStage,
// 			[]controllertest.ReconcileModDBFunc{
// 				controllertest.ReconcileWantReady(true),
// 				controllertest.ReconcileWantDBName(controllertest.NewDatabaseAccountName()),
// 				func(want *v1.DatabaseAccount) {
// 					if initFinalizerExists {
// 						controllertest.ReconcileWantDBFinalizer(want)
// 					}
// 				},
// 			},
// 			[]controllertest.ReconcileModSecretFunc{
// 				controllertest.ReconcileWantSecretInit,
// 				controllertest.ReconcileWantSecretNamePassword,
// 			},
// 		)
// 		ts.reconcileModDBAccount = []controllertest.ReconcileModDBFunc{
// 			controllertest.ReconcileWantStage(v1.ReadyStage),
// 			controllertest.ReconcileWantReady(true),
// 			controllertest.ReconcileWantDBFinalizer,
// 		}
// 		ts.reconcileModSecret = []controllertest.ReconcileModSecretFunc{
// 			controllertest.ReconcileWantSecretOwnerRefs(
// 				ts.ctr.GetDatabaseAccount(),
// 			),
// 			controllertest.ReconcileWantSecretDatabaseDSN,
// 			controllertest.ReconcileWantSecretNamePassword,
// 		}
// 		if initFinalizerExists {
// 			expect.expectClientCallMap = map[string]int{
// 				"MockClientReader.Get": 3,
// 			}
// 		}

// 		t.Run(fmt.Sprintf("InitialFinalizer:%t", initFinalizerExists), func(t *testing.T) {
// 			t.Parallel()

// 			testReconcileResultsTestSet(ts, expect)
// 		})

// 	}

// }
