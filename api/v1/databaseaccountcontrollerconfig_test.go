package v1_test

import (
	"fmt"
	"testing"

	v1 "github.com/dosquad/database-operator/api/v1"
)

func TestDatabaseAccountControllerConfig_DefaultRelayImage(t *testing.T) {
	dbAccConfig := v1.DatabaseAccountControllerConfig{}

	if v := dbAccConfig.GetRelayImage(); v != v1.DefaultRelayImage {
		t.Errorf("DatabaseAccountControllerConfig.GetRelayImage() expected '%v' received '%v'", v1.DefaultRelayImage, v)
	}
}

func TestDatabaseAccountControllerConfig_CustomRelayImage(t *testing.T) {
	customImage := "custom-image"
	dbAccConfig := v1.DatabaseAccountControllerConfig{
		RelayImage: customImage,
	}

	if v := dbAccConfig.GetRelayImage(); v != customImage {
		t.Errorf("DatabaseAccountControllerConfig.GetRelayImage() expected '%v' received '%v'", customImage, v)
	}
}

func TestDatabaseAccountControllerConfig_DSNString(t *testing.T) {
	dsn := "postgresql://user:pass@host/database"
	dbAccConfig := v1.DatabaseAccountControllerConfig{
		DatabaseDSN: v1.PostgreSQLDSN(dsn),
	}

	if v := dbAccConfig.GetDSNHost(); v != "host" {
		t.Errorf("v1.PostgreSQLDSN.Host() expected '%v' received '%v'", "host", v)
	}
}

func TestDatabaseAccountControllerConfig_DSNString_Error(t *testing.T) {
	dsn := fmt.Sprintf("%c", 0x7f)
	dbAccConfig := v1.DatabaseAccountControllerConfig{
		DatabaseDSN: v1.PostgreSQLDSN(dsn),
	}

	if v := dbAccConfig.GetDSNHost(); v != "" {
		t.Errorf("v1.PostgreSQLDSN.Host() expected '%v' received '%v'", "", v)
	}
}
