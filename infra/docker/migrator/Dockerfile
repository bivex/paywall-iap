# Multi-stage build for Migrator service
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make binutils

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -a -installsuffix cgo \
    -o migrator ./cmd/migrator && \
    strip /app/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/migrator .
COPY backend/migrations ./migrations

ENTRYPOINT ["./migrator"]
CMD ["migrate", "up"]
