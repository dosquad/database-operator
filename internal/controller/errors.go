package controller

import "errors"

var (
	// ErrNewSecret is returned when a secret was not able to be retrieved and so a new secret
	// was generated.
	ErrNewSecret = errors.New("creating new secret")

	// ErrSecretImmutable is returned when a secret is immutable and can not be changed.
	ErrSecretImmutable = errors.New("secret is immutable")
)
