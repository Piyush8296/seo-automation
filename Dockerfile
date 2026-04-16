# ─── Stage 1: build frontend ─────────────────────────────────────────────────
FROM node:20-alpine AS ui-builder

WORKDIR /ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ .
RUN npm run build

# ─── Stage 2: build backend ─────────────────────────────────────────────────
# Use Debian-based image — ca-certificates and tzdata are already present,
# no apk/apt install needed, avoids TLS verification issues in restricted networks.
FROM golang:1.22 AS builder

WORKDIR /build

# Copy dependency manifests + pre-vendored sources — zero network calls in Docker.
# Run `go mod vendor` locally whenever go.mod changes, then rebuild the image.
COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -mod=vendor -ldflags="-s -w" -trimpath -o seo-audit .

# ─── Stage 3: minimal runtime ───────────────────────────────────────────────
FROM scratch

# Copy CA certs and timezone data needed for HTTPS crawling
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo                 /usr/share/zoneinfo
COPY --from=builder /build/seo-audit                    /seo-audit

# Copy built frontend assets
COPY --from=ui-builder /ui/dist /ui/dist

# Reports volume — mount a named volume to persist audits across restarts
VOLUME ["/data/reports"]

EXPOSE 8080

ENTRYPOINT ["/seo-audit", "serve"]
CMD ["--port", "8080", "--reports-dir", "/data/reports", "--ui-dir", "/ui/dist"]
