# SEO Audit

Production-grade technical SEO crawler with 175+ checks modelled after Screaming Frog and SEMrush Site Audit. Includes a React dashboard for starting audits, tracking progress in real-time, and viewing results.

---

## Quick start

**Prerequisites:** Go 1.22+, Node.js 18+

```bash
# Install frontend dependencies
make setup

# Build everything and start the server
make run
```

Open **http://localhost:8080** in your browser.

---

## Running the project

### Option 1: Make (recommended)

| Command | Description |
|---|---|
| `make setup` | Install frontend npm dependencies |
| `make run` | Build Go + React, start the server on `:8080` |
| `make dev` | Start Go server + Vite dev server (hot reload) |
| `make docker` | Build and run with Docker Compose |
| `make clean` | Remove build artifacts |

### Option 2: Manual

```bash
# Build the Go binary
go build -o seo-audit .

# Build the React frontend
cd ui && npm install && npm run build && cd ..

# Start the server (API + UI on one port)
./seo-audit serve --port 8080 --ui-dir ui/dist
```

### Option 3: Docker

```bash
docker compose up --build
# → http://localhost:8080
```

No Go or Node.js needed — everything builds inside Docker.

### Development mode

Run the Go backend and Vite dev server separately for hot reload:

```bash
# Terminal 1: Go backend
go build -o seo-audit . && ./seo-audit serve --port 8080

# Terminal 2: Vite dev server (proxies /api → :8080)
cd ui && npm run dev
# → http://localhost:5173
```

Or just run `make dev` to start both.

---

## Web UI

The dashboard at **http://localhost:8080** lets you:

- **Start audits** — enter a URL with config options (depth, concurrency, timeout, platform)
- **Track progress** — real-time updates via SSE with live page count and current URL
- **View results** — health scores with A–F grades (Overall, Desktop, Mobile), error/warning/notice counts
- **Browse reports** — embedded interactive HTML reports with download and fullscreen
- **Compare audits** — diff two runs to track improvements or regressions
- **Manage history** — re-run, delete, or compare past audits

---

## CLI

### Audit a site

```bash
./seo-audit audit --url https://www.cars24.com
```

Reports are written to `./reports/` — open `report.html` in a browser for the full interactive dashboard.

| Flag | Default | Description |
|---|---|---|
| `--url` | *(required)* | Site to crawl |
| `--max-depth` | `-1` (unlimited) | How many link hops from the start URL |
| `--max-pages` | `0` (unlimited) | Stop after N pages crawled |
| `--concurrency` | `5` | Parallel crawl workers |
| `--timeout` | `30s` | Per-request timeout — e.g. `15s`, `1m` |
| `--format` | `json,html,markdown` | Comma-separated output formats |
| `--output-dir` | `./reports` | Where to write report files |
| `--sitemap` | *(auto-detected)* | Override sitemap URL |
| `--platform` | *(not set)* | `desktop`, `mobile`, or omit for both |
| `--no-mobile-check` | `false` | Skip mobile vs desktop comparison |
| `--exit-code` | `false` | Exit 1 if any errors are found (CI) |

**Examples:**

```bash
# Quick shallow audit
./seo-audit audit --url https://example.com --max-depth 3 --max-pages 100

# Deep full-site crawl
./seo-audit audit --url https://example.com --concurrency 10

# Desktop only (faster)
./seo-audit audit --url https://example.com --platform desktop

# Fail in CI if errors exist
./seo-audit audit --url https://example.com --exit-code
```

### HTTP server

```bash
./seo-audit serve [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | HTTP server port |
| `--reports-dir` | `~/.seo-reports` | Root directory for audit reports |
| `--ui-dir` | *(empty)* | Path to built frontend assets (e.g. `ui/dist`) |

### Compare two audits (`diff`)

```bash
./seo-audit diff reports/before.json reports/after.json
```

Shows new issues, resolved issues, and persisting issues between runs.

### CI threshold gate (`check-exit`)

```bash
./seo-audit check-exit --report reports/report.json --max-errors 0 --max-warnings 50
```

---

## GitHub Actions

The workflow runs daily at 02:00 UTC and can be triggered manually via **Actions → SEO Audit → Run workflow**.

| Parameter | Default | Description |
|---|---|---|
| `url` | `https://www.cars24.com` | Site to audit |
| `max_depth` | `3` | Max crawl depth |
| `max_pages` | `500` | Max pages to crawl |
| `concurrency` | `5` | Parallel workers |
| `mobile_check` | `true` | Enable mobile comparison |
| `fail_on_errors` | `false` | Fail workflow on SEO errors |

Reports are uploaded as workflow artifacts and retained for 10 days.

---

## Claude Agent

Ask Claude naturally — no special syntax needed:

```
run seo audit on https://www.cars24.com
audit https://example.com with max-depth 5 and max-pages 500
```

The agent builds the binary, runs the crawl, and returns a prioritized report with fix recommendations.

---

## Project structure

```
seo-automation/
├── cmd/                    # CLI commands (audit, serve, diff, check-exit)
├── internal/
│   ├── server/             # HTTP API, SSE, audit lifecycle, storage
│   ├── crawler/            # BFS crawl engine, workers, robots.txt
│   ├── parser/             # HTML extraction (meta, links, schema, headers)
│   ├── checks/             # 175+ SEO checks across 22 categories
│   ├── models/             # Data structures
│   └── report/             # HTML, JSON, Markdown report generation
├── ui/                     # React frontend (Vite + Tailwind)
│   ├── src/
│   │   ├── pages/          # Home, AuditDetail
│   │   ├── components/     # AuditForm, CrawlProgress, ScoreCard, etc.
│   │   ├── hooks/          # useSSE (Server-Sent Events)
│   │   └── lib/            # API client, utilities
│   └── ...
├── Makefile                # Build & run commands
├── Dockerfile              # Multi-stage build (Node + Go → scratch)
└── docker-compose.yml      # Single-service deployment
```

---

## Output files

| File | Description |
|---|---|
| `reports/report.html` | Interactive dashboard — open in browser |
| `reports/report.json` | Machine-readable full report |
| `reports/report.md` | Markdown summary for docs / PR comments |

---

## Checks (175+ across 22 categories)

| Category | # | What's covered |
|---|---|---|
| Crawlability | 15 | 4xx/5xx, redirect chains, noindex conflicts, robots.txt blocks, orphan pages |
| HTTPS & Security | 10 | Mixed content, HSTS, CSP, X-Frame-Options, Referrer-Policy |
| Performance | 12 | Response time, HTML size, compression, cache headers, render-blocking |
| Internal Linking | 10 | Broken links, redirect targets, nofollow, empty anchors, orphan pages |
| Titles | 5 | Missing, too short/long, duplicates |
| Meta Descriptions | 5 | Missing, too short/long, duplicates |
| Content | 6 | Thin content, lorem ipsum, title = H1, near-duplicate pages |
| Headings | 8 | H1 missing/multiple/empty, hierarchy skipped, duplicates |
| Canonical | 5 | Missing, non-absolute, insecure, points elsewhere, og:url conflict |
| Images | 7 | Alt missing/empty, filename as alt, dimensions missing, srcset missing |
| Structured Data | 10 | Invalid JSON-LD, missing required fields |
| Social / OG | 8 | og:title/description/image/url missing, Twitter card missing |
| URL Structure | 8 | Too long, underscores, uppercase, session params, double slashes |
| Mobile | 4 | Viewport missing/invalid, font size, user-scalable=no |
| Mobile vs Desktop | 10 | Title/meta/H1/canonical/schema mismatch |
| International | 7 | Non-absolute hreflang, invalid BCP47, missing x-default |
| Sitemap | 8 | Missing sitemap, 4xx/redirected/noindex URLs, coverage <80% |
| Pagination | 6 | Self-canonical on paginated pages, thin content, noindex |
| AMP | 3 | Missing canonical, canonical points to AMP |
| Crawl Budget | 6 | Tracking params, faceted nav, internal search indexed |
| Core Web Vitals | 7 | LCP image not preloaded, CLS from fonts/images, FID blocking scripts |
| E-E-A-T | 8 | Author info missing, article dates missing, thin About/Contact/Privacy |

---

## Health score

```
score = 100 − ((errors × 10 + warnings × 3 + notices × 1) × 100 / totalChecks)
```

| Grade | Score |
|---|---|
| A | ≥ 90 |
| B | ≥ 80 |
| C | ≥ 70 |
| D | ≥ 50 |
| F | < 50 |

---

## Requirements

- Go 1.22+
- Node.js 18+ (for frontend build)
- Internet access from the local machine running the audit
