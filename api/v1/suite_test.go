package v1_test

// import (
// 	"os"
// 	"testing"
// 	"time"
// )

// type testLogFuncWrap func(string, ...interface{})

// func tErrorf(t *testing.T, strt time.Time, format string, args ...interface{}) {
// 	t.Helper()

// 	tLogf(t, strt, format, args...)
// 	t.Fail()
// }

// func tLogf(t *testing.T, strt time.Time, format string, args ...interface{}) {
// 	t.Helper()

// 	if useTS := os.Getenv("GO_TEST_TIMESTAMP"); useTS != "" {
// 		t.Logf("%.2fs: "+format, append([]any{time.Since(strt).Seconds()}, args...)...)
// 		return
// 	}

// 	t.Logf(format, args...)
// }
