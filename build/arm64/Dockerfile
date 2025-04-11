FROM golang:1.23 AS builder

RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

ENV GOOS=linux \
    GOARCH=arm64 \
    CC=aarch64-linux-gnu-gcc \
    CGO_ENABLED=1

RUN go clean -cache -modcache && go build -ldflags "-linkmode external -extldflags -static" -o myapp_arm64 .

FROM alpine:latest

COPY --from=builder /app/myapp_arm64 /myapp_arm64