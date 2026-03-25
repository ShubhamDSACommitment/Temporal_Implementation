FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy shared module first
COPY shared/ ./shared/

# Copy payment-service module
COPY payment-service/ ./payment-service/

WORKDIR /build/payment-service
RUN go mod download
RUN CGO_ENABLED=0 go build -o /payment-worker ./cmd/worker

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /payment-worker /payment-worker
ENTRYPOINT ["/payment-worker"]
