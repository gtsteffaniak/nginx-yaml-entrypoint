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

### Dockerfile

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

```bash
docker run -v $(pwd)/config.yaml:/etc/nginx/yaml/main.yaml my-nginx
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.