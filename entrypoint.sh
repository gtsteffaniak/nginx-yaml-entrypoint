#!/bin/sh
# vim:sw=4:ts=4:et

set -e

if test -f /etc/nginx/yaml/main.yaml; then
    nginx-yaml-entrypoint -f /etc/nginx/yaml/main.yaml
fi
