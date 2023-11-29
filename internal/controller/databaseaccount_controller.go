/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dosquad/database-operator/accountsvr"
	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/helper"
	"github.com/go-logr/logr"
	"github.com/oklog/ulid/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DatabaseAccountReconciler reconciles a DatabaseAccount object.
type DatabaseAccountReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Recorder      Recorder
	AccountServer accountsvr.Server
	Config        *dbov1.DatabaseAccountControllerConfig
}

//+kubebuilder:rbac:groups=dbo.dosquad.github.io,resources=databaseaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbo.dosquad.github.io,resources=databaseaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbo.dosquad.github.io,resources=databaseaccounts/finalizers,verbs=update
//+kubebuilder:rbac:groups=dbo.dosquad.github.io,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DatabaseAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	callID := ulid.Make().String()
	logger = logger.WithValues("callID", callID)
	ctx = log.IntoContext(ctx, logger)

	logger.V(1).Info("Reconcile",
		"req.Name", req.Name,
		"req.Namespace", req.Namespace,
	)

	var dbAccount dbov1.DatabaseAccount
	if err := r.Get(ctx, req.NamespacedName, &dbAccount); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to retrieve database account")

		return ctrl.Result{}, err
	} else if apierrors.IsNotFound(err) {
		// record is deleted
		logger.V(1).Info("record is deleted")

		return ctrl.Result{}, nil
	}

	if ok, err := r.handleFinalizers(ctx, &dbAccount); ok {
		return ctrl.Result{}, err
	}

	logger.V(1).Info("entering switch",
		"stage", dbAccount.Status.Stage,
		"resourceVersion", dbAccount.ResourceVersion,
		"generation", dbAccount.Generation,
		// "dump", fmt.Sprintf("%+v", dbAccount),
	)

	if r.Config.Debug.ReconcileSleep > 0 {
		//nolint:gosec,gomnd // simple random for use when debugging.
		n := rand.Intn(r.Config.Debug.ReconcileSleep/2) + r.Config.Debug.ReconcileSleep/2
		logger.V(1).Info(fmt.Sprintf("sleeping %d seconds", n))
		time.Sleep(time.Duration(n) * time.Second)
	}

	switch dbAccount.Status.Stage {
	case dbov1.UnknownStage:
		return r.stageZero(ctx, &dbAccount)
	case dbov1.InitStage:
		return r.stageInit(ctx, &dbAccount)
	case dbov1.UserCreateStage:
		return r.stageUserCreate(ctx, &dbAccount)
	case dbov1.DatabaseCreateStage:
		return r.stageDatabaseCreate(ctx, &dbAccount)
	case dbov1.RelayCreateStage:
		return r.stageRelayCreate(ctx, &dbAccount)
	case dbov1.ReadyStage:
		return r.stageReady(ctx, &dbAccount)
	case dbov1.ErrorStage:
		return r.stageError(ctx, &dbAccount)
	case dbov1.TerminatingStage:
		return r.stageTerminating(ctx, &dbAccount)
	default:
		logger.Error(nil, "Unknown error type", "stage_value", dbAccount.Status.Stage)
	}

	// logger.V(1).Info("Fallback return")
	return ctrl.Result{}, nil
}

// TODO ctrl.Result being the same no matter what, need to look into this during refactor.
func (r *DatabaseAccountReconciler) handleFinalizers(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (bool, error) {
	logger := log.FromContext(ctx)

	if dbAccount.ObjectMeta.DeletionTimestamp.IsZero() {
		// object is not being deleted currently, check for and add finalizers
		// TODO WHAT IS HAPPENING HERE ?
		if controllerutil.AddFinalizer(dbAccount, finalizerName) {
			if err := r.Update(ctx, dbAccount); err != nil {
				return true, err
			}

			logger.V(1).Info("added finalizer to DatabaseAccount")
			return true, nil
		}

		return false, nil
	}

	// object is being deleted
	if controllerutil.ContainsFinalizer(dbAccount, finalizerName) {
		// our finalizer is present, so lets handle any external dependency
		if err := r.deleteExternalResources(ctx, dbAccount); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried
			return true, err
		}

		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(dbAccount, finalizerName)
		if err := r.Update(ctx, dbAccount); err != nil {
			return true, err
		}
	}

	// Stop reconciliation as the item is being deleted
	return true, nil
}

// deleteExternalResources takes the DatabaseAccount and removes any external resources if required.
func (r *DatabaseAccountReconciler) deleteExternalResources(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) error {
	logger := log.FromContext(ctx)

	switch dbAccount.GetSpecOnDelete() {
	case dbov1.OnDeleteDelete:
		name, err := dbAccount.GetDatabaseName()
		if err != nil {
			return err
		}

		return r.deleteDatabase(ctx, name)
	case dbov1.OnDeleteRetain:
		logger.Info("Database record marked for retention, skipping delete")
	}
	return nil
}

// deleteDatabase takes the database name and removes it from the database server.
func (r *DatabaseAccountReconciler) deleteDatabase(ctx context.Context, name string) error {
	logger := log.FromContext(ctx)

	logger.Info("Database record marked for delete, deleting", "databaseName", name)

	dbName, ok, err := r.AccountServer.IsDatabase(ctx, name)
	switch {
	case err == nil && ok:
		logger.V(1).Info("Database exists, deleting")

		if err = r.AccountServer.Delete(ctx, name); err != nil {
			logger.Error(err, "Unable to delete database and/or user")

			return err
		}

		logger.Info("Database account deleted", "databaseName", name)
	case err != nil:
		logger.V(1).Error(err, "Unable to check if database exists")

		return err
	default:
		logger.V(1).Info("IsDatabase returned", "isDB[dbName]", dbName, "isDB[ok]", ok, "isDB[err]", err)
	}

	return nil
}

func (r *DatabaseAccountReconciler) stageZero(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// if !controllerutil.ContainsFinalizer(dbAccount, finalizerName) {
	// 	logger.V(1).Info("adding finalizer to DatabaseAccount")
	// 	if controllerutil.AddFinalizer(dbAccount, finalizerName) {
	// 		logger.V(1).Info("added finalizer to DatabaseAccount")

	// 		if err := dbAccount.Update(ctx, r); err != nil {
	// 			logger.V(1).Error(err, "unable to update DatabaseAccount")

	// 			return ctrl.Result{}, err
	// 		}

	// 		logger.V(1).Info("finalizer added, requeuing.")
	// 		return reconcile.Result{Requeue: false}, nil
	// 	}
	// }

	dbAccount.Status.Stage = dbov1.InitStage
	if len(dbAccount.Status.Name) == 0 {
		dbAccount.Status.Name = NewDatabaseAccountName(ctx)
	}

	if err := dbAccount.UpdateStatus(ctx, r); err != nil {
		logger.V(1).Error(err, "unable to update DatabaseAccount Status")

		return ctrl.Result{}, err
	}

	logger.V(1).Info("return result[ok]", "dbAccount", dbAccount)
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageInit(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	secretErr := SecretRun(ctx, r, r, r.AccountServer, dbAccount, func(secret *corev1.Secret) error {
		if secret.Immutable != nil && *secret.Immutable {
			return ErrSecretImmutable
		}

		name, err := dbAccount.GetDatabaseName()
		if err != nil {
			return err
		}

		r.AccountServer.CopyInitConfigToSecret(dbAccount, secret)
		SetSecretKV(secret, accountsvr.DatabaseKeyUsername, name)
		if dbAccount.GetSpecOnDelete() != dbov1.OnDeleteRetain {
			SecretAddOwnerRefs(secret, dbAccount)
		}
		controllerutil.AddFinalizer(secret, finalizerName)

		return nil
	})

	switch {
	case secretErr != nil && errors.Is(secretErr, ErrSecretImmutable):
		r.Recorder.WarningEvent(dbAccount, ReasonQueued, "Secret already exists and is immutable")
		dbAccount.Status.Error = true
		dbAccount.Status.ErrorMessage = "Secret already exists and is immutable"
		dbAccount.Status.Stage = dbov1.ErrorStage

		if secretErr = dbAccount.UpdateStatus(ctx, r); secretErr != nil {
			logger.V(1).Error(secretErr, "Unable to update DatabaseAccount status")

			return ctrl.Result{}, secretErr
		}

		return ctrl.Result{}, nil
	case secretErr != nil:
		logger.V(1).Error(secretErr, "Unable to create/retrieve secret")

		return ctrl.Result{}, secretErr
	default:
		r.Recorder.NormalEvent(dbAccount, ReasonQueued, "Queued for creation")
	}

	dbAccount.Status.Stage = dbov1.UserCreateStage

	if err := r.Status().Update(ctx, dbAccount); err != nil {
		logger.V(1).Error(err, "Unable to update DatabaseAccount status")

		return ctrl.Result{}, err
	}

	// logger.V(1).Info("return result[ok]")
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageUserCreate(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	r.Recorder.NormalEvent(dbAccount, ReasonUserCreate, "Creating database user")

	if err := SecretRun(ctx, r, r, r.AccountServer, dbAccount, func(secret *corev1.Secret) error {
		name, err := dbAccount.GetDatabaseName()
		if err != nil {
			return err
		}

		usr, pw, err := r.AccountServer.CreateRole(ctx, name)
		if errors.Is(err, accountsvr.ErrRoleExists) {
			r.Recorder.WarningEvent(dbAccount, ReasonUserCreate, "User exists, creating new password")
			usr, pw, err = r.AccountServer.UpdateRolePassword(ctx, name)
		}
		if err != nil {
			r.Recorder.WarningEvent(dbAccount, ReasonUserCreate, fmt.Sprintf("Failed to create user: %s", err))

			return err
		}

		r.Recorder.NormalEvent(dbAccount, ReasonUserCreate, "User created")

		SetSecretKV(secret, accountsvr.DatabaseKeyUsername, usr)
		SetSecretKV(secret, accountsvr.DatabaseKeyPassword, pw)

		return nil
	}); err != nil {
		logger.V(1).Error(err, "Unable to create/retrieve secret")

		return ctrl.Result{}, err
	}

	dbAccount.Status.Stage = dbov1.DatabaseCreateStage

	if err := r.Status().Update(ctx, dbAccount); err != nil {
		logger.V(1).Error(err, "Unable to update DatabaseAccount")

		return ctrl.Result{}, err
	}
	r.Recorder.NormalEvent(dbAccount, ReasonDatabaseCreate, "Creating database")

	// logger.V(1).Info("return result[ok]")
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageDatabaseCreate(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	name, nameErr := dbAccount.GetDatabaseName()
	if nameErr != nil {
		return ctrl.Result{}, nameErr
	}

	dbName, ok, dbErr := r.AccountServer.IsDatabase(ctx, name)
	switch {
	case dbErr != nil:
		r.Recorder.WarningEvent(
			dbAccount,
			ReasonDatabaseCreate,
			fmt.Sprintf("Failed to init check for create database: %s", dbErr),
		)

		return ctrl.Result{}, dbErr
	case ok:
		r.Recorder.WarningEvent(dbAccount, ReasonDatabaseCreate, "Database already exists")
		if err := SecretRun(ctx, r, r, r.AccountServer, dbAccount, func(secret *corev1.Secret) error {
			SetSecretKV(secret, accountsvr.DatabaseKeyDatabase, dbName)
			SetSecretKV(secret, accountsvr.DatabaseKeyDSN, accountsvr.GenerateDSN(secret))
			if dbAccount.GetSpecCreateRelay() {
				AddPGBouncerConf(r.AccountServer, secret)
				SetSecretKV(secret, accountsvr.DatabaseKeyHost, dbAccount.GetSecretName().Name)
			}

			boolTrue := true
			secret.Immutable = &boolTrue
			return nil
		}); err != nil {
			logger.V(1).Error(err, "Unable to update secret")

			return ctrl.Result{}, err
		}
	default:
		dbName, dbErr = r.AccountServer.CreateDatabase(ctx, name, name)
		if dbErr != nil {
			r.Recorder.WarningEvent(dbAccount, ReasonDatabaseCreate, fmt.Sprintf("Failed to create database: %s", dbErr))
			return ctrl.Result{}, dbErr
		}

		if secretErr := SecretRun(ctx, r, r, r.AccountServer, dbAccount, func(secret *corev1.Secret) error {
			SetSecretKV(secret, accountsvr.DatabaseKeyDatabase, dbName)
			SetSecretKV(secret, accountsvr.DatabaseKeyDSN, accountsvr.GenerateDSN(secret))
			if dbAccount.GetSpecCreateRelay() {
				AddPGBouncerConf(r.AccountServer, secret)
				SetSecretKV(secret, accountsvr.DatabaseKeyHost, dbAccount.GetSecretName().Name)
			}

			boolTrue := true
			secret.Immutable = &boolTrue
			return nil
		}); secretErr != nil {
			logger.V(1).Error(secretErr, "Unable to update secret")

			return ctrl.Result{}, secretErr
		}

		r.Recorder.NormalEvent(dbAccount, ReasonDatabaseCreate, "Database created")
	}

	if !dbAccount.GetSpecCreateRelay() {
		dbAccount.Status.Stage = dbov1.ReadyStage
		dbAccount.Status.Ready = true
	} else {
		dbAccount.Status.Stage = dbov1.RelayCreateStage
	}

	if err := r.Status().Update(ctx, dbAccount); err != nil {
		logger.V(1).Error(err, "Unable to update DatabaseAccount")

		return ctrl.Result{}, err
	}

	if !dbAccount.GetSpecCreateRelay() {
		r.Recorder.NormalEvent(dbAccount, ReasonReady, "Ready to use")
	} else {
		r.Recorder.NormalEvent(dbAccount, ReasonRelayCreate, "Creating Relay Pod")
	}
	// logger.Info("Record is marked as ready",
	// 	"databaseUsername", dbAccount.Status.Name,
	// 	"secretName", dbAccount.GetSecretName(),
	// )

	// logger.V(1).Info("return result[ok]")
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageRelayCreate(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// name, nameErr := dbAccount.GetDatabaseName()
	// if nameErr != nil {
	// 	return ctrl.Result{}, nameErr
	// }

	statefulSet, statefulSetErr := StatefulSetGet(ctx, r, r.Config, dbAccount)
	if errors.Is(statefulSetErr, ErrNewStatefulSet) {
		if err := r.Create(ctx, statefulSet); err != nil {
			logger.V(1).Error(err, "unable to create StatefulSet")

			return ctrl.Result{}, err
		}
	} else if statefulSetErr != nil {
		return ctrl.Result{}, statefulSetErr
	}

	r.Recorder.NormalEvent(dbAccount, ReasonRelayCreate, "Created StatefulSet")

	service, serviceErr := ServiceGet(ctx, r, dbAccount)
	if errors.Is(serviceErr, ErrNewService) {
		if err := r.Create(ctx, service); err != nil {
			logger.V(1).Error(err, "unable to create Service")

			return ctrl.Result{}, err
		}
	} else if serviceErr != nil {
		return ctrl.Result{}, serviceErr
	}

	r.Recorder.NormalEvent(dbAccount, ReasonRelayCreate, "Created Service")

	dbAccount.Status.Stage = dbov1.ReadyStage
	dbAccount.Status.Ready = true
	if err := r.Status().Update(ctx, dbAccount); err != nil {
		logger.V(1).Error(err, "Unable to update DatabaseAccount")

		return ctrl.Result{}, err
	}

	r.Recorder.NormalEvent(dbAccount, ReasonReady, "Ready to use")
	// logger.Info("Record is marked as ready",
	// 	"databaseUsername", dbAccount.Status.Name,
	// 	"secretName", dbAccount.GetSecretName(),
	// )

	// // logger.V(1).Info("return result[ok]")
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageReady(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.V(1).Info("Record is marked as ready, nothing to do")

	if _, secretErr := SecretGetByName(ctx, r, dbAccount.GetSecretName()); apierrors.IsNotFound(secretErr) {
		// secret has been deleted, probably sent here from reconcile trigger in secret delete.
		logger.Info("Secret has been deleted, remove DatabaseAccount")

		if err := r.Delete(ctx, dbAccount); err != nil {
			logger.Error(err, "Unable to delete DatabaseAccount in reaction to secret deletion")

			return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
		}

		return ctrl.Result{}, nil
	}

	onDeleteUpdate := false
	if err := SecretRun(ctx, r, r, r.AccountServer, dbAccount, func(secret *corev1.Secret) error {
		// logger.V(1).Info("Checking database account spec")
		// logger.V(1).Info("Database account spec", "dbAccount.Spec.OnDelete", dbAccount.GetSpecOnDelete())
		switch dbAccount.GetSpecOnDelete() {
		case dbov1.OnDeleteDelete:
			if len(secret.ObjectMeta.OwnerReferences) != 1 {
				// logger.V(1).Info("Database account marked for deletion but secret does not have ownerreferences")
				SecretAddOwnerRefs(secret, dbAccount)
				onDeleteUpdate = true
			}
		case dbov1.OnDeleteRetain:
			if len(secret.ObjectMeta.OwnerReferences) != 0 {
				// logger.V(1).Info("Database account marked for retention but secret has ownerreferences")
				secret.OwnerReferences = nil
				onDeleteUpdate = true
			}
		}

		return nil
	}); err != nil {
		logger.V(1).Error(err, "Unable to update secret")

		return ctrl.Result{}, err
	}

	if onDeleteUpdate {
		logger.Info("Database account onDelete changed, updated secret")
	}

	logger.Info("Record is marked as ready",
		"databaseUsername", dbAccount.Status.Name,
		"secretName", dbAccount.GetSecretName(),
	)

	// logger.V(1).Info("return result[ok]")
	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageError(
	ctx context.Context,
	_ *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Record is marked as error, nothing to do")

	return ctrl.Result{}, nil
}

func (r *DatabaseAccountReconciler) stageTerminating(
	ctx context.Context,
	dbAccount *dbov1.DatabaseAccount,
) (ctrl.Result, error) {
	dbName := dbAccount.Status.Name
	if err := r.deleteDatabase(ctx, dbName.String()); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileSecret handles secrets created along side databaseaccount resources.
func (r *DatabaseAccountReconciler) reconcileSecret(ctx context.Context, secretObj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)
	callID := ulid.Make().String()
	logger = logger.WithValues("callID", callID)
	ctx = log.IntoContext(ctx, logger)
	requests := []reconcile.Request{}

	name := types.NamespacedName{
		Name:      secretObj.GetName(),
		Namespace: secretObj.GetNamespace(),
	}
	secret, secretErr := SecretGetByName(ctx, r, name)
	if secretErr != nil {
		return requests
	}

	if !strings.EqualFold(string(secret.Type), secretType) {
		return requests
	}

	logger.V(1).Info("call():reconcileSecret",
		"name", secretObj.GetName(),
		"namespace", secretObj.GetNamespace(),
		"managedFields", secretObj.GetManagedFields(),
		"secretType", secret.Type,
	)

	if secret.ObjectMeta.DeletionTimestamp.IsZero() {
		// object is not being deleted, return early
		return requests
	}

	// object is being deleted
	if controllerutil.ContainsFinalizer(secret, finalizerName) {
		// our finalizer is present, so lets handle any external dependency
		requests = append(requests,
			r.reconcileSecretExternalDependency(ctx, logger, secret)...,
		)
	}

	ownerRefs := secret.GetOwnerReferences()
	for _, ownerRef := range ownerRefs {
		if strings.EqualFold(ownerRef.Kind, dbov1.KindDatabaseAccount) {
			if err := r.setStageOnDatabaseAccount(
				ctx,
				ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      ownerRef.Name,
						Namespace: secret.GetNamespace(),
					},
				},
				dbov1.TerminatingStage,
			); err != nil {
				logger.V(1).Error(err, "Unable to update DatabaseAccount status")
			}
		}
	}

	// Stop reconciliation as the item is being deleted
	return requests
}

func (r *DatabaseAccountReconciler) setStageOnDatabaseAccount(
	ctx context.Context,
	req ctrl.Request,
	stage dbov1.DatabaseAccountCreateStage,
) error {
	logger := log.FromContext(ctx)

	var dbAccount dbov1.DatabaseAccount
	if err := r.Get(ctx, req.NamespacedName, &dbAccount); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to retrieve database account")

		return err
	} else if apierrors.IsNotFound(err) {
		// record is deleted
		logger.V(1).Info("record is deleted")

		return nil
	}

	return dbAccount.SetStage(ctx, r, stage)
}

func (r *DatabaseAccountReconciler) reconcileSecretExternalDependency(
	ctx context.Context,
	logger logr.Logger,
	secret *corev1.Secret,
) []reconcile.Request {
	requests := []reconcile.Request{}

	// our finalizer is present, so lets handle any external dependency
	if dbName, ok := secret.Data[accountsvr.DatabaseKeyDatabase]; ok {
		logger.V(1).Info("returning request for owner reference to be reconciled", "databaseName", dbName)

		if len(secret.GetOwnerReferences()) > 0 {
			ownerRef := secret.GetOwnerReferences()[0]
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secret.GetNamespace(),
					Name:      ownerRef.Name,
				},
			})
		}

		// if err := r.deleteDatabase(ctx, string(dbName)); err != nil {
		// 	return requests
		// }
	}

	logger.V(1).Info("removing finalizer from DatabaseAccount")
	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(secret, finalizerName)
	if err := r.Update(ctx, secret); err != nil {
		return requests
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &dbov1.DatabaseAccount{})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("database_operator").
		For(&dbov1.DatabaseAccount{}).
		Owns(&corev1.Secret{}).
		WithLogConstructor(logCtr).
		Watches(
			&corev1.Secret{},
			helper.EnqueueRequestsFromMapFunc(r.reconcileSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
