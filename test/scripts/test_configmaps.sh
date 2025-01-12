#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base URL and test namespace
BASE_URL="http://localhost:8080"
TEST_NS="default"
TEST_CM="app-config"

echo "Testing ConfigMap API endpoints..."
echo "=============================="

# 1. Create a new ConfigMap
echo -e "\n1. Creating new ConfigMap '${TEST_CM}'..."
curl -s -X POST "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'${TEST_CM}'",
    "data": {
      "database.host": "localhost",
      "database.port": "5432",
      "database.name": "myapp",
      "api.endpoint": "https://api.example.com",
      "log.level": "INFO"
    },
    "labels": {
      "app": "myapp",
      "environment": "development"
    },
    "annotations": {
      "description": "Application configuration"
    }
  }' | jq '.'

# 2. List all ConfigMaps
echo -e "\n2. Listing all ConfigMaps in namespace '${TEST_NS}'..."
curl -s -X GET "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}" | jq '.'

# 3. Get specific ConfigMap
echo -e "\n3. Getting ConfigMap '${TEST_CM}' details..."
curl -s -X GET "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}" | jq '.'

# 4. Update ConfigMap
echo -e "\n4. Updating ConfigMap '${TEST_CM}'..."
curl -s -X PUT "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "database.host": "db.example.com",
      "database.port": "5432",
      "database.name": "myapp",
      "api.endpoint": "https://api.example.com",
      "log.level": "DEBUG",
      "feature.flags": "cache=true,metrics=true"
    },
    "labels": {
      "app": "myapp",
      "environment": "development",
      "version": "1.1"
    }
  }' | jq '.'

# 5. Get ConfigMap usage
echo -e "\n5. Getting ConfigMap '${TEST_CM}' usage..."
curl -s -X GET "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}/usage" | jq '.'

# 6. Delete ConfigMap (commented out for safety)
echo -e "\n6. Delete ConfigMap command (commented out for safety)..."
echo "# curl -X DELETE '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}'"

# Reference commands
echo -e "\n=============================="
echo "Quick reference commands:"
echo -e "${GREEN}# Create ConfigMap${NC}"
echo "curl -X POST '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# List ConfigMaps${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}'"

echo -e "\n${GREEN}# Get ConfigMap details${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}'"

echo -e "\n${GREEN}# Get ConfigMap usage${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}/usage'"

echo -e "\n${GREEN}# Update ConfigMap${NC}"
echo "curl -X PUT '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# Delete ConfigMap${NC}"
echo "curl -X DELETE '${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}/${TEST_CM}'"

# Example with multiple config files
echo -e "\n${GREEN}# Create ConfigMap with multiple files${NC}"
echo 'curl -X POST "${BASE_URL}/api/v1/configmaps/namespaces/${TEST_NS}" -H "Content-Type: application/json" -d '"'{
  "name": "app-config-files",
  "data": {
    "config.yaml": "apiVersion: v1\nkind: Config\nmetadata:\n  name: app-config",
    "settings.json": "{\n  \"debug\": true,\n  \"cache\": {\n    \"enabled\": true,\n    \"ttl\": 3600\n  }\n}",
    ".env": "DB_HOST=localhost\nDB_PORT=5432\nAPI_KEY=secret"
  }
}'"