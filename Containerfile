FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gochatty ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/gochatty .

EXPOSE 8080

CMD ["./gochatty"]
