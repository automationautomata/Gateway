FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 go build -o app cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app /app/app

RUN ln -s /app/app /usr/local/bin/gateway
