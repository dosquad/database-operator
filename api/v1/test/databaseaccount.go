package test

import (
	"encoding/json"
	"strings"

	v1 "github.com/dosquad/database-operator/api/v1"
)

const (
	DBUser       = `k8s_01h97g9exfs6bw874x0k567jr7`
	DBApp        = "testapp"
	DBAccount    = "testaccount"
	DBGroup      = "dbo.dosquad.github.io"
	DBAPIVersion = DBGroup + "/v1"
	DBKind       = "DatabaseAccount"
	DBUUID       = "a6533e70-9716-4c68-91a1-851a6fef074f"
	DBJSONData   = `{
		"apiVersion": "` + DBAPIVersion + `",
		"kind": "` + DBKind + `",
		"metadata": {
			"annotations": {},
			"creationTimestamp": "2023-04-03T05:06:07Z",
			"generation": 1,
			"labels": {
				"app": "` + DBApp + `"
			},
			"name": "` + DBAccount + `",
			"namespace": "default",
			"resourceVersion": "10",
			"uid": "` + DBUUID + `"
		},
		"status": {
			"name": "` + DBUser + `",
			"ready": true,
			"stage": "Ready"
		}
	}`
)

func NewDatabaseAccount() v1.DatabaseAccount {
	dbac := v1.DatabaseAccount{}

	if err := json.NewDecoder(strings.NewReader(DBJSONData)).Decode(&dbac); err != nil {
		panic(err)
	}

	return dbac
}

func NewInitDatabaseAccount() v1.DatabaseAccount {
	dbac := v1.DatabaseAccount{}

	if err := json.NewDecoder(strings.NewReader(DBJSONData)).Decode(&dbac); err != nil {
		panic(err)
	}

	dbac.Status = v1.DatabaseAccountStatus{}

	return dbac
}
