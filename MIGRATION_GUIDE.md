# yourdomain nginx Migration Guide

This guide shows how to migrate from individual `.conf` files to the unified YAML configuration system.

## Before & After Comparison

### Before: 5 Configuration Files (1,043 lines total)

```
conf.d/
├── default.conf     (76 lines)  - Global settings & redirects
├── main.conf        (199 lines) - Main service
├── photos.conf      (134 lines) - Photos service
├── files.conf       (107 lines) - Files service
└── media.conf      (28 lines) - Media service
```

**Issues with the old approach:**
- ❌ **Duplication**: SSL config repeated in every file
- ❌ **Maintenance**: Changes needed in multiple files
- ❌ **Inconsistency**: Auth logic slightly different everywhere
- ❌ **No validation**: Syntax errors only caught at runtime
- ❌ **Complex setup**: Hard to understand the full picture

### After: 1 YAML Configuration (325 lines)

```
myservice-config.yaml  (325 lines) - Everything unified with templates
```

**Benefits of the new approach:**
- ✅ **DRY Principle**: Templates eliminate all duplication
- ✅ **Single source of truth**: All config in one place
- ✅ **Validation**: Catch errors before deployment
- ✅ **Maintainable**: Easy to modify and understand
- ✅ **Consistent**: Same auth/SSL logic everywhere via templates

## Template Mapping

Here's how the old configurations map to templates:

### SSL Configuration
**Before (repeated in every file):**
```nginx
ssl_certificate /etc/nginx/ssl/live/yourdomain.com/fullchain.pem;
ssl_certificate_key /etc/nginx/ssl/live/yourdomain.com/privkey.pem;
```

**After (template applied everywhere):**
```yaml
templateConfigs:
  yourdomainSSL:
    configs:
      ssl_certificate: /etc/nginx/ssl/live/yourdomain.com/fullchain.pem
      ssl_certificate_key: /etc/nginx/ssl/live/yourdomain.com/privkey.pem

servers:
  - server_name: "yourdomain.com"
    applyTemplates:
      - yourdomainSSL
```

### Authentication Logic
**Before (repeated with slight variations):**
```nginx
auth_request_set $user $upstream_http_x_forwarded_user;
proxy_set_header X-Forwarded-User $user;
auth_request /auth/authorize;
auth_request_set $auth_status $upstream_status;
proxy_set_header Authorization $http_authorization;
proxy_pass_header Authorization;
add_header Access-Control-Allow-Origin files.yourdomain.com;
```

**After (template with consistent logic):**
```yaml
templateConfigs:
  auth:
    configs:
      auth_request: "/auth/authorize"
    auth_request_set:
      "$user": "$upstream_http_x_forwarded_user"
      "$auth_status": "$upstream_status"
    proxy_set_header:
      X-Forwarded-User: "$user"
      Authorization: "$http_authorization"
    proxy_pass_header:
      - Authorization
    add_header:
      Access-Control-Allow-Origin: "files.yourdomain.com"

# Applied to any location that needs auth:
locations:
  - path: "/"
    applyTemplates:
      - auth
```

### WebSocket Support
**Before (repeated in multiple locations):**
```nginx
proxy_http_version 1.1;
proxy_set_header Connection $http_connection;
proxy_buffering off;
proxy_set_header Upgrade $http_upgrade;
```

**After (template for consistency):**
```yaml
templateConfigs:
  webSocketSupport:
    configs:
      proxy_http_version: "1.1"
    proxy_set_header:
      Connection: "$http_connection"
      Upgrade: "$http_upgrade"

# Applied wherever WebSockets are needed:
locations:
  - path: "/shinobi/"
    applyTemplates:
      - webSocketSupport
      - auth
```

## Migration Steps

### Step 1: Copy nginx-yaml-entrypoint

```bash
cd /path/to/ingress-proxy
cp -r /path/to/nginx-yaml-entrypoint ./
```

### Step 2: Test the YAML Configuration

```bash
# Build the tool
cd nginx-yaml-entrypoint
go build -o ../nginx-yaml-entrypoint .
cd ..

# Validate the configuration
./nginx-yaml-entrypoint -f yourdomain-config.yaml --validate

# Generate nginx config preview
./nginx-yaml-entrypoint -f yourdomain-config.yaml --dry-run > preview.conf

# Compare with current config
diff /etc/nginx/nginx.conf preview.conf
```

### Step 3: Backup Current Configuration

```bash
# Backup current nginx configuration
cp -r /etc/nginx /etc/nginx.backup.$(date +%Y%m%d)

# Backup current conf.d files
tar -czf nginx-conf-backup-$(date +%Y%m%d).tar.gz conf.d/
```

### Step 4: Deploy with Docker

```bash
# Build and test
docker build -t yourdomain-nginx-test .

# Test run (without exposing ports)
docker run --rm --name nginx-test -d yourdomain-nginx-test

# Check logs
docker logs nginx-test

# If successful, deploy with docker-compose
docker-compose up -d
```

### Step 5: Verify Deployment

```bash
# Check nginx status
curl -s http://localhost/nginx_status

# Test main endpoints
curl -I https://yourdomain.com
curl -I https://photos.yourdomain.com
curl -I https://files.yourdomain.com
curl -I https://media.yourdomain.com

# Check logs
docker logs yourdomain-nginx
```

## Configuration Changes

### Rate Limiting Improvements

**Before:** Different files had different rate limiting setup
```nginx
limit_req zone=global burst=20;
limit_req zone=unlimited burst=1000;  # Different values everywhere
```

**After:** Consistent rate limiting with clear zones
```yaml
limit_req_zones:
  - name: global
    rate: 100r/s
  - name: unlimited
    rate: 10000r/s
  - name: medium
    rate: 25r/s
  - name: strict
    rate: 5r/s
```

### Upstream Management

**Before:** Hard-coded IP addresses scattered throughout
```nginx
proxy_pass http://10.13.13.3:9283;
proxy_pass http://home:8080;
```

**After:** Centralized upstream definitions
```yaml
upstreams:
  - name: photos_service
    servers:
      - address: "10.13.13.3:9283"
  - name: home_service
    servers:
      - address: "home:8080"

# Used as:
proxy_pass: "http://photos_service"
```

### Error Handling

**Before:** Error pages defined in every server block
```nginx
error_page 404 = @notfound;
error_page 401 = @error401;
# ... repeated everywhere
```

**After:** Template for consistent error handling
```yaml
templateConfigs:
  yourdomainErrorPages:
    error_page:
      "404": "@notfound"
      "401": "@error401"
      "429": "@error429"
      "502": "@error502"
```

## Customization Examples

### Adding a New Service

**Before:** Create a new `.conf` file, copy SSL config, auth logic, etc.

**After:** Just add to the YAML:
```yaml
# Add upstream
upstreams:
  - name: new_service
    servers:
      - address: "new-service:8080"

# Add location to main server
servers:
  - server_name: "yourdomain.com"
    locations:
      - path: "/new-service/"
        proxy_pass: "http://new_service/"
        applyTemplates:
          - yourdomainAuth  # Instant auth integration
```

### Modifying SSL Settings

**Before:** Update 4 different files

**After:** Update one template:
```yaml
templateConfigs:
  yourdomainSSL:
    configs:
      ssl_certificate: /new/path/cert.pem
      ssl_certificate_key: /new/path/key.pem
      ssl_protocols: "TLSv1.2 TLSv1.3"  # Add protocols
```

### Custom Authentication for a Service

**Before:** Copy auth logic and modify in place

**After:** Create a new template:
```yaml
templateConfigs:
  customAuth:
    configs:
      auth_request: "/custom/auth"
    proxy_set_header:
      X-Custom-Header: "value"

# Apply to specific service
locations:
  - path: "/special-service/"
    applyTemplates:
      - customAuth
```
