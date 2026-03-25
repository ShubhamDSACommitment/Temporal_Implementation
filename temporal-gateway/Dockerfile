
# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY workflow-designer/package.json workflow-designer/package-lock.json* ./
RUN npm install
COPY workflow-designer/ ./
RUN npx vite build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend
WORKDIR /build
COPY shared/ ./shared/
COPY temporal-gateway/ ./temporal-gateway/
WORKDIR /build/temporal-gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /gateway ./cmd/server

# Stage 3: Final image
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=backend /gateway /gateway
COPY --from=frontend /app/dist /static
ENV STATIC_DIR=/static
ENTRYPOINT ["/gateway"]
