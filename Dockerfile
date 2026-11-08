# ==============================================================================
# Stage 1: Build React SPA Frontend
# ==============================================================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# ==============================================================================
# Stage 2: Compile Go Backend Binary
# ==============================================================================
FROM golang:1.24-bookworm AS backend-builder
WORKDIR /app

COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/embed.go ./web/

# Inject compiled frontend into web/dist for go:embed
COPY --from=frontend-builder /app/web/dist ./web/dist

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/jazz-tools ./cmd/jazz-tools

# ==============================================================================
# Stage 3: Runtime Container
# ==============================================================================
FROM debian:bookworm-slim

# Install LilyPond for companion PDF generation, fonts, and CA certificates
RUN apt-get update && apt-get install -y --no-install-recommends \
    lilypond \
    fonts-dejavu-core \
    ca-certificates \
    fontconfig \
    && rm -rf /var/lib/apt/lists/* \
    && fc-cache -f

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/jazz-tools /app/jazz-tools

# Cloud Run injects the PORT environment variable (default: 8080)
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/jazz-tools"]
CMD ["serve"]
