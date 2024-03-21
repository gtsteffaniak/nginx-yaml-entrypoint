
FROM golang:1.22-alpine as builder
WORKDIR /app/
COPY [ "*.go", "go.*", "./" ]
RUN go build -ldflags='-w -s' .

FROM nginx:mainline-alpine
COPY --from=0 ["/app/nginx-yaml-entrypoint", "/usr/local/bin/"]
COPY ["./entrypoint.sh","/docker-entrypoint.d/"]
