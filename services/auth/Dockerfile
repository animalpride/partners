FROM registry.octagon.home/base-images/golang-base-image:latest AS builder

# Ensure Go toolchain matches module requirement
ENV GOTOOLCHAIN=go1.24.12

WORKDIR /app

# Copy shared module so the replace directive (../shared) resolves correctly
COPY shared/ ./shared/

# Service lives one level down, preserving the ../shared relative replace path
WORKDIR /app/auth

# Copy go mod files first to leverage caching
COPY auth/go.mod auth/go.sum ./
RUN go mod download

# Copy source code
COPY auth/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata bash curl
WORKDIR /opt/auth

COPY --from=builder /app/auth/auth .
COPY --from=builder /app/auth/migrations ./migrations

EXPOSE 8080

CMD ["./auth"]
