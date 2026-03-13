# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder
RUN mkdir /emptydir # Used to make an empty directory in the runtime stage
ARG TARGETARCH
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o ./build/jigs ./cmd/jigs

# ---------- Runtime stage (distroless) ----------
FROM scratch
WORKDIR /mnt
COPY --from=builder /emptydir /mnt
COPY --from=builder /app/build/jigs /jigs
ENTRYPOINT ["/jigs"]
