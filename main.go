package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type FileWriter struct {
	*os.File
}

var (
	filePath      string
	outputPath    string
	dryRun        bool
	validate      bool
	verbose       bool
	listTemplates bool
	version       string = "1.0.0"
	nginxConfig   NginxConfig
	buffer        bytes.Buffer
	acmeChallenge = Location{
		Path: "/.well-known/acme-challenge",
		Configs: map[string]string{
			"root":         "/var/www/html",
			"allow":        "all",
			"default_type": "text/plain",
		},
	}
)

func createMainNginxConfig() {
	// Global worker settings
	if nginxConfig.WorkerProcesses != "" {
		buffer.WriteString(fmt.Sprintf("worker_processes %s;\n", nginxConfig.WorkerProcesses))
	}

	// Events block
	if nginxConfig.WorkerConnections > 0 {
		buffer.WriteString("events {\n")
		buffer.WriteString(fmt.Sprintf("\tworker_connections %d;\n", nginxConfig.WorkerConnections))
		buffer.WriteString("}\n\n")
	}

	// HTTP block
	buffer.WriteString("http {\n")

	// Basic HTTP settings
	if nginxConfig.KeepaliveTimeout != "" {
		buffer.WriteString(fmt.Sprintf("\tkeepalive_timeout %s;\n", nginxConfig.KeepaliveTimeout))
	}
	if nginxConfig.ClientMaxBodySize != "" {
		buffer.WriteString(fmt.Sprintf("\tclient_max_body_size %s;\n", nginxConfig.ClientMaxBodySize))
	}
	if nginxConfig.ServerTokens != "" {
		buffer.WriteString(fmt.Sprintf("\tserver_tokens %s;\n", nginxConfig.ServerTokens))
	}

	// MIME types
	buffer.WriteString("\tinclude /etc/nginx/mime.types;\n")
	buffer.WriteString("\tdefault_type application/octet-stream;\n")

	// Logging
	if nginxConfig.AccessLog != nil {
		createLogConfig("access_log", nginxConfig.AccessLog)
	}
	if nginxConfig.ErrorLog != nil {
		createLogConfig("error_log", nginxConfig.ErrorLog)
	}

	// Gzip compression
	if nginxConfig.Gzip != nil {
		createGzipConfig(nginxConfig.Gzip)
	}

	// Security headers
	if nginxConfig.SecurityHeaders != nil {
		createSecurityHeaders(nginxConfig.SecurityHeaders)
	}

	// Rate limiting configuration
	if nginxConfig.LimitReqLogLevel != "" {
		buffer.WriteString(fmt.Sprintf("\tlimit_req_log_level %s;\n", nginxConfig.LimitReqLogLevel))
	}
	if nginxConfig.LimitReqStatus > 0 {
		buffer.WriteString(fmt.Sprintf("\tlimit_req_status %d;\n", nginxConfig.LimitReqStatus))
	}

	// Rate limiting zones
	for _, z := range nginxConfig.LimitReqZones {
		output := fmt.Sprintf("\tlimit_req_zone %s", z.Key)
		if z.Zone != "" {
			output += fmt.Sprintf(" zone=%s", z.Zone)
		}
		if z.Rate != "" {
			output += fmt.Sprintf(" rate=%s", z.Rate)
		}
		output += ";\n"
		buffer.WriteString(output)
	}

	// Proxy cache paths
	for _, cache := range nginxConfig.ProxyCachePath {
		createProxyCacheConfig(&cache)
	}

	// Upstream blocks
	for _, upstream := range nginxConfig.Upstreams {
		createUpstreamConfig(&upstream)
	}

	// Custom global config
	for _, config := range nginxConfig.CustomGlobalConfig {
		buffer.WriteString(fmt.Sprintf("\t%s;\n", strings.TrimSuffix(config, ";")))
	}

	buffer.WriteString("\n")
}

// setInfo is deprecated - use createLocationContent instead
func setInfo(location Location, indent string) {
	createLocationContent(location, indent+"\t")
}

func createServer(i int, server NginxServer) {
	// Create an Nginx configuration for each server
	if server.Name == "" {
		server.Name = fmt.Sprintf("server_%d", i+1)
	}

	// auto-create 80 redirect to 443 if SSL is configured
	if strings.Contains(server.Listen, "443") || server.SSL != nil {
		buffer.WriteString(fmt.Sprintf("\tserver {\n\t\tlisten 80;\n\t\tserver_name %s;\n\t\treturn 301 https://$host$request_uri;\n\t}\n\n", server.ServerName))
	}

	buffer.WriteString("\tserver {\n")

	// Basic server configuration
	setKey(server.Listen, "\t\tlisten")
	setKey(server.ServerName, "\t\tserver_name")

	// HTTP/2 and HTTP/3 settings
	if server.HTTP2 {
		buffer.WriteString("\t\thttp2 on;\n")
	}
	if server.HTTP3 {
		buffer.WriteString("\t\thttp3 on;\n")
	}

	// SSL configuration
	if server.SSL != nil {
		createSSLConfig(server.SSL)
	}

	// Error pages
	for code, page := range server.ErrorPage {
		buffer.WriteString(fmt.Sprintf("\t\terror_page %s %s;\n", code, page))
	}

	// Return directive (for redirects)
	if server.Return != "" {
		parts := strings.SplitN(server.Return, ":", 2)
		if len(parts) == 2 {
			buffer.WriteString(fmt.Sprintf("\t\treturn %s %s;\n", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
		} else {
			buffer.WriteString(fmt.Sprintf("\t\treturn %s;\n", server.Return))
		}
	}

	// Rewrite rules
	for _, rule := range server.Rewrite {
		rewriteStr := fmt.Sprintf("\t\trewrite %s %s", rule.Regex, rule.Replacement)
		if rule.Flag != "" {
			rewriteStr += " " + rule.Flag
		}
		buffer.WriteString(rewriteStr + ";\n")
	}

	// Rate limiting
	globalZone := getGlobalZone(server.LimitReqZone.Name)
	z := mergeLimitReq(server.LimitReqZone, globalZone)
	if z.Name != "" && z.Zone != "" {
		limitReqStr := fmt.Sprintf("\t\tlimit_req zone=%s", z.Zone)
		if z.Burst > 0 {
			limitReqStr += fmt.Sprintf(" burst=%d", z.Burst)
		}
		buffer.WriteString(limitReqStr + ";\n")
	}

	// Apply templates to server defaults and server-level configs
	if server.ApplyTemplates != nil {
		for _, templateName := range server.ApplyTemplates {
			if template, ok := nginxConfig.TemplateConfigs[templateName]; ok {
				server.Defaults = mergeLocations(server.Defaults, template)
				// Also merge template configs into server-level configs
				server = mergeServerWithTemplate(server, template)
			}
		}
	}

	// Headers
	setKeys(server.AddHeader, "\t\tadd_header")
	setKeys(server.ProxySetHeader, "\t\tproxy_set_header")
	setKeys(server.MoreSetHeaders, "\t\tmore_set_headers")

	// Custom server configurations
	for k, v := range server.Configs {
		buffer.WriteString(fmt.Sprintf("\t\t%s %s;\n", k, v))
	}

	// Custom config strings
	for _, config := range server.CustomConfig {
		buffer.WriteString(fmt.Sprintf("\t\t%s;\n", strings.TrimSuffix(config, ";")))
	}

	// Add ACME challenge location if SSL is configured
	if server.SSL != nil {
		acmePath := "/.well-known/acme-challenge"
		if server.SSL.ACMEChallengeRoot != "" {
			acmeChallenge.Configs["root"] = server.SSL.ACMEChallengeRoot
		}
		buffer.WriteString(fmt.Sprintf("\t\tlocation %s {\n", acmePath))
		setKeys(acmeChallenge.Configs, "\t\t\t")
		buffer.WriteString("\t\t}\n")
	}

	// Apply server defaults
	if hasLocationContent(server.Defaults) {
		createLocationContent(server.Defaults, "\t\t")
	}

	// Location blocks
	for _, location := range server.Locations {
		// Apply templates to location
		if location.ApplyTemplates != nil {
			for _, templateName := range location.ApplyTemplates {
				if template, ok := nginxConfig.TemplateConfigs[templateName]; ok {
					location = mergeLocations(location, template)
				}
			}
		}

		// Skip ACME challenge if already handled
		if location.Path == "/.well-known/acme-challenge" && server.SSL != nil {
			continue
		}

		buffer.WriteString(fmt.Sprintf("\t\tlocation %s {\n", location.Path))
		createLocationContent(location, "\t\t\t")
		buffer.WriteString("\t\t}\n")
	}

	buffer.WriteString("\t}\n\n")
}

func main() {
	// CLI flags
	flag.StringVar(&filePath, "f", "nginx_config.yaml", "Path to YAML configuration file")
	flag.StringVar(&filePath, "file", "nginx_config.yaml", "Path to YAML configuration file (shorthand)")
	flag.StringVar(&outputPath, "o", "/etc/nginx/conf.d/default.conf", "Output path for nginx configuration")
	flag.StringVar(&outputPath, "output", "/etc/nginx/conf.d/default.conf", "Output path for nginx configuration (shorthand)")
	flag.BoolVar(&dryRun, "d", false, "Dry run - output to stdout instead of file")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run - output to stdout instead of file (shorthand)")
	flag.BoolVar(&validate, "v", false, "Validate configuration only")
	flag.BoolVar(&validate, "validate", false, "Validate configuration only (shorthand)")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&listTemplates, "list-templates", false, "List all available built-in templates")
	
	var multiFile bool
	var confDir string
	flag.BoolVar(&multiFile, "multi-file", false, "Generate separate .conf files for each service (like conf.d structure)")
	flag.StringVar(&confDir, "conf-dir", "/etc/nginx/conf.d", "Directory for multi-file output (default: /etc/nginx/conf.d)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Printf("nginx-yaml-entrypoint v%s - Convert YAML to nginx configuration\n\n", version)
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Printf("  %s -f config.yaml -o /etc/nginx/nginx.conf\n", os.Args[0])
		fmt.Printf("  %s -validate -f config.yaml\n", os.Args[0])
		fmt.Printf("  %s -dry-run -verbose -f config.yaml\n", os.Args[0])
		fmt.Printf("  %s -list-templates\n", os.Args[0])
	}

	flag.Parse()

	// Handle special flags
	if listTemplates {
		PrintBuiltInTemplates()
		return
	}

	if verbose {
		log.Printf("nginx-yaml-entrypoint v%s", version)
		log.Printf("Reading YAML configuration from: %s", filePath)
	}

	// Check if input file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Configuration file '%s' does not exist", filePath)
	}

	// Read and parse YAML file
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading YAML file '%s': %v", filePath, err)
	}

	err = yaml.Unmarshal(yamlFile, &nginxConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	if verbose {
		log.Printf("Successfully parsed YAML configuration")
	}

	// Merge user templates with built-in templates
	nginxConfig.TemplateConfigs = MergeWithBuiltInTemplates(nginxConfig.TemplateConfigs)

	if verbose {
		log.Printf("Merged %d built-in templates with user templates", len(BuiltInTemplates))
	}

	// Validate configuration
	if validate || verbose {
		validationErrors := ValidateConfig(&nginxConfig)
		if len(validationErrors) > 0 {
			log.Printf("Configuration validation failed with %d errors:", len(validationErrors))
			for _, err := range validationErrors {
				log.Printf("  - %s", err.Error())
			}
			if validate {
				os.Exit(1)
			}
			log.Printf("Continuing despite validation errors...")
		} else if verbose || validate {
			log.Printf("Configuration validation passed")
			if validate {
				fmt.Println("✓ Configuration is valid")
				return
			}
		}
	}

	if verbose {
		log.Printf("Generating nginx configuration...")
	}

	// Generate nginx configuration
	createMainNginxConfig()

	// Generate server blocks
	for i, server := range nginxConfig.Servers {
		createServer(i, server)
	}

	// Close HTTP block
	buffer.WriteString("}\n")

	if verbose {
		log.Printf("Generated %d bytes of nginx configuration", buffer.Len())
	}

	// Output configuration
	WriteOutput(buffer)
}
