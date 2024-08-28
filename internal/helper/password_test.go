package helper_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dosquad/database-operator/internal/helper"
)

func TestGeneratePassword_Clean(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pw1 := helper.GeneratePassword(ctx)
	t.Logf("Password(1)[%s]", pw1)

	pw2 := helper.GeneratePassword(ctx)
	t.Logf("Password(2)[%s]", pw2)

	if strings.EqualFold(pw1, pw2) {
		t.Errorf("Generated passwords should not be equal: %s == %s", pw1, pw2)
	}
}
