package test

import (
	"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

//nolint:gosec // mock data.
const (
	SecretApp        = "dosquad.github.io/database-account"
	SecretName       = "testaccount"
	SecretAPIVersion = "v1"
	SecretUUID       = "0923871f-c9db-4317-b97c-3099bc3cf84f"
	SecretJSONData   = `{
		"apiVersion": "` + SecretAPIVersion + `",
		"data": {},
		"immutable": true,
		"kind": "Secret",
		"metadata": {
			"creationTimestamp": "2023-04-03T05:06:07Z",
			"name": "` + SecretName + `",
			"namespace": "default",
			"resourceVersion": "123",
			"uid": "` + SecretUUID + `"
		},
		"type": "` + SecretApp + `"
	}`
)

func NewSecret() corev1.Secret {
	secret := corev1.Secret{}

	if err := json.NewDecoder(strings.NewReader(SecretJSONData)).Decode(&secret); err != nil {
		panic(err)
	}

	return secret
}
