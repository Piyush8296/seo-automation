# SEO Automation Tool — Implementation Plan

## Overview
Production-grade Go-based SEO crawler modelled after Screaming Frog + SEMrush Site Audit.
**146 checks across 17 categories.** Runs daily in GitHub Actions, supports manual trigger,
outputs JSON + HTML + Markdown reports, structured as a Claude skill.

---

## Project Structure

```
seo-automation/
├── main.go
├── go.mod                          (module github.com/cars24/seo-automation)
├── go.sum
├── PLAN.md                         (this file)
├── SKILL.md                        (Claude Code skill definition)
├── README.md
├── cmd/
│   ├── root.go                     cobra root + persistent flags
│   ├── audit.go                    `audit` subcommand
│   ├── diff.go                     `diff` — compare two JSON reports
│   └── check_exit.go               `check-exit` — CI pass/fail gating
├── internal/
│   ├── models/
│   │   └── models.go               All shared Go structs
│   ├── crawler/
│   │   ├── crawler.go              BFS orchestrator + worker pool
│   │   ├── worker.go               Per-URL fetch goroutine
│   │   ├── fetcher.go              net/http client, retries, redirect chain
│   │   ├── robots.go               robots.txt fetch, sync.Map cache, singleflight
│   │   ├── sitemap.go              XML sitemap (index + urlset + gzip)
│   │   └── normalize.go            URL normalisation, dedup key
│   ├── parser/
│   │   ├── page.go                 Master extractor — calls all sub-parsers
│   │   ├── meta.go                 <meta>, canonical, viewport, robots meta
│   │   ├── links.go                <a>, <link rel>, internal/external classification
│   │   ├── schema.go               <script type="application/ld+json"> extraction
│   │   └── headers.go              HTTP response header capture
│   ├── checks/
│   │   ├── interface.go            Check + SiteWideCheck interfaces
│   │   ├── registry.go             Register all; RunPage / RunSiteWide
│   │   ├── crawlability/           15 checks
│   │   ├── https_security/         10 checks
│   │   ├── performance/            12 checks
│   │   ├── internal_linking/       10 checks
│   │   ├── content_title/          5 checks
│   │   ├── content_meta_desc/      5 checks
│   │   ├── content_body/           6 checks
│   │   ├── headings/               8 checks
│   │   ├── canonical/              6 checks
│   │   ├── images/                 8 checks
│   │   ├── structured_data/        10 checks
│   │   ├── social/                 8 checks
│   │   ├── url_structure/          8 checks
│   │   ├── mobile/                 4 checks
│   │   ├── mobile_desktop/         10 checks (dual-fetch mobile vs desktop)
│   │   ├── international/          7 checks
│   │   └── sitemap/                8 checks
│   └── report/
│       ├── generator.go            Orchestrates all output formats
│       ├── json.go                 JSON marshal
│       ├── html.go                 html/template renderer
│       ├── markdown.go             Markdown writer
│       └── templates/
│           └── report.html.tmpl    Self-contained HTML (no CDN)
└── .github/
    └── workflows/
        └── seo-audit.yml
```

---

## Go Dependencies

```
github.com/PuerkitoBio/goquery v1.9.2   jQuery-like CSS selectors
github.com/spf13/cobra v1.8.1           CLI framework
github.com/temoto/robotstxt v1.1.2      robots.txt parser
golang.org/x/net v0.26.0               HTML5 tokenizer (transitive)
golang.org/x/sync v0.7.0              singleflight
```

---

## Crawl Depth Logic

| --max-depth | Behaviour |
|---|---|
| -1 (default) | Unlimited — crawl everything reachable, bounded by --max-pages |
| 0 | Seed URL only — no link-following |
| 1 | Seed + direct links (1 hop) |
| N | N hops of link-following |

---

## SEO Checks — 146 Total

### 1. Crawlability (15)
crawl.response.4xx / 5xx / timeout · crawl.redirect.chain / loop / 302_permanent
crawl.robots.page_blocked_but_linked / txt_missing / blocks_all / missing_sitemap_directive / resource_blocked
crawl.noindex.in_sitemap / has_inlinks · crawl.page_depth.too_deep / orphan_external_only

### 2. HTTPS & Security (10)
https.page_insecure / mixed_content / http_not_redirecting
security.hsts.missing / max_age_too_short · security.csp.missing · security.x_frame.missing
security.xcto.missing · security.referrer_policy.missing · security.permissions_policy.missing

### 3. Performance (12)
perf.response_time.slow / critical · perf.html_size.large · perf.compression.missing
perf.cache_control.missing · perf.render_blocking.scripts / css_count · perf.external_scripts.count
perf.lcp_candidate.lazy_loaded · perf.cls_risk.images_no_dimensions
perf.missing_preconnect · perf.inline_css.large

### 4. Internal Linking (10)
links.internal.broken_4xx / broken_5xx / to_redirect / nofollow
links.external.broken / missing_noopener · links.anchor.empty / generic
links.page.orphan (site-wide) · links.page.too_many_outlinks

### 5. Titles (5)
title.missing / empty / too_short / too_long / duplicate (site-wide)

### 6. Meta Descriptions (5)
meta_desc.missing / empty / too_short / too_long / duplicate (site-wide)

### 7. Content Body (6)
body.very_thin / thin / lorem_ipsum · body.near_duplicate (site-wide)
body.title_equals_h1 · body.noindex_meta

### 8. Headings (8)
headings.h1.missing / multiple / empty / too_short / too_long / duplicate (site-wide)
headings.h2.missing · headings.hierarchy.skipped_level

### 9. Canonical (6)
canonical.missing / not_absolute / insecure / points_elsewhere / chain / conflict_og_url

### 10. Images (8)
images.alt.missing / empty_non_decorative / too_long / is_filename
images.dimensions.missing · images.src.broken · images.lazy.above_fold · images.missing_srcset

### 11. Structured Data (10)
schema.jsonld.missing / invalid_json / missing_context / missing_type / duplicate_type
schema.article.missing_fields / product.missing_fields / breadcrumb.invalid
schema.faq.invalid · schema.organization.missing_homepage

### 12. Social (8)
og.title.missing / description.missing / image.missing / url.missing / url.mismatch_canonical
twitter.card.missing / title.missing / image.missing

### 13. URL Structure (8)
url.too_long / has_underscores / has_uppercase / has_spaces / has_session_params
url.too_many_params / double_slash / non_descriptive

### 14. Mobile (4)
mobile.viewport.missing / invalid · mobile.font_size.too_small · mobile.user_scalable.disabled

### 15. Mobile vs Desktop (10) — dual-fetch
mob_desk.content_mismatch / title_mismatch / meta_desc_mismatch / h1_mismatch
mob_desk.canonical_mismatch / schema_mismatch / links_mismatch / status_mismatch
mob_desk.separate_mobile_site / og_image_mismatch

### 16. International / hreflang (7)
i18n.hreflang.url_not_absolute / invalid_lang_code / missing_x_default / missing_self_ref
i18n.hreflang.missing_return_link (site-wide) / non_canonical_url / points_to_redirect

### 17. Sitemap (8)
sitemap.missing / invalid_xml / too_large / url_4xx / url_redirected
sitemap.url_noindex / url_blocked · sitemap.coverage_low (site-wide)

---

## Health Score Formula
```
score = 100 - ((errors × 10 + warnings × 3 + notices × 1) × 100 / totalChecksRun)
```
Grade: A≥90, B≥80, C≥70, D≥50, F<50

---

## CLI
```
seo-audit audit --url URL [--sitemap URL] [--max-depth -1] [--max-pages 0]
                [--concurrency 5] [--timeout 30s] [--no-mobile-check]
                [--format json,html,markdown] [--output-dir ./reports] [--exit-code]

seo-audit diff before.json after.json

seo-audit check-exit --report report.json [--max-errors 0] [--max-warnings 50]
```

---

## GitHub Actions
- Schedule: `0 2 * * *` (daily 02:00 UTC)
- Manual: `workflow_dispatch` with url, max_depth, max_pages, concurrency, mobile_check, fail_on_errors inputs
- Artifacts: JSON + HTML + Markdown, retained 90 days
