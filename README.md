# database-operator

[![CI](https://github.com/dosquad/database-operator/actions/workflows/ci.yml/badge.svg)](https://github.com/dosquad/database-operator/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/dosquad/database-operator/branch/main/graph/badge.svg?token=48QZ7PDBL5)](https://codecov.io/github/dosquad/database-operator)
[![GitHub](https://img.shields.io/github/license/dosquad/database-operator)](LICENSE)

Database account configuration operator

## Usage

### Install

```shell
make install
make deploy
```

Then modify and apply the following configmap.

```yaml
---
apiVersion: v1
data:
  controller_manager_config.yaml: |
    apiVersion: dbo.dosquad.github.io/v1
    kind: DatabaseAccountControllerConfig
    dsn: "postgres://databaseuser:databasepassword@psqlserver.example.com:5432/postgres"
kind: ConfigMap
metadata:
  name: database-operator-manager-config
  namespace: database-operator-system
```

### Database accounts

Modify and apply the following database account resource.

```yaml
---
apiVersion: dbo.dosquad.github.io/v1
kind: DatabaseAccount
metadata:
  labels:
    app: testapp
  name: testaccount
```
