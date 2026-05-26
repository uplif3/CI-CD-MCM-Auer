# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api

# Runtime stage — distroless/static for minimal attack surface
# No shell, no package manager, no Alpine packages (musl, busybox)
# Replaces alpine:3.19 which had 2 HIGH + 5 MEDIUM + 3 LOW vulnerabilities
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /api-server .

EXPOSE 8080

ENTRYPOINT ["/app/api-server"]
