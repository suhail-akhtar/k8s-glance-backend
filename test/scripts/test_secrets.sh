#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base URL and test namespace
BASE_URL="http://localhost:8080"
TEST_NS="default"
TEST_SECRET="app-credentials"

echo "Testing Secrets API endpoints..."
echo "=============================="

# 1. Create a new Secret
echo -e "\n1. Creating new Secret '${TEST_SECRET}'..."
curl -s -X POST "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'${TEST_SECRET}'",
    "type": "Opaque",
    "stringData": {
      "username": "admin",
      "password": "secretpassword123",
      "api.key": "abcdef123456",
      "config.json": "{\"endpoint\":\"https://api.example.com\",\"timeout\":30}"
    },
    "labels": {
      "app": "myapp",
      "environment": "development"
    },
    "annotations": {
      "description": "Application credentials"
    }
  }' | jq '.'

# 2. List all Secrets
echo -e "\n2. Listing all Secrets in namespace '${TEST_NS}'..."
curl -s -X GET "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}" | jq '.'

# 3. Get Secret metadata (no values)
echo -e "\n3. Getting Secret '${TEST_SECRET}' metadata..."
curl -s -X GET "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}" | jq '.'

# 4. Get Secret keys
echo -e "\n4. Getting Secret '${TEST_SECRET}' keys..."
curl -s -X GET "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}/keys" | jq '.'

# 5. Update Secret
echo -e "\n5. Updating Secret '${TEST_SECRET}'..."
curl -s -X PUT "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}" \
  -H "Content-Type: application/json" \
  -d '{
    "stringData": {
      "username": "admin",
      "password": "newpassword456",
      "api.key": "xyz789",
      "config.json": "{\"endpoint\":\"https://api.example.com\",\"timeout\":60}"
    },
    "labels": {
      "app": "myapp",
      "environment": "development",
      "version": "1.1"
    }
  }' | jq '.'

# 6. Get Secret usage
echo -e "\n6. Getting Secret '${TEST_SECRET}' usage..."
curl -s -X GET "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}/usage" | jq '.'

# 7. Delete Secret (commented out for safety)
echo -e "\n7. Delete Secret command (commented out for safety)..."
echo "# curl -X DELETE '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}'"

# Reference commands
echo -e "\n=============================="
echo "Quick reference commands:"
echo -e "${GREEN}# Create Secret${NC}"
echo "curl -X POST '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# List Secrets${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}'"

echo -e "\n${GREEN}# Get Secret metadata${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}'"

echo -e "\n${GREEN}# Get Secret keys${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}/keys'"

echo -e "\n${GREEN}# Get Secret usage${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}/usage'"

echo -e "\n${GREEN}# Update Secret${NC}"
echo "curl -X PUT '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# Delete Secret${NC}"
echo "curl -X DELETE '${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}/${TEST_SECRET}'"

# Example with TLS Secret
echo -e "\n${GREEN}# Create TLS Secret example${NC}"
echo 'curl -X POST "${BASE_URL}/api/v1/secrets/namespaces/${TEST_NS}" -H "Content-Type: application/json" -d '"'{
  "name": "tls-secret",
  "type": "kubernetes.io/tls",
  "stringData": {
    "tls.crt": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "tls.key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
  }
}'"