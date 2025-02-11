FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o exporter cmd/exporter/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/exporter /app/exporter

ENTRYPOINT ["/app/exporter"]

