# build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main ./cmd/server

# final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 5080 5081

ENTRYPOINT ["./main"]