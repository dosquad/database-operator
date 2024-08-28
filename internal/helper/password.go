package helper

import (
	"context"

	"github.com/sethvargo/go-password/password"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	passwordLength         = 28
	passwordComplexDigits  = 10
	passwordComplexSymbols = 1
)

// TODO actually generate password
func GeneratePassword(ctx context.Context) string {
	logger := logr.FromContext(ctx)

	res, err := password.Generate(passwordLength, passwordComplexDigits, passwordComplexSymbols, false, false)
	if err != nil {
		logger.Error(err, "unable to generate password")
		panic(err)
	}
	return res
}
