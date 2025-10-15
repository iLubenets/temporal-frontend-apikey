# Temporal API Key Authorization

[![CI](https://github.com/ilubenets/temporal-frontend-apikey/actions/workflows/ci.yaml/badge.svg)](https://github.com/ilubenets/temporal-frontend-apikey/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilubenets/temporal-frontend-apikey)](https://goreportcard.com/report/github.com/ilubenets/temporal-frontend-apikey)

A custom Temporal frontend server with API key-based authentication using a custom `ClaimMapper`.

[Temporal.io Claim-Mapper Plugin](https://docs.temporal.io/self-hosted-guide/security#claim-mapper)

## Key Components

- **`main.go`**: Simplified server initialization (recommended approach)
- **`api_key_claim_mapper.go`**: Custom ClaimMapper implementation
- Uses **DefaultAuthorizer** for standard Temporal authorization logic
- Only customizes authentication, not the entire serve

## Build

```bash
docker build --build-arg TEMPORAL_VERSION=1.28.1 -t temporal-frontend-apikey:1.28.1 .
```

## Run

The server accepts API keys via the `TEMPORAL_API_KEYS` environment variable:

```bash
TEMPORAL_API_KEYS="key:role:namespace;key2:role2:ns2"
```

**Format:** `key:role:namespace`

- `key` - The API key (used in `Authorization: Bearer <key>`)
- `role` - Role name: `admin`, `write`, `read`, `worker`
- `namespace` - Temporal namespace or `*` for all namespaces (system level)

**Examples:**

```bash
# Admin with all namespaces
admin-secret:admin:*

# Writer for specific namespace
app1-key:writer:app1-namespace

# Multiple keys
admin-key:admin:*;app1:write:ns1;app2:read:ns2
```

### Helm

If you get an error `│ 2025/10/13 11:03:03 config file corrupted: no config files found within /etc/temporal/config`

```yaml
# specify the correct env used by the temporal cluster
# - docker
# - kubernetes
# - some custom one
frontend-external:
  additionalEnv:
    - name: TEMPORAL_ENVIRONMENT
      value: kubernetes
```

## Test (integration)

```bash
# run all tests with docker
make test-all

# start local cluster
make docker-build
make docker-up
make docker-down
```
