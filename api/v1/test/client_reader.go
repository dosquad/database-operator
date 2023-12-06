package test

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClientReader struct {
	calledFunc map[string]int
	OnGet      func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	OnList     func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (m *MockClientReader) Get(
	ctx context.Context, key client.ObjectKey,
	obj client.Object, opts ...client.GetOption,
) error {
	m.calledFunc["Get"]++
	if m.OnGet != nil {
		return m.OnGet(ctx, key, obj, opts...)
	}

	return nil
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (m *MockClientReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	m.calledFunc["List"]++
	if m.OnList != nil {
		return m.OnList(ctx, list, opts...)
	}

	return nil
}
