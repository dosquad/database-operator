package test

import (
	v1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/controller"
)

type MockRecorderMessage struct {
	Reason  controller.RecorderReason
	Message string
}

func NewMockRecorderMessage(reason controller.RecorderReason, message string) MockRecorderMessage {
	return MockRecorderMessage{
		Reason:  reason,
		Message: message,
	}
}

func (m *MockRecorderMessage) GetReason() controller.RecorderReason {
	return m.Reason
}

func (m *MockRecorderMessage) GetMessage() string {
	return m.Message
}

type MockRecorder struct {
	calledFunc     map[string]int
	normalEvents   []MockRecorderMessage
	warningEvents  []MockRecorderMessage
	OnNormalEvent  func(*v1.DatabaseAccount, controller.RecorderReason, string)
	OnWarningEvent func(*v1.DatabaseAccount, controller.RecorderReason, string)
}

func NewRecorder() *MockRecorder {
	m := &MockRecorder{}
	m.CallCountReset()
	m.EventReset()

	return m
}

func (m *MockRecorder) CallCountMap() map[string]int {
	return m.calledFunc
}

func (m *MockRecorder) CallCountReset() {
	m.calledFunc = map[string]int{}
}

func (m *MockRecorder) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (m *MockRecorder) EventReset() {
	m.normalEvents = []MockRecorderMessage{}
	m.warningEvents = []MockRecorderMessage{}
}

func (m *MockRecorder) GetNormalEvents() []MockRecorderMessage {
	return m.normalEvents
}

func (m *MockRecorder) GetWarningEvents() []MockRecorderMessage {
	return m.warningEvents
}

func (m *MockRecorder) NormalEvent(dbAccount *v1.DatabaseAccount, reason controller.RecorderReason, message string) {
	m.calledFunc["NormalEvent"]++
	if m.OnNormalEvent != nil {
		m.OnNormalEvent(dbAccount, reason, message)
		return
	}

	m.normalEvents = append(m.normalEvents, NewMockRecorderMessage(reason, message))
}

func (m *MockRecorder) WarningEvent(dbAccount *v1.DatabaseAccount, reason controller.RecorderReason, message string) {
	m.calledFunc["WarningEvent"]++
	if m.OnWarningEvent != nil {
		m.OnWarningEvent(dbAccount, reason, message)
		return
	}

	m.warningEvents = append(m.warningEvents, NewMockRecorderMessage(reason, message))
}
