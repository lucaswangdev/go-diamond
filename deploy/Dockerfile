FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-diamond ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/go-diamond .
COPY deploy/config.yaml.example /app/config.yaml

EXPOSE 8080

CMD ["./go-diamond", "-config", "/app/config.yaml"]