package testhelp

import (
	"os"
	"testing"
	"time"
)

type TestLogFuncWrap func(string, ...interface{})

func Errorf(t *testing.T, strt time.Time, format string, args ...interface{}) {
	t.Helper()

	Logf(t, strt, format, args...)
	t.Fail()
}

func Logf(t *testing.T, strt time.Time, format string, args ...interface{}) {
	t.Helper()

	if useTS := os.Getenv("GO_TEST_TIMESTAMP"); useTS != "" {
		t.Logf("%.2fs: "+format, append([]any{time.Since(strt).Seconds()}, args...)...)
		return
	}

	t.Logf(format, args...)
}
