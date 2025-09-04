package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func mergeLocations(keep, merge Location) Location {
	// Create a copy of the 'merge' struct
	result := merge
	result.Path = keep.Path
	result.Configs = mergeStringMaps(keep.Configs, merge.Configs)
	result.AddHeader = mergeStringMaps(keep.AddHeader, merge.AddHeader)
	result.ProxySetHeader = mergeStringMaps(keep.ProxySetHeader, merge.ProxySetHeader)
	result.MoreSetHeaders = mergeStringMaps(keep.MoreSetHeaders, merge.MoreSetHeaders)
	result.LimitReqZone = mergeLimitReq(keep.LimitReqZone, merge.LimitReqZone)
	result.Conditions = mergeConditions(keep.Conditions, merge.Conditions)
	result.AuthRequestSet = mergeStringMaps(keep.AuthRequestSet, merge.AuthRequestSet)

	// Merge arrays
	result.ProxyPassHeader = mergeStringArrays(keep.ProxyPassHeader, merge.ProxyPassHeader)
	result.ProxyHideHeader = mergeStringArrays(keep.ProxyHideHeader, merge.ProxyHideHeader)
	result.ProxyIgnoreHeaders = mergeStringArrays(keep.ProxyIgnoreHeaders, merge.ProxyIgnoreHeaders)
	result.Index = mergeStringArrays(keep.Index, merge.Index)
	result.Allow = mergeStringArrays(keep.Allow, merge.Allow)
	result.Deny = mergeStringArrays(keep.Deny, merge.Deny)
	result.DavMethods = mergeStringArrays(keep.DavMethods, merge.DavMethods)
	result.CustomConfig = mergeStringArrays(keep.CustomConfig, merge.CustomConfig)
	result.ProxyCacheBypass = mergeStringArrays(keep.ProxyCacheBypass, merge.ProxyCacheBypass)

	return result
}

// mergeStringArrays merges two string arrays, preferring 'keep' values
func mergeStringArrays(keep, merge []string) []string {
	if len(keep) > 0 {
		return keep
	}
	return merge
}

func mergeStringMaps(keep, merge map[string]string) map[string]string {
	// Create a new map to hold the merged result
	result := make(map[string]string)
	// Copy the contents of 'merge' map to the result
	for k, v := range merge {
		result[k] = v
	}
	// Merge the contents of 'keep' map, overwriting values from 'merge' if keys clash
	for k, v := range keep {
		result[k] = v
	}
	return result
}

func mergeConditions(keep, merge []Condition) []Condition {
	// Create a new slice to hold the merged result
	result := make([]Condition, len(merge))

	// Copy the contents of 'merge' slice to the result
	copy(result, merge)

	// Merge the contents of 'keep' slice, appending any new conditions
	for _, k := range keep {
		var found bool
		for _, r := range result {
			if k.If == r.If {
				// Merge 'then' fields if 'If' conditions match
				r.Then = append(r.Then, k.Then...)
				found = true
				break
			}
		}
		if !found {
			// Append the condition from 'keep' if not found in 'merge'
			result = append(result, k)
		}
	}

	return result
}

func mergeLimitReq(keep, merge LimitReqZone) LimitReqZone {
	// Create a copy of the 'merge' struct
	result := merge

	// Merge 'name' field
	if keep.Name != "" {
		result.Name = keep.Name
	}

	// Merge 'key' field
	if keep.Key != "" {
		result.Key = keep.Key
	}

	// Merge 'zone' field
	if keep.Zone != "" {
		result.Zone = keep.Zone
	}

	// Merge 'rate' field
	if keep.Rate != "" {
		result.Rate = keep.Rate
	}

	// Merge 'burst' field
	if keep.Burst != 0 {
		result.Burst = keep.Burst
	}

	return result
}

func getGlobalZone(name string) LimitReqZone {
	for _, zone := range nginxConfig.LimitReqZones {
		if zone.Name == name {
			return zone
		}
	}
	return LimitReqZone{}
}

func setKeys(m map[string]string, directive string) {
	for k, v := range m {
		buffer.WriteString(fmt.Sprintf("%s %s %s;\n", directive, k, v))
	}
}

func setKey(value string, directive string) {
	if value != "" {
		buffer.WriteString(fmt.Sprintf("%s %s;\n", directive, strings.TrimSuffix(value, ";")))
	}
}

func WriteOutput(output bytes.Buffer) {
	if dryRun {
		fmt.Print(output.String())
	} else {
		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Error creating output directory '%s': %v", dir, err)
		}

		nginxConfFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("Error creating Nginx configuration file '%s': %v", outputPath, err)
		}
		defer func() {
			if closeErr := nginxConfFile.Close(); closeErr != nil {
				log.Printf("Warning: Error closing Nginx configuration file: %v", closeErr)
			}
		}()

		fw := &FileWriter{nginxConfFile}
		_, err = fw.Write(output.Bytes())
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		if verbose {
			log.Printf("Successfully wrote nginx configuration to: %s", outputPath)
		}
	}
}

// createLogConfig generates log configuration
func createLogConfig(directive string, config *LogConfig) {
	logStr := fmt.Sprintf("\t%s", directive)
	if config.Path != "" {
		logStr += " " + config.Path
	} else {
		logStr += " /var/log/nginx/" + strings.Replace(directive, "_", ".", 1) + ".log"
	}

	if config.Format != "" {
		logStr += " " + config.Format
	}

	if config.Level != "" && directive == "error_log" {
		logStr += " " + config.Level
	}

	buffer.WriteString(logStr + ";\n")
}

// createGzipConfig generates gzip compression configuration
func createGzipConfig(gzip *GzipConfig) {
	if !gzip.Enabled {
		buffer.WriteString("\tgzip off;\n")
		return
	}

	buffer.WriteString("\tgzip on;\n")

	if gzip.Vary {
		buffer.WriteString("\tgzip_vary on;\n")
	}

	if gzip.Proxied != "" {
		buffer.WriteString(fmt.Sprintf("\tgzip_proxied %s;\n", gzip.Proxied))
	}

	if gzip.CompLevel > 0 {
		buffer.WriteString(fmt.Sprintf("\tgzip_comp_level %d;\n", gzip.CompLevel))
	}

	if gzip.MinLength > 0 {
		buffer.WriteString(fmt.Sprintf("\tgzip_min_length %d;\n", gzip.MinLength))
	}

	if len(gzip.Types) > 0 {
		buffer.WriteString(fmt.Sprintf("\tgzip_types %s;\n", strings.Join(gzip.Types, " ")))
	}

	if gzip.Disable != "" {
		buffer.WriteString(fmt.Sprintf("\tgzip_disable %s;\n", gzip.Disable))
	}
}

// createSecurityHeaders generates security headers
func createSecurityHeaders(security *SecurityHeaders) {
	// HSTS header
	if security.HSTS != nil {
		hstsValue := fmt.Sprintf("max-age=%d", security.HSTS.MaxAge)
		if security.HSTS.IncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if security.HSTS.Preload {
			hstsValue += "; preload"
		}
		buffer.WriteString(fmt.Sprintf("\tadd_header Strict-Transport-Security \"%s\" always;\n", hstsValue))
	}

	// Content Security Policy
	if security.ContentSecurityPolicy != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header Content-Security-Policy \"%s\" always;\n", security.ContentSecurityPolicy))
	}

	// X-Frame-Options
	if security.XFrameOptions != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header X-Frame-Options \"%s\" always;\n", security.XFrameOptions))
	}

	// X-Content-Type-Options
	if security.XContentTypeOptions != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header X-Content-Type-Options \"%s\" always;\n", security.XContentTypeOptions))
	}

	// Referrer Policy
	if security.ReferrerPolicy != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header Referrer-Policy \"%s\" always;\n", security.ReferrerPolicy))
	}

	// Permissions Policy
	if security.PermissionsPolicy != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header Permissions-Policy \"%s\" always;\n", security.PermissionsPolicy))
	}

	// CORS headers
	if security.CORS != nil {
		createCORSHeaders(security.CORS)
	}
}

// createCORSHeaders generates CORS headers
func createCORSHeaders(cors *CORSConfig) {
	if cors.AllowOrigin != "" {
		buffer.WriteString(fmt.Sprintf("\tadd_header Access-Control-Allow-Origin \"%s\" always;\n", cors.AllowOrigin))
	}

	if len(cors.AllowMethods) > 0 {
		buffer.WriteString(fmt.Sprintf("\tadd_header Access-Control-Allow-Methods \"%s\" always;\n", strings.Join(cors.AllowMethods, ", ")))
	}

	if len(cors.AllowHeaders) > 0 {
		buffer.WriteString(fmt.Sprintf("\tadd_header Access-Control-Allow-Headers \"%s\" always;\n", strings.Join(cors.AllowHeaders, ", ")))
	}

	if len(cors.ExposeHeaders) > 0 {
		buffer.WriteString(fmt.Sprintf("\tadd_header Access-Control-Expose-Headers \"%s\" always;\n", strings.Join(cors.ExposeHeaders, ", ")))
	}

	if cors.AllowCredentials {
		buffer.WriteString("\tadd_header Access-Control-Allow-Credentials \"true\" always;\n")
	}

	if cors.MaxAge > 0 {
		buffer.WriteString(fmt.Sprintf("\tadd_header Access-Control-Max-Age \"%d\" always;\n", cors.MaxAge))
	}
}

// createProxyCacheConfig generates proxy cache path configuration
func createProxyCacheConfig(cache *ProxyCachePath) {
	cacheStr := fmt.Sprintf("\tproxy_cache_path %s", cache.Path)

	if cache.Levels != "" {
		cacheStr += fmt.Sprintf(" levels=%s", cache.Levels)
	}

	cacheStr += fmt.Sprintf(" keys_zone=%s", cache.KeysZone)

	if cache.MaxSize != "" {
		cacheStr += fmt.Sprintf(" max_size=%s", cache.MaxSize)
	}

	if cache.Inactive != "" {
		cacheStr += fmt.Sprintf(" inactive=%s", cache.Inactive)
	}

	if cache.UseStaleError {
		cacheStr += " use_temp_path=off"
	}

	buffer.WriteString(cacheStr + ";\n")
}

// createUpstreamConfig generates upstream block configuration
func createUpstreamConfig(upstream *Upstream) {
	buffer.WriteString(fmt.Sprintf("\tupstream %s {\n", upstream.Name))

	// Load balancing method
	if upstream.Method != "" && upstream.Method != "round_robin" {
		buffer.WriteString(fmt.Sprintf("\t\t%s;\n", upstream.Method))
	}

	// Keepalive settings
	if upstream.KeepaliveConnections > 0 {
		buffer.WriteString(fmt.Sprintf("\t\tkeepalive %d;\n", upstream.KeepaliveConnections))
	}

	if upstream.KeepaliveRequests > 0 {
		buffer.WriteString(fmt.Sprintf("\t\tkeepalive_requests %d;\n", upstream.KeepaliveRequests))
	}

	if upstream.KeepaliveTimeout != "" {
		buffer.WriteString(fmt.Sprintf("\t\tkeepalive_timeout %s;\n", upstream.KeepaliveTimeout))
	}

	// Upstream servers
	for _, server := range upstream.Servers {
		serverStr := fmt.Sprintf("\t\tserver %s", server.Address)

		if server.Weight > 0 {
			serverStr += fmt.Sprintf(" weight=%d", server.Weight)
		}

		if server.MaxFails > 0 {
			serverStr += fmt.Sprintf(" max_fails=%d", server.MaxFails)
		}

		if server.FailTimeout != "" {
			serverStr += fmt.Sprintf(" fail_timeout=%s", server.FailTimeout)
		}

		if server.Backup {
			serverStr += " backup"
		}

		if server.Down {
			serverStr += " down"
		}

		buffer.WriteString(serverStr + ";\n")
	}

	buffer.WriteString("\t}\n\n")
}

// createSSLConfig generates SSL configuration
func createSSLConfig(ssl *SSLConfig) {
	buffer.WriteString(fmt.Sprintf("\t\tssl_certificate %s;\n", ssl.Certificate))
	buffer.WriteString(fmt.Sprintf("\t\tssl_certificate_key %s;\n", ssl.CertificateKey))

	if len(ssl.Protocols) > 0 {
		buffer.WriteString(fmt.Sprintf("\t\tssl_protocols %s;\n", strings.Join(ssl.Protocols, " ")))
	}

	if ssl.Ciphers != "" {
		buffer.WriteString(fmt.Sprintf("\t\tssl_ciphers %s;\n", ssl.Ciphers))
	}

	if ssl.DHParam != "" {
		buffer.WriteString(fmt.Sprintf("\t\tssl_dhparam %s;\n", ssl.DHParam))
	}

	if ssl.SessionCache != "" {
		buffer.WriteString(fmt.Sprintf("\t\tssl_session_cache %s;\n", ssl.SessionCache))
	}

	if ssl.SessionTimeout != "" {
		buffer.WriteString(fmt.Sprintf("\t\tssl_session_timeout %s;\n", ssl.SessionTimeout))
	}

	if !ssl.SessionTickets {
		buffer.WriteString("\t\tssl_session_tickets off;\n")
	}

	if ssl.StaplingEnabled {
		buffer.WriteString("\t\tssl_stapling on;\n")
	}

	if ssl.StaplingVerify {
		buffer.WriteString("\t\tssl_stapling_verify on;\n")
	}

	if ssl.TrustedCert != "" {
		buffer.WriteString(fmt.Sprintf("\t\tssl_trusted_certificate %s;\n", ssl.TrustedCert))
	}
}

// hasLocationContent checks if location has any content to render
func hasLocationContent(location Location) bool {
	return location.ProxyPass != "" ||
		len(location.ProxySetHeader) > 0 ||
		len(location.AddHeader) > 0 ||
		len(location.MoreSetHeaders) > 0 ||
		len(location.Configs) > 0 ||
		len(location.CustomConfig) > 0 ||
		len(location.Conditions) > 0 ||
		location.Root != "" ||
		len(location.Index) > 0 ||
		location.TryFiles != "" ||
		location.AuthRequest != "" ||
		len(location.AuthRequestSet) > 0 ||
		location.LimitReqZone.Name != ""
}

// createLocationContent generates location block content
func createLocationContent(location Location, indent string) {
	// Proxy settings
	if location.ProxyPass != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_pass %s;\n", indent, location.ProxyPass))
	}

	// Headers
	setKeys(location.ProxySetHeader, indent+"proxy_set_header")
	setKeys(location.AddHeader, indent+"add_header")
	setKeys(location.MoreSetHeaders, indent+"more_set_headers")

	// Proxy pass/hide headers
	for _, header := range location.ProxyPassHeader {
		buffer.WriteString(fmt.Sprintf("%sproxy_pass_header %s;\n", indent, header))
	}
	for _, header := range location.ProxyHideHeader {
		buffer.WriteString(fmt.Sprintf("%sproxy_hide_header %s;\n", indent, header))
	}
	for _, header := range location.ProxyIgnoreHeaders {
		buffer.WriteString(fmt.Sprintf("%sproxy_ignore_headers %s;\n", indent, header))
	}

	// Proxy buffering
	if location.ProxyBuffering != nil {
		if *location.ProxyBuffering {
			buffer.WriteString(fmt.Sprintf("%sproxy_buffering on;\n", indent))
		} else {
			buffer.WriteString(fmt.Sprintf("%sproxy_buffering off;\n", indent))
		}
	}

	if location.ProxyBuffers != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_buffers %s;\n", indent, location.ProxyBuffers))
	}
	if location.ProxyBufferSize != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_buffer_size %s;\n", indent, location.ProxyBufferSize))
	}
	if location.ProxyBusyBuffersSize != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_busy_buffers_size %s;\n", indent, location.ProxyBusyBuffersSize))
	}

	// Proxy timeouts
	if location.ProxyConnectTimeout != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_connect_timeout %s;\n", indent, location.ProxyConnectTimeout))
	}
	if location.ProxySendTimeout != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_send_timeout %s;\n", indent, location.ProxySendTimeout))
	}
	if location.ProxyReadTimeout != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_read_timeout %s;\n", indent, location.ProxyReadTimeout))
	}

	// Caching
	if location.ProxyCache != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_cache %s;\n", indent, location.ProxyCache))
	}

	for status, time := range location.ProxyCacheValid {
		buffer.WriteString(fmt.Sprintf("%sproxy_cache_valid %s %s;\n", indent, status, time))
	}

	if location.ProxyCacheKey != "" {
		buffer.WriteString(fmt.Sprintf("%sproxy_cache_key %s;\n", indent, location.ProxyCacheKey))
	}

	for _, bypass := range location.ProxyCacheBypass {
		buffer.WriteString(fmt.Sprintf("%sproxy_cache_bypass %s;\n", indent, bypass))
	}

	if location.ProxyCacheLock {
		buffer.WriteString(fmt.Sprintf("%sproxy_cache_lock on;\n", indent))
	}

	// Rate limiting
	globalZone := getGlobalZone(location.LimitReqZone.Name)
	limitZone := mergeLimitReq(location.LimitReqZone, globalZone)
	if limitZone.Name != "" && limitZone.Zone != "" {
		limitReqStr := fmt.Sprintf("%slimit_req zone=%s", indent, limitZone.Zone)
		if limitZone.Burst > 0 {
			limitReqStr += fmt.Sprintf(" burst=%d", limitZone.Burst)
		}
		buffer.WriteString(limitReqStr + ";\n")
	}

	// Static file serving
	if location.Root != "" {
		buffer.WriteString(fmt.Sprintf("%sroot %s;\n", indent, location.Root))
	}

	if len(location.Index) > 0 {
		buffer.WriteString(fmt.Sprintf("%sindex %s;\n", indent, strings.Join(location.Index, " ")))
	}

	if location.TryFiles != "" {
		buffer.WriteString(fmt.Sprintf("%stry_files %s;\n", indent, location.TryFiles))
	}

	if location.Expires != "" {
		buffer.WriteString(fmt.Sprintf("%sexpires %s;\n", indent, location.Expires))
	}

	// Access control
	for _, allow := range location.Allow {
		buffer.WriteString(fmt.Sprintf("%sallow %s;\n", indent, allow))
	}
	for _, deny := range location.Deny {
		buffer.WriteString(fmt.Sprintf("%sdeny %s;\n", indent, deny))
	}

	// Authentication
	if location.AuthRequest != "" {
		buffer.WriteString(fmt.Sprintf("%sauth_request %s;\n", indent, location.AuthRequest))
	}

	for variable, value := range location.AuthRequestSet {
		buffer.WriteString(fmt.Sprintf("%sauth_request_set %s %s;\n", indent, variable, value))
	}

	// WebDAV
	if len(location.DavMethods) > 0 {
		buffer.WriteString(fmt.Sprintf("%sdav_methods %s;\n", indent, strings.Join(location.DavMethods, " ")))
	}
	if location.DavAccess != "" {
		buffer.WriteString(fmt.Sprintf("%sdav_access %s;\n", indent, location.DavAccess))
	}

	// Custom configurations
	for k, v := range location.Configs {
		buffer.WriteString(fmt.Sprintf("%s%s %s;\n", indent, k, v))
	}

	// Custom config strings
	for _, config := range location.CustomConfig {
		buffer.WriteString(fmt.Sprintf("%s%s;\n", indent, strings.TrimSuffix(config, ";")))
	}

	// Conditions
	setConditions(location.Conditions, indent)
}

func setConditions(conditions []Condition, indent string) {
	for _, c := range conditions {
		buffer.WriteString(fmt.Sprintf("%sif (%s) {\n", indent, strings.TrimSuffix(c.If, ";")))
		for _, statement := range c.Then {
			buffer.WriteString(fmt.Sprintf("%s\t%s;\n", indent, strings.TrimSuffix(statement, ";")))
		}
		buffer.WriteString(fmt.Sprintf("%s}\n", indent))
	}
}

// mergeServerWithTemplate merges template configurations into server-level settings
func mergeServerWithTemplate(server NginxServer, template Location) NginxServer {
	// Merge headers
	if server.AddHeader == nil {
		server.AddHeader = make(map[string]string)
	}
	for k, v := range template.AddHeader {
		server.AddHeader[k] = v
	}

	if server.ProxySetHeader == nil {
		server.ProxySetHeader = make(map[string]string)
	}
	for k, v := range template.ProxySetHeader {
		server.ProxySetHeader[k] = v
	}

	if server.MoreSetHeaders == nil {
		server.MoreSetHeaders = make(map[string]string)
	}
	for k, v := range template.MoreSetHeaders {
		server.MoreSetHeaders[k] = v
	}

	// Merge configs - handle SSL configs specially for backward compatibility
	if server.Configs == nil {
		server.Configs = make(map[string]string)
	}
	for k, v := range template.Configs {
		// Handle SSL configs for backward compatibility
		if k == "ssl_certificate" && server.SSL == nil {
			server.SSL = &SSLConfig{}
		}
		if k == "ssl_certificate_key" && server.SSL == nil {
			server.SSL = &SSLConfig{}
		}

		switch k {
		case "ssl_certificate":
			if server.SSL != nil {
				server.SSL.Certificate = v
			}
		case "ssl_certificate_key":
			if server.SSL != nil {
				server.SSL.CertificateKey = v
			}
		case "ssl_protocols":
			if server.SSL != nil {
				server.SSL.Protocols = strings.Fields(v)
			}
		case "ssl_ciphers":
			if server.SSL != nil {
				server.SSL.Ciphers = v
			}
		case "ssl_session_cache":
			if server.SSL != nil {
				server.SSL.SessionCache = v
			}
		case "ssl_session_timeout":
			if server.SSL != nil {
				server.SSL.SessionTimeout = v
			}
		case "ssl_session_tickets":
			if server.SSL != nil {
				server.SSL.SessionTickets = v == "on"
			}
		case "ssl_stapling":
			if server.SSL != nil {
				server.SSL.StaplingEnabled = v == "on"
			}
		case "ssl_stapling_verify":
			if server.SSL != nil {
				server.SSL.StaplingVerify = v == "on"
			}
		case "ssl_dhparam":
			if server.SSL != nil {
				server.SSL.DHParam = v
			}
		case "ssl_trusted_certificate":
			if server.SSL != nil {
				server.SSL.TrustedCert = v
			}
		case "http2":
			server.HTTP2 = v == "on"
		case "http3":
			server.HTTP3 = v == "on"
		default:
			// For non-SSL and non-HTTP2/3 configs, merge into server configs
			server.Configs[k] = v
		}
	}

	// Merge auth_request_set
	if len(template.AuthRequestSet) > 0 {
		// Convert the old format if needed
		for k, v := range template.AuthRequestSet {
			if server.Configs == nil {
				server.Configs = make(map[string]string)
			}
			server.Configs["auth_request_set"] = fmt.Sprintf("%s %s", k, v)
		}
	}

	return server
}
