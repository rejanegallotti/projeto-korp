# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Cache dependency downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o http-server-projeto-korp .

# ── Stage 2: Run ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /build/http-server-projeto-korp /http-server-projeto-korp

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/http-server-projeto-korp"]
