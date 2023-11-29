package test

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClientWriter struct {
	calledFunc    map[string]int
	OnCreate      func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	OnDelete      func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	OnUpdate      func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	OnPatch       func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error
	OnDeleteAllOf func(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error
}

// Create saves the object obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (m *MockClientWriter) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.calledFunc["Create"]++
	if m.OnCreate != nil {
		return m.OnCreate(ctx, obj, opts...)
	}

	return nil
}

// Delete deletes the given obj from Kubernetes cluster.
func (m *MockClientWriter) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.calledFunc["Delete"]++
	if m.OnDelete != nil {
		return m.OnDelete(ctx, obj, opts...)
	}

	return nil
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (m *MockClientWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.calledFunc["Update"]++
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, obj, opts...)
	}

	return nil
}

// Patch client.patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (m *MockClientWriter) Patch(
	ctx context.Context, obj client.Object,
	patch client.Patch, opts ...client.PatchOption,
) error {
	m.calledFunc["Patch"]++
	if m.OnPatch != nil {
		return m.OnPatch(ctx, obj, patch, opts...)
	}

	return nil
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (m *MockClientWriter) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	m.calledFunc["DeleteAllOf"]++
	if m.OnDeleteAllOf != nil {
		return m.OnDeleteAllOf(ctx, obj, opts...)
	}

	return nil
}
