package main

import "fmt"

// BuiltInTemplates provides common nginx configuration templates
var BuiltInTemplates = map[string]Location{
	// SSL/TLS modern security template
	"modern_ssl": {
		Configs: map[string]string{
			"ssl_protocols":             "TLSv1.2 TLSv1.3",
			"ssl_ciphers":               "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384",
			"ssl_prefer_server_ciphers": "off",
			"ssl_session_cache":         "shared:SSL:10m",
			"ssl_session_timeout":       "10m",
			"ssl_session_tickets":       "off",
			"ssl_stapling":              "on",
			"ssl_stapling_verify":       "on",
		},
	},

	// WebSocket support template
	"websocket": {
		ProxySetHeader: map[string]string{
			"Upgrade":           "$http_upgrade",
			"Connection":        "$connection_upgrade",
			"Host":              "$host",
			"X-Real-IP":         "$remote_addr",
			"X-Forwarded-For":   "$proxy_add_x_forwarded_for",
			"X-Forwarded-Proto": "$scheme",
		},
		Configs: map[string]string{
			"proxy_http_version":    "1.1",
			"proxy_cache_bypass":    "$http_upgrade",
			"proxy_read_timeout":    "86400s",
			"proxy_send_timeout":    "86400s",
			"proxy_connect_timeout": "30s",
		},
	},

	// Standard reverse proxy template
	"reverse_proxy": {
		ProxySetHeader: map[string]string{
			"Host":              "$host",
			"X-Real-IP":         "$remote_addr",
			"X-Forwarded-For":   "$proxy_add_x_forwarded_for",
			"X-Forwarded-Proto": "$scheme",
			"X-Forwarded-Host":  "$host",
			"X-Forwarded-Port":  "$server_port",
		},
		Configs: map[string]string{
			"proxy_buffering":       "on",
			"proxy_buffer_size":     "4k",
			"proxy_buffers":         "8 4k",
			"proxy_connect_timeout": "30s",
			"proxy_send_timeout":    "30s",
			"proxy_read_timeout":    "30s",
			"proxy_redirect":        "off",
		},
	},

	// Static file serving template
	"static_files": {
		Configs: map[string]string{
			"expires":                  "1y",
			"access_log":               "off",
			"tcp_nopush":               "on",
			"tcp_nodelay":              "on",
			"open_file_cache":          "max=1000 inactive=20s",
			"open_file_cache_valid":    "30s",
			"open_file_cache_min_uses": "2",
			"open_file_cache_errors":   "on",
		},
		AddHeader: map[string]string{
			"Cache-Control": "public, no-transform",
			"Vary":          "Accept-Encoding",
		},
	},

	// Security headers template
	"security_headers": {
		AddHeader: map[string]string{
			"X-Frame-Options":           "SAMEORIGIN",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
			"Referrer-Policy":           "strict-origin-when-cross-origin",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		},
	},

	// CORS template for APIs
	"cors": {
		AddHeader: map[string]string{
			"Access-Control-Allow-Origin":      "$http_origin",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
			"Access-Control-Allow-Headers":     "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization",
			"Access-Control-Expose-Headers":    "Content-Length,Content-Range",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Max-Age":           "1728000",
		},
		Conditions: []Condition{
			{
				If: "$request_method = OPTIONS",
				Then: []string{
					"add_header Access-Control-Allow-Origin $http_origin always",
					"add_header Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS, PATCH' always",
					"add_header Access-Control-Allow-Headers 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization' always",
					"add_header Access-Control-Max-Age 1728000 always",
					"add_header Content-Type 'text/plain; charset=utf-8' always",
					"add_header Content-Length 0 always",
					"return 204",
				},
			},
		},
	},

	// PHP-FPM template
	"php_fpm": {
		Configs: map[string]string{
			"fastcgi_pass":                 "unix:/var/run/php/php8.2-fpm.sock",
			"fastcgi_index":                "index.php",
			"include":                      "fastcgi_params",
			"fastcgi_param":                "SCRIPT_FILENAME $document_root$fastcgi_script_name",
			"fastcgi_connect_timeout":      "60s",
			"fastcgi_send_timeout":         "180s",
			"fastcgi_read_timeout":         "180s",
			"fastcgi_buffer_size":          "128k",
			"fastcgi_buffers":              "4 256k",
			"fastcgi_busy_buffers_size":    "256k",
			"fastcgi_temp_file_write_size": "256k",
			"fastcgi_intercept_errors":     "on",
		},
	},

	// Caching template for proxied content
	"cache_proxy": {
		Configs: map[string]string{
			"proxy_cache":                   "default",
			"proxy_cache_key":               "$scheme$request_method$host$request_uri",
			"proxy_cache_lock":              "on",
			"proxy_cache_use_stale":         "error timeout updating http_500 http_502 http_503 http_504",
			"proxy_cache_background_update": "on",
		},
		ProxyCacheValid: map[string]string{
			"200 302": "10m",
			"404":     "1m",
		},
		AddHeader: map[string]string{
			"X-Cache-Status": "$upstream_cache_status",
		},
	},

	// Authentication template (using auth_request module)
	"auth_required": {
		Configs: map[string]string{
			"auth_request": "/auth/verify",
		},
		AuthRequestSet: map[string]string{
			"$user":        "$upstream_http_x_user",
			"$auth_status": "$upstream_status",
		},
		ProxySetHeader: map[string]string{
			"X-Forwarded-User": "$user",
			"X-Auth-Status":    "$auth_status",
		},
	},

	// Rate limiting template
	"rate_limited": {
		LimitReqZone: LimitReqZone{
			Name:  "default",
			Zone:  "default:10m",
			Rate:  "10r/s",
			Burst: 20,
		},
	},

	// Gzip compression for proxied content
	"gzip_proxy": {
		Configs: map[string]string{
			"gzip":            "on",
			"gzip_vary":       "on",
			"gzip_min_length": "1000",
			"gzip_proxied":    "any",
			"gzip_comp_level": "6",
			"gzip_types":      "text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/atom+xml image/svg+xml",
		},
	},

	// Maintenance mode template
	"maintenance": {
		Conditions: []Condition{
			{
				If: "-f $document_root/maintenance.html",
				Then: []string{
					"return 503",
				},
			},
		},
		Configs: map[string]string{
			"error_page": "503 @maintenance",
		},
		CustomConfig: []string{
			"location @maintenance",
			"root /var/www/html",
			"rewrite ^(.*)$ /maintenance.html break",
		},
	},

	// Load balancer health check template
	"health_check": {
		Configs: map[string]string{
			"access_log": "off",
			"return":     "200 'OK'",
		},
		AddHeader: map[string]string{
			"Content-Type": "text/plain",
		},
	},

	// Bot protection template
	"bot_protection": {
		Conditions: []Condition{
			{
				If: "$http_user_agent ~* (bot|crawler|spider|scraper)",
				Then: []string{
					"access_log /var/log/nginx/bot.log",
					"limit_req zone=bot burst=5 nodelay",
				},
			},
		},
		LimitReqZone: LimitReqZone{
			Name:  "bot",
			Zone:  "bot:10m",
			Rate:  "1r/s",
			Burst: 5,
		},
	},
}

// GetBuiltInTemplate returns a built-in template by name
func GetBuiltInTemplate(name string) (Location, bool) {
	template, exists := BuiltInTemplates[name]
	return template, exists
}

// ListBuiltInTemplates returns a list of all available built-in template names
func ListBuiltInTemplates() []string {
	names := make([]string, 0, len(BuiltInTemplates))
	for name := range BuiltInTemplates {
		names = append(names, name)
	}
	return names
}

// PrintBuiltInTemplates prints all available built-in templates with descriptions
func PrintBuiltInTemplates() {
	fmt.Println("Available built-in templates:")
	fmt.Println()

	descriptions := map[string]string{
		"modern_ssl":       "Modern SSL/TLS configuration with strong security",
		"websocket":        "WebSocket proxy support with proper headers",
		"reverse_proxy":    "Standard reverse proxy configuration",
		"static_files":     "Optimized static file serving with caching",
		"security_headers": "Common security headers (HSTS, CSP, etc.)",
		"cors":             "CORS headers for API endpoints",
		"php_fpm":          "PHP-FPM fastcgi configuration",
		"cache_proxy":      "Proxy caching for improved performance",
		"auth_required":    "Authentication using auth_request module",
		"rate_limited":     "Rate limiting for abuse protection",
		"gzip_proxy":       "Gzip compression for proxied content",
		"maintenance":      "Maintenance mode handling",
		"health_check":     "Simple health check endpoint",
		"bot_protection":   "Basic bot protection and rate limiting",
	}

	for _, name := range ListBuiltInTemplates() {
		description := descriptions[name]
		if description == "" {
			description = "No description available"
		}
		fmt.Printf("  %-16s - %s\n", name, description)
	}
	fmt.Println()
}

// MergeWithBuiltInTemplates merges user templates with built-in templates
func MergeWithBuiltInTemplates(userTemplates map[string]Location) map[string]Location {
	merged := make(map[string]Location)

	// Add built-in templates first
	for name, template := range BuiltInTemplates {
		merged[name] = template
	}

	// Override with user templates (user templates take precedence)
	for name, template := range userTemplates {
		merged[name] = template
	}

	return merged
}
