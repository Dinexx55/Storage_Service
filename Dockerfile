FROM golang:1.21.0 AS builder

WORKDIR /storage/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o storage-service cmd/app/main.go

FROM debian:bookworm AS runner

WORKDIR /usr/bin

COPY --from=builder /storage/app/storage-service .
COPY --from=builder /storage/app/configs/ /usr/bin/configs

EXPOSE 8085

ENTRYPOINT ["storage-service"]