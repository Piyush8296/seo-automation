package report

import (
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// Issue is a flattened check result for template rendering.
type Issue struct {
	CheckID  string
	Category string
	Severity models.Severity
	Message  string
	URL      string
	Details  string
	Platform models.Platform
}

// CategoryStat holds aggregated per-category stats.
type CategoryStat struct {
	Name     string
	Errors   int
	Warnings int
	Notices  int
	Total    int
	Score    float64
	Icon     string
}

type templateData struct {
	*models.SiteAudit
	AllIssues     []Issue
	TopIssues     []Issue
	CategoryStats []CategoryStat
	ErrorIssues   []Issue
	WarnIssues    []Issue
	NoticeIssues  []Issue
	HasMobileData bool
}

var catIcons = map[string]string{
	"Crawlability": "🕷", "HTTPS &amp; Security": "🔒", "HTTPS & Security": "🔒",
	"Performance": "⚡", "Internal Linking": "🔗", "Titles": "📝",
	"Meta Descriptions": "📋", "Content": "📄", "Headings": "🏷",
	"Canonical": "🎯", "Images": "🖼", "Structured Data": "🧩",
	"Social": "📣", "URL Structure": "🌐", "Mobile": "📱",
	"Mobile vs Desktop": "💻", "International": "🌍", "Sitemap": "🗺",
	"Pagination": "📑", "AMP": "⚡", "Crawl Budget": "💰",
	"Core Web Vitals": "🏎", "E-E-A-T": "⭐",
}

// WriteHTML renders a comprehensive self-contained HTML audit report.
func WriteHTML(audit *models.SiteAudit, outputDir string) (string, error) {
	var allIssues []Issue
	hasMobileData := false
	for _, page := range audit.Pages {
		if page.MobileData != nil {
			hasMobileData = true
		}
		for _, r := range page.CheckResults {
			u := r.URL
			if u == "" {
				u = page.URL
			}
			allIssues = append(allIssues, Issue{r.ID, r.Category, r.Severity, r.Message, u, r.Details, r.Platform})
		}
	}
	for _, r := range audit.SiteChecks {
		allIssues = append(allIssues, Issue{r.ID, r.Category, r.Severity, r.Message, r.URL, r.Details, r.Platform})
	}

	var errI, warnI, noticeI []Issue
	for _, i := range allIssues {
		switch i.Severity {
		case models.SeverityError:
			errI = append(errI, i)
		case models.SeverityWarning:
			warnI = append(warnI, i)
		default:
			noticeI = append(noticeI, i)
		}
	}

	top := append([]Issue{}, errI...)
	top = append(top, warnI...)
	if len(top) > 10 {
		top = top[:10]
	}

	catMap := map[string]*CategoryStat{}
	for _, issue := range allIssues {
		cat := issue.Category
		if cat == "" {
			cat = "Other"
		}
		if catMap[cat] == nil {
			ic := catIcons[cat]
			if ic == "" {
				ic = "📌"
			}
			catMap[cat] = &CategoryStat{Name: cat, Icon: ic}
		}
		s := catMap[cat]
		s.Total++
		switch issue.Severity {
		case models.SeverityError:
			s.Errors++
		case models.SeverityWarning:
			s.Warnings++
		default:
			s.Notices++
		}
	}
	for _, s := range catMap {
		if s.Total == 0 {
			s.Score = 100
		} else {
			p := float64(s.Errors*10+s.Warnings*3+s.Notices) * 100.0 / float64(max(s.Total*10, 1))
			s.Score = math.Round(math.Max(0, math.Min(100, 100-p)))
		}
	}
	var cats []CategoryStat
	for _, s := range catMap {
		cats = append(cats, *s)
	}
	sort.Slice(cats, func(i, j int) bool {
		if cats[i].Errors != cats[j].Errors {
			return cats[i].Errors > cats[j].Errors
		}
		return cats[i].Warnings > cats[j].Warnings
	})

	data := templateData{
		SiteAudit:     audit,
		AllIssues:     allIssues,
		TopIssues:     top,
		CategoryStats: cats,
		ErrorIssues:   errI,
		WarnIssues:    warnI,
		NoticeIssues:  noticeI,
		HasMobileData: hasMobileData,
	}

	funcMap := template.FuncMap{
		"sevClass": func(s models.Severity) string {
			switch s {
			case models.SeverityError:
				return "e"
			case models.SeverityWarning:
				return "w"
			default:
				return "n"
			}
		},
		"sevLabel": func(s models.Severity) string {
			switch s {
			case models.SeverityError:
				return "Error"
			case models.SeverityWarning:
				return "Warning"
			default:
				return "Notice"
			}
		},
		"gradeCol": func(g string) string {
			switch g {
			case "A":
				return "#10b981"
			case "B":
				return "#3b82f6"
			case "C":
				return "#f59e0b"
			case "D":
				return "#f97316"
			default:
				return "#ef4444"
			}
		},
		"scoreCol": func(score float64) string {
			if score >= 80 {
				return "#10b981"
			} else if score >= 60 {
				return "#f59e0b"
			}
			return "#ef4444"
		},
		"statusBadge": func(code int) string {
			if code >= 200 && code < 300 {
				return "ok"
			} else if code >= 300 && code < 400 {
				return "redir"
			}
			return "err"
		},
		"short": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "…"
		},
		"catSlug": func(s string) string {
			r := strings.NewReplacer(" ", "-", "&", "", "/", "-", "(", "", ")", "")
			return strings.ToLower(r.Replace(s))
		},
		"pct": func(n, total int) float64 {
			if total == 0 {
				return 0
			}
			return math.Round(float64(n) * 100.0 / float64(total))
		},
		"add": func(a, b int) int { return a + b },
		"platClass": func(p models.Platform) string {
			switch p {
			case models.PlatformMobile:
				return "plat-m"
			case models.PlatformDesktop:
				return "plat-d"
			case models.PlatformDiff:
				return "plat-diff"
			default:
				return "plat-both"
			}
		},
		"platLabel": func(p models.Platform) string {
			switch p {
			case models.PlatformMobile:
				return "📱 Mobile"
			case models.PlatformDesktop:
				return "🖥 Desktop"
			case models.PlatformDiff:
				return "🔄 M↔D"
			default:
				return ""
			}
		},
		"iterate": func(n int) []int {
			if n <= 0 {
				return nil
			}
			r := make([]int, n)
			for i := range r {
				r[i] = i
			}
			return r
		},
	}

	tmpl, err := template.New("r").Funcs(funcMap).Parse(htmlTmpl)
	if err != nil {
		return "", err
	}

	path := filepath.Join(outputDir, "report.html")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return path, tmpl.Execute(f, data)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const htmlTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>SEO Audit — {{.SiteURL}}</title>
<style>
/* ── Reset & tokens ───────────────────────────────────────── */
*, *::before, *::after{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#0d0f18;--s1:#141620;--s2:#1c1f2e;--s3:#242840;
  --border:#2a2d42;--border2:#343756;
  --text:#e2e8f5;--muted:#7c839e;--faint:#4a5068;
  --accent:#6366f1;--accent2:#818cf8;
  --red:#ef4444;--redBg:rgba(239,68,68,.12);
  --orange:#f97316;--orangeBg:rgba(249,115,22,.12);
  --yellow:#f59e0b;--yellowBg:rgba(245,158,11,.12);
  --blue:#3b82f6;--blueBg:rgba(59,130,246,.12);
  --green:#10b981;--greenBg:rgba(16,185,129,.12);
  --purple:#a855f7;--purpleBg:rgba(168,85,247,.12);
  --r:8px;--r2:12px;--r3:16px;
  --shadow:0 2px 8px rgba(0,0,0,.4);
  --shadow2:0 4px 20px rgba(0,0,0,.5);
}
html{scroll-behavior:smooth}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',system-ui,sans-serif;background:var(--bg);color:var(--text);line-height:1.55;font-size:14px}
a{color:var(--accent2);text-decoration:none}a:hover{text-decoration:underline}
h1{font-size:26px;font-weight:800;letter-spacing:-.5px}
h2{font-size:19px;font-weight:700;letter-spacing:-.3px}
h3{font-size:15px;font-weight:600}
code{font-family:'JetBrains Mono',Menlo,monospace;font-size:11px;background:var(--s3);padding:2px 6px;border-radius:4px;color:var(--accent2)}

/* ── Layout ───────────────────────────────────────────────── */
.shell{display:flex;min-height:100vh}
aside{width:240px;flex-shrink:0;background:var(--s1);border-right:1px solid var(--border);position:sticky;top:0;height:100vh;overflow-y:auto;display:flex;flex-direction:column}
aside::-webkit-scrollbar{width:4px}aside::-webkit-scrollbar-track{background:transparent}aside::-webkit-scrollbar-thumb{background:var(--border2);border-radius:2px}
.brand{padding:20px 20px 12px;border-bottom:1px solid var(--border)}
.brand-logo{display:flex;align-items:center;gap:8px;font-weight:800;font-size:15px;color:var(--text)}
.brand-logo span{background:linear-gradient(135deg,var(--accent),#a855f7);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
.nav-section{padding:12px 0}
.nav-label{padding:6px 20px;font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.1em;color:var(--faint)}
.nav-link{display:flex;align-items:center;gap:8px;padding:7px 20px;color:var(--muted);font-size:13px;transition:all .15s;cursor:pointer;border-left:3px solid transparent}
.nav-link:hover,.nav-link.active{background:var(--s2);color:var(--text);border-left-color:var(--accent)}
.nav-link .icon{font-size:14px;width:18px;text-align:center}
.nav-count{margin-left:auto;font-size:10px;background:var(--s3);border-radius:10px;padding:2px 7px;color:var(--muted)}
.nav-count.err{background:var(--redBg);color:var(--red)}
main{flex:1;min-width:0;padding:32px 36px}

/* ── Cover / Header ───────────────────────────────────────── */
.cover{background:linear-gradient(135deg,var(--s2) 0%,var(--s1) 100%);border:1px solid var(--border);border-radius:var(--r3);padding:36px 40px;margin-bottom:28px;display:flex;align-items:flex-start;justify-content:space-between;gap:24px}
.cover-left h1{margin-bottom:6px}
.cover-meta{display:flex;flex-wrap:wrap;gap:12px;margin-top:16px}
.cover-chip{display:flex;align-items:center;gap:6px;background:var(--s3);border:1px solid var(--border2);border-radius:20px;padding:5px 12px;font-size:12px;color:var(--muted)}
.cover-chip strong{color:var(--text)}
.cover-right{display:flex;flex-direction:column;align-items:center;gap:16px;flex-shrink:0}
.score-ring{position:relative;display:flex;align-items:center;justify-content:center}
.score-ring svg{transform:rotate(-90deg)}
.score-inner{position:absolute;text-align:center;pointer-events:none}
.score-num{font-size:36px;font-weight:900;line-height:1}
.score-lbl{font-size:10px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin-top:2px}
.grade-pill{padding:6px 20px;border-radius:24px;font-size:22px;font-weight:900;letter-spacing:-.5px}

/* ── Metric cards ─────────────────────────────────────────── */
.metric-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:14px;margin-bottom:28px}
.metric-card{background:var(--s1);border:1px solid var(--border);border-radius:var(--r2);padding:18px 20px;position:relative;overflow:hidden}
.metric-card::before{content:'';position:absolute;left:0;top:0;bottom:0;width:3px;border-radius:2px}
.metric-card.err::before{background:var(--red)}
.metric-card.warn::before{background:var(--yellow)}
.metric-card.notice::before{background:var(--blue)}
.metric-card.ok::before{background:var(--green)}
.mc-label{font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);margin-bottom:8px}
.mc-value{font-size:30px;font-weight:800;line-height:1}
.mc-value.err{color:var(--red)}
.mc-value.warn{color:var(--yellow)}
.mc-value.notice{color:var(--blue)}
.mc-value.ok{color:var(--green)}
.mc-sub{font-size:11px;color:var(--muted);margin-top:4px}

/* ── Section layout ───────────────────────────────────────── */
.section{margin-bottom:36px}
.section-header{display:flex;align-items:center;gap:10px;margin-bottom:18px;padding-bottom:10px;border-bottom:1px solid var(--border)}
.section-header h2{flex:1}
.section-badge{background:var(--s3);border:1px solid var(--border2);border-radius:20px;padding:3px 12px;font-size:12px;color:var(--muted)}

/* ── Category grid ────────────────────────────────────────── */
.cat-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:14px}
.cat-card{background:var(--s1);border:1px solid var(--border);border-radius:var(--r2);padding:16px 18px;cursor:pointer;transition:all .2s;text-decoration:none;color:inherit;display:block}
.cat-card:hover{border-color:var(--border2);background:var(--s2);transform:translateY(-1px);box-shadow:var(--shadow)}
.cat-top{display:flex;align-items:center;gap:8px;margin-bottom:12px}
.cat-icon{font-size:18px;line-height:1}
.cat-name{font-size:13px;font-weight:600;flex:1}
.cat-score{font-size:13px;font-weight:700}
.cat-bar{height:4px;background:var(--s3);border-radius:2px;overflow:hidden;margin-bottom:10px}
.cat-bar-fill{height:100%;border-radius:2px;transition:width .4s}
.cat-counts{display:flex;gap:8px}
.cc{font-size:11px;font-weight:600;padding:2px 7px;border-radius:4px}
.cc.e{background:var(--redBg);color:var(--red)}
.cc.w{background:var(--yellowBg);color:var(--yellow)}
.cc.n{background:var(--blueBg);color:var(--blue)}

/* ── Priority issues ──────────────────────────────────────── */
.priority-list{display:flex;flex-direction:column;gap:10px}
.priority-item{background:var(--s1);border:1px solid var(--border);border-radius:var(--r2);padding:16px 20px;display:flex;gap:14px;align-items:flex-start}
.priority-num{font-size:20px;font-weight:900;color:var(--faint);line-height:1;flex-shrink:0;width:28px;text-align:right}
.priority-body{flex:1;min-width:0}
.priority-title{font-weight:600;margin-bottom:4px;display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.priority-url{font-size:12px;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;margin-bottom:4px}
.priority-detail{font-size:12px;color:var(--muted);font-style:italic}

/* ── Severity badges ──────────────────────────────────────── */
.badge{display:inline-flex;align-items:center;gap:4px;padding:2px 8px;border-radius:5px;font-size:11px;font-weight:700;text-transform:uppercase;letter-spacing:.04em;white-space:nowrap}
.badge.e{background:var(--redBg);color:var(--red);border:1px solid rgba(239,68,68,.25)}
.badge.w{background:var(--yellowBg);color:var(--yellow);border:1px solid rgba(245,158,11,.25)}
.badge.n{background:var(--blueBg);color:var(--blue);border:1px solid rgba(59,130,246,.25)}
.badge::before{content:'●';font-size:7px}

/* ── Collapsible category sections ───────────────────────── */
.cat-section{background:var(--s1);border:1px solid var(--border);border-radius:var(--r2);margin-bottom:10px;overflow:hidden}
.cat-section-hdr{display:flex;align-items:center;gap:10px;padding:14px 18px;cursor:pointer;user-select:none;transition:background .15s}
.cat-section-hdr:hover{background:var(--s2)}
.cat-section-hdr .arrow{margin-left:auto;color:var(--muted);transition:transform .2s;font-size:12px}
.cat-section-hdr.open .arrow{transform:rotate(90deg)}
.cat-section-body{display:none;border-top:1px solid var(--border)}
.cat-section-body.open{display:block}

/* ── Issue table ──────────────────────────────────────────── */
.tbl-wrap{overflow-x:auto}
table{width:100%;border-collapse:collapse;font-size:13px}
thead th{padding:9px 12px;background:var(--s2);border-bottom:1px solid var(--border2);font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.07em;color:var(--muted);text-align:left;white-space:nowrap}
tbody td{padding:9px 12px;border-bottom:1px solid var(--border);vertical-align:middle}
tbody tr:last-child td{border-bottom:none}
tbody tr:hover td{background:var(--s2)}
.td-id code{font-size:10px}
.td-url{max-width:260px}
.td-url a{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--accent2);font-size:12px}
.td-msg{max-width:280px}
.td-detail{max-width:180px;font-size:11px;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

/* ── Toolbar ──────────────────────────────────────────────── */
.toolbar{display:flex;gap:8px;margin-bottom:14px;flex-wrap:wrap;align-items:center}
.search-box{flex:1;min-width:200px;max-width:320px;background:var(--s2);border:1px solid var(--border2);border-radius:var(--r);padding:7px 12px;color:var(--text);font-size:13px;outline:none;transition:border-color .15s}
.search-box:focus{border-color:var(--accent)}
.search-box::placeholder{color:var(--faint)}
.filt-btn{background:var(--s2);border:1px solid var(--border);border-radius:var(--r);padding:6px 12px;color:var(--muted);font-size:12px;font-weight:600;cursor:pointer;transition:all .15s;white-space:nowrap}
.filt-btn:hover,.filt-btn.on{border-color:var(--accent);color:var(--text);background:rgba(99,102,241,.12)}
.pg-bar{display:flex;gap:8px;margin-top:12px;align-items:center;font-size:12px;color:var(--muted)}
.pg-bar button{background:var(--s2);border:1px solid var(--border);border-radius:5px;padding:4px 10px;color:var(--text);cursor:pointer;font-size:12px}
.pg-bar button:disabled{opacity:.35;cursor:default}

/* ── Pages table ──────────────────────────────────────────── */
.status-ok{color:var(--green);font-weight:700}
.status-redir{color:var(--yellow);font-weight:700}
.status-err{color:var(--red);font-weight:700}
.depth-pip{display:inline-flex;gap:2px}
.dp{width:6px;height:6px;border-radius:50%;background:var(--s3)}
.dp.fill{background:var(--accent)}

/* ── Charts ───────────────────────────────────────────────── */
.charts-row{display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:28px}
.chart-card{background:var(--s1);border:1px solid var(--border);border-radius:var(--r2);padding:20px 24px}
.chart-title{font-size:13px;font-weight:600;color:var(--muted);margin-bottom:16px;text-transform:uppercase;letter-spacing:.06em}
.donut-wrap{display:flex;align-items:center;gap:24px}
.donut-legend{display:flex;flex-direction:column;gap:10px}
.dl-item{display:flex;align-items:center;gap:8px;font-size:13px}
.dl-dot{width:10px;height:10px;border-radius:50%;flex-shrink:0}
.dl-val{margin-left:auto;font-weight:700;min-width:36px;text-align:right}
.bar-chart{display:flex;flex-direction:column;gap:8px}
.bc-row{display:flex;align-items:center;gap:10px;font-size:12px}
.bc-label{width:130px;text-align:right;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex-shrink:0}
.bc-track{flex:1;height:16px;background:var(--s3);border-radius:4px;overflow:hidden}
.bc-fill{height:100%;border-radius:4px;min-width:4px;transition:width .5s}
.bc-num{width:36px;text-align:right;font-weight:600;color:var(--text)}

/* ── Platform badges ──────────────────────────────────────── */
.plat{display:inline-flex;align-items:center;padding:1px 7px;border-radius:5px;font-size:10px;font-weight:700;white-space:nowrap;letter-spacing:.02em}
.plat-m{background:rgba(16,185,129,.12);color:#10b981;border:1px solid rgba(16,185,129,.25)}
.plat-d{background:rgba(59,130,246,.12);color:#3b82f6;border:1px solid rgba(59,130,246,.25)}
.plat-diff{background:rgba(168,85,247,.12);color:#a855f7;border:1px solid rgba(168,85,247,.25)}
.plat-both{display:none}

/* ── Score grid ───────────────────────────────────────────── */
.score-grid{display:flex;gap:24px;align-items:center;flex-wrap:wrap}
.score-block{display:flex;flex-direction:column;align-items:center;gap:8px}
.score-block-label{font-size:11px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);font-weight:700}

/* ── Misc ─────────────────────────────────────────────────── */
.empty{text-align:center;padding:40px;color:var(--faint);font-size:13px}
.hidden{display:none!important}
.pill{display:inline-block;padding:1px 8px;border-radius:10px;font-size:11px;font-weight:600}

/* ── Print ────────────────────────────────────────────────── */
@media print{
  aside{display:none}
  main{padding:0}
  .toolbar,.pg-bar{display:none}
  .cat-section-body{display:block!important}
  body{background:#fff;color:#111}
  .cover,.metric-card,.cat-card,.cat-section,.chart-card,.priority-item{background:#fff;border-color:#ddd}
}
</style>
</head>
<body>
<div class="shell">

<!-- ── Sidebar ──────────────────────────────────────────── -->
<aside>
  <div class="brand">
    <div class="brand-logo"><span>SEO</span>Audit</div>
    <div style="font-size:11px;color:var(--muted);margin-top:4px;word-break:break-all">{{.SiteURL}}</div>
  </div>
  <nav class="nav-section">
    <div class="nav-label">Report</div>
    <a class="nav-link" href="#overview"><span class="icon">🏠</span>Overview</a>
    <a class="nav-link" href="#categories"><span class="icon">📊</span>Categories</a>
    <a class="nav-link" href="#priority"><span class="icon">🔥</span>Priority Issues<span class="nav-count err">{{len .TopIssues}}</span></a>
    <a class="nav-link" href="#all-issues"><span class="icon">📋</span>All Issues<span class="nav-count">{{len .AllIssues}}</span></a>
    <a class="nav-link" href="#pages"><span class="icon">📄</span>Pages<span class="nav-count">{{.PagesCrawled}}</span></a>
  </nav>
  <nav class="nav-section" style="border-top:1px solid var(--border)">
    <div class="nav-label">By Severity</div>
    <a class="nav-link" href="#sec-errors"><span class="icon">🔴</span>Errors<span class="nav-count err">{{len .ErrorIssues}}</span></a>
    <a class="nav-link" href="#sec-warnings"><span class="icon">🟡</span>Warnings<span class="nav-count">{{len .WarnIssues}}</span></a>
    <a class="nav-link" href="#sec-notices"><span class="icon">🔵</span>Notices<span class="nav-count">{{len .NoticeIssues}}</span></a>
  </nav>
  <div style="margin-top:auto;padding:16px 20px;font-size:11px;color:var(--faint);border-top:1px solid var(--border)">
    Generated {{.CrawledAt.Format "Jan 02, 2006"}}<br>
    {{.Stats.TotalChecksRun}} checks run
  </div>
</aside>

<!-- ── Main ─────────────────────────────────────────────── -->
<main>

<!-- Cover -->
<div id="overview" class="cover">
  <div class="cover-left">
    <h1>SEO Audit Report</h1>
    <div style="color:var(--muted);font-size:15px;margin-top:4px">{{.SiteURL}}</div>
    <div class="cover-meta">
      <div class="cover-chip">📅 <strong>{{.CrawledAt.Format "Jan 02, 2006 · 15:04 UTC"}}</strong></div>
      <div class="cover-chip">📄 <strong>{{.PagesCrawled}} pages</strong> crawled</div>
      <div class="cover-chip">✅ <strong>{{.Stats.TotalChecksRun}} checks</strong> run</div>
      {{if gt .SitemapPageCount 0}}<div class="cover-chip">🗺 <strong>{{.SitemapPageCount}} URLs</strong> in sitemap</div>{{end}}
    </div>
  </div>
  <div class="cover-right">
    <div class="score-grid">
      <!-- Overall score -->
      <div class="score-block">
        <div class="score-block-label">Overall</div>
        <div class="score-ring">
          <svg width="100" height="100" viewBox="0 0 120 120">
            <circle cx="60" cy="60" r="54" fill="none" stroke="var(--s3)" stroke-width="10"/>
            <circle id="scoreArc" cx="60" cy="60" r="54" fill="none" stroke="#10b981" stroke-width="10"
              stroke-dasharray="339.3 339.3" stroke-linecap="round"
              style="transition:stroke-dasharray .8s ease"/>
          </svg>
          <div class="score-inner">
            <div class="score-num" id="scoreNum" style="font-size:28px">{{printf "%.0f" .HealthScore}}</div>
            <div class="score-lbl">/ 100</div>
          </div>
        </div>
        <div class="grade-pill" id="gradePill">{{.Grade}}</div>
      </div>
      {{if .HasMobileData}}
      <!-- Desktop score -->
      <div class="score-block">
        <div class="score-block-label">🖥 Desktop</div>
        <div class="score-ring">
          <svg width="80" height="80" viewBox="0 0 120 120">
            <circle cx="60" cy="60" r="54" fill="none" stroke="var(--s3)" stroke-width="10"/>
            <circle id="desktopArc" cx="60" cy="60" r="54" fill="none" stroke="#3b82f6" stroke-width="10"
              stroke-dasharray="339.3 339.3" stroke-linecap="round"
              style="transition:stroke-dasharray .8s ease"/>
          </svg>
          <div class="score-inner">
            <div class="score-num" id="desktopNum" style="font-size:22px">{{printf "%.0f" .DesktopHealthScore}}</div>
            <div class="score-lbl" style="font-size:9px">/ 100</div>
          </div>
        </div>
        <div style="font-size:13px;font-weight:800;color:#3b82f6">{{.DesktopGrade}}</div>
      </div>
      <!-- Mobile score -->
      <div class="score-block">
        <div class="score-block-label">📱 Mobile</div>
        <div class="score-ring">
          <svg width="80" height="80" viewBox="0 0 120 120">
            <circle cx="60" cy="60" r="54" fill="none" stroke="var(--s3)" stroke-width="10"/>
            <circle id="mobileArc" cx="60" cy="60" r="54" fill="none" stroke="#10b981" stroke-width="10"
              stroke-dasharray="339.3 339.3" stroke-linecap="round"
              style="transition:stroke-dasharray .8s ease"/>
          </svg>
          <div class="score-inner">
            <div class="score-num" id="mobileNum" style="font-size:22px">{{printf "%.0f" .MobileHealthScore}}</div>
            <div class="score-lbl" style="font-size:9px">/ 100</div>
          </div>
        </div>
        <div style="font-size:13px;font-weight:800;color:#10b981">{{.MobileGrade}}</div>
      </div>
      {{end}}
    </div>
  </div>
</div>

<!-- Metric cards -->
<div class="metric-grid">
  <div class="metric-card err"><div class="mc-label">Critical Errors</div><div class="mc-value err">{{len .ErrorIssues}}</div><div class="mc-sub">Must fix immediately</div></div>
  <div class="metric-card warn"><div class="mc-label">Warnings</div><div class="mc-value warn">{{len .WarnIssues}}</div><div class="mc-sub">Should fix soon</div></div>
  <div class="metric-card notice"><div class="mc-label">Notices</div><div class="mc-value notice">{{len .NoticeIssues}}</div><div class="mc-sub">Improvements available</div></div>
  <div class="metric-card ok"><div class="mc-label">Pages Crawled</div><div class="mc-value ok">{{.PagesCrawled}}</div><div class="mc-sub">Depth-first BFS</div></div>
  <div class="metric-card"><div class="mc-label">Overall Score</div><div class="mc-value" id="scoreCard" style="color:var(--green)">{{printf "%.1f" .HealthScore}}</div><div class="mc-sub">Grade: {{.Grade}}</div></div>
  {{if .HasMobileData}}
  <div class="metric-card" style="border-left:3px solid #3b82f6"><div class="mc-label">🖥 Desktop Score</div><div class="mc-value" style="color:#3b82f6">{{printf "%.1f" .DesktopHealthScore}}</div><div class="mc-sub">Grade: {{.DesktopGrade}} · {{.DesktopStats.Errors}} err / {{.DesktopStats.Warnings}} warn</div></div>
  <div class="metric-card" style="border-left:3px solid #10b981"><div class="mc-label">📱 Mobile Score</div><div class="mc-value" style="color:#10b981">{{printf "%.1f" .MobileHealthScore}}</div><div class="mc-sub">Grade: {{.MobileGrade}} · {{.MobileStats.Errors}} err / {{.MobileStats.Warnings}} warn</div></div>
  {{else}}
  <div class="metric-card"><div class="mc-label">Categories</div><div class="mc-value">{{len .CategoryStats}}</div><div class="mc-sub">Checked</div></div>
  {{end}}
</div>

<!-- Charts row -->
<div class="charts-row">
  <div class="chart-card">
    <div class="chart-title">Issue Distribution</div>
    <div class="donut-wrap">
      <canvas id="donut" width="120" height="120"></canvas>
      <div class="donut-legend">
        <div class="dl-item"><div class="dl-dot" style="background:var(--red)"></div>Errors<div class="dl-val" style="color:var(--red)">{{len .ErrorIssues}}</div></div>
        <div class="dl-item"><div class="dl-dot" style="background:var(--yellow)"></div>Warnings<div class="dl-val" style="color:var(--yellow)">{{len .WarnIssues}}</div></div>
        <div class="dl-item"><div class="dl-dot" style="background:var(--blue)"></div>Notices<div class="dl-val" style="color:var(--blue)">{{len .NoticeIssues}}</div></div>
      </div>
    </div>
  </div>
  <div class="chart-card">
    <div class="chart-title">Issues by Category (Top 8)</div>
    <div class="bar-chart" id="barChart"></div>
  </div>
</div>

<!-- Category overview -->
<div id="categories" class="section">
  <div class="section-header">
    <h2>Category Performance</h2>
    <span class="section-badge">{{len .CategoryStats}} categories</span>
  </div>
  <div class="cat-grid">
    {{range .CategoryStats}}
    <a class="cat-card" href="#cat-{{catSlug .Name}}">
      <div class="cat-top">
        <span class="cat-icon">{{.Icon}}</span>
        <span class="cat-name">{{.Name}}</span>
        <span class="cat-score" style="color:{{scoreCol .Score}}">{{printf "%.0f" .Score}}</span>
      </div>
      <div class="cat-bar"><div class="cat-bar-fill" style="width:{{printf "%.0f" .Score}}%;background:{{scoreCol .Score}}"></div></div>
      <div class="cat-counts">
        {{if gt .Errors 0}}<span class="cc e">{{.Errors}} err</span>{{end}}
        {{if gt .Warnings 0}}<span class="cc w">{{.Warnings}} warn</span>{{end}}
        {{if gt .Notices 0}}<span class="cc n">{{.Notices}} notice</span>{{end}}
        {{if eq .Total 0}}<span style="font-size:11px;color:var(--green)">✓ Clean</span>{{end}}
      </div>
    </a>
    {{end}}
  </div>
</div>

<!-- Priority issues -->
<div id="priority" class="section">
  <div class="section-header">
    <h2>🔥 Priority Issues</h2>
    <span class="section-badge">Top {{len .TopIssues}} to fix</span>
  </div>
  {{if eq (len .TopIssues) 0}}
  <div class="empty">🎉 No critical issues found!</div>
  {{else}}
  <div class="priority-list">
    {{range $i, $issue := .TopIssues}}
    <div class="priority-item">
      <div class="priority-num">{{add $i 1}}</div>
      <div class="priority-body">
        <div class="priority-title">
          <span class="badge {{sevClass .Severity}}">{{sevLabel .Severity}}</span>
          <code>{{.CheckID}}</code>
          <span style="color:var(--muted);font-size:12px">{{.Category}}</span>
        </div>
        <div style="font-size:14px;margin-bottom:4px">{{.Message}}</div>
        {{if .URL}}<div class="priority-url">🔗 <a href="{{.URL}}" target="_blank" rel="noopener">{{short .URL 80}}</a></div>{{end}}
        {{if .Details}}<div class="priority-detail">{{short .Details 120}}</div>{{end}}
      </div>
    </div>
    {{end}}
  </div>
  {{end}}
</div>

<!-- All Issues by Category -->
<div id="all-issues" class="section">
  <div class="section-header">
    <h2>All Issues by Category</h2>
    <span class="section-badge">{{len .AllIssues}} total</span>
  </div>

  <!-- Global filter -->
  <div class="toolbar">
    <input class="search-box" id="gSearch" type="text" placeholder="Search URL, check ID, message…" oninput="globalFilter()">
    <button class="filt-btn on" id="fb-all" onclick="setGFilt('all',this)">All</button>
    <button class="filt-btn" id="fb-e" onclick="setGFilt('e',this)">Errors</button>
    <button class="filt-btn" id="fb-w" onclick="setGFilt('w',this)">Warnings</button>
    <button class="filt-btn" id="fb-n" onclick="setGFilt('n',this)">Notices</button>
    {{if .HasMobileData}}
    <span style="width:1px;height:20px;background:var(--border);margin:0 4px"></span>
    <button class="filt-btn on" id="fp-all" onclick="setPFilt('all',this)">🌐 All</button>
    <button class="filt-btn" id="fp-d" onclick="setPFilt('desktop',this)">🖥 Desktop</button>
    <button class="filt-btn" id="fp-m" onclick="setPFilt('mobile',this)">📱 Mobile</button>
    <button class="filt-btn" id="fp-diff" onclick="setPFilt('diff',this)">🔄 M↔D Diff</button>
    {{end}}
  </div>

  {{range .CategoryStats}}
  {{if gt .Total 0}}
  <div class="cat-section" id="cat-{{catSlug .Name}}">
    <div class="cat-section-hdr" onclick="toggleCat(this)">
      <span style="font-size:16px">{{.Icon}}</span>
      <span style="font-weight:600">{{.Name}}</span>
      <span class="nav-count" style="margin-left:8px">{{.Total}}</span>
      {{if gt .Errors 0}}<span class="cc e" style="margin-left:4px">{{.Errors}} err</span>{{end}}
      {{if gt .Warnings 0}}<span class="cc w" style="margin-left:4px">{{.Warnings}} warn</span>{{end}}
      <span class="arrow">▶</span>
    </div>
    <div class="cat-section-body">
      <div class="tbl-wrap">
        <table>
          <thead><tr><th>Severity</th><th>Check ID</th><th>URL</th><th>Message</th><th>Details</th><th>Platform</th></tr></thead>
          <tbody>
            {{$catName := .Name}}
            {{range $.AllIssues}}{{if eq .Category $catName}}
            <tr data-sev="{{sevClass .Severity}}" data-plat="{{.Platform}}" class="issue-row">
              <td><span class="badge {{sevClass .Severity}}">{{sevLabel .Severity}}</span></td>
              <td class="td-id"><code>{{.CheckID}}</code></td>
              <td class="td-url"><a href="{{.URL}}" target="_blank" rel="noopener" title="{{.URL}}">{{short .URL 60}}</a></td>
              <td class="td-msg">{{.Message}}</td>
              <td class="td-detail" title="{{.Details}}">{{short .Details 80}}</td>
              <td>{{if platLabel .Platform}}<span class="plat {{platClass .Platform}}">{{platLabel .Platform}}</span>{{end}}</td>
            </tr>
            {{end}}{{end}}
          </tbody>
        </table>
      </div>
    </div>
  </div>
  {{end}}
  {{end}}
</div>

<!-- Errors section -->
<div id="sec-errors" class="section">
  <div class="section-header"><h2>🔴 Errors <span style="font-weight:400;color:var(--muted)">({{len .ErrorIssues}})</span></h2></div>
  {{if eq (len .ErrorIssues) 0}}<div class="empty">✅ No errors found</div>
  {{else}}
  <div id="errTable">
  <div class="tbl-wrap"><table><thead><tr><th>Check ID</th><th>Category</th><th>Platform</th><th>URL</th><th>Message</th><th>Details</th></tr></thead><tbody>
  {{range .ErrorIssues}}
  <tr data-plat="{{.Platform}}"><td><code>{{.CheckID}}</code></td><td>{{.Category}}</td>
  <td>{{if platLabel .Platform}}<span class="plat {{platClass .Platform}}">{{platLabel .Platform}}</span>{{end}}</td>
  <td class="td-url"><a href="{{.URL}}" target="_blank" rel="noopener">{{short .URL 60}}</a></td>
  <td class="td-msg">{{.Message}}</td><td class="td-detail" title="{{.Details}}">{{short .Details 80}}</td></tr>
  {{end}}
  </tbody></table></div>
  </div>{{end}}
</div>

<!-- Warnings section -->
<div id="sec-warnings" class="section">
  <div class="section-header"><h2>🟡 Warnings <span style="font-weight:400;color:var(--muted)">({{len .WarnIssues}})</span></h2></div>
  {{if eq (len .WarnIssues) 0}}<div class="empty">✅ No warnings found</div>
  {{else}}
  <div class="tbl-wrap"><table><thead><tr><th>Check ID</th><th>Category</th><th>Platform</th><th>URL</th><th>Message</th><th>Details</th></tr></thead><tbody>
  {{range .WarnIssues}}
  <tr data-plat="{{.Platform}}"><td><code>{{.CheckID}}</code></td><td>{{.Category}}</td>
  <td>{{if platLabel .Platform}}<span class="plat {{platClass .Platform}}">{{platLabel .Platform}}</span>{{end}}</td>
  <td class="td-url"><a href="{{.URL}}" target="_blank" rel="noopener">{{short .URL 60}}</a></td>
  <td class="td-msg">{{.Message}}</td><td class="td-detail" title="{{.Details}}">{{short .Details 80}}</td></tr>
  {{end}}
  </tbody></table></div>{{end}}
</div>

<!-- Notices section -->
<div id="sec-notices" class="section">
  <div class="section-header"><h2>🔵 Notices <span style="font-weight:400;color:var(--muted)">({{len .NoticeIssues}})</span></h2></div>
  {{if eq (len .NoticeIssues) 0}}<div class="empty">No notices</div>
  {{else}}
  <div class="tbl-wrap"><table><thead><tr><th>Check ID</th><th>Category</th><th>Platform</th><th>URL</th><th>Message</th></tr></thead><tbody>
  {{range .NoticeIssues}}
  <tr data-plat="{{.Platform}}"><td><code>{{.CheckID}}</code></td><td>{{.Category}}</td>
  <td>{{if platLabel .Platform}}<span class="plat {{platClass .Platform}}">{{platLabel .Platform}}</span>{{end}}</td>
  <td class="td-url"><a href="{{.URL}}" target="_blank" rel="noopener">{{short .URL 60}}</a></td>
  <td class="td-msg">{{.Message}}</td></tr>
  {{end}}
  </tbody></table></div>{{end}}
</div>

<!-- Pages -->
<div id="pages" class="section">
  <div class="section-header">
    <h2>Pages Crawled</h2>
    <span class="section-badge">{{.PagesCrawled}} pages</span>
  </div>
  <div class="toolbar">
    <input class="search-box" id="pSearch" type="text" placeholder="Filter by URL or title…" oninput="filterPages()">
  </div>
  <div class="tbl-wrap">
    <table id="pageTable">
      <thead><tr><th>#</th><th>URL</th><th>Status</th><th>Title</th><th>Words</th><th>Issues</th><th>Depth</th><th>In Sitemap</th></tr></thead>
      <tbody>
        {{range $i, $p := .Pages}}
        <tr class="page-row">
          <td style="color:var(--faint);font-size:12px">{{add $i 1}}</td>
          <td class="td-url" style="max-width:320px"><a href="{{$p.URL}}" target="_blank" rel="noopener" title="{{$p.URL}}">{{short $p.URL 70}}</a></td>
          <td><span class="{{statusBadge $p.StatusCode}}">{{$p.StatusCode}}</span></td>
          <td style="max-width:220px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--muted);font-size:12px">{{short $p.Title 50}}</td>
          <td style="font-variant-numeric:tabular-nums">{{$p.WordCount}}</td>
          <td>{{if gt (len $p.CheckResults) 0}}<span class="pill" style="background:var(--redBg);color:var(--red)">{{len $p.CheckResults}}</span>{{else}}<span style="color:var(--green)">✓</span>{{end}}</td>
          <td>
            <div class="depth-pip">
              {{range $d := iterate $p.Depth}}<span class="dp fill"></span>{{end}}
            </div>
            <span style="font-size:11px;color:var(--muted);margin-left:4px">{{$p.Depth}}</span>
          </td>
          <td>{{if $p.InSitemap}}<span style="color:var(--green)">✓</span>{{else}}<span style="color:var(--faint)">—</span>{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  <div class="pg-bar" id="pagePg"></div>
</div>

</main>
</div>

<script>
// ── Score ring animation ──────────────────────────────────────
(function(){
  var circ = 339.3;
  var colors = {A:'#10b981',B:'#3b82f6',C:'#f59e0b',D:'#f97316',F:'#ef4444'};

  function animateRing(arcId, score, grade) {
    var arc = document.getElementById(arcId);
    if(!arc) return;
    var col = colors[grade] || '#ef4444';
    arc.style.stroke = col;
    arc.style.strokeDasharray = (circ * score / 100) + ' ' + circ;
  }

  animateRing('scoreArc', {{printf "%.1f" .HealthScore}}, '{{.Grade}}');
  animateRing('desktopArc', {{printf "%.1f" .DesktopHealthScore}}, '{{.DesktopGrade}}');
  animateRing('mobileArc', {{printf "%.1f" .MobileHealthScore}}, '{{.MobileGrade}}');

  // Style overall grade pill
  var score = {{printf "%.1f" .HealthScore}};
  var grade = '{{.Grade}}';
  var col = colors[grade] || '#ef4444';
  var pill = document.getElementById('gradePill');
  var sc = document.getElementById('scoreCard');
  if(pill){
    pill.style.background = col+'26';
    pill.style.color = col;
    pill.style.border = '2px solid '+col+'66';
  }
  if(sc) sc.style.color = col;
})();

// ── Donut chart ───────────────────────────────────────────────
(function(){
  var c = document.getElementById('donut');
  if(!c) return;
  var ctx = c.getContext('2d');
  var E = {{len .ErrorIssues}}, W = {{len .WarnIssues}}, N = {{len .NoticeIssues}};
  var total = E + W + N;
  if(total === 0){ ctx.fillStyle='#2a2d42'; ctx.beginPath(); ctx.arc(60,60,45,0,Math.PI*2); ctx.fill(); return; }
  var data = [{v:E,c:'#ef4444'},{v:W,c:'#f59e0b'},{v:N,c:'#3b82f6'}];
  var start = -Math.PI/2;
  data.forEach(function(d){
    if(d.v===0) return;
    var angle = (d.v/total)*Math.PI*2;
    ctx.beginPath(); ctx.moveTo(60,60);
    ctx.arc(60,60,54,start,start+angle);
    ctx.closePath(); ctx.fillStyle=d.c; ctx.fill();
    start+=angle;
  });
  ctx.beginPath(); ctx.arc(60,60,35,0,Math.PI*2);
  ctx.fillStyle='#141620'; ctx.fill();
})();

// ── Bar chart ─────────────────────────────────────────────────
(function(){
  var el = document.getElementById('barChart');
  if(!el) return;
  var cats = [
    {{range .CategoryStats}}
    {name:'{{short .Name 20}}',total:{{.Total}},errors:{{.Errors}},warnings:{{.Warnings}}},
    {{end}}
  ];
  cats.sort(function(a,b){return b.total-a.total});
  var top = cats.slice(0,8);
  var maxVal = top.reduce(function(m,c){return Math.max(m,c.total)},0) || 1;
  top.forEach(function(cat){
    var pct = Math.round(cat.total/maxVal*100);
    var fillCol = cat.errors>0?'#ef4444':cat.warnings>0?'#f59e0b':'#3b82f6';
    el.innerHTML += '<div class="bc-row">' +
      '<div class="bc-label">'+cat.name+'</div>' +
      '<div class="bc-track"><div class="bc-fill" style="width:'+pct+'%;background:'+fillCol+'"></div></div>' +
      '<div class="bc-num">'+cat.total+'</div></div>';
  });
})();

// ── Category toggles ──────────────────────────────────────────
function toggleCat(hdr){
  hdr.classList.toggle('open');
  var body = hdr.nextElementSibling;
  body.classList.toggle('open');
}
// Auto-open sections with errors
document.querySelectorAll('.cat-section-hdr').forEach(function(hdr){
  if(hdr.querySelector('.cc.e')) {
    hdr.classList.add('open');
    hdr.nextElementSibling.classList.add('open');
  }
});

// ── Global issue filter ───────────────────────────────────────
var gFilt = 'all', pFilt = 'all';
function setGFilt(f,btn){
  gFilt=f;
  document.querySelectorAll('[id^="fb-"]').forEach(function(b){b.classList.remove('on')});
  btn.classList.add('on');
  globalFilter();
}
function setPFilt(f,btn){
  pFilt=f;
  document.querySelectorAll('[id^="fp-"]').forEach(function(b){b.classList.remove('on')});
  btn.classList.add('on');
  globalFilter();
}
function platMatch(rowPlat, filter) {
  if(filter === 'all') return true;
  if(filter === 'desktop') return rowPlat==='' || rowPlat==='both' || rowPlat==='desktop';
  if(filter === 'mobile')  return rowPlat==='' || rowPlat==='both' || rowPlat==='mobile';
  if(filter === 'diff')    return rowPlat==='diff';
  return true;
}
function globalFilter(){
  var q = (document.getElementById('gSearch').value||'').toLowerCase();
  document.querySelectorAll('.issue-row').forEach(function(row){
    var sev = row.dataset.sev;
    var plat = row.dataset.plat || '';
    var sevOk = gFilt==='all' || sev===gFilt;
    var platOk = platMatch(plat, pFilt);
    var qOk = !q || row.textContent.toLowerCase().includes(q);
    row.classList.toggle('hidden', !(sevOk&&platOk&&qOk));
  });
}

// ── Pages filter & pagination ─────────────────────────────────
var pagePgSize = 50, pagePgCur = 0, pageRows = [];
function initPages(){
  pageRows = Array.from(document.querySelectorAll('.page-row'));
  renderPagePg();
}
function filterPages(){
  var q = (document.getElementById('pSearch').value||'').toLowerCase();
  pageRows.forEach(function(r){ r.classList.toggle('hidden', q && !r.textContent.toLowerCase().includes(q)); });
  pagePgCur=0; renderPagePg();
}
function renderPagePg(){
  var vis = pageRows.filter(function(r){return !r.classList.contains('hidden')});
  var s = pagePgCur*pagePgSize, e = s+pagePgSize;
  vis.forEach(function(r,i){ r.style.display=(i>=s&&i<e)?'':'none'; });
  var pg = document.getElementById('pagePg');
  var total = Math.ceil(vis.length/pagePgSize);
  if(total<=1){pg.innerHTML='';return;}
  pg.innerHTML='<button onclick="goPgP(-1)" '+(pagePgCur===0?'disabled':'')+'>← Prev</button>'+
    '<span>'+(pagePgCur+1)+' / '+total+' ('+vis.length+' pages)</span>'+
    '<button onclick="goPgP(1)" '+(pagePgCur>=total-1?'disabled':'')+'>Next →</button>';
}
function goPgP(d){
  var vis=pageRows.filter(function(r){return !r.classList.contains('hidden')});
  var t=Math.ceil(vis.length/pagePgSize);
  pagePgCur=Math.max(0,Math.min(t-1,pagePgCur+d)); renderPagePg();
}

// ── Active nav highlight ──────────────────────────────────────
var sections = document.querySelectorAll('[id]');
window.addEventListener('scroll', function(){
  var pos = window.scrollY + 120;
  sections.forEach(function(s){
    var link = document.querySelector('.nav-link[href="#'+s.id+'"]');
    if(!link) return;
    var top = s.offsetTop, bot = top + s.offsetHeight;
    link.classList.toggle('active', pos >= top && pos < bot);
  });
}, {passive:true});

initPages();
</script>
</body>
</html>`
