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
	dryRun        bool
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
	for _, z := range nginxConfig.LimitReqZones {
		output := fmt.Sprintf("limit_req_zone %s", z.Key)
		if z.Zone != "" {
			output += fmt.Sprintf(" zone=%s", z.Zone)
		}
		if z.Rate != "" {
			output += fmt.Sprintf(" rate=%s", z.Rate)
		}
		output += ";\n"
		buffer.WriteString(output)
	}
	//for _, server := range nginxConfig.Servers {
	//	buffer.WriteString("include " + server.Name + ".conf;\n")
	//}
}

func setInfo(location Location, indent string) {
	setConditions(location.Conditions)
	globalZone := getGlobalZone(location.LimitReqZone.Name)
	limit_zone := mergeLimitReq(location.LimitReqZone, globalZone)
	if globalZone != limit_zone {
		buffer.WriteString(fmt.Sprintf(indent+"\tlimit_req %s zone=%s;\n", limit_zone.Rate, limit_zone.Zone))
	}
	setKeys(location.AddHeader, indent+"add_header")
	setKeys(location.ProxySetHeader, indent+"proxy_set_header")
	for k, v := range location.Configs {
		buffer.WriteString(fmt.Sprintf(indent+"\t%s %s;\n", k, v))
	}
	if location.ProxyPass != "" {
		buffer.WriteString(fmt.Sprintf(indent+"\tproxy_pass %s;\n", location.ProxyPass))
	}
}

func createServer(i int, server NginxServer) {
	// Create an Nginx configuration for each server
	if server.Name == "" {
		server.Name = fmt.Sprintf("server_%d", i+1)
	}
	// auto-create 80 redirect to 443
	if strings.Contains(server.Listen, "443") {
		// auto-create 80 redirect to 443
		buffer.WriteString(fmt.Sprintf("server {\n\tlisten 80;\n\tserver_name %s;\n\treturn 301 https://$host$request_uri;\n}\n", server.ServerName))
	}
	buffer.WriteString("server {\n")
	setKey(server.ServerName, "server_name")
	setKey(server.Listen, "listen")
	globalZone := getGlobalZone(server.LimitReqZone.Name)
	z := mergeLimitReq(server.LimitReqZone, globalZone)
	if globalZone != z {
		buffer.WriteString(fmt.Sprintf("\tlimit_req %s zone=%s;\n", z.Rate, z.Zone))
	}
	for _, v := range server.CustomConfig {
		buffer.WriteString(fmt.Sprintf("\t%s;\n", v))
	}
	if server.ApplyTemplates != nil {
		for _, v := range server.ApplyTemplates {
			if _, ok := nginxConfig.TemplateConfigs[v]; ok {
				server.Defaults = mergeLocations(server.Defaults, nginxConfig.TemplateConfigs[v])
			}
		}
	}
	setKeys(server.AddHeader, "add_header")
	setKeys(server.ProxySetHeader, "proxy_set_header")
	for k, v := range server.Configs {
		buffer.WriteString(fmt.Sprintf("\t%s %s;\n", k, v))
	}
	if server.Configs["ssl_certificate"] != "" {
		buffer.WriteString(fmt.Sprintf("\tlocation %s {\n", acmeChallenge.Path))
		setKeys(acmeChallenge.Configs, "\t")
		buffer.WriteString("\t}\n")
	}
	setInfo(server.Defaults, "")
	for _, location := range server.Locations {
		if location.ApplyTemplates != nil {
			for _, v := range location.ApplyTemplates {
				if _, ok := nginxConfig.TemplateConfigs[v]; ok {
					location = mergeLocations(location, nginxConfig.TemplateConfigs[v])
				}
			}
		}
		if location.Path == acmeChallenge.Path {
			continue
		}
		buffer.WriteString(fmt.Sprintf("\tlocation %s {\n", location.Path))
		setInfo(location, "\t")
		buffer.WriteString("\t}\n")
	}
	buffer.WriteString("}\n")
}

func main() {
	flag.StringVar(&filePath, "f", "nginx_config.yaml", "Filepath for nginx_config.yaml")
	flag.StringVar(&filePath, "file", "nginx_config.yaml", "Filepath for nginx_config.yaml (shorthand)")
	flag.StringVar(&filePath, "o", "etc/nginx/conf.d/default.conf", "Filepath for nginx_config.yaml (shorthand)")
	flag.BoolVar(&dryRun, "d", false, "Dry run")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run (shorthand)")
	flag.Parse()
	if dryRun {
		fmt.Printf("Dry run: Nginx configuration file for server\n")
	}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &nginxConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}
	createMainNginxConfig()

	for i, server := range nginxConfig.Servers {
		createServer(i, server)
	}

	WriteOutput(buffer)
}
