# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder
ARG TARGETARCH
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o ./build/jigs ./cmd/jigs

# ---------- Runtime stage ----------
FROM scratch
COPY --from=builder /app/build/jigs /jigs
ENTRYPOINT ["/jigs"]
