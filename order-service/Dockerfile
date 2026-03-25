FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy shared module first
COPY shared/ ./shared/

# Copy order-service module
COPY order-service/ ./order-service/

WORKDIR /build/order-service
RUN go mod download
RUN CGO_ENABLED=0 go build -o /order-worker ./cmd/worker

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /order-worker /order-worker
ENTRYPOINT ["/order-worker"]
