package controller

import "time"

const (
	finalizerName      = "kubeprj.github.io/database-account-finalizer"
	defaultRequeueTime = 5 * time.Second
)
