package controller

import "errors"

var (
	// ErrNewSecret is returned when a secret was not able to be retrieved and so a new secret
	// was generated.
	ErrNewSecret = errors.New("creating new secret")

	// ErrNewStatefulSet is returned when a statefulset was not able to be retrieved and
	// so a new statefulset was generated.
	ErrNewStatefulSet = errors.New("creating new statefulset")

	// ErrNewService is returned when a service was not able to be retrieved and
	// so a new service was generated.
	ErrNewService = errors.New("creating new service")

	// ErrSecretImmutable is returned when a secret is immutable and can not be changed.
	ErrSecretImmutable = errors.New("secret is immutable")
)
