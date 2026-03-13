# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o ./build/jigs ./cmd/jigs

# ---------- Runtime stage ----------
FROM scratch
COPY --from=builder /app/build/jigs /jigs
ENTRYPOINT ["/jigs"]
