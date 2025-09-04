#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DURATION=30  # Run each test for 30 seconds
COMPOSE_FILE="docker-compose.test.yml"

echo -e "${BLUE}🧪 Starting nginx-yaml-entrypoint regression tests${NC}"
echo "========================================================"

# Clean up any existing containers
echo -e "${YELLOW}🧹 Cleaning up existing test containers...${NC}"
docker-compose -f $COMPOSE_FILE down --remove-orphans 2>/dev/null || true

# Create test directories and files
echo -e "${YELLOW}📁 Setting up test environment...${NC}"
mkdir -p test/ssl test/mock-responses

# Create mock SSL certificates for testing
if [ ! -f test/ssl/cert.pem ]; then
    echo -e "${YELLOW}🔐 Generating test SSL certificates...${NC}"
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout test/ssl/key.pem \
        -out test/ssl/cert.pem \
        -subj "/C=US/ST=Test/L=Test/O=Test/OU=Test/CN=localhost" \
        2>/dev/null
fi

# Create mock response files
cat > test/mock-responses/index.html <<EOF
<!DOCTYPE html>
<html>
<head><title>Mock Backend</title></head>
<body>
    <h1>Mock Backend Response</h1>
    <p>This is a test backend service</p>
    <p>Timestamp: $(date)</p>
</body>
</html>
EOF

# Build the nginx image
echo -e "${YELLOW}🏗️  Building nginx image...${NC}"
docker-compose -f $COMPOSE_FILE build --no-cache

# List of services to test
SERVICES=(
    "test-simple-proxy"
    "test-api-gateway" 
    "test-comprehensive"
    "test-load-balancer"
    "test-ssl-website"
    "test-websocket-app"
    "test-template-system"
    "test-simple-templates"
)

# Track test results
declare -a PASSED_TESTS
declare -a FAILED_TESTS

# Function to test a service
test_service() {
    local service=$1
    local config_file="./examples/$(echo $service | sed 's/test-//' | sed 's/-/_/g').yaml"
    
    echo -e "${BLUE}🔍 Testing: $service${NC}"
    echo "   Config: $config_file"
    
    # Start the service and dependencies
    echo "   Starting service..."
    if docker-compose -f $COMPOSE_FILE up -d $service mock-backend-1 mock-backend-2 mock-backend-3 2>/dev/null; then
        echo "   ✓ Service started"
        
        # Wait for health check to pass
        echo "   Waiting for health check..."
        local attempts=0
        local max_attempts=12  # 60 seconds total
        
        while [ $attempts -lt $max_attempts ]; do
            if docker-compose -f $COMPOSE_FILE ps | grep -q "$service.*healthy"; then
                echo "   ✓ Health check passed"
                break
            elif docker-compose -f $COMPOSE_FILE ps | grep -q "$service.*unhealthy"; then
                echo "   ✗ Health check failed"
                docker-compose -f $COMPOSE_FILE logs $service
                return 1
            fi
            
            attempts=$((attempts + 1))
            sleep 5
        done
        
        if [ $attempts -eq $max_attempts ]; then
            echo "   ✗ Health check timeout"
            docker-compose -f $COMPOSE_FILE logs $service
            return 1
        fi
        
        # Run for test duration
        echo "   Running for ${TEST_DURATION}s..."
        sleep $TEST_DURATION
        
        # Check if still running and healthy
        if docker-compose -f $COMPOSE_FILE ps | grep -q "$service.*healthy"; then
            echo -e "   ${GREEN}✓ Test passed${NC}"
            return 0
        else
            echo -e "   ${RED}✗ Service became unhealthy${NC}"
            docker-compose -f $COMPOSE_FILE logs $service
            return 1
        fi
        
    else
        echo -e "   ${RED}✗ Failed to start service${NC}"
        return 1
    fi
}

# Run tests
echo -e "\n${BLUE}🚀 Running tests...${NC}"
echo "========================================================"

for service in "${SERVICES[@]}"; do
    if test_service $service; then
        PASSED_TESTS+=($service)
    else
        FAILED_TESTS+=($service)
    fi
    
    # Clean up after each test
    echo "   Stopping service..."
    docker-compose -f $COMPOSE_FILE stop $service 2>/dev/null || true
    docker-compose -f $COMPOSE_FILE rm -f $service 2>/dev/null || true
    echo ""
done

# Clean up
echo -e "${YELLOW}🧹 Cleaning up...${NC}"
docker-compose -f $COMPOSE_FILE down --remove-orphans 2>/dev/null || true

# Print results
echo "========================================================"
echo -e "${BLUE}📊 Test Results${NC}"
echo "========================================================"

echo -e "${GREEN}✓ Passed (${#PASSED_TESTS[@]}):${NC}"
for test in "${PASSED_TESTS[@]}"; do
    echo "  - $test"
done

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "\n${RED}✗ Failed (${#FAILED_TESTS[@]}):${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo ""
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
fi
