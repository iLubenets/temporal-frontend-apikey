#!/bin/bash
# Test script for Temporal API Key authentication
# Uses grpcurl to test gRPC endpoints

set -e

TEMPORAL_ADDRESS="${TEMPORAL_ADDRESS:-localhost:7233}"
ADMIN_KEY="${ADMIN_KEY:-admin-key}"
INVALID_KEY="invalid-key-12345"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔑 Testing Temporal API Key Authentication"
echo "   Endpoint: $TEMPORAL_ADDRESS"
echo ""

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${YELLOW}⚠️  grpcurl not found. Installing...${NC}"
    echo "   Run: brew install grpcurl  (macOS)"
    echo "   Or: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    echo ""
fi

# Test 1: Health check (no auth required)
echo "Test 1: Health check (no authentication required)"
if grpcurl -plaintext "$TEMPORAL_ADDRESS" grpc.health.v1.Health/Check 2>&1 | grep -q "SERVING"; then
    echo -e "${GREEN}✅ PASS: Health check successful${NC}"
else
    echo -e "${RED}❌ FAIL: Health check failed${NC}"
fi
echo ""

# Test 2: ListNamespaces WITHOUT API key (should fail)
echo "Test 2: ListNamespaces without API key (should fail)"
if grpcurl -plaintext \
    -d '{}' \
    "$TEMPORAL_ADDRESS" \
    temporal.api.workflowservice.v1.WorkflowService/ListNamespaces 2>&1 | grep -q "PermissionDenied"; then
    echo -e "${GREEN}✅ PASS: Request rejected without API key${NC}"
else
    echo -e "${YELLOW}⚠️  WARNING: Request not rejected (may allow unauthenticated access)${NC}"
fi
echo ""

# Test 3: ListNamespaces WITH invalid API key (should fail)
echo "Test 3: ListNamespaces with invalid API key (should fail)"
if grpcurl -plaintext \
    -H "Authorization: Bearer $INVALID_KEY" \
    -d '{}' \
    "$TEMPORAL_ADDRESS" \
    temporal.api.workflowservice.v1.WorkflowService/ListNamespaces 2>&1 | grep -q -E "(PermissionDenied|Unauthenticated|invalid)"; then
    echo -e "${GREEN}✅ PASS: Invalid API key rejected${NC}"
else
    echo -e "${RED}❌ FAIL: Invalid API key not rejected${NC}"
fi
echo ""

# Test 4: ListNamespaces WITH valid API key (should succeed)
echo "Test 4: ListNamespaces with valid API key (should succeed)"
RESPONSE=$(grpcurl -plaintext \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -d '{}' \
    "$TEMPORAL_ADDRESS" \
    temporal.api.workflowservice.v1.WorkflowService/ListNamespaces 2>&1)

if echo "$RESPONSE" | grep -q '"name"'; then
    echo -e "${GREEN}✅ PASS: Valid API key accepted${NC}"
    echo "   Namespaces found:"
    echo "$RESPONSE" | grep '"name"' | head -3 | sed 's/^/   /'
else
    echo -e "${RED}❌ FAIL: Valid API key rejected${NC}"
    echo "   Response: $RESPONSE"
fi
echo ""

# Test 5: Using Go test client (if available)
if [ -f "test-client/test_client.go" ]; then
    echo "Test 5: Running Go SDK test client"
    cd test-client
    if go run test_client.go 2>&1 | grep -q "Connected"; then
        echo -e "${GREEN}✅ PASS: Go SDK client connected successfully${NC}"
    else
        echo -e "${RED}❌ FAIL: Go SDK client connection failed${NC}"
    fi
    cd ..
    echo ""
fi

echo "🏁 Test suite completed"
