# nginx-yaml-entrypoint

Convert YAML configuration files to nginx configuration - because nginx config files are cumbersome and prone to errors!

## Overview

This tool allows you to write nginx configurations in YAML format with templates, reusable components, and comprehensive validation, then automatically converts them to nginx `.conf` files. Perfect for Docker containers and configuration management.

## Features

### ✅ Comprehensive nginx Support
- **Global Settings**: Worker processes, connections, timeouts, body size limits
- **SSL/TLS**: Modern SSL configuration with security best practices
- **Load Balancing**: Multiple upstream algorithms (round-robin, least_conn, ip_hash, etc.)
- **Caching**: Proxy caching with configurable zones and invalidation
- **Rate Limiting**: Advanced rate limiting zones with burst support
- **Security Headers**: HSTS, CSP, CORS, X-Frame-Options, etc.
- **Compression**: Gzip configuration with type-specific settings
- **WebSocket Support**: Proper WebSocket proxy configuration
- **Static File Serving**: Optimized static file delivery
- **Authentication**: auth_request module support
- **Error Pages**: Custom error page configuration
- **URL Rewrites**: Flexible rewrite rules with flags
- **Access Control**: IP-based allow/deny rules
- **Logging**: Configurable access and error logs

### 🎨 Template System
- **Built-in Templates**: 14 pre-configured templates for common patterns
- **Custom Templates**: Define your own reusable configurations
- **Template Inheritance**: Apply multiple templates to servers/locations
- **Template Override**: User templates override built-in ones

### 🛡️ Validation & Safety
- **YAML Validation**: Comprehensive configuration validation
- **nginx Syntax Checking**: Validates nginx directive syntax
- **Dependency Checking**: Ensures referenced upstreams/zones exist
- **Error Reporting**: Clear, actionable error messages

### 🚀 Developer Experience
- **Dry Run Mode**: Preview generated config without writing files
- **Verbose Output**: Detailed logging for debugging
- **Template Listing**: Explore available built-in templates
- **CLI Interface**: Rich command-line options
- **Docker Ready**: Designed for container environments

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/nginx-yaml-entrypoint
cd nginx-yaml-entrypoint

# Build the binary
go build -o nginx-yaml-entrypoint .
```

### Basic Usage

```bash
# Convert YAML to nginx config
./nginx-yaml-entrypoint -f config.yaml -o /etc/nginx/nginx.conf

# Validate configuration only
./nginx-yaml-entrypoint -validate -f config.yaml

# Dry run (output to stdout)
./nginx-yaml-entrypoint -dry-run -f config.yaml

# List available templates
./nginx-yaml-entrypoint -list-templates
```

### Docker Usage

#### Ready-to-Use Docker Image

For immediate use without building, use the pre-built image:

```bash
# Pull the image with nginx-yaml-entrypoint built-in
docker pull gtstef/nginx

# Run with your YAML configuration
docker run -d \
  -p 80:80 -p 443:443 \
  -v $(pwd)/config.yaml:/etc/nginx/yaml/main.yaml \
  gtstef/nginx
```

The `gtstef/nginx` image includes:
- Latest nginx with alpine base
- `nginx-yaml-entrypoint` tool pre-installed
- Automatic YAML-to-nginx conversion on startup
- All features ready to use out of the box

#### Building Your Own Image

```dockerfile
FROM nginx:alpine
COPY --from=builder /app/nginx-yaml-entrypoint /usr/local/bin/
COPY entrypoint.sh /docker-entrypoint.d/
```

## Configuration Examples

### Simple Reverse Proxy

```yaml
servers:
  - listen: "80"
    server_name: "example.com"
    locations:
      - path: "/"
        proxy_pass: "http://backend:3000"
        applyTemplates:
          - reverse_proxy
```

### SSL-enabled Site with Caching

```yaml
# Global rate limiting
limit_req_zones:
  - name: global
    key: $binary_remote_addr
    zone: global:10m
    rate: 10r/s

# Caching configuration
proxy_cache_path:
  - name: app_cache
    path: /var/cache/nginx
    keys_zone: "app:100m"
    max_size: 1g
    inactive: 60m

servers:
  - listen: "443 ssl http2"
    server_name: "example.com"

    ssl:
      certificate: "/etc/ssl/cert.pem"
      certificate_key: "/etc/ssl/key.pem"

    applyTemplates:
      - modern_ssl
      - security_headers

    locations:
      - path: "/"
        proxy_pass: "http://app:3000"
        applyTemplates:
          - reverse_proxy
          - cache_proxy
        limit_req:
          zone: global
          burst: 20
```

### Multi-service Setup with Load Balancing

```yaml
upstreams:
  - name: api_backend
    method: least_conn
    servers:
      - address: "api1:8080"
        weight: 2
      - address: "api2:8080"
        weight: 1
      - address: "api3:8080"
        backup: true

servers:
  - listen: "443 ssl"
    server_name: "api.example.com"

    ssl:
      certificate: "/etc/ssl/api-cert.pem"
      certificate_key: "/etc/ssl/api-key.pem"

    locations:
      - path: "/api/v1"
        proxy_pass: "http://api_backend"
        applyTemplates:
          - reverse_proxy
          - cors
          - rate_limited

      - path: "/health"
        applyTemplates:
          - health_check
```

## Built-in Templates

Run `nginx-yaml-entrypoint -list-templates` to see all available templates:

- `modern_ssl` - Modern SSL/TLS with strong security
- `websocket` - WebSocket proxy support
- `reverse_proxy` - Standard reverse proxy
- `static_files` - Optimized static file serving
- `security_headers` - Common security headers
- `cors` - CORS headers for APIs
- `cache_proxy` - Proxy caching for performance
- `auth_required` - Authentication support
- `rate_limited` - Rate limiting protection
- And more...

## Advanced Features

### Custom Templates

Define reusable configurations:

```yaml
templateConfigs:
  my_api_template:
    proxy_set_header:
      X-API-Key: "$http_x_api_key"
      Authorization: "$http_authorization"
    configs:
      proxy_timeout: 30s
    limit_req:
      zone: api
      burst: 10

servers:
  - listen: "80"
    server_name: "api.local"
    locations:
      - path: "/v1"
        proxy_pass: "http://backend"
        applyTemplates:
          - my_api_template
          - cors
```

### Conditional Logic

```yaml
servers:
  - listen: "80"
    server_name: "example.com"
    locations:
      - path: "/"
        proxy_pass: "http://backend"
        conditions:
          - if: "$request_method = POST"
            then:
              - "limit_req zone=strict burst=1"
              - "access_log /var/log/nginx/post.log"
```

### Environment-specific Configuration

```yaml
servers:
  - listen: "80"
    server_name: "app.local"
    custom_config:
      - "add_header X-Environment development always"
      - "error_log /var/log/nginx/dev.log debug"
    locations:
      - path: "/"
        proxy_pass: "http://dev-backend"
```

## Configuration Reference

See [`examples/comprehensive.yaml`](examples/comprehensive.yaml) for a complete configuration example showcasing all features.

### Global Configuration

```yaml
# Worker settings
worker_processes: auto
worker_connections: 1024

# Timeouts and limits
keepalive_timeout: 65s
client_max_body_size: 100m
server_tokens: "off"

# Logging
access_log:
  path: /var/log/nginx/access.log
  format: main
error_log:
  path: /var/log/nginx/error.log
  level: warn

# Compression
gzip:
  enabled: true
  comp_level: 6
  types: ["text/plain", "text/css", "application/json"]

# Security
security_headers:
  hsts:
    max_age: 31536000
    include_subdomains: true
  csp: "default-src 'self'"
```

### Server Configuration

```yaml
servers:
  - listen: "443 ssl http2"
    server_name: "example.com"

    # SSL configuration
    ssl:
      certificate: "/path/to/cert.pem"
      certificate_key: "/path/to/key.pem"
      protocols: ["TLSv1.2", "TLSv1.3"]

    # HTTP settings
    http2: true

    # Error pages
    error_page:
      "404": "/404.html"
      "500 502 503 504": "/50x.html"

    # Templates
    applyTemplates:
      - modern_ssl
      - security_headers

    # Custom config
    configs:
      server_name_in_redirect: "off"

    # Locations
    locations:
      - path: "/"
        proxy_pass: "http://backend"
        applyTemplates:
          - reverse_proxy
```

## CLI Options

```
Usage: nginx-yaml-entrypoint [options]

Options:
  -f, -file string          Path to YAML configuration file (default "nginx_config.yaml")
  -o, -output string        Output path for nginx configuration (default "/etc/nginx/conf.d/default.conf")
  -d, -dry-run              Dry run - output to stdout instead of file
  -v, -validate             Validate configuration only
  -verbose                  Verbose output
  -list-templates           List all available built-in templates
```

## Validation

The tool performs comprehensive validation:

- **YAML Syntax**: Valid YAML structure
- **Required Fields**: Essential nginx directives
- **Value Validation**: Port ranges, time formats, size values
- **Reference Checking**: Upstream/zone references exist
- **Syntax Checking**: nginx directive syntax
- **Duplicate Detection**: Prevent conflicting configurations

Example validation output:
```
Configuration validation failed with 2 errors:
  - validation error in field 'servers[0].listen' (value: '99999'): invalid listen format
  - validation error in field 'servers[0].locations[0].proxy_pass' (value: 'invalid-url'): invalid proxy_pass URL format
```

## Docker Integration

### Pre-built Image

The easiest way to use nginx-yaml-entrypoint is with the ready-to-use Docker image:

```bash
docker run -d \
  --name my-nginx \
  -p 80:80 -p 443:443 \
  -v $(pwd)/nginx.yaml:/etc/nginx/yaml/main.yaml \
  -v $(pwd)/ssl:/etc/ssl \
  gtstef/nginx
```

### Custom Dockerfile

```dockerfile
FROM golang:1.22-alpine as builder
WORKDIR /app
COPY *.go go.* ./
RUN go build -ldflags='-w -s' -o nginx-yaml-entrypoint .

FROM nginx:alpine
COPY --from=builder /app/nginx-yaml-entrypoint /usr/local/bin/
COPY entrypoint.sh /docker-entrypoint.d/99-yaml-config.sh
RUN chmod +x /docker-entrypoint.d/99-yaml-config.sh
```

### Entrypoint Script

```bash
#!/bin/sh
set -e

if test -f /etc/nginx/yaml/main.yaml; then
    echo "Converting YAML configuration to nginx config..."
    nginx-yaml-entrypoint -f /etc/nginx/yaml/main.yaml -o /etc/nginx/conf.d/default.conf
    echo "Configuration generated successfully"
fi
```

### Usage

Using the pre-built image:
```bash
docker run -v $(pwd)/config.yaml:/etc/nginx/yaml/main.yaml gtstef/nginx
```

Using your custom image:
```bash
docker run -v $(pwd)/config.yaml:/etc/nginx/yaml/main.yaml my-nginx
```

### Working with Multiple YAML Files

The tool currently processes a single YAML file, but you can work with multiple configuration files using several approaches:

#### Option 1: External Merge with yq (Recommended)

Use `yq` to merge multiple YAML files before running the container:

```bash
# Install yq if not available
# macOS: brew install yq
# Ubuntu: sudo apt install yq

# Merge multiple YAML files
yq eval-all '. as $item ireduce ({}; . *+ $item)' \
  configs/global.yaml \
  configs/upstreams.yaml \
  configs/servers.yaml > merged.yaml

# Run with merged configuration
docker run -d \
  -p 80:80 -p 443:443 \
  -v $(pwd)/merged.yaml:/etc/nginx/yaml/main.yaml \
  gtstef/nginx
```

**Example file structure:**
```
configs/
├── global.yaml      # Global settings, gzip, security
├── upstreams.yaml   # Backend server definitions  
├── servers.yaml     # Virtual hosts and locations
└── ssl.yaml         # SSL certificates and config
```

**configs/global.yaml:**
```yaml
worker_processes: auto
worker_connections: 1024
gzip:
  enabled: true
  comp_level: 6
  types: ["text/plain", "text/css", "application/json"]
security_headers:
  hsts:
    max_age: 31536000
```

**configs/upstreams.yaml:**
```yaml
upstreams:
  - name: api_backend
    method: least_conn
    servers:
      - address: "api1:8080"
      - address: "api2:8080"
```

**configs/servers.yaml:**
```yaml
servers:
  - listen: "443 ssl http2"
    server_name: "example.com"
    locations:
      - path: "/api"
        proxy_pass: "http://api_backend"
        applyTemplates:
          - reverse_proxy
```

#### Option 2: Docker Compose with Custom Merge Script

Create a custom entrypoint that automatically merges multiple files:

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  nginx:
    image: gtstef/nginx
    ports:
      - "80:80"
      - "443:443"  
    volumes:
      - ./configs:/etc/nginx/yaml:ro
      - ./scripts/merge-entrypoint.sh:/docker-entrypoint.d/98-merge-yaml.sh:ro
      - ./ssl:/etc/ssl:ro
```

**scripts/merge-entrypoint.sh:**
```bash
#!/bin/sh
set -e

# Install yq for YAML merging
if ! command -v yq > /dev/null; then
    apk add --no-cache yq
fi

# Check if multiple YAML files exist
YAML_COUNT=$(find /etc/nginx/yaml -name "*.yaml" -o -name "*.yml" | wc -l)

if [ "$YAML_COUNT" -gt 1 ]; then
    echo "Found $YAML_COUNT YAML files, merging..."
    
    # Merge all YAML files in alphabetical order
    yq eval-all '. as $item ireduce ({}; . *+ $item)' \
        /etc/nginx/yaml/*.yaml /etc/nginx/yaml/*.yml 2>/dev/null > /tmp/merged.yaml || \
        yq eval-all '. as $item ireduce ({}; . *+ $item)' \
        /etc/nginx/yaml/*.yaml > /tmp/merged.yaml 2>/dev/null || \
        yq eval-all '. as $item ireduce ({}; . *+ $item)' \
        /etc/nginx/yaml/*.yml > /tmp/merged.yaml 2>/dev/null
    
    # Process merged configuration
    nginx-yaml-entrypoint -f /tmp/merged.yaml -o /etc/nginx/conf.d/default.conf
    echo "Successfully processed merged YAML configuration"
    
elif [ -f /etc/nginx/yaml/main.yaml ]; then
    # Fallback to single file processing
    echo "Processing single YAML file..."
    nginx-yaml-entrypoint -f /etc/nginx/yaml/main.yaml -o /etc/nginx/conf.d/default.conf
    echo "Successfully processed YAML configuration"
else
    echo "No YAML configuration files found in /etc/nginx/yaml/"
    exit 1
fi
```

Make the script executable:
```bash
chmod +x scripts/merge-entrypoint.sh
```

#### Option 3: Organized File Structure

Structure your configuration files by concern:

```
project/
├── docker-compose.yml
├── configs/
│   ├── 01-global.yaml       # Global settings (processed first)
│   ├── 02-upstreams.yaml    # Backend definitions  
│   ├── 03-ssl.yaml          # SSL configuration
│   ├── 04-servers.yaml      # Virtual hosts
│   └── 99-custom.yaml       # Custom overrides (processed last)
├── scripts/
│   └── merge-entrypoint.sh
└── ssl/
    ├── cert.pem
    └── key.pem
```

**Benefits of Multiple Files:**
- **Maintainability**: Separate concerns (SSL, upstreams, servers)  
- **Reusability**: Share common configurations across environments
- **Team Collaboration**: Different team members can work on different files
- **Environment Specific**: Override settings per environment

**File Naming Convention:**
- Use numeric prefixes to control merge order: `01-global.yaml`, `02-upstreams.yaml`
- Group by function: `global.yaml`, `upstreams.yaml`, `servers.yaml`
- Environment suffixes: `servers-prod.yaml`, `servers-dev.yaml`

**Merge Behavior:**
- Later files override earlier files for conflicting keys
- Arrays are replaced, not merged (design decision for predictability)
- Maps are merged recursively

**Complete Example:**
```bash
# Merge and run
yq eval-all '. as $item ireduce ({}; . *+ $item)' \
  configs/01-global.yaml \
  configs/02-upstreams.yaml \
  configs/03-servers.yaml > merged.yaml

docker run -d --name nginx \
  -p 80:80 -p 443:443 \
  -v $(pwd)/merged.yaml:/etc/nginx/yaml/main.yaml \
  -v $(pwd)/ssl:/etc/ssl \
  gtstef/nginx

# Verify configuration
docker logs nginx
```

## Testing

### Regression Tests

The project includes a comprehensive test suite that validates all example configurations against the actual nginx image to ensure they don't have syntax errors and start successfully.

#### Running All Tests

```bash
# Full test suite - runs each example for 30 seconds
./test/run-tests.sh
```

**What it tests:**
- ✅ All example YAML files parse correctly
- ✅ Generated nginx configs have valid syntax  
- ✅ Nginx starts successfully with each configuration
- ✅ Services remain healthy during test duration
- ✅ No error logs are generated

#### Quick Tests (CI/CD)

```bash
# Fast parallel tests - 10 second validation
./test/quick-test.sh
```

#### Test Output

```
🧪 Starting nginx-yaml-entrypoint regression tests
========================================================
🧹 Cleaning up existing test containers...
📁 Setting up test environment...
🔐 Generating test SSL certificates...
🏗️  Building nginx image...

🚀 Running tests...
========================================================
🔍 Testing: test-simple-proxy
   Config: ./examples/simple_proxy.yaml
   Starting service...
   ✓ Service started
   Waiting for health check...
   ✓ Health check passed
   Running for 30s...
   ✓ Test passed

📊 Test Results
========================================================
✓ Passed (8):
  - test-simple-proxy
  - test-api-gateway
  - test-comprehensive
  - test-load-balancer
  - test-ssl-website
  - test-websocket-app
  - test-template-system
  - test-simple-templates

🎉 All tests passed!
```

#### Test Architecture

The test suite uses Docker Compose to:

1. **Build** the nginx image with your current code
2. **Start** each example in a separate container
3. **Health check** using curl to ensure nginx responds
4. **Run** for a specified duration to catch delayed failures
5. **Collect logs** if any service fails
6. **Clean up** automatically after tests

**Test Infrastructure:**
```
test/
├── run-tests.sh           # Full regression test suite  
├── quick-test.sh          # Fast CI/CD tests
├── ssl/                   # Generated test certificates
└── mock-responses/        # Mock backend responses

docker-compose.test.yml    # Test service definitions
.github/workflows/test.yaml # CI/CD integration
```

#### Individual Example Testing

Test a specific configuration manually:

```bash
# Test one example
docker-compose -f docker-compose.test.yml up test-simple-proxy

# View logs
docker-compose -f docker-compose.test.yml logs test-simple-proxy

# Check configuration
docker exec test-simple-proxy nginx -t

# View generated config  
docker exec test-simple-proxy cat /etc/nginx/nginx.conf
```

#### Adding New Tests

When adding new examples:

1. **Add the YAML file** to `examples/`
2. **Add test service** to `docker-compose.test.yml`:
```yaml
test-my-example:
  build: .
  container_name: test-my-example
  volumes:
    - ./examples/my-example.yaml:/etc/nginx/yaml/main.yaml:ro
  ports:
    - "8009:80"
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:80/", "||", "exit 1"]
    interval: 5s
    timeout: 3s
    retries: 3
    start_period: 10s
  networks:
    - test-network
```
3. **Update test script** to include the new service

#### Continuous Integration

Tests run automatically on:
- ✅ Pull requests
- ✅ Pushes to main/develop branches
- ✅ Manual workflow dispatch

**GitHub Actions Integration:**
```yaml
- name: Run regression tests
  run: ./test/run-tests.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. **Run the test suite**: `./test/run-tests.sh`
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details.