package v1_test

import (
	"fmt"
	"testing"

	v1 "github.com/dosquad/database-operator/api/v1"
)

func TestDSNString(t *testing.T) {
	dsn := "postgresql://user:pass@host/database"

	psqlDSN := v1.PostgreSQLDSN(dsn)

	if psqlDSN.String() != dsn {
		t.Errorf("v1.PostgreSQLDSN.String() expected '%v' received '%v'", dsn, psqlDSN.String())
	}

	if psqlDSN.Host() != "host" {
		t.Errorf("v1.PostgreSQLDSN.Host() expected '%v' received '%v'", "host", psqlDSN.Host())
	}
}

func TestDSNString_Error(t *testing.T) {
	dsn := fmt.Sprintf("%c", 0x7f)

	psqlDSN := v1.PostgreSQLDSN(dsn)

	if psqlDSN.String() != dsn {
		t.Errorf("v1.PostgreSQLDSN.String() expected '%v' received '%v'", dsn, psqlDSN.String())
	}

	if psqlDSN.Host() != "" {
		t.Errorf("v1.PostgreSQLDSN.Host() expected '' received '%v'", psqlDSN.Host())
	}
}
