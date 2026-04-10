# SEO Audit Skill

## Trigger phrases
- "run seo audit on [url]"
- "audit [url] for seo"
- "seo check [url]"
- "crawl [url] seo"
- "seo report for [url]"
- "technical seo audit [url]"
- "check seo issues on [url]"
- "run the seo tool on [url]"

## What this skill does

Runs a production-grade SEO site crawl with **175+ checks** across 22 categories modelled after Screaming Frog + SEMrush Site Audit. Produces JSON + HTML + Markdown reports.

## How to invoke

When triggered, extract the URL from the user's message and run:

```bash
cd /path/to/seo-automation
go build -o seo-audit . 2>/dev/null || true
./seo-audit audit \
  --url <URL> \
  --max-depth 3 \
  --max-pages 200 \
  --concurrency 5 \
  --format json,html,markdown \
  --output-dir ./reports
```

Then summarize the results from the JSON report.

## Output summary format

After the audit completes, present a summary like:

```
## SEO Audit: <URL>

**Health Score:** XX.X / 100  (Grade: X)
**Pages Crawled:** N
**Total Issues:** N (E errors · W warnings · N notices)

### Top Issues
1. [error] <check.id> — <message> (<url>)
2. [warning] ...
...

### Reports
- HTML: reports/report.html
- JSON: reports/report.json
- Markdown: reports/report.md
```

## Categories checked (22)

| Category | Checks |
|---|---|
| Crawlability | 15 |
| HTTPS & Security | 10 |
| Performance | 12 |
| Internal Linking | 10 |
| Titles | 5 |
| Meta Descriptions | 5 |
| Content Body | 6 |
| Headings | 8 |
| Canonical | 5 |
| Images | 7 |
| Structured Data | 10 |
| Social / OG | 8 |
| URL Structure | 8 |
| Mobile | 4 |
| Mobile vs Desktop | 10 |
| International / hreflang | 7 |
| Sitemap | 8 |
| Pagination | 6 |
| AMP | 3 |
| Crawl Budget | 6 |
| Core Web Vitals | 7 |
| E-E-A-T | 8 |

## Expert context

Think of yourself as a senior Google SEO engineer with 15+ years of experience. When presenting results:

- **Errors** = Critical, must fix now. These actively harm rankings or user experience.
- **Warnings** = Should fix within 2 weeks. Meaningful ranking opportunity.
- **Notices** = Best-practice improvements. Fix in next sprint.

Prioritize recommendations by:
1. Crawlability/indexing issues (nothing matters if Google can't crawl)
2. Core Web Vitals (direct ranking factor since 2021)
3. Content quality & E-E-A-T (dominant ranking factor post-HCU)
4. Technical SEO (canonicals, hreflang, sitemaps)
5. Social/OG (CTR improvement, not ranking)

Always explain *why* each issue matters and give specific, actionable fixes.
