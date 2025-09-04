package main

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ValidateConfig validates the entire nginx configuration
func ValidateConfig(config *NginxConfig) []ValidationError {
	var errors []ValidationError
	
	// Validate global settings
	errors = append(errors, validateGlobalSettings(config)...)
	
	// Validate upstreams
	for i, upstream := range config.Upstreams {
		errors = append(errors, validateUpstream(&upstream, fmt.Sprintf("upstreams[%d]", i))...)
	}
	
	// Validate rate limiting zones
	for i, zone := range config.LimitReqZones {
		errors = append(errors, validateLimitReqZone(&zone, fmt.Sprintf("limit_req_zones[%d]", i))...)
	}
	
	// Validate proxy cache paths
	for i, cache := range config.ProxyCachePath {
		errors = append(errors, validateProxyCachePath(&cache, fmt.Sprintf("proxy_cache_path[%d]", i))...)
	}
	
	// Validate servers
	serverNames := make(map[string]bool)
	for i, server := range config.Servers {
		serverPath := fmt.Sprintf("servers[%d]", i)
		errors = append(errors, validateServer(&server, serverPath)...)
		
		// Check for duplicate server names
		if server.ServerName != "" {
			if serverNames[server.ServerName] {
				errors = append(errors, ValidationError{
					Field:   serverPath + ".server_name",
					Value:   server.ServerName,
					Message: "duplicate server name",
				})
			}
			serverNames[server.ServerName] = true
		}
	}
	
	// Validate templates
	for name, template := range config.TemplateConfigs {
		errors = append(errors, validateLocation(&template, fmt.Sprintf("templateConfigs[%s]", name))...)
	}
	
	return errors
}

// validateGlobalSettings validates global nginx settings
func validateGlobalSettings(config *NginxConfig) []ValidationError {
	var errors []ValidationError
	
	// Validate worker connections
	if config.WorkerConnections < 0 {
		errors = append(errors, ValidationError{
			Field:   "worker_connections",
			Value:   fmt.Sprintf("%d", config.WorkerConnections),
			Message: "worker_connections must be non-negative",
		})
	}
	
	// Validate keepalive timeout
	if config.KeepaliveTimeout != "" {
		if !isValidTimeValue(config.KeepaliveTimeout) {
			errors = append(errors, ValidationError{
				Field:   "keepalive_timeout",
				Value:   config.KeepaliveTimeout,
				Message: "invalid time format (e.g., '75s', '10m')",
			})
		}
	}
	
	// Validate client max body size
	if config.ClientMaxBodySize != "" {
		if !isValidSizeValue(config.ClientMaxBodySize) {
			errors = append(errors, ValidationError{
				Field:   "client_max_body_size",
				Value:   config.ClientMaxBodySize,
				Message: "invalid size format (e.g., '1m', '50k')",
			})
		}
	}
	
	// Validate server tokens
	if config.ServerTokens != "" {
		validTokens := map[string]bool{"on": true, "off": true, "build": true}
		if !validTokens[config.ServerTokens] {
			errors = append(errors, ValidationError{
				Field:   "server_tokens",
				Value:   config.ServerTokens,
				Message: "must be 'on', 'off', or 'build'",
			})
		}
	}
	
	return errors
}

// validateUpstream validates upstream configuration
func validateUpstream(upstream *Upstream, path string) []ValidationError {
	var errors []ValidationError
	
	if upstream.Name == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".name",
			Value:   "",
			Message: "upstream name is required",
		})
	}
	
	if len(upstream.Servers) == 0 {
		errors = append(errors, ValidationError{
			Field:   path + ".servers",
			Value:   "[]",
			Message: "at least one upstream server is required",
		})
	}
	
	// Validate load balancing method
	if upstream.Method != "" {
		validMethods := map[string]bool{
			"round_robin": true, "least_conn": true, "ip_hash": true,
			"hash": true, "random": true, "least_time": true,
		}
		if !validMethods[upstream.Method] {
			errors = append(errors, ValidationError{
				Field:   path + ".method",
				Value:   upstream.Method,
				Message: "invalid load balancing method",
			})
		}
	}
	
	// Validate servers
	for i, server := range upstream.Servers {
		serverPath := fmt.Sprintf("%s.servers[%d]", path, i)
		errors = append(errors, validateUpstreamServer(&server, serverPath)...)
	}
	
	return errors
}

// validateUpstreamServer validates individual upstream server
func validateUpstreamServer(server *UpstreamServer, path string) []ValidationError {
	var errors []ValidationError
	
	if server.Address == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".address",
			Value:   "",
			Message: "server address is required",
		})
	} else {
		// Validate address format (host:port or just host)
		if !isValidUpstreamAddress(server.Address) {
			errors = append(errors, ValidationError{
				Field:   path + ".address",
				Value:   server.Address,
				Message: "invalid address format (expected host:port or just host)",
			})
		}
	}
	
	if server.Weight < 0 {
		errors = append(errors, ValidationError{
			Field:   path + ".weight",
			Value:   fmt.Sprintf("%d", server.Weight),
			Message: "weight must be non-negative",
		})
	}
	
	if server.MaxFails < 0 {
		errors = append(errors, ValidationError{
			Field:   path + ".max_fails",
			Value:   fmt.Sprintf("%d", server.MaxFails),
			Message: "max_fails must be non-negative",
		})
	}
	
	if server.FailTimeout != "" && !isValidTimeValue(server.FailTimeout) {
		errors = append(errors, ValidationError{
			Field:   path + ".fail_timeout",
			Value:   server.FailTimeout,
			Message: "invalid time format",
		})
	}
	
	return errors
}

// validateLimitReqZone validates rate limiting zone
func validateLimitReqZone(zone *LimitReqZone, path string) []ValidationError {
	var errors []ValidationError
	
	if zone.Name == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".name",
			Value:   "",
			Message: "zone name is required",
		})
	}
	
	if zone.Key == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".key",
			Value:   "",
			Message: "zone key is required",
		})
	}
	
	if zone.Zone == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".zone",
			Value:   "",
			Message: "zone definition is required",
		})
	}
	
	if zone.Rate == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".rate",
			Value:   "",
			Message: "rate is required",
		})
	} else {
		// Validate rate format (e.g., "10r/s", "100r/m")
		if !isValidRateValue(zone.Rate) {
			errors = append(errors, ValidationError{
				Field:   path + ".rate",
				Value:   zone.Rate,
				Message: "invalid rate format (e.g., '10r/s', '100r/m')",
			})
		}
	}
	
	if zone.Burst < 0 {
		errors = append(errors, ValidationError{
			Field:   path + ".burst",
			Value:   fmt.Sprintf("%d", zone.Burst),
			Message: "burst must be non-negative",
		})
	}
	
	return errors
}

// validateProxyCachePath validates proxy cache path configuration
func validateProxyCachePath(cache *ProxyCachePath, path string) []ValidationError {
	var errors []ValidationError
	
	if cache.Name == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".name",
			Value:   "",
			Message: "cache name is required",
		})
	}
	
	if cache.Path == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".path",
			Value:   "",
			Message: "cache path is required",
		})
	}
	
	if cache.KeysZone == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".keys_zone",
			Value:   "",
			Message: "keys_zone is required",
		})
	}
	
	// Validate optional size and time values
	if cache.MaxSize != "" && !isValidSizeValue(cache.MaxSize) {
		errors = append(errors, ValidationError{
			Field:   path + ".max_size",
			Value:   cache.MaxSize,
			Message: "invalid size format",
		})
	}
	
	if cache.Inactive != "" && !isValidTimeValue(cache.Inactive) {
		errors = append(errors, ValidationError{
			Field:   path + ".inactive",
			Value:   cache.Inactive,
			Message: "invalid time format",
		})
	}
	
	return errors
}

// validateServer validates server block configuration
func validateServer(server *NginxServer, path string) []ValidationError {
	var errors []ValidationError
	
	if server.Listen == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".listen",
			Value:   "",
			Message: "listen directive is required",
		})
	} else {
		if !isValidListenValue(server.Listen) {
			errors = append(errors, ValidationError{
				Field:   path + ".listen",
				Value:   server.Listen,
				Message: "invalid listen format (e.g., '80', '443 ssl', '[::]:80')",
			})
		}
	}
	
	if server.ServerName == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".server_name",
			Value:   "",
			Message: "server_name is required",
		})
	}
	
	// Validate SSL configuration if present
	if server.SSL != nil {
		errors = append(errors, validateSSLConfig(server.SSL, path+".ssl")...)
	}
	
	// Validate locations
	locationPaths := make(map[string]bool)
	for i, location := range server.Locations {
		locationPath := fmt.Sprintf("%s.locations[%d]", path, i)
		errors = append(errors, validateLocation(&location, locationPath)...)
		
		// Check for duplicate location paths
		if locationPaths[location.Path] {
			errors = append(errors, ValidationError{
				Field:   locationPath + ".path",
				Value:   location.Path,
				Message: "duplicate location path",
			})
		}
		locationPaths[location.Path] = true
	}
	
	// Validate defaults location
	if server.Defaults.Path != "" || len(server.Defaults.Configs) > 0 {
		errors = append(errors, validateLocation(&server.Defaults, path+".defaults")...)
	}
	
	return errors
}

// validateSSLConfig validates SSL configuration
func validateSSLConfig(ssl *SSLConfig, path string) []ValidationError {
	var errors []ValidationError
	
	if ssl.Certificate == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".certificate",
			Value:   "",
			Message: "SSL certificate path is required",
		})
	}
	
	if ssl.CertificateKey == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".certificate_key",
			Value:   "",
			Message: "SSL certificate key path is required",
		})
	}
	
	// Validate SSL protocols
	if len(ssl.Protocols) > 0 {
		validProtocols := map[string]bool{
			"SSLv2": true, "SSLv3": true, "TLSv1": true,
			"TLSv1.1": true, "TLSv1.2": true, "TLSv1.3": true,
		}
		for _, protocol := range ssl.Protocols {
			if !validProtocols[protocol] {
				errors = append(errors, ValidationError{
					Field:   path + ".protocols",
					Value:   protocol,
					Message: "invalid SSL protocol",
				})
			}
		}
	}
	
	return errors
}

// validateLocation validates location block configuration
func validateLocation(location *Location, path string) []ValidationError {
	var errors []ValidationError
	
	// Path is required unless this is a template or defaults
	if location.Path == "" && !strings.Contains(path, "templateConfigs") && !strings.Contains(path, "defaults") {
		errors = append(errors, ValidationError{
			Field:   path + ".path",
			Value:   "",
			Message: "location path is required",
		})
	}
	
	// Validate proxy_pass URL if present
	if location.ProxyPass != "" {
		if !isValidProxyPass(location.ProxyPass) {
			errors = append(errors, ValidationError{
				Field:   path + ".proxy_pass",
				Value:   location.ProxyPass,
				Message: "invalid proxy_pass URL format",
			})
		}
	}
	
	// Validate timeout values
	timeoutFields := map[string]string{
		"proxy_connect_timeout": location.ProxyConnectTimeout,
		"proxy_send_timeout":    location.ProxySendTimeout,
		"proxy_read_timeout":    location.ProxyReadTimeout,
	}
	
	for field, value := range timeoutFields {
		if value != "" && !isValidTimeValue(value) {
			errors = append(errors, ValidationError{
				Field:   path + "." + field,
				Value:   value,
				Message: "invalid time format",
			})
		}
	}
	
	// Validate size values
	sizeFields := map[string]string{
		"proxy_buffer_size":       location.ProxyBufferSize,
		"proxy_busy_buffers_size": location.ProxyBusyBuffersSize,
	}
	
	for field, value := range sizeFields {
		if value != "" && !isValidSizeValue(value) {
			errors = append(errors, ValidationError{
				Field:   path + "." + field,
				Value:   value,
				Message: "invalid size format",
			})
		}
	}
	
	return errors
}

// Helper functions for validation

func isValidTimeValue(value string) bool {
	// Match patterns like "30s", "5m", "1h", "1d"
	re := regexp.MustCompile(`^\d+[smhd]?$`)
	return re.MatchString(value)
}

func isValidSizeValue(value string) bool {
	// Match patterns like "1m", "50k", "1g"
	re := regexp.MustCompile(`^\d+[kmg]?$`)
	return re.MatchString(strings.ToLower(value))
}

func isValidRateValue(value string) bool {
	// Match patterns like "10r/s", "100r/m"
	re := regexp.MustCompile(`^\d+r/[sm]$`)
	return re.MatchString(value)
}

func isValidUpstreamAddress(address string) bool {
	// Check if it's just a hostname/IP
	if net.ParseIP(address) != nil {
		return true
	}
	
	// Check if it's host:port format
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// Could be just a hostname without port
		return isValidHostname(address)
	}
	
	// Validate port number
	if portNum, err := strconv.Atoi(port); err != nil || portNum < 1 || portNum > 65535 {
		return false
	}
	
	// Validate host part
	return isValidHostname(host) || net.ParseIP(host) != nil
}

func isValidHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 253 {
		return false
	}
	
	// Simple regex for hostname validation
	re := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return re.MatchString(hostname)
}

func isValidListenValue(listen string) bool {
	// Parse different listen formats
	parts := strings.Fields(listen)
	if len(parts) == 0 {
		return false
	}
	
	addressPort := parts[0]
	
	// Check for IPv6 format [::]:80
	if strings.HasPrefix(addressPort, "[") {
		if !strings.Contains(addressPort, "]:") {
			return false
		}
		// Extract and validate IPv6
		closeBracket := strings.Index(addressPort, "]")
		if closeBracket == -1 {
			return false
		}
		ipv6 := addressPort[1:closeBracket]
		if net.ParseIP(ipv6) == nil {
			return false
		}
		// Validate port
		portStr := addressPort[closeBracket+2:]
		if portNum, err := strconv.Atoi(portStr); err != nil || portNum < 1 || portNum > 65535 {
			return false
		}
	} else if strings.Contains(addressPort, ":") {
		// host:port format
		host, port, err := net.SplitHostPort(addressPort)
		if err != nil {
			return false
		}
		if net.ParseIP(host) == nil && !isValidHostname(host) && host != "*" {
			return false
		}
		if portNum, err := strconv.Atoi(port); err != nil || portNum < 1 || portNum > 65535 {
			return false
		}
	} else {
		// Just port number
		if portNum, err := strconv.Atoi(addressPort); err != nil || portNum < 1 || portNum > 65535 {
			return false
		}
	}
	
	// Validate additional options (ssl, http2, etc.)
	validOptions := map[string]bool{
		"ssl": true, "http2": true, "spdy": true, "proxy_protocol": true,
		"default_server": true, "default": true, "bind": true, "ipv6only": true,
	}
	for i := 1; i < len(parts); i++ {
		option := parts[i]
		if strings.Contains(option, "=") {
			option = strings.Split(option, "=")[0]
		}
		if !validOptions[option] {
			return false
		}
	}
	
	return true
}

func isValidProxyPass(proxyPass string) bool {
	// Must be a valid URL or upstream name
	if strings.HasPrefix(proxyPass, "http://") || strings.HasPrefix(proxyPass, "https://") {
		_, err := url.Parse(proxyPass)
		return err == nil
	}
	
	// Could be an upstream name
	return isValidHostname(strings.TrimPrefix(proxyPass, "http://"))
}
