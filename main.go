package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/goccy/go-yaml"
)

type FileWriter struct {
	*os.File
}

var (
	filePath    string
	dryRun      bool
	nginxConfig NginxConfig
	buffer      bytes.Buffer
)

func WriteOutput(output bytes.Buffer) {
	if dryRun {
		fmt.Print(string(output.String()))
	} else {
		nginxConfFile, err := os.Create("default.conf")
		if err != nil {
			log.Fatalf("Error creating Nginx configuration file: %v", err)
		}
		defer nginxConfFile.Close()
		fw := &FileWriter{nginxConfFile}
		_, err = fw.Write(output.Bytes())
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}

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
	for _, server := range nginxConfig.Servers {
		buffer.WriteString("include " + server.Name + ".conf;\n")
	}
}

func setKeys(m map[string]string, key string) {
	for k, v := range m {
		buffer.WriteString(fmt.Sprintf("\t%s %s \"%s\";\n", key, k, v))
	}
}
func setKey(s string, key string) {
	buffer.WriteString(fmt.Sprintf("\t%s %s;\n", key, s))
}

func createServer(i int, server NginxServer) {
	// Create an Nginx configuration file for each server
	if server.Name == "" {
		server.Name = fmt.Sprintf("server_%d", i+1)
	}
	buffer.WriteString("server {\n")
	setKey(server.ServerName, "server_name")
	setKey(server.Listen, "listen")
	setKey(server.SslCertificate, "ssl_certificate")
	globalZone := getGlobalZone(server.LimitReqZone.Name)
	z := mergeStruct(server.LimitReqZone, globalZone).(LimitReqZone)
	buffer.WriteString(fmt.Sprintf("\tlimit_req %s zone=%s;\n", z.Rate, z.Zone))
	for _, v := range server.CustomConfig {
		buffer.WriteString(fmt.Sprintf("\t%s;\n", v))
	}
	setKeys(server.AddHeader, "add_header")
	setKeys(server.ProxySetHeader, "proxy_set_header")

	for _, location := range server.Locations {
		buffer.WriteString(fmt.Sprintf("\tlocation %s {\n", location.Path))
		globalZone := getGlobalZone(location.LimitReqZone.Name)
		if globalZone != (LimitReqZone{}) {
			z := mergeStruct(location.LimitReqZone, globalZone).(LimitReqZone)
			buffer.WriteString(fmt.Sprintf("\t\tlimit_req %s zone=%s;\n", z.Rate, z.Zone))
		}
		for k, v := range location.Configs {
			buffer.WriteString(fmt.Sprintf("\t\t%s %s;\n", k, v))
		}
		if location.ProxyPass != "" {
			buffer.WriteString(fmt.Sprintf("\t\tproxy_pass %s;\n", location.ProxyPass))
		}
		buffer.WriteString("\t}\n")
	}
	buffer.WriteString("}\n")
}

// mergeStruct merges two structs using reflection
func mergeStruct(keep, merge interface{}) interface{} {
	keepValue := reflect.ValueOf(keep)
	mergeValue := reflect.ValueOf(merge)

	// Make sure both 'keep' and 'merge' are structs
	if keepValue.Kind() != reflect.Struct || mergeValue.Kind() != reflect.Struct {
		return nil
	}

	// Create a copy of the 'merge' struct
	result := reflect.New(mergeValue.Type()).Elem()

	// Iterate over fields of the 'merge' struct
	for i := 0; i < mergeValue.NumField(); i++ {
		field := mergeValue.Field(i)
		fieldName := mergeValue.Type().Field(i).Name

		// If the field is non-zero, use its value
		if !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			result.FieldByName(fieldName).Set(field)
		} else {
			// If the field is zero, use the value from 'keep'
			keepField := keepValue.FieldByName(fieldName)
			if keepField.IsValid() {
				result.FieldByName(fieldName).Set(keepField)
			}
		}
	}

	return result.Interface()
}

func getGlobalZone(name string) LimitReqZone {
	for _, zone := range nginxConfig.LimitReqZones {
		if zone.Name == name {
			return zone
		}
	}
	return LimitReqZone{}
}

func main() {
	flag.StringVar(&filePath, "f", "nginx_config.yaml", "Filepath for nginx_config.yaml")
	flag.StringVar(&filePath, "file", "nginx_config.yaml", "Filepath for nginx_config.yaml (shorthand)")
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
