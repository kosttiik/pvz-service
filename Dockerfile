FROM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/server

FROM alpine:3.21.3

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
