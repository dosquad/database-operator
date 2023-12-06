package test

import (
	"bytes"
	"context"
	"testing"

	v1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/testhelp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	Hostname = "testhost"
	Port     = 1337
)

type MockDB struct {
	dsn        v1.PostgreSQLDSN
	logr       testhelp.TestLogFuncWrap
	calledFunc map[string]int
	OnConfig   func() *pgx.ConnConfig
	OnClose    func(context.Context) error
	OnExec     func(context.Context, string, ...any) (pgconn.CommandTag, error)
	OnQuery    func(context.Context, string, ...any) (pgx.Rows, error)
	OnIsClosed func() bool
}

func NewMockDB(_ *testing.T, logr testhelp.TestLogFuncWrap, dsn v1.PostgreSQLDSN) *MockDB {
	m := &MockDB{
		dsn:  dsn,
		logr: logr,
	}

	m.CallCountReset()

	return m
}

func (m *MockDB) SetLogger(logr testhelp.TestLogFuncWrap) {
	m.logr = logr
}

func (m *MockDB) CallCountMap() map[string]int {
	return m.calledFunc
}

func (m *MockDB) CallCount(name string) (int, bool) {
	if v, ok := m.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (m *MockDB) CallCountReset() {
	m.calledFunc = map[string]int{
		// "Config":   0,
		// "Close":    0,
		// "Exec":     0,
		// "Query":    0,
		// "IsClosed": 0,
	}
}

func (m *MockDB) Config() *pgx.ConnConfig {
	m.calledFunc["Config"]++
	if m.OnConfig != nil {
		return m.OnConfig()
	}

	return &pgx.ConnConfig{
		Config: pgconn.Config{
			Host: Hostname,
			Port: Port,
		},
	}
}

func (m *MockDB) Close(ctx context.Context) error {
	m.calledFunc["Close"]++
	if m.OnClose != nil {
		return m.OnClose(ctx)
	}

	return nil
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.calledFunc["Exec"]++
	if m.OnExec != nil {
		return m.OnExec(ctx, sql, args...)
	}

	return pgconn.NewCommandTag(""), nil
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.calledFunc["Query"]++
	if m.OnQuery != nil {
		return m.OnQuery(ctx, sql, args...)
	}

	return NewMockRows(m, m.logr, []string{"foo", "bar"}), nil
}

func (m *MockDB) IsClosed() bool {
	m.calledFunc["IsClosed"]++
	if m.OnIsClosed != nil {
		return m.OnIsClosed()
	}

	return false
}

type MockRows struct {
	idx                int
	rows               []string
	logr               testhelp.TestLogFuncWrap
	calledFunc         map[string]int
	OnErr              func() error
	OnCommandTag       func() pgconn.CommandTag
	OnFieldDescription func() []pgconn.FieldDescription
	OnNext             func() bool
	OnScan             func(dest ...any) error
	OnValues           func() ([]any, error)
	OnRawValues        func() [][]byte
	OnConn             func() *pgx.Conn
	OnClose            func()
}

func NewMockRows(_ *MockDB, logr testhelp.TestLogFuncWrap, o []string) *MockRows {
	m := &MockRows{
		idx:  -1,
		rows: o,
		logr: logr,
	}
	m.CallCountReset()

	return m
}

func (r *MockRows) CallCountMap() map[string]int {
	return r.calledFunc
}

func (r *MockRows) CallCount(name string) (int, bool) {
	if v, ok := r.calledFunc[name]; ok {
		return v, true
	}

	return 0, false
}

func (r *MockRows) CallCountReset() {
	r.calledFunc = map[string]int{}
}

func (r *MockRows) Close() {
	if r.OnClose != nil {
		r.OnClose()
	}
}

func (r *MockRows) Err() error {
	if r.OnErr != nil {
		return r.OnErr()
	}

	return nil
}

func (r *MockRows) CommandTag() pgconn.CommandTag {
	if r.OnCommandTag != nil {
		return r.OnCommandTag()
	}

	return pgconn.NewCommandTag("")
}

func (r *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	if r.OnFieldDescription != nil {
		return r.OnFieldDescription()
	}

	return nil
}

func (r *MockRows) Next() bool {
	if r.logr != nil {
		r.logr("Next(); r.idx:%d", r.idx)
	}
	if r.OnNext != nil {
		return r.OnNext()
	}

	if r.idx >= len(r.rows)-1 {
		return false
	}

	r.idx++
	return true
}

func (r *MockRows) Scan(dest ...any) error {
	if r.OnScan != nil {
		return r.OnScan(dest...)
	}

	return nil
}

func (r *MockRows) Values() ([]any, error) {
	if r.logr != nil {
		r.logr("Values(); r.idx:%d", r.idx)
	}
	if r.OnValues != nil {
		return r.OnValues()
	}

	if r.idx >= 0 && r.idx < len(r.rows) {
		return []any{r.rows[r.idx]}, nil
	}

	return nil, pgx.ErrNoRows
}

func (r *MockRows) RawValues() [][]byte {
	if r.logr != nil {
		r.logr("RawValues(); r.idx:%d", r.idx)
	}

	if r.OnRawValues != nil {
		return r.OnRawValues()
	}

	if r.idx >= 0 && r.idx < len(r.rows) {
		buf := bytes.NewBuffer(nil)
		buf.WriteString(r.rows[r.idx])

		return [][]byte{
			buf.Bytes(),
		}
	}

	return nil
}

func (r *MockRows) Conn() *pgx.Conn {
	if r.OnConn != nil {
		return r.OnConn()
	}

	return nil
}
