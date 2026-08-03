# Build stage
FROM golang:1.26-alpine AS builder

# Set working directory
WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

FROM builder AS build-server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=.,target=. \
    go build -o /bin/server ./cmd/server

RUN --mount=type=cache,target=/go/pkg/mod \
    go build -o /bin/goose github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.22 AS server
RUN apk add --no-cache ca-certificates

WORKDIR /app

# 拷贝生成的二进制文件
COPY --from=build-server /bin/server ./
COPY --from=build-server /bin/goose ./

# 拷贝迁移 SQL 脚本（确保 ./migrations 文件夹在 Dockerfile 同级目录下）
COPY ./migrations ./migrations 

EXPOSE 8080
ENTRYPOINT [ "./server" ]
