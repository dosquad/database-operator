package controller

import "errors"

var ErrNewSecret = errors.New("creating new secret")
var ErrSecretImmutable = errors.New("secret is immutable")
