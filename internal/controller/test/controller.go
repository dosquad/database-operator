package test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	v1 "github.com/dosquad/database-operator/api/v1"
	v1test "github.com/dosquad/database-operator/api/v1/test"
	"github.com/dosquad/database-operator/internal/testhelp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControllerMockWrapper struct {
	t                            *testing.T
	start                        time.Time
	calledFunc                   map[string]int
	Secret, OriginalSecret       *corev1.Secret
	DBAccount, OriginalDBAccount *v1.DatabaseAccount
	Client                       *v1test.MockClient
}

func NewControllerMockWrapper(t *testing.T, start time.Time, client *v1test.MockClient) *ControllerMockWrapper {
	c := &ControllerMockWrapper{
		t:      t,
		start:  start,
		Client: client,
	}

	c.CallCountReset()
	c.Init()

	return c
}

func (c *ControllerMockWrapper) CallCountMap() map[string]int {
	return c.calledFunc
}

func (c *ControllerMockWrapper) CallCountReset() {
	c.calledFunc = map[string]int{}
}

func (c *ControllerMockWrapper) CallCount(name string) (int, bool) {
	if v, ok := c.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (c *ControllerMockWrapper) IncCallCount(name string) {
	c.calledFunc[name]++
}

func (c *ControllerMockWrapper) initMockClientReaderOnGet() {
	c.Client.MockClientReader.OnGet = func(
		_ context.Context, key client.ObjectKey,
		obj client.Object, opts ...client.GetOption,
	) error {
		testhelp.Logf(c.t, c.start, "MockClientReader.Get(ctx, '%+v', '%+v', '%+v')", key, obj, opts)
		c.IncCallCount(
			fmt.Sprintf("MockClientReader.Get(%s)", reflect.TypeOf(obj).String()),
		)
		if key.String() == fmt.Sprintf("%s/%s", c.DBAccount.GetNamespace(), c.DBAccount.GetName()) {
			switch v := obj.(type) {
			case *v1.DatabaseAccount:
				testhelp.Logf(c.t, c.start, "Object is v1.DatabaseAccount: %+v", obj)
				if c.DBAccount != nil {
					*v = *c.DBAccount
					return nil
				}
			case *corev1.Secret:
				testhelp.Logf(c.t, c.start, "Object is corev1.Secret: %+v", obj)
				if c.Secret != nil {
					*v = *c.Secret
					return nil
				}
			}
		}

		return apierrors.NewNotFound(
			schema.GroupResource{Resource: "databaseaccounts"},
			v1test.DBAccount,
		)
	}
}

func (c *ControllerMockWrapper) initMockClientWriterOnCreate() {
	c.Client.MockClientWriter.OnCreate = func(_ context.Context, obj client.Object, opts ...client.CreateOption) error {
		testhelp.Logf(c.t, c.start, "MockClientWriter.Create(ctx, '%+v', '%+v')", obj, opts)
		c.IncCallCount(
			fmt.Sprintf("MockClientWriter.Create(%s)", reflect.TypeOf(obj).String()),
		)
		switch v := obj.(type) {
		case *v1.DatabaseAccount:
			testhelp.Logf(c.t, c.start, "Object is v1.DatabaseAccount: %+v", obj)
			// c.IncCallCount("MockClientWriter.Create(*v1.DatabaseAccount)")
			c.DBAccount = v
			return nil
		case *corev1.Secret:
			testhelp.Logf(c.t, c.start, "Object is corev1.Secret: %+v", obj)
			// c.IncCallCount("MockClientWriter.Create(*corev1.Secret)")
			c.Secret = v
			return nil
		case *appsv1.StatefulSet:
			// c.IncCallCount("MockClientWriter.Create(*appsv1.StatefulSet)")
			return nil
		case *corev1.Service:
			// c.IncCallCount("MockClientWriter.Create(*corev1.Service)")
			return nil
		}
		return nil
	}
}

func (c *ControllerMockWrapper) initMockClientWriterOnUpdate() {
	c.Client.MockClientWriter.OnUpdate = func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
		testhelp.Logf(c.t, c.start, "MockClientWriter.Update(ctx, obj): %+v", obj)
		c.IncCallCount(
			fmt.Sprintf("MockClientWriter.Update(%s)", reflect.TypeOf(obj).String()),
		)
		switch v := obj.(type) {
		case *v1.DatabaseAccount:
			if v != nil {
				c.DBAccount = v
			}
		case *corev1.Secret:
			if v != nil {
				c.Secret = v
			}
		default:
			testhelp.Errorf(c.t, c.start, "c.MockClientWriter.Update: unknown object type: %+v", v)
		}

		return nil
	}
}

func (c *ControllerMockWrapper) initTestStatusWriterOnUpdate() {
	c.Client.TestStatusWriter.OnUpdate = func(
		_ context.Context, o client.Object,
		_ ...client.SubResourceUpdateOption,
	) error {
		testhelp.Logf(c.t, c.start, "TestStatusWriter.OnUpdate(ctx, obj): %+v", o)
		switch v := o.(type) {
		case *v1.DatabaseAccount:
			if v != nil {
				c.DBAccount = v
			}
		case *corev1.Secret:
			if v != nil {
				c.Secret = v
			}
		default:
			testhelp.Errorf(c.t, c.start, "c.TestStatusWriter.OnUpdate: unknown object type: %+v", v)
		}

		return nil
	}
}

func (c *ControllerMockWrapper) Init() {
	c.initMockClientReaderOnGet()
	c.initMockClientWriterOnCreate()
	c.initMockClientWriterOnUpdate()
	c.initTestStatusWriterOnUpdate()
}

func (c *ControllerMockWrapper) InitDatabaseAccount(f func(dba *v1.DatabaseAccount)) {
	dba := v1test.NewInitDatabaseAccount()
	c.DBAccount = &dba
	c.OriginalDBAccount = &v1.DatabaseAccount{}
	if f != nil {
		f(c.DBAccount)
	}
	c.DBAccount.DeepCopyInto(c.OriginalDBAccount)
}

func (c *ControllerMockWrapper) InitSecrets(f func(secret *corev1.Secret)) {
	secret := v1test.NewSecret()
	c.Secret = &secret
	c.OriginalSecret = &corev1.Secret{}
	if f != nil {
		f(c.Secret)
	}
	c.Secret.DeepCopyInto(c.OriginalSecret)
}

func (c *ControllerMockWrapper) GetDatabaseAccount() v1.DatabaseAccount {
	if c.DBAccount != nil {
		return *c.DBAccount
	}

	return v1.DatabaseAccount{}
}

func (c *ControllerMockWrapper) GetDatabaseAccountOriginal() v1.DatabaseAccount {
	if c.OriginalDBAccount != nil {
		return *c.OriginalDBAccount
	}

	return v1.DatabaseAccount{}
}

func (c *ControllerMockWrapper) GetDatabaseAccountNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: c.DBAccount.GetNamespace(),
		Name:      c.DBAccount.GetName(),
	}
}

func (c *ControllerMockWrapper) GetExpectDatabaseAccount(ff ...ReconcileModDBFunc) v1.DatabaseAccount {
	out := &v1.DatabaseAccount{}
	c.OriginalDBAccount.DeepCopyInto(out)
	for _, f := range ff {
		f(out)
	}

	return *out
}

func (c *ControllerMockWrapper) GetSecret() corev1.Secret {
	if c.Secret != nil {
		return *c.Secret
	}

	return corev1.Secret{}
}

func (c *ControllerMockWrapper) GetExpectSecret(ff ...ReconcileModSecretFunc) corev1.Secret {
	out := &corev1.Secret{}
	if c.OriginalSecret == nil && c.Secret != nil {
		c.Secret.DeepCopyInto(out)
	}
	if c.OriginalSecret != nil {
		c.OriginalSecret.DeepCopyInto(out)
	}
	for _, f := range ff {
		f(out)
	}

	return *out
}
