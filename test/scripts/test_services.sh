#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base URL and test namespace
BASE_URL="http://localhost:8080"
TEST_NS="default"
TEST_SERVICE="nginx-service"

echo "Testing Service API endpoints..."
echo "=============================="

# 1. Create a new service
echo -e "\n1. Creating new service '${TEST_SERVICE}'..."
curl -s -X POST "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'${TEST_SERVICE}'",
    "type": "ClusterIP",
    "ports": [
      {
        "name": "http",
        "port": 80,
        "targetPort": 8080,
        "protocol": "TCP"
      }
    ],
    "selector": {
      "app": "nginx"
    },
    "labels": {
      "app": "nginx",
      "environment": "test"
    }
  }' | jq '.'

# 2. List all services
echo -e "\n2. Listing all services in namespace '${TEST_NS}'..."
curl -s -X GET "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}" | jq '.'

# 3. Get specific service
echo -e "\n3. Getting service '${TEST_SERVICE}' details..."
curl -s -X GET "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}" | jq '.'

# 4. Get service status
echo -e "\n4. Getting service '${TEST_SERVICE}' status..."
curl -s -X GET "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}/status" | jq '.'

# 5. Update service
echo -e "\n5. Updating service '${TEST_SERVICE}'..."
curl -s -X PUT "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}" \
  -H "Content-Type: application/json" \
  -d '{
    "ports": [
      {
        "name": "http",
        "port": 80,
        "targetPort": 80,
        "protocol": "TCP"
      },
      {
        "name": "https",
        "port": 443,
        "targetPort": 443,
        "protocol": "TCP"
      }
    ],
    "selector": {
      "app": "nginx",
      "environment": "prod"
    }
  }' | jq '.'

# 6. Delete service (commented out for safety)
echo -e "\n6. Delete service command (commented out for safety)..."
echo "# curl -X DELETE '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}'"

# Reference commands
echo -e "\n=============================="
echo "Quick reference commands:"
echo -e "${GREEN}# Create service${NC}"
echo "curl -X POST '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# List services${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}'"

echo -e "\n${GREEN}# Get service details${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}'"

echo -e "\n${GREEN}# Get service status${NC}"
echo "curl -X GET '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}/status'"

echo -e "\n${GREEN}# Update service${NC}"
echo "curl -X PUT '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}' -H 'Content-Type: application/json' -d '{...}'"

echo -e "\n${GREEN}# Delete service${NC}"
echo "curl -X DELETE '${BASE_URL}/api/v1/services/namespaces/${TEST_NS}/${TEST_SERVICE}'"

# Example LoadBalancer service
echo -e "\n${GREEN}# Create LoadBalancer service example${NC}"
echo 'curl -X POST "${BASE_URL}/api/v1/services/namespaces/${TEST_NS}" -H "Content-Type: application/json" -d '"'{
  \"name\": \"nginx-lb\",
  \"type\": \"LoadBalancer\",
  \"ports\": [
    {
      \"name\": \"http\",
      \"port\": 80,
      \"targetPort\": 80,
      \"protocol\": \"TCP\"
    }
  ],
  \"selector\": {
    \"app\": \"nginx\"
  }
}'"