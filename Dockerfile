FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o goapp cmd/server/main.go 

FROM debian:slim AS runner

WORKDIR /app

COPY --from=builder /app/goapp .

COPY frontend ./frontend

EXPOSE 8080

CMD ./goapp
