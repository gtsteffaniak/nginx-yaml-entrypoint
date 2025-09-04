#!/bin/sh
# nginx-yaml-entrypoint automatic configuration detection
# vim:sw=4:ts=4:et

set -e

# Function to process YAML files
process_yaml_configs() {
    local yaml_dir="/etc/nginx/yaml"
    local output_file="/etc/nginx/nginx.conf"
    local found_yaml=false

    echo "Scanning for YAML configurations in $yaml_dir..."

    if [ -d "$yaml_dir" ]; then
        # Look for YAML files in priority order
        for yaml_file in "$yaml_dir/main.yaml" "$yaml_dir/nginx.yaml" "$yaml_dir/config.yaml" "$yaml_dir"/*.yaml "$yaml_dir"/*.yml; do
            if [ -f "$yaml_file" ]; then
                echo "Found YAML configuration: $yaml_file"

                # Process the YAML file
                echo "Converting YAML to nginx configuration..."
                if nginx-yaml-entrypoint -f "$yaml_file" -o "$output_file" --verbose; then
                    echo "Successfully generated nginx configuration from $yaml_file"
                    found_yaml=true
                    break  # Use the first valid YAML file found
                else
                    echo "Failed to process $yaml_file, trying next..."
                fi
            fi
        done
    fi

    if [ "$found_yaml" = "true" ]; then
        # Validate the generated configuration
        echo "Validating generated nginx configuration..."
        if nginx -t; then
            echo "nginx configuration validation passed"
        else
            echo "nginx configuration validation failed"
            return 1
        fi
    else
        echo "No YAML configuration files found in $yaml_dir"
        echo "nginx will use existing configuration"
    fi
}

# Auto-detect and process YAML configurations
if command -v nginx-yaml-entrypoint >/dev/null 2>&1; then
    echo "nginx-yaml-entrypoint detected, processing YAML configurations..."
    process_yaml_configs
else
    echo "nginx-yaml-entrypoint not available, skipping YAML processing"
fi
