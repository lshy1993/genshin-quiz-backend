# Build stage
FROM golang:1.25.0-alpine AS builder

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
    --mount=type=bind,source=.,target=. \
    go build -o /bin/console ./cmd/console

FROM alpine:3.22 AS server
COPY --from=build-server /bin/server /bin/
COPY --from=build-server /bin/console /bin/
ENTRYPOINT [ "/bin/server" ]