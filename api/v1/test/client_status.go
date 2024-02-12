package test

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockStatusClient struct {
	calledFunc       map[string]int
	TestStatusWriter *MockSubResourceWriter
	OnStatus         func() client.SubResourceWriter
}

func (m *MockStatusClient) CallCountMap() map[string]int {
	out := map[string]int{}
	for k, v := range m.TestStatusWriter.calledFunc {
		out["TestStatusWriter."+k] = v
	}
	for k, v := range m.calledFunc {
		out[k] = v
	}
	return out
}

func (m *MockStatusClient) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (m *MockStatusClient) CallCountReset() {
	m.calledFunc = map[string]int{}
}

func (m *MockStatusClient) Status() client.SubResourceWriter {
	m.calledFunc["Status"]++
	if m.OnStatus != nil {
		return m.OnStatus()
	}

	return m.TestStatusWriter
}
