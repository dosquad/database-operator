package test

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClient struct {
	MockClientReader
	MockClientWriter
	MockStatusClient
	MockSubResourceClientConstructor

	calledFunc            map[string]int
	OnScheme              func() *runtime.Scheme
	OnRESTMapper          func() meta.RESTMapper
	OnGroupVersionKindFor func(obj runtime.Object) (schema.GroupVersionKind, error)
	OnIsObjectNamespaced  func(obj runtime.Object) (bool, error)
}

func NewMockClient() *MockClient {
	m := &MockClient{
		MockStatusClient: MockStatusClient{
			TestStatusWriter: NewMockSubResourceWriter(),
		},
	}
	m.CallCountReset()

	return m
}

func (m *MockClient) CallCountMap() map[string]int {
	out := map[string]int{}
	for k, v := range m.MockClientReader.calledFunc {
		out[fmt.Sprintf("MockClientReader.%s", k)] = v
	}
	for k, v := range m.MockClientWriter.calledFunc {
		out[fmt.Sprintf("MockClientWriter.%s", k)] = v
	}
	for k, v := range m.MockStatusClient.CallCountMap() {
		out[fmt.Sprintf("MockStatusClient.%s", k)] = v
	}
	for k, v := range m.MockSubResourceClientConstructor.calledFunc {
		out[fmt.Sprintf("MockSubResourceClientConstructor.%s", k)] = v
	}
	for k, v := range m.calledFunc {
		out[k] = v
	}
	return out
}

func (m *MockClient) CallCountReset() {
	m.calledFunc = map[string]int{}
	m.MockClientReader.calledFunc = map[string]int{}
	m.MockClientWriter.calledFunc = map[string]int{}
	// m.MockStatusClient.calledFunc = map[string]int{}
	m.MockStatusClient.CallCountReset()
	m.MockSubResourceClientConstructor.calledFunc = map[string]int{}
}

func (m *MockClient) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

// Scheme returns the scheme this client is using.
func (m *MockClient) Scheme() *runtime.Scheme {
	m.calledFunc["Scheme"]++
	if m.OnScheme != nil {
		return m.OnScheme()
	}

	return nil
}

// RESTMapper returns the rest this client is using.
func (m *MockClient) RESTMapper() meta.RESTMapper {
	m.calledFunc["RESTMapper"]++
	if m.OnRESTMapper != nil {
		return m.OnRESTMapper()
	}

	return nil
}

// GroupVersionKindFor returns the GroupVersionKind for the given object.
func (m *MockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	m.calledFunc["GroupVersionKindFor"]++
	if m.OnGroupVersionKindFor != nil {
		return m.OnGroupVersionKindFor(obj)
	}

	return schema.GroupVersionKind{}, nil
}

// IsObjectNamespaced returns true if the GroupVersionKind of the object is namespaced.
func (m *MockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	m.calledFunc["IsObjectNamespaced"]++
	if m.OnIsObjectNamespaced != nil {
		return m.OnIsObjectNamespaced(obj)
	}

	return true, nil
}

type MockSubResourceWriter struct {
	calledFunc map[string]int
	OnCreate   func(context.Context, client.Object, client.Object, ...client.SubResourceCreateOption) error
	OnUpdate   func(context.Context, client.Object, ...client.SubResourceUpdateOption) error
	OnPatch    func(context.Context, client.Object, client.Patch, ...client.SubResourcePatchOption) error
}

func NewMockSubResourceWriter() *MockSubResourceWriter {
	m := &MockSubResourceWriter{}
	m.CallCountReset()
	return m
}

func (m *MockSubResourceWriter) CallCountMap() map[string]int {
	return m.calledFunc
}

func (m *MockSubResourceWriter) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (m *MockSubResourceWriter) CallCountReset() {
	m.calledFunc = map[string]int{}
}

// Create saves the subResource object in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (m *MockSubResourceWriter) Create(
	ctx context.Context, obj client.Object,
	subResource client.Object, opts ...client.SubResourceCreateOption,
) error {
	m.calledFunc["Create"]++
	if m.OnCreate != nil {
		return m.OnCreate(ctx, obj, subResource, opts...)
	}

	return nil
}

// Update updates the fields corresponding to the status subresource for the
// given obj. obj must be a struct pointer so that obj can be updated
// with the content returned by the Server.
func (m *MockSubResourceWriter) Update(
	ctx context.Context, obj client.Object,
	opts ...client.SubResourceUpdateOption,
) error {
	m.calledFunc["Update"]++
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, obj, opts...)
	}

	return nil
}

// Patch patches the given object's subresource. obj must be a struct
// pointer so that obj can be updated with the content returned by the
// Server.
func (m *MockSubResourceWriter) Patch(
	ctx context.Context, obj client.Object,
	patch client.Patch, opts ...client.SubResourcePatchOption,
) error {
	m.calledFunc["Patch"]++
	if m.OnPatch != nil {
		return m.OnPatch(ctx, obj, patch, opts...)
	}

	return nil
}
