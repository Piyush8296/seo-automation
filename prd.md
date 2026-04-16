# PRD: SEO Audit Automation Platform

**Version:** 1.1  
**Date:** April 12, 2026  
**Author:** SEO Product Team  
**Status:** Draft — Pending Engineering Review

> **v1.1 changes:** (1) Added 3-stream unified platform architecture (Static CI Gate + Periodic Prod Scan + DebugBear RUM); (2) Replaced CrUX real-time dependency with proactive DebugBear-native stack; (3) Added FR-15 (CI Gate blocker), FR-16 (Unified Dashboard); (4) Added Appendix C — Check ID reference map for engineering teams.

---

## Table of Contents

1. [Executive Summary & Problem Statement](#1-executive-summary--problem-statement)
2. [Goals & Success Metrics](#2-goals--success-metrics)
3. [User Personas](#3-user-personas)
4. [Functional Requirements](#4-functional-requirements)
5. [The 17 Automation Modules](#5-the-17-automation-modules)
6. [Non-Functional Requirements](#6-non-functional-requirements)
7. [Out of Scope (v1)](#7-out-of-scope-v1)
8. [Appendix A — Manual & Monitoring Checklist](#appendix-a--manual--monitoring-checklist)
9. [Appendix B — Suggested Tech Stack](#appendix-b--suggested-tech-stack)
10. [Appendix C — Check ID Reference Map](#appendix-c--check-id-reference-map)

---

## 1. Executive Summary & Problem Statement

### The Problem

SEO audits today are slow, expensive, and inconsistent. A thorough audit of a large site — covering canonical tags, structured data, Core Web Vitals, JavaScript rendering, mobile responsiveness, and more — takes a skilled SEO engineer 8–40 hours of manual work. That work is not repeatable, not version-controlled, and not actionable at scale.

Specifically:

- **Inconsistency**: Two SEO auditors reviewing the same site will check different things and produce incompatible reports. There is no standard format, no Check ID system, no shared taxonomy.
- **Speed**: By the time a manual audit is complete, the site may have already shipped the next deploy. Issues found are stale on arrival.
- **Coverage**: Manual auditors routinely skip lower-priority checks when under time pressure, leaving known-dangerous issues (redirect chains, index bloat, broken hreflang) undetected.
- **Cost**: At typical agency rates, a comprehensive 400-point audit costs $3,000–$12,000. Most teams run them quarterly at best.

### The Opportunity

Our internal 407-point SEO checklist — built from years of auditing automotive and e-commerce marketplaces — reveals a critical insight:

| Check Type | Count | Automatable? |
|---|---|---|
| Crawl (HTML/HTTP parsing) | 263 | ✅ Yes |
| Performance (Lighthouse/CrUX) | 53 | ✅ Yes |
| API (SSL certificate inspection) | 6 | ✅ Yes |
| Monitoring (recurring alerts) | 14 | 📡 Semi-automated |
| Manual (requires human + GSC/Ahrefs) | 59 | 🔧 No |
| Content (E-E-A-T, editorial review) | 13 | ✍️ No |
| UX (visual/responsive testing) | 3 | 👁️ No |
| **Total** | **407** | |

**322 checks (79%) can be automated.** A platform running all 17 automation modules can deliver a complete audit in under 5 minutes — compared to 8–40 hours manually. The remaining 85 checks (Manual + Content + UX + Monitoring) are surfaced as a structured guided checklist so nothing falls through the cracks.

### Product Vision

Build a developer-friendly, AIO-native SEO audit platform that:
- Runs 322 automated checks against any URL or domain in under 5 minutes
- Returns structured findings with Check IDs, severity ratings, and specific code-level fixes
- Adapts its check language and schema expectations to the vertical (automotive, e-commerce, SaaS, media)
- Acts as a **CI/CD gate** — Critical SEO failures block PR merges before they reach production
- Runs the same 322 checks on a **scheduled prod scan** to catch unintentional drift (CMS updates, CDN changes, content team edits that bypass CI)
- Combines static check results, periodic prod audits, and **DebugBear RUM** signals into a single unified dashboard — one place for every team to slice and dice SEO health
- Includes AI Optimization (AIO) scoring for Google SGE / Bing Copilot readiness — a differentiator no existing tool covers

---

## 2. Goals & Success Metrics

### Business Goals

1. Reduce time-to-audit from 8–40 hours to under 5 minutes for 79% of checks
2. Create a scalable SaaS product usable by SEO PMs, agencies, and engineering teams
3. Differentiate on: Check ID traceability, AIO readiness scoring, and CI/CD integration

### Success Metrics

| Metric | Target | Measurement Method |
|---|---|---|
| Audit latency — single URL | < 60 seconds | P95 server-side timing |
| Audit latency — full-site crawl | < 5 minutes (up to 500 pages) | P95 server-side timing |
| Critical issue detection rate | > 95% (vs. manual audit baseline) | Monthly eval against reference sites |
| False positive rate | < 5% per audit | Manual spot-check of 10% of audits |
| Critical issue fix velocity | 70% of Critical findings actioned ≤ 7 days | Tracked via re-audit delta |
| Organic traffic lift post-fix | Measurable improvement for teams fixing ≥5 Critical issues | 30-day GSC comparison |
| Monthly active audits at launch | 500 | Product analytics |
| Net Promoter Score (NPS) | ≥ 50 at 90 days post-launch | In-product survey |
| Performance regression alert SLA | < 30 minutes from regression to Slack/webhook | DebugBear RUM webhook latency + platform routing |
| Static CI Gate miss rate | 0 Critical issues ship to Prod without CI gate detection | Monthly PR audit vs. Prod diff |

### Acceptance Criteria (Overall)

- [ ] A full single-URL audit completes in under 60 seconds on a standard e-commerce homepage
- [ ] A 500-page crawl completes in under 5 minutes with all 322 automated checks applied
- [ ] Every finding in the output cites a Check ID (e.g., `CANONICAL-003`), status (`PASS` / `FAIL` / `WARNING`), and a specific fix recommendation
- [ ] The SEO Health Score (0–100, grade A–F) is reproducible: re-running the same URL within 60 seconds produces a score within ±2 points
- [ ] The AIO Readiness rating (`Strong` / `Moderate` / `Weak`) is present in every report
- [ ] A PR introducing a Critical SEO failure (e.g., `noindex` on all pages) is blocked by the CI gate before merge — verified via test PR against staging
- [ ] The unified dashboard shows all three data streams (CI Gate, Prod Health, RUM Trends) under a single authenticated view with working slice-and-dice filters

---

## 3. User Personas

### P1 — SEO Product Manager

**Context:** Manages SEO roadmap for a 10,000+ page marketplace (e.g., automotive, real estate). Needs a weekly scorecard to brief engineering on what to fix next.

**Jobs to be done:**
- Get a prioritized list of critical issues with business impact estimates
- Track score improvement over time (trend view, week-over-week delta)
- Export a clean report for stakeholder presentations

**Key features needed:** Health Score dashboard, Fix Roadmap sorted by priority, trend charts, PDF export

---

### P2 — Technical SEO Developer

**Context:** Backend or frontend engineer who owns SEO implementation. Thinks in Check IDs and code diffs, not prose summaries.

**Jobs to be done:**
- Get machine-readable audit output (JSON) to pipe into existing dashboards
- See exact code-level fixes per finding (e.g., "add `<link rel='canonical' href='...' />` to `<head>`")
- Block deploys that introduce Critical SEO regressions

**Key features needed:** JSON API output, Check ID references, code snippets per finding, CI/CD webhook integration

---

### P3 — Marketing Manager / Content Lead

**Context:** Owns content calendar and keyword strategy. Not technical but needs to know which pages are underperforming and why.

**Jobs to be done:**
- Identify pages with missing or duplicate title tags and meta descriptions
- See which pages are failing AIO readiness (missing FAQ schema, E-E-A-T signals)
- Understand keyword gaps without needing to run a separate rank-tracking tool

**Key features needed:** On-Page SEO summary, AIO Readiness section, content quality flags, non-technical language in fix recommendations

---

### P4 — SEO Agency

**Context:** Runs audits for 20–50 client sites. Needs white-label reports and a multi-site management view.

**Jobs to be done:**
- Run bulk audits across client sites and compare scores
- Brand the report with client logo before delivery
- Automate monthly re-audits and email results to clients

**Key features needed:** Multi-site dashboard, white-label PDF/JSON export, scheduled audit automation, client portal access

---

### P5 — DevOps / SRE

**Context:** Owns site reliability and CI/CD pipelines. Cares about SEO only when it causes traffic drops or on-call alerts.

**Jobs to be done:**
- Get alerted when a deploy causes a Critical SEO regression (e.g., accidentally noindexing the site)
- See Core Web Vitals trends alongside other performance metrics
- Integrate SEO health into existing monitoring stacks (PagerDuty, Datadog, Grafana)

**Key features needed:** Webhook alerts for Critical findings, Core Web Vitals API, Prometheus/JSON metrics endpoint

---

## 4. Functional Requirements

### Phase 1 — MVP: Single-URL Audit Engine

| ID | Requirement | Acceptance Criteria |
|---|---|---|
| FR-01 | Accept a URL input and run all 17 Phase 1 automation modules | Audit completes in <60s; all 14 Phase 1 modules return a result (PASS/FAIL/WARNING/SKIP) |
| FR-02 | Output structured report in standard format | Report contains: Executive Summary → Category Scores → Findings Table → Fix Roadmap → AIO Section → Manual Checklist |
| FR-03 | Every finding cites Check ID, status, severity, and a specific fix | Zero findings without a Check ID in any report; fix text is actionable (not "fix this issue") |
| FR-04 | SEO Health Score (0–100) with grade A–F | Score uses weighted formula: Critical=3pts, High=2pts, Medium=1pt; Technical SEO and On-Page SEO weighted 2×; grade thresholds: A≥90, B≥75, C≥60, D≥45, F<45 |
| FR-05 | AIO Readiness rating with top-3 recommendations | Rating is `Strong` / `Moderate` / `Weak`; checks include: FAQPage schema, Speakable schema, E-E-A-T signals, featured snippet eligibility |
| FR-06 | Vertical detection — auto-adapt check language and schema expectations | Automotive sites flagged for missing `Vehicle` schema; e-commerce for `Product`/`Review` schema; SaaS for `SoftwareApplication`; media for `Article`/`NewsArticle` |

### Phase 2 — Site-Wide Crawl & Scheduling

| ID | Requirement | Acceptance Criteria |
|---|---|---|
| FR-07 | Accept a domain; crawl up to N pages (configurable, default 500); aggregate findings | Crawl respects `robots.txt` Crawl-Delay; duplicate findings de-duplicated at domain level; summary shows affected page count per issue |
| FR-08 | Scheduled Prod Scan — run all 322 checks against Production on a configurable schedule (default: daily); show delta vs. prior run | Score delta shown as ±N; new issues (not in prior run) flagged RED; resolved issues GREEN; schedule configurable per domain (daily/weekly/monthly); manual trigger available from dashboard |
| FR-09 | Unified Alerting — notify via Slack/email/webhook when (a) new Critical Static check fails in Prod scan, or (b) RUM P75 metric breaches threshold for ≥15 min | Alert fires within 5 min of detection for static (Prod Scan); within 30 min for RUM (DebugBear webhook → platform); payload includes: stream type, Check ID or metric name, affected URL, severity, suggested fix; all alerts visible in unified Alert Feed |

### Phase 3 — Integrations & API

| ID | Requirement | Acceptance Criteria |
|---|---|---|
| FR-10 | Google Search Console integration | GSC OAuth flow; index coverage data overlaid on crawl findings; query data (impressions/clicks/CTR) shown per page |
| FR-11 | Google Search Console + CrUX integration for **historical** trend context | CrUX data used for 28-day trend overlay in GSC view only — explicitly not for real-time alerting; real-time field data handled by DebugBear RUM (FR-09); GSC OAuth flow with index coverage + query performance overlay |
| FR-12 | Programmatic JSON API for CI/CD integration — Static CI Gate | API returns full audit as structured JSON; `?fail_on=critical` exits non-zero (code 1) to block PR merges; `?fail_on=high` also available; response includes per-check status for each of the 322 checks |
| FR-13 | Jira / Linear ticket auto-creation per Critical finding | One ticket per unique Check ID; ticket includes: description, affected URLs, fix recommendation, severity, stream source (CI/Prod/RUM) |
| FR-14 | White-label report export (PDF + JSON) | PDF branded with configurable logo, color scheme, and agency name; JSON conforms to published schema; exportable at any filter state from the unified dashboard |
| FR-15 | Static CI Gate — block PR merge on any Critical check failure | GitHub / GitLab PR status check set to "failed" when any of the 322 automated checks returns Critical FAIL; engineer sees which Check ID failed and why inline on the PR; gate must pass before merge to protected branch is allowed |
| FR-16 | Unified Dashboard — single view across all three data streams | Dashboard shows: (1) SEO Health Score per domain with 30/60/90-day trend; (2) Stream tabs: CI Gate (PR history) / Prod Health (scheduled scan history) / RUM Trends (DebugBear live feed); (3) Alert Feed: unified alarms from all three streams; (4) Slice-and-dice filters: URL/path, check category, severity, stream, device type, date range; (5) Drill-down: click any Check ID → full detail panel with history, fix, and linked ticket |

---

## 5. The 17 Automation Modules

### Platform Architecture: Three Data Streams → One Unified Dashboard

All 17 modules feed into a single PostgreSQL data store. Data arrives via three distinct streams — each serving a different purpose:

```
┌─────────────────────────────────────────────────────────────────┐
│                SEO PLATFORM — 3-STREAM DATA MODEL               │
├─────────────────┬───────────────────────┬───────────────────────┤
│  STREAM 1       │     STREAM 2          │     STREAM 3          │
│  Static CI Gate │  Periodic Prod Scan   │  RUM / Live Feed      │
│  (per PR)       │  (scheduled)          │  (DebugBear)          │
├─────────────────┼───────────────────────┼───────────────────────┤
│ 322 checks on   │ Same 322 checks on    │ LCP, INP, CLS, TTFB,  │
│ staging branch  │ live Prod, daily by   │ FCP from real users;  │
│ on every PR;    │ default; diff vs      │ P50/P75/P90 per page; │
│ Critical FAIL   │ last run; catches     │ immediate post-deploy │
│ = PR blocked    │ CMS/CDN drift that    │ data; breach alerts   │
│ (FR-15)         │ bypasses CI (FR-08)   │ via webhook (FR-09)   │
└─────────────────┴───────────────────────┴───────────────────────┘
                              │
               [PostgreSQL — unified SEO data store]
               check_runs (stream, url, check_id, status, ts)
               rum_snapshots (page, metric, p50/p75/p90, ts)
               alert_events (type, stream, severity, resolved_at)
                              │
               [Unified Dashboard (FR-16)]
               Tabs: CI Gate | Prod Health | RUM Trends | Alert Feed
               Filters: URL · Category · Severity · Stream · Date
```

> **Why three streams?** CI (Stream 1) catches regressions introduced by code changes. Prod Scan (Stream 2) catches drift from CMS edits, CDN changes, and content updates that never touch the CI pipeline. RUM (Stream 3) catches real-user impact from latency regressions, third-party script slowdowns, and geographic CDN issues that synthetic tools miss.

---

### Phase 1 — MVP (Single-URL)

| Module | Check IDs | Checks | Technique | Key Libraries | Reference File |
|---|---|---|---|---|---|
| **M1: Canonical & Duplicates** | CANONICAL-001–020, DUPE-001–003 | 23 | HTTP crawl + chain follow | `requests`, `BeautifulSoup` | `references/technical-seo.md`, `references/content-seo.md` |
| **M2: Robots & Sitemap** | ROBOTS-001–018, SITEMAP-001–017 | 35 | Text/XML parse | `urllib.robotparser`, `xmltodict` | `references/technical-seo.md` |
| **M3: HTTPS & SSL** | SSL-001–010 | 10 | SSL certificate inspection | `ssl`, `certifi`, `cryptography` | `references/technical-seo.md` |
| **M4: Redirect Chains** | REDIRECT-001–015, STATUS-001–002 | 17 | Redirect chain following (max 10 hops) | `requests` (allow_redirects=False) | `references/technical-seo.md` |
| **M5: HTTP Status Codes** | STATUS-001–008, CRAWL-001–001 | 9 | HTTP crawl | `requests`, `httpx` | `references/technical-seo.md` |
| **M6: URL Structure** | URL-001–015, CRAWL-002–004 | 17 | URL parsing + regex | `urllib.parse`, `tldextract` | `references/technical-seo.md` |
| **M7: Title Tags & Meta** | TITLE-001–020, META-001–008 | 28 | DOM parsing | `BeautifulSoup`, `lxml` | `references/onpage-seo.md` |
| **M8: Heading Hierarchy** | H1-001–013 | 13 | DOM tree traversal | `BeautifulSoup` | `references/onpage-seo.md` |
| **M9: Image Optimisation** | IMG-001–015 | 15 | Image crawl + metadata | `Pillow`, `requests` | `references/onpage-seo.md` |
| **M10: Internal Linking** | INTLINK-001–017 | 17 | Link graph analysis | `networkx`, `BeautifulSoup` | `references/onpage-seo.md` |
| **M11: Structured Data** | SCHEMA-001–020 | 20 | JSON-LD + Microdata parsing | `extruct`, `jsonschema` | `references/onpage-seo.md` |
| **M12: JavaScript SEO** | JS-001–010 | 10 | Headless browser rendering | `playwright` (Chromium) | `references/advanced-technical.md` |
| **M13: Mobile Responsiveness** | MOBILE-001–013 | 13 | Responsive viewport testing | `playwright` (mobile emulation) | `references/mobile-offpage-local.md` |
| **M14: Core Web Vitals** | CWV-001–010, SPEED-001–028 | 38 | **3-layer proactive stack** (see below) | Lighthouse CI, DebugBear RUM API, Playwright/CDP | `references/performance.md` |

#### M14 — Performance Measurement: 3-Layer Proactive Stack

CrUX has a **28-day rolling window** — it is a retrospective data source, not a proactive one. M14 uses a layered approach that catches regressions within minutes:

| Layer | When | Tool | Regression Detected In | CI/CD Gate? |
|---|---|---|---|---|
| **L1 — Pre-Deploy Lab** | Every PR to main/staging | Lighthouse CI (`@lhci/cli`) + Playwright/CDP | ~2 min after commit | ✅ Yes — exits non-zero |
| **L2 — Post-Deploy RUM** | Continuously from real users | **DebugBear RUM** (existing org integration) | ~5–30 min post-deploy | ❌ Alerting only |
| **L3 — Continuous Synthetic** | 24/7 between deploys | **DebugBear Synthetic** (30+ global locations) | ~5–15 min from CDN/server regression | ❌ Alerting only |

- **L1 (Lighthouse CI):** Runs `@lhci/cli` in CI job; enforces budgets (LCP <2.5s, CLS <0.1, INP <200ms, TBT <300ms); exits non-zero on failure; free, no vendor dependency
- **L2 (DebugBear RUM):** Existing org integration; <10KB async snippet; captures LCP, INP, CLS, TTFB, FCP from real users at P50/P75/P90; data surfaces immediately after deploy (not 28 days); DebugBear webhook → platform alert engine → Slack + Alert Feed
- **L3 (DebugBear Synthetic):** Scheduled Lighthouse from 30+ CDN locations; catches third-party script regressions and CDN-level slowdowns; no additional vendor required (DebugBear covers both L2 and L3)
- **CrUX retained for:** Historical GSC overlay only (FR-11) — its 28-day lag is acceptable for long-term trend analysis, not for operational alerting

### Phase 2 — Site-Wide

| Module | Check IDs | Checks | Technique | Key Libraries | Reference File |
|---|---|---|---|---|---|
| **M15: Crawl Efficiency** | CRAWL-001–015, INDEX-001–012 | 12 | Full crawl simulation + index analysis | `scrapy`, `networkx` | `references/technical-seo.md`, `references/advanced-technical.md` |
| **M16: Hreflang** | HREFLANG-001–007 | 7 | Tag parsing + reciprocal validation | `BeautifulSoup`, graph traversal | `references/advanced-technical.md` |
| **M17: Local SEO** | LOCAL-001, LOCAL-007, LOCAL-015 | 5 | Schema + map embed detection | `extruct`, DOM parsing | `references/mobile-offpage-local.md` |

**Total automatable checks covered: 290 across 17 modules**
*(The remaining 32 automatable checks in Crawl/Performance/API categories are distributed as sub-checks within modules above.)*

---

## 6. Non-Functional Requirements

### Performance & Reliability

| Requirement | Target |
|---|---|
| Single-URL audit latency (P95) | < 60 seconds |
| Full-site crawl latency (500 pages, P95) | < 5 minutes |
| API uptime SLA | 99.9% monthly |
| Maximum crawl concurrency per audit | 10 parallel requests (configurable) |
| Crawl rate limiting | Respect `Crawl-Delay` in `robots.txt`; default 1 req/sec if not specified |

### Security & Privacy

- No raw HTML stored beyond 24 hours after audit completion
- All audit findings encrypted at rest (AES-256)
- API keys scoped per project; no cross-project data access
- GDPR-compliant: audit data deletable on request
- SSL/TLS for all API endpoints; no HTTP fallback

### Scalability & Architecture

- Multi-tenant SaaS; complete data isolation per tenant
- Audit history retained for 12 months for trend analysis
- Horizontal scaling: audit workers stateless; scale via queue depth
- Rate limiting on public API: 100 audits/hour per API key (configurable per tier)

### Unified Data Model

The platform's PostgreSQL schema must support all three streams in a single queryable store:

```sql
check_runs   (id, stream ENUM('ci','prod','manual'), domain, url, check_id, 
              status ENUM('PASS','FAIL','WARNING','SKIP'), severity, run_ts, pr_ref)
rum_snapshots(id, domain, url, metric, p50, p75, p90, sample_count, snapshot_ts)
alert_events (id, stream, alert_type, check_id_or_metric, url, severity, 
              triggered_at, resolved_at, notified_channels)
```

All dashboard queries filter across streams using the same schema — no per-stream silos.

### Alerting SLAs

| Alert Type | Source | Target SLA |
|---|---|---|
| Critical Static check new failure (Prod Scan) | Stream 2 | < 5 minutes |
| RUM metric threshold breach (P75 CWV) | Stream 3 — DebugBear webhook | < 30 minutes |
| CI Gate failure (PR blocked) | Stream 1 — synchronous | Immediate (blocks PR) |

### Observability

- All audit runs emit structured logs (JSON) with Check ID, status, latency, module name, and stream
- Prometheus metrics endpoint for crawl queue depth, audit latency percentiles, error rates, alert backlog
- Alerting on: audit failure rate >1%, P95 latency >90s, queue depth >500, DebugBear webhook delivery failures

---

## 7. Out of Scope (v1)

The following are intentionally excluded from v1 to maintain launch focus:

| Item | Reason | Planned Phase |
|---|---|---|
| Backlink profile analysis (BACKLINK-001–020) | Requires licensed third-party API (Ahrefs/SEMrush) | Phase 3+ |
| Log file analysis (LOG-001–005) | Requires server log upload pipeline; high engineering complexity | Phase 3+ |
| Keyword research & rank tracking | Separate product surface; not audit-adjacent | Future product |
| Content quality scoring (KW, CONTENT, DUPE checks beyond duplicate detection) | Requires LLM integration for E-E-A-T analysis | Phase 2+ |
| Manual checks automated (ANALYTICS-001–017, BRAND-001–008) | Require authenticated access to GA4/GSC/third-party tools | Phase 3+ |

These 85 checks (Manual=59, Content=13, UX=3, plus Monitoring=14 as alerts) are surfaced in **Appendix A** as a structured guided checklist — the product presents them to the user as "things a human needs to verify," not as automated findings.

---

## Appendix A — Manual & Monitoring Checklist

The following 85 checks cannot be automated in v1. The product surfaces these as a structured checklist within each audit report, grouped by category, with guidance on which tool to use for each check.

### Analytics & Tracking (17 checks — Manual/Monitoring)

| Check ID | Check Name | Tool Required | Priority |
|---|---|---|---|
| ANALYTICS-001 | GA4 properly installed | Chrome DevTools / Tag Assistant | Critical |
| ANALYTICS-002 | GA4 not double-firing | Chrome DevTools → Network tab | Critical |
| ANALYTICS-003 | GSC property verified | Google Search Console | Critical |
| ANALYTICS-004 | GSC and GA4 linked | GA4 Admin → Search Console Links | High |
| ANALYTICS-005 | Bing Webmaster Tools set up | Bing Webmaster Tools dashboard | Medium |
| ANALYTICS-006 | Goal/conversion tracking set up in GA4 | GA4 Events → Conversions | Critical |
| ANALYTICS-007 | Internal IPs filtered from analytics | GA4 Data Filters | High |
| ANALYTICS-008 | Bot traffic filtered | GA4 Data Filters | High |
| ANALYTICS-009 | UTM parameters used consistently | GA4 Traffic Acquisition report | High |
| ANALYTICS-010 | Ecommerce/lead tracking configured | GA4 Ecommerce setup | Critical |
| ANALYTICS-011 | Scroll depth tracking enabled | GA4 Enhanced Measurement | Medium |
| ANALYTICS-012 | Heatmap tool configured | Hotjar / Microsoft Clarity | Medium |
| ANALYTICS-013 | Keyword rank tracking set up | SEMrush / Ahrefs / GSC | Critical |
| ANALYTICS-014 | Organic traffic baseline established | GA4 + GSC | Critical |
| ANALYTICS-015 | Competitor ranking tracked | SEMrush / Ahrefs | High |
| ANALYTICS-016 | Monthly Screaming Frog crawl scheduled | Screaming Frog | High |
| ANALYTICS-017 | Uptime monitoring in place | Pingdom / UptimeRobot | Critical |

### Backlink Profile (20 checks — Manual)

All BACKLINK-001–020 checks require Ahrefs, SEMrush, or Moz. See `references/mobile-offpage-local.md` for full descriptions. Key checks:

- **BACKLINK-003 / BACKLINK-004**: Identify and disavow toxic/spammy backlinks — Critical
- **BACKLINK-010**: Acquire high-authority editorial links — Critical
- **BACKLINK-005**: Ensure anchor text distribution is natural — High

### Brand Signals (8 checks — Manual/Monitoring)

All BRAND-001–008 checks. Key: Google Business Profile claimed (BRAND-001 — Critical), reviews on third-party platforms (BRAND-006 — Critical).

### Local SEO — Manual Subset (10 checks)

LOCAL-002 through LOCAL-006, LOCAL-008 through LOCAL-011, LOCAL-013 through LOCAL-014. Key: city pages must be unique and not templated duplicates (LOCAL-008 — Critical).

### Content Quality (13 checks — Editorial)

KW-001–015 (keyword targeting and cannibalization), CONTENT-001–020 (E-E-A-T, freshness, readability, plagiarism). These require human editorial judgment and/or an LLM-assisted review.

### UX (3 checks — Visual)

MOBILE-003 (touch targets ≥48×48px), MOBILE-004 (font size ≥16px), MOBILE-012 (mobile navigation usability). Require manual visual testing or automated visual regression tools (Percy, Chromatic).

---

## Appendix B — Suggested Tech Stack

| Module | Recommended Stack | Rationale |
|---|---|---|
| HTTP Crawling (M1–M6, M10, M15) | Python + `httpx` (async) + `BeautifulSoup`/`lxml` | Fast async crawling; excellent HTML parsing ecosystem |
| JavaScript Rendering (M12, M13) | `playwright` (Python or Node) + Chromium | Required for SPA/React sites; Google's own rendering uses Chrome |
| Core Web Vitals (M14) — L1 Pre-Deploy | Lighthouse CI (`@lhci/cli`) + Playwright/CDP | Free, open-source; CI gate; lab data per PR; no vendor dependency |
| Core Web Vitals (M14) — L2 Post-Deploy RUM | **DebugBear RUM API** (`/api/v1/project/{id}/rumMetrics`) | Existing org integration; immediate post-deploy data; P75 per page; Slack webhook native |
| Core Web Vitals (M14) — L3 Continuous Synthetic | **DebugBear Synthetic** (30+ global locations) | Existing org integration; scheduled Lighthouse; CDN/third-party regression detection |
| CrUX (historical context only) | Google CrUX REST API | 28-day trend for GSC overlay (FR-11); NOT used for alerting |
| Structured Data (M11) | `extruct` (Python) | Extracts JSON-LD, Microdata, RDFa in one pass |
| SSL/TLS (M3) | Python `ssl` + `cryptography` library | Native; no external API dependency |
| Crawl Queue | Redis + Celery (Python) or BullMQ (Node) | Reliable distributed task queue; supports priority lanes |
| API Layer | FastAPI (Python) or Express (Node) | High performance; OpenAPI schema auto-generation |
| Data Storage | PostgreSQL (findings) + S3 (raw HTML, 24h TTL) | Relational for query/trend; object store for ephemeral raw data |
| Scheduling | Celery Beat or cron-based triggers via the API | Reuses existing worker infrastructure |
| Alerting | Webhooks (Slack/PagerDuty/custom) + SMTP | Stateless; no additional infrastructure required |
| Reporting | Markdown → PDF via `weasyprint` or Puppeteer | White-label PDF with CSS theming; JSON native from API |

### Architecture Diagram (v1.1 — 3-stream unified platform)

```
                   ┌──────────────────────────────────────────┐
                   │            INPUT SOURCES                  │
         ┌─────────┴──────────┐  ┌────────────┐  ┌──────────┐│
         │  PR / CI Trigger   │  │  Scheduler  │  │DebugBear ││
         │  (every PR/commit) │  │  (daily/    │  │ RUM API  ││
         └─────────┬──────────┘  │  weekly)    │  │(webhook) ││
                   │             └──────┬───────┘  └────┬─────┘│
                   │                   │               │       │
                   ▼                   ▼               ▼       │
         ┌─────────────────┐  ┌────────────────┐  ┌─────────┐ │
         │  STREAM 1       │  │  STREAM 2      │  │STREAM 3 │ │
         │  Static CI Gate │  │  Prod Scan     │  │RUM Feed │ │
         │  (M1–M14 checks)│  │  (M1–M17)      │  │(DebugBr)│ │
         └────────┬────────┘  └───────┬────────┘  └────┬────┘ │
                  │  Critical FAIL     │                │      │
                  │  = block PR ────►[GitHub/GitLab     │      │
                  │    status check]   │                │      │
                  └──────────┬─────────┘                │      │
                             │              ┌────────────┘      │
                             ▼              ▼                   │
                    ┌──────────────────────────────┐           │
                    │  API Layer (FastAPI)          │           │
                    │  + Redis Task Queue           │           │
                    └──────────────┬───────────────┘           │
                                   │                            │
              ┌────────────────────▼────────────────────────┐  │
              │           Audit Orchestrator                 │  │
              │  M1–M11: httpx Crawl Workers                 │  │
              │  M12–M13: Playwright (Chromium)              │  │
              │  M14-L1: Lighthouse CI (@lhci/cli)           │  │
              │  M15–M17: Scrapy Domain Crawl                │  │
              └────────────────────┬────────────────────────┘  │
                                   │                            │
                    ┌──────────────▼─────────────┐             │
                    │   PostgreSQL (unified store) │◄───────────┘
                    │   check_runs · rum_snapshots │
                    │   alert_events               │
                    └──────────────┬──────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              ▼                    ▼                     ▼
    [Unified Dashboard]    [Alert Engine]        [Report Generator]
    CI Gate | Prod Health  Slack / email /        JSON API + PDF
    RUM Trends | Alerts    webhook / PagerDuty    white-label export
    Slice & dice filters   (< 5min static,
                            < 30min RUM SLA)
```

---

*This PRD is a living document. Update Check ID references when the master checklist is revised. Next review: 30 days post-engineering kickoff.*

---

## Appendix C — Check ID Reference Map

This appendix is the **source-of-truth index for engineering teams**. Every Check ID cited in this PRD (in module tables, FR acceptance criteria, and ticket payloads) is defined in one of the seven reference files below. Each check record contains: `Check ID | Check Name | Description | Priority | Type`.

> **File location prefix:** `~/.claude/skills/seo-auditor/references/`  
> **Priority values:** Critical · High · Medium  
> **Type values:** Crawl · Performance · API · Manual · Monitoring · Content · UX

---

### C1 — Technical SEO

**File:** `references/technical-seo.md`  
**Total checks:** 121  
**Used by modules:** M1, M2, M3, M4, M5, M6, M15

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Canonical Tags | CANONICAL-001–020 | 20 | Includes AMP canonicals, hreflang consistency, URL parameter handling |
| Robots.txt | ROBOTS-001–018 | 18 | Includes staging blocking, crawl-delay, JS/CSS allow rules |
| XML Sitemap | SITEMAP-001–020 | 20 | Includes image/video/news sitemaps, compression, hreflang sitemap |
| HTTPS & SSL | SSL-001–010 | 10 | Includes HSTS, mixed content, TLS version, certificate transparency |
| Redirects | REDIRECT-001–015 | 15 | Includes chain length, meta refresh, JS redirects, destination indexability |
| HTTP Status Codes | STATUS-001–008 | 8 | Includes soft 404s, 410 usage, external link validation |
| URL Structure | URL-001–015 | 15 | Includes stop words, session IDs, breadcrumb matching |
| Crawlability | CRAWL-001–015 | 15 | Includes JS renderability, infinite scroll, orphan pages, crawl budget |

---

### C2 — On-Page SEO

**File:** `references/onpage-seo.md`  
**Total checks:** 95  
**Used by modules:** M7, M8, M9, M10, M11

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Title Tags | TITLE-001–020 | 20 | Includes dynamic titles, CTR optimisation, GSC title rewriting |
| Meta Descriptions | META-001–010 | 10 | Includes dynamic descriptions, Google rewriting detection |
| Heading Tags | H1-001–013 | 13 | Includes H1/H2/H3 hierarchy, CSS-hidden H1 detection |
| Image Optimisation | IMG-001–015 | 15 | Includes WebP/AVIF, lazy loading, preload, width/height |
| Internal Linking | INTLINK-001–017 | 17 | Includes link graph, anchor text, orphan detection, JS-only links |
| Structured Data / Schema | SCHEMA-001–020 | 20 | Includes Vehicle, Product, FAQPage, Speakable, SpeakableSpecification |

---

### C3 — Content SEO

**File:** `references/content-seo.md`  
**Total checks:** 45  
**Used by modules:** M1 (duplicate detection); remainder are Manual/Editorial

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Keyword Optimisation | KW-001–015 | 15 | Includes cannibalization, intent matching, featured snippets — **Manual** |
| Content Quality | CONTENT-001–020 | 20 | E-E-A-T, freshness, readability, AI content review — **Manual/Editorial** |
| Thin & Duplicate Content | DUPE-001–010 | 10 | Parameterised URLs, printer-friendly canonicals, near-duplicate pages |

---

### C4 — Speed & Performance

**File:** `references/performance.md`  
**Total checks:** 38  
**Used by modules:** M14

| Check Group | Check IDs | Count | Thresholds |
|---|---|---|---|
| Core Web Vitals | CWV-001–010 | 10 | LCP <2.5s · INP <200ms · CLS <0.1 · FCP <1.8s · TTFB <600ms |
| Page Speed | SPEED-001–028 | 28 | PageSpeed score ≥90, page size <3MB, TTFB <200ms, render-blocking JS/CSS |

> All 38 checks are measured via the **3-layer M14 stack**: Lighthouse CI (L1) + DebugBear RUM (L2) + DebugBear Synthetic (L3). See Section 5 — M14 for the full breakdown.

---

### C5 — Mobile, Off-Page & Local SEO

**File:** `references/mobile-offpage-local.md`  
**Total checks:** 58  
**Used by modules:** M13 (Mobile), M17 (Local)

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Mobile Optimisation | MOBILE-001–015 | 15 | Includes viewport, touch targets, interstitials, mobile CWV — M13 covers MOBILE-001–013 |
| Backlink Profile | BACKLINK-001–020 | 20 | Requires Ahrefs/SEMrush — **Manual, Out of Scope v1** |
| Brand Signals | BRAND-001–008 | 8 | GBP, social profiles, review strategy — **Manual** |
| Local SEO | LOCAL-001–015 | 15 | M17 covers LOCAL-001, LOCAL-007, LOCAL-015; remainder are Manual |

---

### C6 — Advanced Technical SEO

**File:** `references/advanced-technical.md`  
**Total checks:** 34  
**Used by modules:** M12 (JS SEO), M15 (Crawl), M16 (Hreflang)

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Indexing & Coverage | INDEX-001–012 | 12 | Includes index bloat, crawled-not-indexed, indexing speed — M15 |
| Log File Analysis | LOG-001–005 | 5 | Requires server log upload — **Manual, Out of Scope v1** |
| JavaScript SEO | JS-001–010 | 10 | SSR/pre-rendering, History API, lazy-loaded content — M12 |
| Hreflang / International | HREFLANG-001–007 | 7 | Reciprocal validation, x-default, sitemap hreflang — M16 |

---

### C7 — Analytics & Monitoring

**File:** `references/analytics.md`  
**Total checks:** 17  
**Used by modules:** Surfaced as Appendix A Guided Checklist (all Manual/Monitoring)

| Check Group | Check IDs | Count | Notes |
|---|---|---|---|
| Analytics & Tracking | ANALYTICS-001–017 | 17 | GA4, GSC, Bing Webmaster, rank tracking, uptime monitoring — all **Manual** |

---

### Quick Lookup: Check ID → Reference File

| If you see a Check ID starting with… | Look in… |
|---|---|
| `CANONICAL`, `ROBOTS`, `SITEMAP`, `SSL`, `REDIRECT`, `STATUS`, `URL`, `CRAWL` | `references/technical-seo.md` |
| `TITLE`, `META`, `H1`, `IMG`, `INTLINK`, `SCHEMA` | `references/onpage-seo.md` |
| `KW`, `CONTENT`, `DUPE` | `references/content-seo.md` |
| `CWV`, `SPEED` | `references/performance.md` |
| `MOBILE`, `BACKLINK`, `BRAND`, `LOCAL` | `references/mobile-offpage-local.md` |
| `INDEX`, `LOG`, `JS`, `HREFLANG` | `references/advanced-technical.md` |
| `ANALYTICS` | `references/analytics.md` |
