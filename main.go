package main

import (
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
)

func (fw *FileWriter) WriteOutput(output string) {
	if dryRun {
		fmt.Printf(output)
	} else {
		_, err := fw.Write([]byte(output))
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}

func createMainNginxConfig() {
	var nginxConfFile *os.File
	if !dryRun {
		nginxConfFile, err := os.Create("default.conf")
		if err != nil {
			log.Fatalf("Error creating Nginx configuration file: %v", err)
		}
		defer nginxConfFile.Close()
	}
	fw := &FileWriter{nginxConfFile}
	for _, z := range nginxConfig.LimitReqZones {
		output := fmt.Sprintf("limit_req_zone %s", z.Key)
		if z.Zone != "" {
			output += fmt.Sprintf(" zone=%s", z.Zone)
		}
		if z.Rate != "" {
			output += fmt.Sprintf(" rate=%s", z.Rate)
		}
		output += ";\n"
		fw.WriteOutput(output)
	}
	for _, server := range nginxConfig.Servers {
		fw.WriteOutput("include " + server.Name + ".conf;\n")
	}
}

func createServer(i int, server NginxServer) {
	// Create an Nginx configuration file for each server
	if server.Name == "" {
		server.Name = fmt.Sprintf("server_%d", i+1)
	}
	var nginxConfFile *os.File
	if !dryRun {
		nginxConfFile, err := os.Create(server.Name + ".conf")
		if err != nil {
			log.Fatalf("Error creating Nginx configuration file: %v", err)
		}
		defer nginxConfFile.Close()
	}
	fw := &FileWriter{nginxConfFile}
	fw.WriteOutput("server {\n")
	fw.WriteOutput(fmt.Sprintf("\tserver_name %s;\n", server.ServerName))
	fw.WriteOutput(fmt.Sprintf("\tlisten %s;\n", server.Listen))
	fw.WriteOutput(fmt.Sprintf("\tssl_certificate %s;\n", server.SslCertificate))
	globalZone := getGlobalZone(server.LimitReqZone.Name)
	z := mergeStruct(server.LimitReqZone, globalZone).(LimitReqZone)
	fw.WriteOutput(fmt.Sprintf("\tlimit_req %s zone=%s;\n", z.Rate, z.Zone))

	for _, location := range server.Locations {
		location = mergeStruct(location, server.LocationDefaults).(Location)
		fw.WriteOutput(fmt.Sprintf("\tlocation %s {\n", location.Path))
		globalZone := getGlobalZone(location.LimitReqZone.Name)
		if globalZone != (LimitReqZone{}) {
			z := mergeStruct(location.LimitReqZone, globalZone).(LimitReqZone)
			fw.WriteOutput(fmt.Sprintf("\t\tlimit_req %s zone=%s;\n", z.Rate, z.Zone))
		}
		for k, v := range location.Configs {
			fw.WriteOutput(fmt.Sprintf("\t\t%s %s;\n", k, v))
		}
		fw.WriteOutput("\t}\n")
	}
	fw.WriteOutput("}\n")
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
}
