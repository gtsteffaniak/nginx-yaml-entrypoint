# nginx-yaml-entrypoint Feature Matrix

This document provides a comprehensive overview of what nginx features are supported directly vs what requires custom configuration.

## Legend
- ✅ **Fully Supported**: Native YAML configuration with validation
- 🔧 **Custom Config**: Supported via `configs` or `custom_config` fields
- ⚠️ **Limited**: Partial support or requires workarounds
- ❌ **Not Supported**: Not implemented

## Core nginx Features

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| **Core Module** | | | |
| worker_processes | ✅ | `worker_processes` | auto, number, or formula |
| worker_connections | ✅ | `worker_connections` | Integer value |
| keepalive_timeout | ✅ | `keepalive_timeout` | Time format (e.g., "65s") |
| client_max_body_size | ✅ | `client_max_body_size` | Size format (e.g., "100m") |
| server_tokens | ✅ | `server_tokens` | on/off/build |
| sendfile | 🔧 | `custom_global_config` | Add as custom directive |
| tcp_nopush | 🔧 | `custom_global_config` | Add as custom directive |
| tcp_nodelay | 🔧 | `custom_global_config` | Add as custom directive |

## HTTP Module

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| **Basic HTTP** | | | |
| listen | ✅ | `servers[].listen` | Full syntax support with SSL, HTTP/2 |
| server_name | ✅ | `servers[].server_name` | Multiple names supported |
| return | ✅ | `servers[].return` | Redirects and status codes |
| rewrite | ✅ | `servers[].rewrite` | Array of rewrite rules |
| error_page | ✅ | `servers[].error_page` | Map of codes to pages |
| try_files | ✅ | `locations[].try_files` | Standard try_files syntax |
| root | ✅ | `locations[].root` | Document root |
| index | ✅ | `locations[].index` | Array of index files |
| **Headers** | | | |
| add_header | ✅ | `add_header` | Map of header: value |
| more_set_headers | ✅ | `more_set_headers` | Requires headers-more module |
| proxy_set_header | ✅ | `proxy_set_header` | Map of header: value |
| proxy_pass_header | ✅ | `proxy_pass_header` | Array of headers |
| proxy_hide_header | ✅ | `proxy_hide_header` | Array of headers |

## SSL/TLS Module

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| ssl_certificate | ✅ | `ssl.certificate` | File path |
| ssl_certificate_key | ✅ | `ssl.certificate_key` | File path |
| ssl_protocols | ✅ | `ssl.protocols` | Array of protocols |
| ssl_ciphers | ✅ | `ssl.ciphers` | Cipher suite string |
| ssl_session_cache | ✅ | `ssl.session_cache` | Cache configuration |
| ssl_session_timeout | ✅ | `ssl.session_timeout` | Time format |
| ssl_session_tickets | ✅ | `ssl.session_tickets` | Boolean |
| ssl_stapling | ✅ | `ssl.stapling` | Boolean |
| ssl_stapling_verify | ✅ | `ssl.stapling_verify` | Boolean |
| ssl_trusted_certificate | ✅ | `ssl.trusted_certificate` | File path |
| ssl_dhparam | ✅ | `ssl.dhparam` | DH parameters file |
| ssl_prefer_server_ciphers | 🔧 | `configs` | Add as custom directive |
| ssl_ecdh_curve | 🔧 | `configs` | Add as custom directive |

## Proxy Module

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| **Basic Proxy** | | | |
| proxy_pass | ✅ | `proxy_pass` | URL or upstream name |
| proxy_set_header | ✅ | `proxy_set_header` | Map of headers |
| proxy_redirect | 🔧 | `configs` | Add as custom directive |
| **Buffering** | | | |
| proxy_buffering | ✅ | `proxy_buffering` | Boolean pointer |
| proxy_buffers | ✅ | `proxy_buffers` | Buffer configuration |
| proxy_buffer_size | ✅ | `proxy_buffer_size` | Size format |
| proxy_busy_buffers_size | ✅ | `proxy_busy_buffers_size` | Size format |
| **Timeouts** | | | |
| proxy_connect_timeout | ✅ | `proxy_connect_timeout` | Time format |
| proxy_send_timeout | ✅ | `proxy_send_timeout` | Time format |
| proxy_read_timeout | ✅ | `proxy_read_timeout` | Time format |
| **Caching** | | | |
| proxy_cache_path | ✅ | `proxy_cache_path[]` | Full cache configuration |
| proxy_cache | ✅ | `proxy_cache` | Cache zone name |
| proxy_cache_valid | ✅ | `proxy_cache_valid` | Map of status: time |
| proxy_cache_key | ✅ | `proxy_cache_key` | Cache key pattern |
| proxy_cache_bypass | ✅ | `proxy_cache_bypass` | Array of conditions |
| proxy_cache_lock | ✅ | `proxy_cache_lock` | Boolean |
| proxy_cache_use_stale | 🔧 | `configs` | Add as custom directive |

## Upstream Module

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| upstream | ✅ | `upstreams[]` | Full upstream blocks |
| server | ✅ | `upstreams[].servers[]` | Upstream servers |
| weight | ✅ | `servers[].weight` | Server weight |
| max_fails | ✅ | `servers[].max_fails` | Failure threshold |
| fail_timeout | ✅ | `servers[].fail_timeout` | Failure timeout |
| backup | ✅ | `servers[].backup` | Backup server flag |
| down | ✅ | `servers[].down` | Disabled server flag |
| **Load Balancing** | | | |
| round_robin | ✅ | `upstreams[].method` | Default method |
| least_conn | ✅ | `upstreams[].method` | Connection-based |
| ip_hash | ✅ | `upstreams[].method` | IP-based |
| hash | ✅ | `upstreams[].method` | Custom hash |
| random | ✅ | `upstreams[].method` | Random selection |
| least_time | ✅ | `upstreams[].method` | nginx Plus only |
| **Keepalive** | | | |
| keepalive | ✅ | `upstreams[].keepalive` | Connection count |
| keepalive_requests | ✅ | `upstreams[].keepalive_requests` | Request limit |
| keepalive_timeout | ✅ | `upstreams[].keepalive_timeout` | Timeout value |

## Rate Limiting

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| limit_req_zone | ✅ | `limit_req_zones[]` | Rate limiting zones |
| limit_req | ✅ | `limit_req` | Apply rate limiting |
| limit_req_log_level | ✅ | `limit_req_log_level` | Global log level |
| limit_req_status | ✅ | `limit_req_status` | Global status code |
| limit_conn_zone | 🔧 | `custom_global_config` | Connection limiting |
| limit_conn | 🔧 | `configs` | Apply conn limiting |
| burst | ✅ | `limit_req.burst` | Burst allowance |
| nodelay | 🔧 | `configs` | Add to limit_req directive |

## Compression

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| gzip | ✅ | `gzip.enabled` | Enable/disable |
| gzip_vary | ✅ | `gzip.vary` | Vary header |
| gzip_proxied | ✅ | `gzip.proxied` | Proxy condition |
| gzip_comp_level | ✅ | `gzip.comp_level` | Compression level (1-9) |
| gzip_min_length | ✅ | `gzip.min_length` | Minimum size |
| gzip_types | ✅ | `gzip.types` | Array of MIME types |
| gzip_disable | ✅ | `gzip.disable` | User agent patterns |
| gzip_buffers | 🔧 | `custom_global_config` | Buffer configuration |
| gzip_http_version | 🔧 | `custom_global_config` | HTTP version |

## Security

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| **Security Headers** | | | |
| HSTS | ✅ | `security_headers.hsts` | Full HSTS configuration |
| CSP | ✅ | `security_headers.csp` | Content Security Policy |
| X-Frame-Options | ✅ | `security_headers.x_frame_options` | Clickjacking protection |
| X-Content-Type-Options | ✅ | `security_headers.x_content_type_options` | MIME sniffing |
| Referrer-Policy | ✅ | `security_headers.referrer_policy` | Referrer control |
| Permissions-Policy | ✅ | `security_headers.permissions_policy` | Feature policy |
| **CORS** | | | |
| CORS Headers | ✅ | `security_headers.cors` | Full CORS support |
| Preflight Handling | ✅ | Built into CORS template | OPTIONS requests |
| **Access Control** | | | |
| allow | ✅ | `allow[]` | Array of IP/CIDR |
| deny | ✅ | `deny[]` | Array of IP/CIDR |
| auth_basic | 🔧 | `configs` | Basic authentication |
| auth_basic_user_file | 🔧 | `configs` | User file path |
| auth_request | ✅ | `auth_request` | External auth |
| auth_request_set | ✅ | `auth_request_set` | Set variables |

## Logging

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| access_log | ✅ | `access_log` | Path, format, level |
| error_log | ✅ | `error_log` | Path, level |
| log_format | 🔧 | `custom_global_config` | Custom log formats |
| log_subrequest | 🔧 | `configs` | Subrequest logging |
| open_log_file_cache | 🔧 | `custom_global_config` | Log file caching |

## Advanced Features

| Feature | Status | YAML Field | Notes |
|---------|--------|------------|-------|
| **Conditions** | | | |
| if statements | ✅ | `conditions[]` | Conditional logic |
| Complex conditions | ✅ | `conditions[].if` | Full nginx syntax |
| **Variables** | | | |
| set | 🔧 | `configs` or conditions | Variable assignment |
| map | 🔧 | `custom_global_config` | Variable mapping |
| **WebDAV** | | | |
| dav_methods | ✅ | `dav_methods[]` | WebDAV methods |
| dav_access | ✅ | `dav_access` | Access permissions |
| **FastCGI** | | | |
| fastcgi_pass | 🔧 | `configs` | PHP/FastCGI backend |
| fastcgi_param | 🔧 | `configs` | FastCGI parameters |
| fastcgi_cache | 🔧 | `configs` | FastCGI caching |
| **Real IP** | | | |
| real_ip_header | 🔧 | `custom_global_config` | Real IP header |
| set_real_ip_from | 🔧 | `custom_global_config` | Trusted proxies |

## Third-party Modules

| Module | Status | Configuration | Notes |
|--------|--------|---------------|-------|
| **headers-more** | | | |
| more_set_headers | ✅ | `more_set_headers` | Enhanced header control |
| more_clear_headers | 🔧 | `configs` | Clear headers |
| **lua** | | | |
| lua_code_cache | 🔧 | `custom_global_config` | Lua configuration |
| content_by_lua | 🔧 | `configs` | Lua content handler |
| **njs** | | | |
| js_import | 🔧 | `custom_global_config` | JavaScript modules |
| js_content | 🔧 | `configs` | JavaScript content |
| **ModSecurity** | | | |
| modsecurity | 🔧 | `configs` | WAF integration |
| modsecurity_rules_file | 🔧 | `configs` | Rules file |

## Built-in Templates

| Template | Purpose | Includes |
|----------|---------|----------|
| `modern_ssl` | Modern SSL/TLS security | Strong protocols, ciphers, HSTS |
| `websocket` | WebSocket proxy support | Upgrade headers, timeouts |
| `reverse_proxy` | Standard reverse proxy | Headers, timeouts, buffering |
| `static_files` | Static file serving | Caching, compression, optimization |
| `security_headers` | Security headers | HSTS, CSP, XSS protection |
| `cors` | CORS API support | Full CORS with preflight |
| `php_fpm` | PHP-FPM integration | FastCGI configuration |
| `cache_proxy` | Proxy caching | Cache zones, validation |
| `auth_required` | Authentication | auth_request integration |
| `rate_limited` | Rate limiting | Request limiting |
| `gzip_proxy` | Compression | Gzip for proxied content |
| `maintenance` | Maintenance mode | Maintenance page handling |
| `health_check` | Health endpoint | Simple health checks |
| `bot_protection` | Bot protection | Bot detection and limiting |

## Configuration Flexibility

### 1. Native YAML Support (✅)
- Type-safe configuration with validation
- IDE autocompletion and syntax highlighting
- Template inheritance and composition
- Built-in best practices

### 2. Custom Configuration (🔧)
For features not directly supported:

```yaml
# Global custom config
custom_global_config:
  - "map $http_upgrade $connection_upgrade"
  - "default upgrade"
  - "''"      close"

# Server-level custom config  
servers:
  - listen: "80"
    configs:
      client_header_timeout: "60s"
      large_client_header_buffers: "4 16k"
    custom_config:
      - "set $maintenance 0"
      - "if (-f /var/www/maintenance.html) { set $maintenance 1; }"

# Location-level custom config
locations:
  - path: "/api"
    configs:
      fastcgi_pass: "unix:/var/run/php/php8.2-fpm.sock"
      fastcgi_index: "index.php"
    custom_config:
      - "fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name"
      - "include fastcgi_params"
```

### 3. Template System
- **Built-in Templates**: 14 pre-configured templates for common patterns
- **Custom Templates**: Define reusable configuration blocks
- **Template Inheritance**: Apply multiple templates to any location/server
- **Override Support**: Custom configurations override template defaults

### 4. Validation Benefits
- **Syntax Validation**: Prevents invalid nginx configurations
- **Reference Checking**: Ensures upstreams, zones, and caches exist
- **Type Safety**: Validates data types, formats, and ranges
- **Dependency Checking**: Validates template and reference dependencies

## Migration Path

### From nginx.conf to YAML

1. **Direct Migration**: Most common directives map directly to YAML fields
2. **Template Usage**: Replace repetitive blocks with templates
3. **Custom Config**: Use custom_config for advanced/uncommon directives
4. **Validation**: Catch errors early with built-in validation

### Example Migration

**Before (nginx.conf):**
```nginx
server {
    listen 443 ssl http2;
    server_name example.com;
    
    ssl_certificate /etc/ssl/cert.pem;
    ssl_certificate_key /etc/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options SAMEORIGIN always;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**After (YAML):**
```yaml
servers:
  - listen: "443 ssl http2"
    server_name: "example.com"
    ssl:
      certificate: "/etc/ssl/cert.pem"
      certificate_key: "/etc/ssl/key.pem"
      protocols: ["TLSv1.2", "TLSv1.3"]
      ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256"
    applyTemplates:
      - security_headers
    locations:
      - path: "/"
        proxy_pass: "http://backend"
        applyTemplates:
          - reverse_proxy
```

This dramatically reduces configuration complexity while maintaining full functionality and adding validation, templates, and better maintainability.
