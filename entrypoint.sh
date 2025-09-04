#!/bin/sh
# vim:sw=4:ts=4:et

set -e

if test -f /etc/nginx/yaml/main.yaml; then
    echo "Converting YAML configuration to nginx config..."
    nginx-yaml-entrypoint -f /etc/nginx/yaml/main.yaml -o /etc/nginx/nginx.conf
    echo "Configuration generated successfully"
fi
