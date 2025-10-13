# Temporal API Key Authorization

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
- `role` - Role name: `admin`, `writer`, `reader`, `worker`
- `namespace` - Temporal namespace or `*` for all namespaces

**Examples:**

```bash
# Admin with all namespaces
admin-secret:admin:*

# Writer for specific namespace
app1-key:writer:app1-namespace

# Multiple keys
admin-key:admin:*;app1:writer:ns1;app2:reader:ns2
```

### Helm

If you get an error `│ 2025/10/13 11:03:03 config file corrupted: no config files found within /etc/temporal/config`

```yaml
# specify config pass used by the temporal cluster
frontend-apikey:
  additionalEnv:
    - name: TEMPORAL_ENVIRONMENT
      value: kubernetes
    - name: TEMPORAL_ROOT
      value: /etc/temporal
    - name: TEMPORAL_CONFIG_DIR
      value: config
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
