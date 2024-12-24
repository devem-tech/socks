FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -o bin/entrypoint cmd/main.go

FROM alpine:latest

WORKDIR /opt

COPY --from=builder /opt/bin/entrypoint .

CMD ["./entrypoint"]
