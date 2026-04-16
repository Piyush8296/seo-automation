# PRD Gap Analysis: Current Build vs `prd.md`

## Scope

This document compares the current implementation in this repository against the product requirements described in `prd.md`.

It separates the analysis into two views:

1. Full PRD comparison
2. Feature-only website check coverage

The second view intentionally ignores infrastructure concerns such as DB choice, queues, parallelism, concurrency, tenancy, auth, and deployment architecture. It focuses only on what SEO checks/features we can run against a website today versus what the PRD expects.

---

## Executive Summary

The current repository is a strong SEO audit crawler and report generator, but it is not yet the 3-stream platform described in the PRD.

What exists today:

- CLI audit runner for a single site/domain crawl
- Configurable crawl depth, max pages, concurrency, timeout
- JSON, HTML, and Markdown reports
- Local HTTP server for starting/listing audits and serving reports
- GitHub Actions workflow for scheduled/manual audit runs
- Roughly 160 unique implemented check IDs across 22 check groups

What the PRD expects in addition:

- 322 automated checks across 17 modules
- Standardized Check ID taxonomy and per-check status model
- AIO readiness scoring
- Vertical-aware SEO expectations
- JavaScript rendering checks
- Measured Core Web Vitals via Lighthouse + DebugBear RUM/Synthetic
- Static CI gate semantics based on critical failures
- Scheduled prod scan with history/delta
- Unified dashboard across CI, Prod, and RUM streams
- Integrations such as GSC/CrUX, Jira/Linear, webhooks, white-label PDF

Bottom line:

- The current build overlaps materially with the PRD's crawl-and-audit engine.
- The current build does not yet satisfy the PRD's platform, data model, integration, and measurement requirements.
- From a pure website-check feature perspective, the repo already covers many technical SEO checks, but there are major gaps in JavaScript SEO, measured performance, AIO, vertical awareness, SSL certificate inspection, and local SEO.

---

## What We Have Today

Current product shape:

- SEO crawler and analyzer
- Issue-based reporting
- Health score and grade
- Single-run reports plus simple before/after diffing
- Basic CI threshold gate

Current check families implemented in code:

- Crawlability
- HTTPS and security headers
- Performance heuristics
- Internal linking
- Title tags
- Meta descriptions
- Content body / duplicate heuristics
- Headings
- Canonicals
- Images
- Structured data
- Social tags
- URL structure
- Mobile meta checks
- Mobile vs desktop diff checks
- International / hreflang
- Sitemap checks
- Pagination
- AMP
- Crawl budget heuristics
- Core Web Vitals heuristics
- E-E-A-T heuristics

Approximate implemented check count:

- About 160 unique check IDs found in `internal/checks`

---

## Full PRD Comparison

### Summary

| Area | PRD Expectation | Current Status | Notes |
|---|---|---|---|
| Audit engine | 322 automated checks / 17 modules | Partial | Current code implements about 160 unique check IDs |
| Report schema | Standard report with status + roadmap + AIO + manual checklist | Partial | Reports are issue-oriented, not PRD-shaped |
| Scoring | PRD-specific weighted formula | No | Current formula is different |
| AIO readiness | Required in every report | No | Not implemented |
| Vertical detection | Automotive / ecommerce / SaaS / media adaptation | No | Not implemented |
| Site-wide crawl | Required | Yes / Partial | Crawl exists, but not at PRD completeness |
| Scheduled prod scan | Required | Partial | Daily GitHub Action exists, not a productized prod scan flow |
| CI gate | Critical check based PR blocker | Partial | Threshold-based gate exists, not PRD semantics |
| Unified alerting | Slack/email/webhook + RUM | No | Not implemented |
| GSC / CrUX | Required | No | Not implemented |
| Ticketing | Jira / Linear | No | Not implemented |
| White-label export | PDF + branded output | No | Not implemented |
| Unified dashboard | CI + Prod + RUM single view | No | Not implemented |
| Unified data model | 3-stream platform store | No | Current storage is filesystem-based |

### Functional Requirements

| FR | Requirement | Status | Notes |
|---|---|---|---|
| FR-01 | Single-URL audit engine | Partial | URL input and audit exist, but not all PRD modules and no per-check PASS/FAIL/WARNING/SKIP matrix |
| FR-02 | Standard report format | Partial | Reports exist, but missing executive summary, category scores, fix roadmap, AIO section, manual checklist |
| FR-03 | Check ID + status + severity + specific fix | Partial | Check ID and severity exist; explicit status model does not; taxonomy differs from PRD |
| FR-04 | PRD health score formula | No | Current scoring model is different |
| FR-05 | AIO readiness | No | Missing |
| FR-06 | Vertical detection | No | Missing |
| FR-07 | Domain crawl + aggregation | Partial | Crawl exists; aggregation is not PRD-complete |
| FR-08 | Scheduled prod scan with deltas | Partial | Daily workflow exists, but not productized scan history/delta behavior |
| FR-09 | Unified alerting | No | Missing |
| FR-10 | GSC integration | No | Missing |
| FR-11 | GSC + CrUX historical overlay | No | Missing |
| FR-12 | Programmatic JSON API for CI gate | Partial | API exists, but not PRD gate semantics or per-check status payload |
| FR-13 | Jira / Linear ticket creation | No | Missing |
| FR-14 | White-label PDF + JSON export | No | Missing |
| FR-15 | Static CI gate block on critical failure | Partial | Threshold gate exists, but not PRD-style critical-check PR status flow |
| FR-16 | Unified dashboard | No | Missing |

### Architecture Alignment

| PRD Theme | Status | Notes |
|---|---|---|
| 3 streams: CI / Prod / RUM | No | Current code supports individual audit runs, not 3 unified streams |
| PostgreSQL unified store | No | Current implementation stores reports on local filesystem |
| Browser rendering | No | Current crawler is HTTP + HTML parsing, not Playwright/Chromium rendering |
| DebugBear integration | No | Missing |
| Lighthouse CI measurement | No | Missing |
| Multi-tenant SaaS | No | Missing |

---

## Feature-Only View: Website Checks in PRD vs What We Have

This section ignores platform and infra. It only answers:

- What website checks/features from the PRD can we do right now?
- What is only partially covered?
- What is missing?

### 1. In PRD and Currently Present

These feature areas clearly exist today in the repo and overlap with the PRD:

| PRD Area | Current Coverage |
|---|---|
| Canonical basics | Missing canonical, non-absolute canonical, insecure canonical, canonical mismatch/conflict |
| Redirect basics | Redirect chains, redirect loops, temporary-vs-permanent style issue detection |
| Basic status/crawlability | 4xx, 5xx, timeout, deep pages, robots-blocked-but-linked, noindex in sitemap |
| URL structure | Long URLs, underscores, uppercase, spaces, too many params, session params, double slash, non-descriptive URLs |
| Title checks | Missing, too short, too long, duplicate titles |
| Meta description checks | Missing, too short, too long, duplicate meta descriptions |
| Heading checks | Missing H1, multiple H1s, empty H1, short/long H1, missing H2, skipped hierarchy, duplicate H1 |
| Image checks | Missing/empty alt, filename-as-alt, missing dimensions, lazy above fold, missing srcset |
| Internal linking | Broken links, redirect links, nofollow internal links, empty/generic anchors, orphan pages, too many outlinks |
| Structured data basics | Missing JSON-LD, invalid JSON, missing `@context`, missing `@type`, duplicate type, basic Article/Product/Breadcrumb/FAQ validation |
| Mobile basics | Viewport missing/invalid, font size too small, user-scalable disabled |
| Mobile vs desktop diffs | Status/title/meta/H1/canonical/schema/link/content mismatches between desktop and mobile responses |
| Hreflang basics | Non-absolute hreflang URL, invalid language code, missing `x-default`, missing self-reference, missing return link, non-canonical hreflang URL |
| Sitemap basics | Missing sitemap, sitemap URL 4xx, redirected, noindex, blocked, too large, low coverage |
| Crawl budget basics | Tracking parameters, faceted navigation, low-value pages, sitemap/noindex conflicts, search pages indexable |

### 2. In PRD and Partially Present

These are feature areas where the repo overlaps with the PRD, but only partially.

| PRD Area | Current Status | Gap |
|---|---|---|
| Robots.txt | Partial | We detect missing/block-all/sitemap directive and use robots allow/deny, but PRD expects a richer robots check set |
| HTTPS & SSL | Partial | HTTPS/security headers and mixed content exist, but SSL certificate inspection is missing |
| Duplicate content | Partial | Near-duplicate content heuristics exist, but not the fuller PRD duplicate taxonomy |
| Structured data | Partial | Good generic coverage exists, but PRD expects broader schema coverage and vertical-aware expectations |
| Core Web Vitals | Partial | Repo has CWV heuristics, but not actual measurement-based CWV checks |
| Page speed | Partial | Response time, compression, cache, render-blocking, HTML size heuristics exist, but not PRD-style measured SPEED suite |
| Mobile responsiveness | Partial | Basic meta/mobile checks exist, but not full responsive/browser-based testing |
| Crawl efficiency / indexing | Partial | There is crawl-budget and orphan/depth logic, but not the full PRD indexing/coverage feature set |
| E-E-A-T signals | Partial | Some heuristics exist, but not in the PRD's formal AIO/content scoring model |

### 3. In PRD but Missing

These feature areas are expected by the PRD but are not currently implemented as website-check capabilities.

| Missing Feature Area | Notes |
|---|---|
| AIO Readiness rating | No `Strong / Moderate / Weak` score or recommendation set |
| AIO-specific checks | No featured-snippet readiness model, no dedicated AIO scoring layer |
| Speakable schema checks | Missing |
| Vertical-aware expectations | No automotive/ecommerce/SaaS/media-aware rules |
| Vehicle schema expectations | Missing as an explicit feature |
| SoftwareApplication schema expectations | Missing |
| Review schema expectations by vertical | Missing as explicit logic |
| JavaScript SEO checks | No Playwright/browser rendering, no SSR/hydration/rendered-content checks |
| Real CWV measurement | No Lighthouse, no field data, no real metric thresholds like LCP/INP/CLS from measured runs |
| Local SEO checks | No M17-style local SEO automation |
| SSL certificate checks | No cert expiry / issuer / chain / TLS inspection |
| Advanced index coverage checks | No PRD-style indexing/coverage suite |
| Manual checklist surfacing in report | PRD expects manual checks to appear in audit output; current reports do not include them |

---

## Module-by-Module Feature Coverage

### M1: Canonical & Duplicates

- Status: Partial
- Present:
  - Canonical basics
  - Near-duplicate content heuristics
  - Duplicate title/meta/H1 checks
- Missing:
  - PRD-scale canonical taxonomy
  - Full duplicate-content taxonomy tied to PRD Check IDs

### M2: Robots & Sitemap

- Status: Partial
- Present:
  - Robots missing/block-all/sitemap directive detection
  - Robots allow/deny usage during crawl
  - Sitemap discovery and sitemap URL validation
- Missing:
  - Rich robots issue taxonomy from PRD
  - Broader sitemap variants and coverage expected in PRD

### M3: HTTPS & SSL

- Status: Partial
- Present:
  - HTTP page detection
  - Mixed content
  - HSTS/CSP/XFO/XCTO/Referrer-Policy/Permissions-Policy checks
- Missing:
  - SSL certificate inspection
  - TLS-level checks

### M4: Redirect Chains

- Status: Partial
- Present:
  - Redirect chain and loop detection
- Missing:
  - Full PRD redirect suite

### M5: HTTP Status Codes

- Status: Partial
- Present:
  - 4xx, 5xx, timeout, broken internal links
- Missing:
  - Soft 404 logic
  - 410-specific logic
  - Rich status-code taxonomy expected in PRD

### M6: URL Structure

- Status: Strong Partial / Mostly Present
- Present:
  - Strong basic URL quality checks
- Missing:
  - PRD-specific URL taxonomy breadth

### M7: Title Tags & Meta

- Status: Strong Partial / Mostly Present
- Present:
  - Missing, length, duplicate checks
- Missing:
  - PRD's broader count and richer optimization coverage

### M8: Heading Hierarchy

- Status: Strong Partial / Mostly Present
- Present:
  - Most foundational heading checks
- Missing:
  - Full PRD breadth

### M9: Image Optimisation

- Status: Partial
- Present:
  - Alt text, dimensions, lazy above fold, srcset
- Missing:
  - PRD-level richer optimization set

### M10: Internal Linking

- Status: Partial
- Present:
  - Strong foundational internal link checks
- Missing:
  - Full PRD graph-analysis breadth

### M11: Structured Data

- Status: Partial
- Present:
  - Generic JSON-LD extraction and validation
  - Basic Article/Product/Breadcrumb/FAQ validation
- Missing:
  - Broader schema families
  - PRD taxonomy alignment
  - Vertical-adaptive schema expectations

### M12: JavaScript SEO

- Status: Missing
- Missing:
  - Browser rendering
  - JS-only content/indexability validation
  - SSR/rendered DOM comparison

### M13: Mobile Responsiveness

- Status: Partial but limited
- Present:
  - Viewport and some mobile-only HTML checks
  - Mobile-vs-desktop fetch comparison
- Missing:
  - Playwright/mobile emulation-based responsive validation
  - Touch target / nav usability / richer mobile UX checks

### M14: Core Web Vitals

- Status: Partial but limited
- Present:
  - HTML heuristics that suggest likely CWV issues
  - Some basic page speed heuristics
- Missing:
  - Lighthouse-based lab measurement
  - RUM-based CWV
  - Synthetic CWV
  - Real threshold enforcement per PRD

### M15: Crawl Efficiency

- Status: Partial
- Present:
  - Crawl-budget style heuristics
  - Orphan/depth/sitemap conflict checks
- Missing:
  - Full indexing and crawl-efficiency feature set expected in PRD

### M16: Hreflang

- Status: Strong Partial / Mostly Present
- Present:
  - Core hreflang consistency and reciprocity checks
- Missing:
  - PRD taxonomy alignment and any remaining advanced cases

### M17: Local SEO

- Status: Missing
- Missing:
  - Local SEO feature checks described in PRD

---

## Important Note: Useful Checks We Have That Are Not First-Class in the PRD

The current repo includes several useful feature checks that are not cleanly represented as standalone modules in the PRD:

- Social / Open Graph / Twitter card checks
- AMP checks
- Pagination checks
- E-E-A-T heuristics

These are not wasted work. They are useful and should either:

- be mapped into the PRD taxonomy, or
- be carried forward as "extra implementation coverage" beyond the initial PRD language

---

## Net Assessment

### If the question is:

"Do we already have the core website audit engine?"

Answer:

- Yes, for a meaningful subset of technical SEO checks.

### If the question is:

"Are we feature-complete against the PRD's website-check vision?"

Answer:

- No.

### If the question is:

"What are the biggest feature gaps, ignoring infrastructure?"

Answer:

- JavaScript SEO
- Real performance/CWV measurement
- AIO readiness
- Vertical-aware check logic
- SSL certificate inspection
- Local SEO
- Full PRD check taxonomy and check-count coverage

---

## Suggested Next Use Of This Document

This document can be used as the base for:

- a product review with engineering
- a backlog split into "taxonomy alignment" vs "new feature work"
- a Phase 1 roadmap focused only on website-check coverage before platform build-out
