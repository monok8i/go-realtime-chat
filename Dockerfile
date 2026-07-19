FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /build/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /build/worker ./cmd/worker

FROM alpine:3.21 AS api

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/api .

EXPOSE 8000

CMD ["./api"]

FROM alpine:3.21 AS worker

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/worker .

CMD ["./worker"]
