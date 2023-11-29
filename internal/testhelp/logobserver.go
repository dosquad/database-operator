package testhelp

import "github.com/go-logr/logr"

type ObservableLogs struct {
	Output map[string]map[int][]interface{}
	r      logr.RuntimeInfo
}

func NewObservableLogs() *ObservableLogs {
	return &ObservableLogs{
		make(map[string]map[int][]interface{}),
		logr.RuntimeInfo{CallDepth: 1},
	}
}

func (t *ObservableLogs) doLog(level int, msg string, keysAndValues ...interface{}) {
	m := make(map[int][]interface{}, len(keysAndValues))
	m[level] = keysAndValues
	t.Output[msg] = m
}
func (t *ObservableLogs) Init(info logr.RuntimeInfo) { t.r = info }
func (t *ObservableLogs) Enabled(_ int) bool         { return true }
func (t *ObservableLogs) Info(level int, msg string, keysAndValues ...interface{}) {
	t.doLog(level, msg, keysAndValues...)
}
func (t *ObservableLogs) Error(err error, msg string, keysAndValues ...interface{}) {
	t.doLog(0, msg, append(keysAndValues, err)...)
}
func (t *ObservableLogs) WithValues(_ ...interface{}) logr.LogSink { return t }
func (t *ObservableLogs) WithName(_ string) logr.LogSink           { return t }
