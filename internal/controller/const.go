package controller

import "time"

const (
	// finalizerName is used as the constant to bind the controller as the finalizer for the
	// resources it creates.
	finalizerName = "dbo.dosquad.github.io/database-account-finalizer"

	// defaultRequeueTime is the time to requeue a reconciliation request if there was an error.
	defaultRequeueTime = 5 * time.Second

	// secretType is the type used to indicate a secret has been created by this operator.
	//
	//nolint:gosec // not credentials.
	secretType = `dosquad.github.io/database-account`
)
