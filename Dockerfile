FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/.. internal/.. server/.. config/.. ./

RUN CGO_ENABLED=1 go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app /app/app

RUN apk add --no-cache libcap && \
    setcap 'cap_net_bind_service=+ep' /app/app && \
    adduser -D gatewayuser && \
    chown gatewayuser:gatewayuser /app/app && \
    ln -s /app/app /usr/bin/gateway

ENV PORT=80 \
    HOST=0.0.0.0 \
    READ_TIMEOUT=10s \
    WRITE_TIMEOUT=10s \
    LOG_LEVEL=ERROR \
    CONFIG_PATH=/usr/etc/gateway/config.yaml

EXPOSE 80

USER gatewayuser

ENTRYPOINT gateway --config="$CONFIG_PATH"
