
FROM golang:1.25-alpine AS builder
WORKDIR /app/
COPY [ "*.go", "go.*", "./" ]
RUN go build -ldflags='-w -s' .

FROM nginx:mainline-alpine
COPY --from=0 ["/app/nginx-yaml-entrypoint", "/usr/local/bin/"]
COPY ["./entrypoint.sh","/docker-entrypoint.d/"]
