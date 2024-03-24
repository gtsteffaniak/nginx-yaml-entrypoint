package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

func mergeLocations(keep, merge Location) Location {
	// Create a copy of the 'merge' struct
	result := merge
	result.Path = keep.Path
	result.Configs = mergeStringMaps(keep.Configs, merge.Configs)
	result.AddHeader = mergeStringMaps(keep.AddHeader, merge.AddHeader)
	result.ProxySetHeader = mergeStringMaps(keep.ProxySetHeader, merge.ProxySetHeader)
	result.LimitReqZone = mergeLimitReq(keep.LimitReqZone, merge.LimitReqZone)
	result.Conditions = mergeConditions(keep.Conditions, merge.Conditions)
	return result
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

func setKeys(m map[string]string, key string) {
	for k, v := range m {
		setKey(key, k+v)
	}
}

func setKey(key string, s string) {
	writeString := fmt.Sprintf("\t%s %s;\n", key, strings.TrimSuffix(s, ";"))
	buffer.WriteString(writeString)
}

func WriteOutput(output bytes.Buffer) {
	if dryRun {
		fmt.Print(string(output.String()))
	} else {
		nginxConfFile, err := os.Create("etc/nginx/conf.d/default.conf")
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

func setConditions(conditions []Condition) {
	for _, c := range conditions {
		buffer.WriteString(fmt.Sprintf("\t\tif (%s) {\n", strings.TrimSuffix(c.If, ";")))
		for _, statement := range c.Then {
			buffer.WriteString(fmt.Sprintf("\t\t\t%s;\n", strings.TrimSuffix(statement, ";")))
		}
		buffer.WriteString("\t\t}\n")
	}
}
