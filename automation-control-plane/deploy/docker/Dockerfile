FROM golang:1.21-alpine AS builder

WORKDIR /build

# Build context is at demo root, so we need to copy from automation-control-plane
# Copy go.mod first for better caching
COPY automation-control-plane/go.mod ./
# Copy go.sum if it exists (optional)
COPY automation-control-plane/go.sum* ./
# Download dependencies
RUN go mod download

# Copy all source code
COPY automation-control-plane .

# Generate go.sum if missing and build
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o control-plane ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /build/control-plane .

EXPOSE 8080

CMD ["./control-plane"]
