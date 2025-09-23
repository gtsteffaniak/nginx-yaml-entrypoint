#!/bin/bash
set -e

# Script to easily test individual examples

if [ $# -eq 0 ]; then
    echo "Usage: $0 <example-name>"
    echo ""
    echo "Available examples:"
    ls examples/*.yaml | sed 's|examples/||' | sed 's|\.yaml||'
    exit 1
fi

EXAMPLE=$1
EXAMPLE_FILE="examples/${EXAMPLE}.yaml"

if [ ! -f "$EXAMPLE_FILE" ]; then
    echo "❌ Example file not found: $EXAMPLE_FILE"
    echo ""
    echo "Available examples:"
    ls examples/*.yaml | sed 's|examples/||' | sed 's|\.yaml||'
    exit 1
fi

echo "🧪 Testing example: $EXAMPLE"
echo "📄 File: $EXAMPLE_FILE"

# Create test directories
mkdir -p test/ssl test/mock-responses

# Generate SSL certs if needed
if [ ! -f test/ssl/cert.pem ]; then
    echo "🔐 Generating test SSL certificates..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout test/ssl/key.pem \
        -out test/ssl/cert.pem \
        -subj "/C=US/ST=Test/L=Test/O=Test/OU=Test/CN=localhost" \
        2>/dev/null
fi

# Create mock response
cat > test/mock-responses/index.html <<EOF
<!DOCTYPE html>
<html>
<head><title>Mock Backend - $EXAMPLE</title></head>
<body style="font-family: Arial, sans-serif; margin: 40px;">
    <h1>✅ Mock Backend Response</h1>
    <p><strong>Example:</strong> $EXAMPLE</p>
    <p><strong>Timestamp:</strong> $(date)</p>
    <p><strong>Status:</strong> nginx-yaml-entrypoint working correctly!</p>
    
    <h2>🔗 Test URLs:</h2>
    <ul>
        <li><a href="http://localhost:8080/">HTTP (port 8080)</a></li>
        <li><a href="https://localhost:8443/">HTTPS (port 8443)</a></li>
        <li><a href="http://localhost:8080/health">Health Check</a></li>
    </ul>
</body>
</html>
EOF

# Update docker-compose to use the specified example
sed "s|./examples/simple-proxy.yaml|./examples/${EXAMPLE}.yaml|" docker-compose.manual-test.yml > /tmp/docker-compose-test.yml

echo "🚀 Starting services..."
docker-compose -f /tmp/docker-compose-test.yml up --build -d

echo ""
echo "⏳ Waiting for services to start..."
sleep 5

# Check if nginx is running
if docker-compose -f /tmp/docker-compose-test.yml ps | grep -q "nginx-test.*Up"; then
    echo "✅ Services started successfully!"
    echo ""
    echo "🌐 Test URLs:"
    echo "   HTTP:  http://localhost:8080"
    echo "   HTTPS: https://localhost:8443"
    echo ""
    echo "📊 View logs:"
    echo "   docker-compose -f /tmp/docker-compose-test.yml logs nginx-test"
    echo ""
    echo "🔧 View generated config:"
    echo "   docker exec nginx-yaml-entrypoint-nginx-test-1 cat /etc/nginx/nginx.conf"
    echo ""
    echo "🛑 Stop services:"
    echo "   docker-compose -f /tmp/docker-compose-test.yml down"
    
    # Test connectivity
    echo "🔍 Testing connectivity..."
    if curl -f -s http://localhost:8080 > /dev/null; then
        echo "✅ HTTP test passed"
    else
        echo "❌ HTTP test failed"
        docker-compose -f /tmp/docker-compose-test.yml logs nginx-test
    fi
    
else
    echo "❌ Services failed to start"
    docker-compose -f /tmp/docker-compose-test.yml logs nginx-test
    docker-compose -f /tmp/docker-compose-test.yml down
    rm -f /tmp/docker-compose-test.yml
    exit 1
fi

# Clean up temp file
rm -f /tmp/docker-compose-test.yml
