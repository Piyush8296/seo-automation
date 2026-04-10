# SEO Audit

Production-grade technical SEO crawler with 175+ checks modelled after Screaming Frog and SEMrush Site Audit. Run it locally, in GitHub Actions, or as a Claude agent.

---

## Quick start

```bash
# Build
go build -o seo-audit .

# Audit a site
./seo-audit audit --url https://www.cars24.com
```

Reports are written to `./reports/` — open `report.html` in a browser for the full interactive dashboard.

---

## Modes

### 1. Local CLI

```bash
./seo-audit audit --url <URL> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--url` | *(required)* | Site to crawl |
| `--max-depth` | `-1` (unlimited) | How many link hops from the start URL |
| `--max-pages` | `0` (unlimited) | Stop after N pages crawled |
| `--concurrency` | `5` | Parallel crawl workers (raise for speed, lower for polite crawling) |
| `--timeout` | `30s` | Per-request timeout — e.g. `15s`, `1m` |
| `--format` | `json,html,markdown` | Comma-separated output formats |
| `--output-dir` | `./reports` | Where to write report files |
| `--sitemap` | *(auto-detected)* | Override sitemap URL if auto-detection fails |
| `--platform` | *(not set)* | `desktop`, `mobile`, or omit for bifurcated report showing both |
| `--no-mobile-check` | `false` | Skip the mobile vs desktop comparison (same as `--platform desktop`) |
| `--exit-code` | `false` | Exit 1 if any errors are found (useful in CI) |

**Examples:**

```bash
# Quick shallow audit — top 3 levels, max 100 pages
./seo-audit audit --url https://example.com --max-depth 3 --max-pages 100

# Deep full-site crawl with higher concurrency
./seo-audit audit --url https://example.com --max-depth -1 --max-pages 0 --concurrency 10

# Default: bifurcated report with desktop + mobile scores side by side
./seo-audit audit --url https://example.com

# Desktop only (skips mobile fetch — faster)
./seo-audit audit --url https://example.com --platform desktop

# Mobile focus — shows mobile-specific and M↔D diff issues only
./seo-audit audit --url https://example.com --platform mobile

# JSON only, custom timeout
./seo-audit audit --url https://example.com --format json --timeout 15s

# Fail the process if any errors exist (for CI gates)
./seo-audit audit --url https://example.com --exit-code
```

---

### 2. Compare two audits (`diff`)

Track regressions and improvements between runs:

```bash
./seo-audit diff reports/before.json reports/after.json
```

Shows:
- **New issues** — problems introduced since the last run
- **Resolved issues** — problems that were fixed
- **Persisting issues** — still present in both reports

**Tip:** Rename the report before re-running so you can compare:
```bash
cp reports/report.json reports/before.json
./seo-audit audit --url https://example.com
./seo-audit diff reports/before.json reports/report.json
```

---

### 3. CI threshold gate (`check-exit`)

Use after an audit to fail a build if too many issues exist:

```bash
./seo-audit check-exit --report reports/report.json --max-errors 0 --max-warnings 50
```

| Flag | Default | Description |
|---|---|---|
| `--report` | *(required)* | Path to `report.json` |
| `--max-errors` | `0` | Exit 1 if errors exceed this count |
| `--max-warnings` | `50` | Exit 1 if warnings exceed this count |

---

### 4. GitHub Actions

The workflow runs automatically every day at 02:00 UTC, and can also be triggered manually via **Actions → SEO Audit → Run workflow**.

**Manual dispatch parameters:**

| Parameter | Default | Description |
|---|---|---|
| `url` | `https://www.cars24.com` | Site to audit |
| `max_depth` | `3` | Max crawl depth |
| `max_pages` | `500` | Max pages to crawl |
| `concurrency` | `5` | Parallel workers |
| `mobile_check` | `true` | Enable mobile vs desktop comparison |
| `fail_on_errors` | `false` | Fail the workflow if SEO errors are found |

Reports are uploaded as a workflow artifact (`seo-reports-<run_number>`) and retained for 10 days.

To change the schedule, edit the cron expression in `.github/workflows/seo-audit.yml`:
```yaml
schedule:
  - cron: '0 2 * * *'   # Daily at 02:00 UTC
```

---

### 5. Claude Agent

The `seo-audit` agent is defined in `.claude/agents/seo-audit.md`. Claude will automatically delegate to it when you ask about SEO auditing.

**How to invoke:**

Just ask Claude naturally — no special syntax needed:

```
run seo audit on https://www.cars24.com
audit https://example.com for seo
check seo issues on https://example.com
technical seo audit https://example.com
```

**Passing parameters to the agent:**

```
audit https://example.com with max-depth 5 and max-pages 500
audit https://example.com, skip mobile check, concurrency 10
deep crawl https://example.com — unlimited depth, no page limit
quick audit https://example.com — depth 2, max 50 pages, json only
```

The agent understands natural language parameters and maps them to the correct CLI flags. It then:
1. Builds the binary if needed
2. Runs the crawl
3. Reads the JSON output
4. Returns a prioritized expert report with actionable fix recommendations

---

## Output files

| File | Description |
|---|---|
| `reports/report.html` | Interactive dashboard — open in browser |
| `reports/report.json` | Machine-readable full report |
| `reports/report.md` | Markdown summary for docs / PR comments |

---

## Checks (175+ across 22 categories)

| Category | Checks | What's covered |
|---|---|---|
| Crawlability | 15 | 4xx/5xx, redirect chains, noindex conflicts, robots.txt blocks, orphan pages |
| HTTPS & Security | 10 | Mixed content, HSTS, CSP, X-Frame-Options, Referrer-Policy |
| Performance | 12 | Response time, HTML size, compression, cache headers, render-blocking resources |
| Internal Linking | 10 | Broken links, redirect targets, nofollow, empty anchors, orphan pages |
| Titles | 5 | Missing, too short (<10 chars), too long (>60 chars), duplicates |
| Meta Descriptions | 5 | Missing, too short (<50 chars), too long (>160 chars), duplicates |
| Content | 6 | Thin content, lorem ipsum, title = H1, near-duplicate pages |
| Headings | 8 | H1 missing/multiple/empty, hierarchy skipped, duplicates |
| Canonical | 5 | Missing, non-absolute, insecure, points elsewhere, og:url conflict |
| Images | 7 | Alt missing/empty, filename as alt, dimensions missing, srcset missing |
| Structured Data | 10 | Invalid JSON-LD, missing required fields (Article, Product, Breadcrumb, FAQ) |
| Social / OG | 8 | og:title/description/image/url missing, Twitter card missing |
| URL Structure | 8 | Too long, underscores, uppercase, session params, double slashes |
| Mobile | 4 | Viewport missing/invalid, font size, user-scalable=no |
| Mobile vs Desktop | 10 | Title/meta/H1/canonical/schema mismatch between mobile and desktop responses |
| International | 7 | Non-absolute hreflang, invalid BCP47 lang codes, missing x-default, no self-ref |
| Sitemap | 8 | Missing sitemap, 4xx/redirected/noindex URLs in sitemap, coverage <80% |
| Pagination | 6 | Self-canonical on paginated pages, thin paginated content, noindex |
| AMP | 3 | AMP missing canonical, canonical points to AMP, regular page missing amphtml |
| Crawl Budget | 6 | Tracking params, faceted nav, internal search pages indexed, sitemap+noindex conflict |
| Core Web Vitals | 7 | LCP image not preloaded, CLS from fonts/images, FID from blocking scripts |
| E-E-A-T | 8 | Author info missing, article dates missing, thin About/Contact/Privacy pages |

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
- Internet access from the machine running the audit
