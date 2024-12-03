package testhelp

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
)

type TestingLogWrapper struct {
	t     *testing.T
	start time.Time
	r     logr.RuntimeInfo
}

func NewTestingLogWrapper(t *testing.T, start time.Time) *TestingLogWrapper {
	t.Helper()

	return &TestingLogWrapper{
		t,
		start,
		logr.RuntimeInfo{CallDepth: 1},
	}
}

func (t *TestingLogWrapper) Init(info logr.RuntimeInfo) { t.r = info }
func (t *TestingLogWrapper) Enabled(_ int) bool         { return true }
func (t *TestingLogWrapper) Info(_ int, msg string, keysAndValues ...interface{}) {
	t.t.Helper()
	Logf(t.t, t.start, "%s [%+v]", msg, keysAndValues)
}
func (t *TestingLogWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	t.t.Helper()
	Errorf(t.t, t.start, "%s [%+v]", msg, append([]interface{}{"error", err}, keysAndValues...))
}
func (t *TestingLogWrapper) WithValues(_ ...interface{}) logr.LogSink { return t }
func (t *TestingLogWrapper) WithName(_ string) logr.LogSink           { return t }
