FROM registry.octagon.home/base-images/golang-base-image:latest AS builder

WORKDIR /app

# Copy shared module so the replace directive (../shared) resolves correctly
COPY shared/ ./shared/

WORKDIR /app/core

COPY core/go.mod core/go.sum* ./
RUN go mod download

COPY core/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o core ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata bash coreutils curl
WORKDIR /opt/core

COPY --from=builder /app/core/core .
COPY --from=builder /app/core/migrations ./migrations

EXPOSE 8080

CMD ["./core"]
