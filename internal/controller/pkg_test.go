package controller_test

import (
	"context"
	"testing"
	"time"

	accountsvrtest "github.com/dosquad/database-operator/accountsvr/test"
	dbov1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
	"github.com/dosquad/database-operator/internal/controller"
	controllertest "github.com/dosquad/database-operator/internal/controller/test"
	"github.com/dosquad/database-operator/internal/testhelp"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type expectSet struct {
	expectResult          reconcile.Result
	expectNormalMessage   []v1test.MockRecorderMessage
	expectWarningMessage  []v1test.MockRecorderMessage
	expectClientCallMap   map[string]int
	expectServerCallMap   map[string]int
	expectRecorderCallMap map[string]int
	// expectTestCallMap only includes mocked functions that do something (overridden).
	expectTestCallMap map[string]int
}

type testSet struct {
	t                     *testing.T
	start                 time.Time
	stage                 dbov1.DatabaseAccountCreateStage
	reconcileLoop         int
	ctx                   context.Context
	ctr                   *controllertest.ControllerMockWrapper
	c                     *v1test.MockClient
	rcd                   *v1test.MockRecorder
	svr                   *accountsvrtest.MockServer
	rec                   *controller.DatabaseAccountReconciler
	reconcileModDBAccount []controllertest.ReconcileModDBFunc
	reconcileModSecret    []controllertest.ReconcileModSecretFunc
}

func newTestSet(
	t *testing.T,
	stage dbov1.DatabaseAccountCreateStage,
	initDBFunc []controllertest.ReconcileModDBFunc,
	initSecretFunc []controllertest.ReconcileModSecretFunc,
) testSet {
	t.Helper()

	ts := testSet{
		t:             t,
		start:         time.Now(),
		stage:         stage,
		reconcileLoop: 1,
	}
	ts.ctx, _ = newLoggerContext(ts.t, ts.start)
	ts.c, ts.rcd, ts.svr, ts.rec = newReconciler(ts.t, ts.start)

	ts.ctr = controllertest.NewControllerMockWrapper(ts.t, ts.start, ts.c)
	ts.ctr.InitDatabaseAccount(func(dba *dbov1.DatabaseAccount) {
		controllertest.ReconcileWantStage(stage)(dba)
		for _, f := range initDBFunc {
			f(dba)
		}
	})
	if initSecretFunc != nil {
		ts.ctr.InitSecrets(func(secret *corev1.Secret) {
			for _, f := range initSecretFunc {
				f(secret)
			}
		})
	}

	return ts
}

func newReconciler(_ *testing.T, _ time.Time) (
	*v1test.MockClient,
	*v1test.MockRecorder,
	*accountsvrtest.MockServer,
	*controller.DatabaseAccountReconciler,
) {
	scheme := runtime.NewScheme()
	ctrlConfig := dbov1.DatabaseAccountControllerConfig{}
	rcd := v1test.NewRecorder()
	c := v1test.NewMockClient()
	svr := accountsvrtest.NewMockServer(accountsvrtest.TestDSN)
	rec := &controller.DatabaseAccountReconciler{
		Client:        c,
		Scheme:        scheme,
		Recorder:      rcd,
		AccountServer: svr,
		Config:        &ctrlConfig,
	}

	return c, rcd, svr, rec
}

func newLoggerContext(t *testing.T, start time.Time) (context.Context, *testhelp.TestingLogWrapper) {
	t.Helper()

	ctx := context.TODO()
	l := testhelp.NewTestingLogWrapper(t, start)
	tl := logr.New(l)
	ctx = logr.NewContext(ctx, tl)

	return ctx, l
}

func diffOriginalDatabaseAccount(
	t *testing.T,
	start time.Time,
	ctr *controllertest.ControllerMockWrapper,
) {
	diff := cmp.Diff(ctr.GetDatabaseAccountOriginal(), ctr.GetDatabaseAccount())
	testhelp.Errorf(t, start, "rec.Reconcile(): DatabaseAccount object -original +final:\n%s", diff)
}

func testReconcileResultsTestSet(
	ts testSet,
	expect expectSet,
) {
	ts.t.Helper()

	if ts.reconcileLoop > 1 {
		for recCount := 1; recCount < ts.reconcileLoop; recCount++ {
			_, resErr := ts.rec.Reconcile(ts.ctx, ctrl.Request{
				NamespacedName: ts.ctr.GetDatabaseAccountNamespacedName(),
			})

			if resErr != nil {
				testhelp.Errorf(
					ts.t, ts.start,
					"DatabaseAccountReconciler.Reconcile()[loop:%d]: error, got '%v', want 'nil'",
					recCount, resErr,
				)
			}
		}
	}
	testReconcileResults(
		ts.ctx, ts.t, ts.start,
		expect,
		ts.ctr, ts.c,
		ts.rcd, ts.svr, ts.rec,
		ts.reconcileModDBAccount, ts.reconcileModSecret,
	)
}

func testReconcileResults(
	ctx context.Context,
	t *testing.T, start time.Time,
	expect expectSet,
	ctr *controllertest.ControllerMockWrapper,
	c *v1test.MockClient,
	rcd *v1test.MockRecorder,
	svr *accountsvrtest.MockServer,
	rec *controller.DatabaseAccountReconciler,
	reconcileModDBAccount []controllertest.ReconcileModDBFunc,
	reconcileModSecret []controllertest.ReconcileModSecretFunc,
) {
	t.Helper()

	res, resErr := rec.Reconcile(ctx, ctrl.Request{
		NamespacedName: ctr.GetDatabaseAccountNamespacedName(),
	})

	if resErr != nil {
		testhelp.Errorf(t, start, "DatabaseAccountReconciler.Reconcile(): error, got '%v', want 'nil'", resErr)
	}

	if diff := cmp.Diff(res, expect.expectResult); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): result -got +want:\n%s", diff)
	}

	if diff := cmp.Diff(c.CallCountMap(), expect.expectClientCallMap); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): client called functions -got +want:\n%s", diff)
	}

	if diff := cmp.Diff(svr.CallCountMap(), expect.expectServerCallMap); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): server called functions -got +want:\n%s", diff)
	}

	if diff := cmp.Diff(rcd.CallCountMap(), expect.expectRecorderCallMap); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): recorder called functions -got +want:\n%s", diff)
	}

	if expect.expectTestCallMap != nil {
		if diff := cmp.Diff(ctr.CallCountMap(), expect.expectTestCallMap); diff != "" {
			testhelp.Errorf(t, start, "rec.Reconcile(): test called functions -got +want:\n%s", diff)
		}
	}

	if diff := cmp.Diff(
		ctr.GetDatabaseAccount(), ctr.GetExpectDatabaseAccount(reconcileModDBAccount...),
		controllertest.CompareAccountsIgnore(),
	); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): DatabaseAccount object -got +want:\n%s", diff)
		diffOriginalDatabaseAccount(t, start, ctr)
	}

	if diff := cmp.Diff(
		ctr.GetSecret(), ctr.GetExpectSecret(reconcileModSecret...),
		controllertest.CompareAccountsIgnore(),
	); diff != "" {
		testhelp.Errorf(t, start, "rec.Reconcile(): Secret object -got +want:\n%s", diff)
	}

	if expect.expectNormalMessage != nil {
		if diff := cmp.Diff(
			rcd.GetNormalEvents(), expect.expectNormalMessage,
			controllertest.CompareEventsIgnore(),
		); diff != "" {
			testhelp.Errorf(t, start, "rec.Reconcile(): recorder normal messages -got +want:\n%s", diff)
		}
	}

	if expect.expectWarningMessage != nil {
		if diff := cmp.Diff(
			rcd.GetWarningEvents(), expect.expectWarningMessage,
			controllertest.CompareEventsIgnore(),
		); diff != "" {
			testhelp.Errorf(t, start, "rec.Reconcile(): recorder warning messages -got +want:\n%s", diff)
		}
	}
}
