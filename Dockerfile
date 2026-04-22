FROM golang:1.23-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# SQLite driver needs CGO enabled.
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/lambdavault ./cmd/api

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/lambdavault /app/lambdavault

EXPOSE 8080

CMD ["/app/lambdavault"]
