package accountsvr_test

// import (
// 	"bytes"
// 	"context"
// 	"testing"

// 	v1 "github.com/dosquad/database-operator/api/v1"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgconn"
// )

// type MockDB struct {
// 	dsn        v1.PostgreSQLDSN
// 	logr       testLogFuncWrap
// 	calledFunc map[string]int
// 	onConfig   func() *pgx.ConnConfig
// 	onClose    func(context.Context) error
// 	onExec     func(context.Context, string, ...any) (pgconn.CommandTag, error)
// 	onQuery    func(context.Context, string, ...any) (pgx.Rows, error)
// 	onIsClosed func() bool
// }

// func NewMockDB(_ *testing.T, logr testLogFuncWrap, dsn v1.PostgreSQLDSN) *MockDB {
// 	m := &MockDB{
// 		dsn:  dsn,
// 		logr: logr,
// 	}

// 	m.CallCountReset()

// 	return m
// }

// func (m *MockDB) SetLogger(logr testLogFuncWrap) {
// 	m.logr = logr
// }

// func (m *MockDB) CallCount(name string) (int, bool) {
// 	if v, ok := m.calledFunc[name]; ok {
// 		return v, true
// 	}

// 	return 0, false
// }

// func (m *MockDB) CallCountReset() {
// 	m.calledFunc = map[string]int{
// 		// "Config":   0,
// 		// "Close":    0,
// 		// "Exec":     0,
// 		// "Query":    0,
// 		// "IsClosed": 0,
// 	}
// }

// func (m *MockDB) Config() *pgx.ConnConfig {
// 	m.calledFunc["Config"]++
// 	if m.onConfig != nil {
// 		return m.onConfig()
// 	}

// 	return &pgx.ConnConfig{
// 		Config: pgconn.Config{
// 			Host: "testhost",
// 			Port: 1337,
// 		},
// 	}
// }

// func (m *MockDB) Close(ctx context.Context) error {
// 	m.calledFunc["Close"]++
// 	if m.onClose != nil {
// 		return m.onClose(ctx)
// 	}

// 	return nil
// }

// func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
// 	m.calledFunc["Exec"]++
// 	if m.onExec != nil {
// 		return m.onExec(ctx, sql, args...)
// 	}

// 	return pgconn.NewCommandTag(""), nil
// }

// func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
// 	m.calledFunc["Query"]++
// 	if m.onQuery != nil {
// 		return m.onQuery(ctx, sql, args...)
// 	}

// 	return newMockRows(m, m.logr, []string{"foo", "bar"}), nil
// }

// func (m *MockDB) IsClosed() bool {
// 	m.calledFunc["IsClosed"]++
// 	if m.onIsClosed != nil {
// 		return m.onIsClosed()
// 	}

// 	return false
// }

// type mockRows struct {
// 	idx                int
// 	rows               []string
// 	logr               testLogFuncWrap
// 	onErr              func() error
// 	onCommandTag       func() pgconn.CommandTag
// 	onFieldDescription func() []pgconn.FieldDescription
// 	onNext             func() bool
// 	onScan             func(dest ...any) error
// 	onValues           func() ([]any, error)
// 	onRawValues        func() [][]byte
// 	onConn             func() *pgx.Conn
// 	onClose            func()
// }

// func newMockRows(_ *MockDB, logr testLogFuncWrap, o []string) *mockRows {
// 	return &mockRows{
// 		idx:  -1,
// 		rows: o,
// 		logr: logr,
// 	}
// }

// func (r *mockRows) Close() {
// 	if r.onClose != nil {
// 		r.onClose()
// 	}
// }

// func (r *mockRows) Err() error {
// 	if r.onErr != nil {
// 		return r.onErr()
// 	}

// 	return nil
// }

// func (r *mockRows) CommandTag() pgconn.CommandTag {
// 	if r.onCommandTag != nil {
// 		return r.onCommandTag()
// 	}

// 	return pgconn.NewCommandTag("")
// }

// func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription {
// 	if r.onFieldDescription != nil {
// 		return r.onFieldDescription()
// 	}

// 	return nil
// }

// func (r *mockRows) Next() bool {
// 	if r.logr != nil {
// 		r.logr("Next(); r.idx:%d", r.idx)
// 	}
// 	if r.onNext != nil {
// 		return r.onNext()
// 	}

// 	if r.idx >= len(r.rows)-1 {
// 		return false
// 	}

// 	r.idx++
// 	return true
// }

// func (r *mockRows) Scan(dest ...any) error {
// 	if r.onScan != nil {
// 		return r.onScan(dest...)
// 	}

// 	return nil
// }

// func (r *mockRows) Values() ([]any, error) {
// 	if r.logr != nil {
// 		r.logr("Values(); r.idx:%d", r.idx)
// 	}
// 	if r.onValues != nil {
// 		return r.onValues()
// 	}

// 	if r.idx >= 0 && r.idx < len(r.rows) {
// 		return []any{r.rows[r.idx]}, nil
// 	}

// 	return nil, pgx.ErrNoRows
// }

// func (r *mockRows) RawValues() [][]byte {
// 	if r.logr != nil {
// 		r.logr("RawValues(); r.idx:%d", r.idx)
// 	}

// 	if r.onRawValues != nil {
// 		return r.onRawValues()
// 	}

// 	if r.idx >= 0 && r.idx < len(r.rows) {
// 		buf := bytes.NewBuffer(nil)
// 		buf.WriteString(r.rows[r.idx])

// 		return [][]byte{
// 			buf.Bytes(),
// 		}
// 	}

// 	return nil
// }

// func (r *mockRows) Conn() *pgx.Conn {
// 	if r.onConn != nil {
// 		return r.onConn()
// 	}

// 	return nil
// }
