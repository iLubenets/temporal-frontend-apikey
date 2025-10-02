# Temporal API Key Authorization


## Setup

```bash
docker build --build-arg TEMPORAL_VERSION=1.28.1 -t temporal-frontend-apikey:1.28.1 .
```

## Configuration

Run `temporal-frontend-apikey` service with:
```bash
TEMPORAL_API_KEYS="key:role:namespace;key2:role2:ns2"
```

**Format:** `key:role:namespace`
- `key` - The API key (used in `Authorization: Bearer <key>`)
- `role` - Role name (for logging)
- `namespace` - Temporal namespace or `*` for all

**Examples:**
```bash
# Admin with all namespaces
admin-secret:admin:*

# Service with specific namespace
app1-key:service:app1-namespace

# Multiple keys
admin:admin:*;app1:svc:ns1;app2:svc:ns2
```
## Testing

```bash
docker build --build-arg TEMPORAL_VERSION=1.28.1 -t temporal-frontend-apikey:1.28.1 .

cd test-docker-compose

# no -d to see logs
docker compose up

# use your temporal client wiht 
# eg spring boot
#   api-key: admin-key
#   enable-https: false
# or run simple test:
go run test_client.go

# open localhost:8080 to see WF

```
