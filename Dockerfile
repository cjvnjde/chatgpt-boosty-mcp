# syntax=docker/dockerfile:1.27.0@sha256:bde3983e9c939224420ddaf6b784cc30e09b035a4dea01f581230c50809f372e
FROM golang:1.27.1-trixie@sha256:9baa6b4187bbb98d240372a8a235ac0bb6b5ddd52bba1431dc2f7c0705862728 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/boosty-mcp ./cmd/boosty-mcp

FROM debian:stable-slim@sha256:04634311a8d5fc442b6eb06d792293c4f3e2268652ca7634e00ce8ef5cc0a28a
RUN mkdir -p /app/data /home/app && chown -R 10001:10001 /app /home/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/boosty-mcp /usr/local/bin/boosty-mcp
ENV HOME=/home/app
WORKDIR /app
USER 10001:10001
EXPOSE 8080 8081
ENTRYPOINT ["/usr/local/bin/boosty-mcp"]
CMD ["serve"]
