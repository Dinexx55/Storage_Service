FROM golang:1.21.0 AS builder

WORKDIR /storage/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/app/main.go

FROM golang:1.21.0

COPY --from=builder /storage/app/main /
COPY --from=builder /storage/app/configs/ /configs

EXPOSE 8080

CMD /main