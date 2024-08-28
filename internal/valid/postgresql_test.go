package valid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dosquad/database-operator/internal/valid"
)

func TestPGIdentifier_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value  string
		expect string
	}{
		{"00000", "00000"},
		{"$00000", "00000"},
		{"00000$", "00000"},
		{"$00$000$", "00000"},
		{"__", "__"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			if v := valid.PGIdentifier(tt.value).String(); v != tt.expect {
				t.Errorf("String(): got:'%s' want:'%s'", v, tt.expect)
			}

			if v := valid.PGIdentifier(tt.value).Sanitize(); v != `"`+tt.expect+`"` {
				t.Errorf("Sanitize(): got:'%s' want:'%s'", v, `"`+tt.expect+`"`)
			}
		})
	}
}

func TestPGIdentifier_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value       string
		expect      string
		expectError error
	}{
		{"00000", "00000", valid.ErrInvalidName},
		{"$00000", "00000", valid.ErrInvalidName},
		{"00000$", "00000", valid.ErrInvalidName},
		{"$00$000$", "00000", valid.ErrInvalidName},
		{"__", "__", valid.ErrInvalidName},
		{
			strings.Repeat("0", valid.PostgreSQLNameDataLen),
			strings.Repeat("0", valid.PostgreSQLNameDataLen),
			valid.ErrInvalidName,
		},
		{
			strings.Repeat("a", valid.PostgreSQLNameDataLen),
			strings.Repeat("a", valid.PostgreSQLNameDataLen),
			valid.ErrInvalidName,
		},
		{
			strings.Repeat("a", valid.PostgreSQLNameDataLen-1),
			strings.Repeat("a", valid.PostgreSQLNameDataLen-1),
			nil,
		},
		{"psql", "psql", valid.ErrInvalidName},
		{"postgres", "postgres", valid.ErrInvalidName},
		{"root", "root", valid.ErrInvalidName},
		{"a___", "a___", nil},
		{"___a", "___a", nil},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			v, err := valid.PGIdentifier(tt.value).Validate()
			if !errors.Is(err, tt.expectError) {
				t.Errorf("Validate(): got:'%+v' want:'%+v'", err, tt.expectError)
			}

			if v != tt.expect {
				t.Errorf("Validate(): got:'%s' want:'%s'", v, tt.expect)
			}
		})
	}
}

func TestPGValue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value  string
		expect string
	}{
		{"00000", "00000"},
		{"$00000", "$00000"},
		{"00000$", "00000$"},
		{"$00$000$", "$00$000$"},
		{"__", "__"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			if v := valid.PGValue(tt.value).String(); v != tt.expect {
				t.Errorf("String(): got:'%s' want:'%s'", v, tt.expect)
			}

			if v := valid.PGValue(tt.value).Sanitize(); v != `'`+tt.expect+`'` {
				t.Errorf("Sanitize(): got:'%s' want:'%s'", v, `'`+tt.expect+`'`)
			}
		})
	}
}
