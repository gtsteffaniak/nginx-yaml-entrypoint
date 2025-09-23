#!/bin/bash
# Quick test script for CI/CD - tests all examples in parallel with shorter duration

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="docker-compose.test.yml"
TEST_DURATION=10  # Quick 10-second tests

echo -e "${BLUE}⚡ Quick regression tests${NC}"

# Setup
mkdir -p test/ssl test/mock-responses

# Generate SSL certs if needed
if [ ! -f test/ssl/cert.pem ]; then
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout test/ssl/key.pem \
        -out test/ssl/cert.pem \
        -subj "/C=US/ST=Test/L=Test/O=Test/OU=Test/CN=localhost" \
        2>/dev/null
fi

# Create mock response
echo "<h1>Mock Backend</h1>" > test/mock-responses/index.html

# Build
docker-compose -f $COMPOSE_FILE build --no-cache

# Start all services
echo "Starting all test services..."
docker-compose -f $COMPOSE_FILE up -d

# Wait and check
sleep $TEST_DURATION

# Check which services are healthy
FAILED=0
echo -e "\n${BLUE}Service Status:${NC}"
while IFS= read -r line; do
    if echo "$line" | grep -q "test-.*healthy"; then
        service=$(echo "$line" | awk '{print $1}')
        echo -e "  ${GREEN}✓${NC} $service"
    elif echo "$line" | grep -q "test-.*"; then
        service=$(echo "$line" | awk '{print $1}')
        echo -e "  ${RED}✗${NC} $service"
        FAILED=1
    fi
done < <(docker-compose -f $COMPOSE_FILE ps)

# Cleanup
docker-compose -f $COMPOSE_FILE down --remove-orphans 2>/dev/null

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ All services started successfully!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some services failed to start!${NC}"
    exit 1
fi
