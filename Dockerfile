# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o brok ./cmd/main.go

FROM ubuntu:22.04

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/brok .
COPY .env .env

CMD ["./brok"]
