package v1_test

import (
	"testing"
	"time"

	v1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/testhelp"
)

func TestDatabaseAccountOnDelete_String(t *testing.T) {
	t.Parallel()
	start := time.Now()

	tests := []struct {
		name   string
		item   v1.DatabaseAccountOnDelete
		expect string
	}{
		{"OnDeleteRetain", v1.OnDeleteRetain, "retain"},
		{"OnDeleteDelete", v1.OnDeleteDelete, "delete"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if v := tt.item.String(); v != tt.expect {
				testhelp.Errorf(t, start, "v1.DatabaseAccountOnDelete.String(): got '%s', want '%s'", v, tt.expect)
			}
		})
	}
}

func TestDatabaseAccountCreateStage_String(t *testing.T) {
	t.Parallel()
	start := time.Now()

	tests := []struct {
		name   string
		item   v1.DatabaseAccountCreateStage
		expect string
	}{
		{"UnknownStage", v1.UnknownStage, ""},
		{"InitStage", v1.InitStage, "Init"},
		{"UserCreateStage", v1.UserCreateStage, "UserCreate"},
		{"DatabaseCreateStage", v1.DatabaseCreateStage, "DatabaseCreate"},
		{"RelayCreateStage", v1.RelayCreateStage, "RelayCreate"},
		{"ErrorStage", v1.ErrorStage, "Error"},
		{"ReadyStage", v1.ReadyStage, "Ready"},
		{"TerminatingStage", v1.TerminatingStage, "Terminating"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if v := tt.item.String(); v != tt.expect {
				testhelp.Errorf(t, start, "v1.DatabaseAccountCreateStage.String(): got '%s', want '%s'", v, tt.expect)
			}
		})
	}
}
