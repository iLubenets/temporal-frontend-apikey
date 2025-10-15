#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FRONTEND_ADDRESS="localhost:7233"
OAUTH_SERVER="http://localhost:8081"
OAUTH_ISSUER="default"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Multi-Claim Mapper${NC}"
echo -e "${BLUE}(API Key + OAuth JWT)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to test API key authentication
test_api_key() {
    local key=$1
    local description=$2
    
    echo -e "${YELLOW}Testing API Key: ${description}${NC}"
    
    # Use tctl to list namespaces with API key
    docker exec temporal-admin-tools bash -c "
        tctl --address ${FRONTEND_ADDRESS} \
             --tls-disable \
             --auth-header 'Bearer ${key}' \
             namespace list 2>&1
    " | head -n 5
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ API Key authentication successful${NC}"
    else
        echo -e "${RED}✗ API Key authentication failed${NC}"
    fi
    echo ""
}

# Function to get OAuth token
get_oauth_token() {
    local grant_type=$1
    local client_id=${2:-"temporal-ui"}
    local client_secret=${3:-"temporal-ui-secret"}
    
    if [ "$grant_type" == "client_credentials" ]; then
        response=$(curl -s -X POST "${OAUTH_SERVER}/token" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "grant_type=client_credentials" \
            -d "client_id=${client_id}" \
            -d "client_secret=${client_secret}" \
            -d "scope=openid profile")
    else
        # For testing, we'll simulate getting a token with authorization_code flow
        # In real scenario, this would require browser interaction
        echo -e "${YELLOW}Note: Authorization code flow requires browser interaction${NC}"
        echo -e "${YELLOW}Using client_credentials for testing instead${NC}"
        response=$(curl -s -X POST "${OAUTH_SERVER}/token" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "grant_type=client_credentials" \
            -d "client_id=${client_id}" \
            -d "client_secret=${client_secret}" \
            -d "scope=openid profile")
    fi
    
    echo "$response" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//' | sed 's/"$//'
}

# Function to test JWT authentication
test_jwt() {
    local description=$1
    local token=$2
    
    echo -e "${YELLOW}Testing JWT: ${description}${NC}"
    
    if [ -z "$token" ]; then
        echo -e "${RED}✗ Failed to obtain JWT token${NC}"
        echo ""
        return 1
    fi
    
    # Decode JWT to show claims (just for visibility)
    echo -e "${BLUE}JWT Claims:${NC}"
    echo "$token" | awk -F'.' '{print $2}' | base64 -d 2>/dev/null | jq '.' 2>/dev/null || echo "Could not decode token"
    echo ""
    
    # Test with JWT token
    docker exec temporal-admin-tools bash -c "
        tctl --address ${FRONTEND_ADDRESS} \
             --tls-disable \
             --auth-header 'Bearer ${token}' \
             namespace list 2>&1
    " | head -n 5
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ JWT authentication successful${NC}"
    else
        echo -e "${RED}✗ JWT authentication failed${NC}"
    fi
    echo ""
}

# Function to check OAuth server health
check_oauth_server() {
    echo -e "${YELLOW}Checking OAuth server...${NC}"
    response=$(curl -s "${OAUTH_SERVER}/.well-known/openid-configuration")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ OAuth server is running${NC}"
        echo -e "${BLUE}Issuer: $(echo $response | jq -r '.issuer')${NC}"
        echo -e "${BLUE}JWKS URI: $(echo $response | jq -r '.jwks_uri')${NC}"
    else
        echo -e "${RED}✗ OAuth server is not accessible${NC}"
        exit 1
    fi
    echo ""
}

# Main test execution
echo -e "${BLUE}Step 1: Verify OAuth Server${NC}"
check_oauth_server

echo -e "${BLUE}Step 2: Test API Key Authentication${NC}"
echo -e "${BLUE}========================================${NC}"
test_api_key "admin-key" "Admin API Key (admin:*)"
test_api_key "test-key" "Service API Key (service:default)"
test_api_key "invalid-key" "Invalid API Key (should fail)"

echo -e "${BLUE}Step 3: Test OAuth JWT Authentication${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${YELLOW}Obtaining JWT token from OAuth server...${NC}"
jwt_token=$(get_oauth_token "client_credentials" "temporal-frontend" "temporal-frontend-secret")

if [ -n "$jwt_token" ]; then
    test_jwt "OAuth Client Credentials" "$jwt_token"
else
    echo -e "${RED}✗ Failed to obtain JWT token${NC}"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Summary:${NC}"
echo -e "1. API keys are tested first by multi-claim mapper"
echo -e "2. If API key not found, JWT validation is attempted"
echo -e "3. Both authentication methods can coexist"
echo ""
echo -e "${YELLOW}Additional Testing:${NC}"
echo -e "- Visit http://localhost:8080 to test UI OAuth login"
echo -e "- Visit http://localhost:8081/default/.well-known/openid-configuration for OAuth config"
echo -e "- Check logs: docker compose logs temporal-frontend-external"
