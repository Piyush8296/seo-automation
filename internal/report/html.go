package report

import (
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cars24/seo-automation/internal/checks"
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

// CheckCategoryGroup groups check descriptors by category for template rendering.
type CheckCategoryGroup struct {
	Category string
	Icon     string
	Checks   []models.CheckDescriptor
}

type templateData struct {
	*models.SiteAudit
	AllIssues      []Issue
	TopIssues      []Issue
	CategoryStats  []CategoryStat
	ErrorIssues    []Issue
	WarnIssues     []Issue
	NoticeIssues   []Issue
	HasMobileData  bool
	CheckGroups    []CheckCategoryGroup
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

	var checkGroups []CheckCategoryGroup
	for _, d := range checks.GetCheckDescriptors() {
		if len(checkGroups) == 0 || checkGroups[len(checkGroups)-1].Category != d.Category {
			ic := catIcons[d.Category]
			if ic == "" {
				ic = "📌"
			}
			checkGroups = append(checkGroups, CheckCategoryGroup{Category: d.Category, Icon: ic})
		}
		checkGroups[len(checkGroups)-1].Checks = append(checkGroups[len(checkGroups)-1].Checks, d)
	}

	data := templateData{
		SiteAudit:     audit,
		AllIssues:     allIssues,
		TopIssues:     top,
		CategoryStats: cats,
		ErrorIssues:   errI,
		WarnIssues:    warnI,
		NoticeIssues:  noticeI,
		HasMobileData: hasMobileData,
		CheckGroups:   checkGroups,
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
				return "#3fe56c"
			case "B":
				return "#8ed793"
			case "C":
				return "#ffb7ae"
			case "D":
				return "#ffb4ab"
			default:
				return "#ffb4ab"
			}
		},
		"scoreCol": func(score float64) string {
			if score >= 80 {
				return "#3fe56c"
			} else if score >= 60 {
				return "#ffb7ae"
			}
			return "#ffb4ab"
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
		"checkGroupCount": func(groups []CheckCategoryGroup) int {
			n := 0
			for _, g := range groups {
				n += len(g.Checks)
			}
			return n
		},
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
<title>SEO Observatory — {{.SiteURL}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=Inter:wght@300;400;500;600&display=swap" rel="stylesheet">
<style>
/* ── Reset & design tokens ────────────────────────────────── */
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#0e131e;
  --s1:#161c26;
  --s2:#1a202a;
  --s3:#242a35;
  --s4:#2f3540;
  --bright:#343945;
  --text:#dde2f1;
  --muted:#bbcbb8;
  --faint:#869583;
  --ov:#3c4a3c;
  --primary:#3fe56c;
  --primary-dim:#00c853;
  --on-primary:#003912;
  --secondary:#8ed793;
  --error:#ffb4ab;   --error-bg:rgba(147,0,10,.25);
  --warn:#ffb7ae;    --warn-bg:rgba(118,37,31,.22);
  --notice:#8ed793;  --notice-bg:rgba(2,83,30,.22);
  --r:6px;--r2:10px;--r3:14px;
}
html{scroll-behavior:smooth}
body{font-family:'Inter',system-ui,sans-serif;background:var(--bg);color:var(--text);line-height:1.6;font-size:14px;-webkit-font-smoothing:antialiased}
a{color:var(--primary);text-decoration:none}a:hover{opacity:.8}
h1{font-family:'Space Grotesk',sans-serif;font-size:24px;font-weight:700;letter-spacing:-.5px}
h2{font-family:'Space Grotesk',sans-serif;font-size:17px;font-weight:700;letter-spacing:-.3px}
h3{font-family:'Space Grotesk',sans-serif;font-size:14px;font-weight:600}
code{font-family:'JetBrains Mono',Menlo,monospace;font-size:10px;background:var(--s4);padding:2px 6px;border-radius:4px;color:var(--primary)}

/* Scrollbar */
::-webkit-scrollbar{width:5px;height:5px}
::-webkit-scrollbar-track{background:var(--bg)}
::-webkit-scrollbar-thumb{background:var(--s4);border-radius:3px}
::-webkit-scrollbar-thumb:hover{background:var(--bright)}

/* ── Layout ───────────────────────────────────────────────── */
.shell{display:flex;min-height:100vh}
aside{width:236px;flex-shrink:0;background:var(--s1);position:sticky;top:0;height:100vh;overflow-y:auto;display:flex;flex-direction:column}
.brand{padding:20px 16px 14px}
.brand-logo{display:flex;align-items:center;gap:8px;font-family:'Space Grotesk',sans-serif;font-weight:700;font-size:14px;color:var(--text);margin-bottom:4px}
.brand-dot{width:28px;height:28px;border-radius:7px;background:linear-gradient(135deg,#3fe56c,#00c853);display:flex;align-items:center;justify-content:center;font-size:13px;color:#003912;font-weight:700;flex-shrink:0}
.brand-url{font-size:10px;color:var(--muted);word-break:break-all;padding-left:36px;line-height:1.4}
.nav-sep{height:1px;background:rgba(60,74,60,.35);margin:6px 0}
.nav-label{padding:8px 16px 4px;font-size:9px;font-weight:700;text-transform:uppercase;letter-spacing:.12em;color:var(--faint)}
.nav-link{display:flex;align-items:center;gap:8px;padding:7px 16px;color:var(--muted);font-size:12px;transition:all .15s;cursor:pointer;border-right:2px solid transparent;text-decoration:none}
.nav-link:hover{background:var(--bright);color:var(--text)}
.nav-link.active{background:var(--s2);color:var(--primary);border-right-color:var(--primary-dim)}
.nav-link .icon{font-size:13px;width:16px;text-align:center;flex-shrink:0}
.nav-count{margin-left:auto;font-size:9px;background:var(--s3);border-radius:8px;padding:1px 6px;color:var(--muted);font-weight:600}
.nav-count.err{background:var(--error-bg);color:var(--error)}
.nav-footer{margin-top:auto;padding:14px 16px;font-size:10px;color:var(--faint);border-top:1px solid rgba(60,74,60,.3);line-height:1.6}
main{flex:1;min-width:0;padding:28px 32px}

/* ── Cover ────────────────────────────────────────────────── */
.cover{background:var(--s1);border-radius:var(--r3);padding:32px 36px;margin-bottom:24px;display:flex;align-items:flex-start;justify-content:space-between;gap:24px;position:relative;overflow:hidden}
.cover::before{content:'';position:absolute;inset:0;opacity:.03;background-image:radial-gradient(#00C853 .5px,transparent .5px);background-size:18px 18px;pointer-events:none}
.cover-left h1{margin-bottom:6px}
.cover-meta{display:flex;flex-wrap:wrap;gap:8px;margin-top:14px}
.cover-chip{display:flex;align-items:center;gap:5px;background:var(--s3);border-radius:20px;padding:4px 11px;font-size:11px;color:var(--muted)}
.cover-chip strong{color:var(--text)}
.cover-right{display:flex;flex-direction:column;align-items:center;gap:14px;flex-shrink:0}
.score-ring{position:relative;display:flex;align-items:center;justify-content:center}
.score-ring svg{transform:rotate(-90deg)}
.score-inner{position:absolute;text-align:center;pointer-events:none}
.score-num{font-family:'Space Grotesk',sans-serif;font-size:32px;font-weight:700;line-height:1}
.score-lbl{font-size:9px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin-top:2px}
.grade-pill{padding:5px 18px;border-radius:20px;font-family:'Space Grotesk',sans-serif;font-size:20px;font-weight:700}
.score-grid{display:flex;gap:20px;align-items:center;flex-wrap:wrap}
.score-block{display:flex;flex-direction:column;align-items:center;gap:6px}
.score-block-label{font-size:10px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);font-weight:600}

/* ── Metric cards ─────────────────────────────────────────── */
.metric-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(145px,1fr));gap:12px;margin-bottom:24px}
.metric-card{background:var(--s1);border-radius:var(--r2);padding:16px 18px;position:relative;overflow:hidden}
.metric-card::before{content:'';position:absolute;left:0;top:0;bottom:0;width:2px;border-radius:0 1px 1px 0}
.metric-card.err::before{background:var(--error)}
.metric-card.warn::before{background:var(--warn)}
.metric-card.notice::before{background:var(--notice)}
.metric-card.ok::before{background:var(--primary)}
.mc-label{font-size:9px;font-weight:600;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin-bottom:8px}
.mc-value{font-family:'Space Grotesk',sans-serif;font-size:28px;font-weight:700;line-height:1}
.mc-value.err{color:var(--error)}
.mc-value.warn{color:var(--warn)}
.mc-value.notice{color:var(--notice)}
.mc-value.ok{color:var(--primary)}
.mc-sub{font-size:10px;color:var(--muted);margin-top:4px}

/* ── Sections ─────────────────────────────────────────────── */
.section{margin-bottom:32px}
.section-header{display:flex;align-items:center;gap:10px;margin-bottom:16px;padding-bottom:10px;border-bottom:1px solid rgba(60,74,60,.35)}
.section-header h2{flex:1}
.section-badge{background:var(--s3);border-radius:20px;padding:2px 10px;font-size:11px;color:var(--muted)}

/* ── Category grid ────────────────────────────────────────── */
.cat-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(210px,1fr));gap:12px}
.cat-card{background:var(--s1);border-radius:var(--r2);padding:14px 16px;cursor:pointer;transition:background .15s,transform .15s;text-decoration:none;color:inherit;display:block}
.cat-card:hover{background:var(--s2);transform:translateY(-1px)}
.cat-top{display:flex;align-items:center;gap:8px;margin-bottom:10px}
.cat-icon{font-size:16px;line-height:1}
.cat-name{font-size:12px;font-weight:600;flex:1}
.cat-score{font-size:12px;font-weight:700}
.cat-bar{height:3px;background:var(--s4);border-radius:2px;overflow:hidden;margin-bottom:8px}
.cat-bar-fill{height:100%;border-radius:2px;transition:width .4s}
.cat-counts{display:flex;gap:6px}
.cc{font-size:10px;font-weight:600;padding:2px 6px;border-radius:4px}
.cc.e{background:var(--error-bg);color:var(--error)}
.cc.w{background:var(--warn-bg);color:var(--warn)}
.cc.n{background:var(--notice-bg);color:var(--notice)}

/* ── Priority items ───────────────────────────────────────── */
.priority-list{display:flex;flex-direction:column;gap:8px}
.priority-item{background:var(--s1);border-radius:var(--r2);padding:14px 18px;display:flex;gap:12px;align-items:flex-start;position:relative;overflow:hidden}
.priority-item::before{content:'';position:absolute;left:0;top:0;bottom:0;width:3px;background:var(--error);border-radius:0 2px 2px 0}
.priority-item.warn-item::before{background:var(--warn)}
.priority-num{font-family:'Space Grotesk',sans-serif;font-size:18px;font-weight:700;color:var(--faint);line-height:1;flex-shrink:0;width:24px;text-align:right}
.priority-body{flex:1;min-width:0}
.priority-title{font-weight:600;margin-bottom:4px;display:flex;align-items:center;gap:6px;flex-wrap:wrap}
.priority-url{font-size:11px;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;margin-bottom:3px}
.priority-detail{font-size:11px;color:var(--faint);font-style:italic}

/* ── Badges ───────────────────────────────────────────────── */
.badge{display:inline-flex;align-items:center;gap:3px;padding:2px 7px;border-radius:4px;font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.04em;white-space:nowrap}
.badge.e{background:var(--error-bg);color:var(--error)}
.badge.w{background:var(--warn-bg);color:var(--warn)}
.badge.n{background:var(--notice-bg);color:var(--notice)}
.badge::before{content:'●';font-size:6px}

/* ── Collapsible category sections ───────────────────────── */
.cat-section{background:var(--s1);border-radius:var(--r2);margin-bottom:8px;overflow:hidden}
.cat-section-hdr{display:flex;align-items:center;gap:10px;padding:12px 16px;cursor:pointer;user-select:none;transition:background .15s}
.cat-section-hdr:hover{background:var(--bright)}
.cat-section-hdr .arrow{margin-left:auto;color:var(--faint);transition:transform .2s;font-size:11px}
.cat-section-hdr.open .arrow{transform:rotate(90deg)}
.cat-section-body{display:none;border-top:1px solid rgba(60,74,60,.3)}
.cat-section-body.open{display:block}

/* ── Tables ───────────────────────────────────────────────── */
.tbl-wrap{overflow-x:auto}
table{width:100%;border-collapse:collapse;font-size:12px}
thead th{padding:8px 12px;background:var(--s3);font-size:9px;font-weight:700;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);text-align:left;white-space:nowrap;border-bottom:1px solid rgba(60,74,60,.3)}
tbody td{padding:8px 12px;border-bottom:1px solid rgba(60,74,60,.2);vertical-align:middle}
tbody tr:last-child td{border-bottom:none}
tbody tr:hover td{background:var(--bright)}
.td-id code{font-size:9px}
.td-url{max-width:260px}
.td-url a{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--primary);font-size:11px}
.td-msg{max-width:280px}
.td-detail{max-width:180px;font-size:10px;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

/* ── Toolbar ──────────────────────────────────────────────── */
.toolbar{display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap;align-items:center}
.search-box{flex:1;min-width:200px;max-width:300px;background:var(--s4);border:1px solid transparent;border-radius:var(--r);padding:7px 12px;color:var(--text);font-size:12px;outline:none;transition:border-color .15s;font-family:'Inter',sans-serif}
.search-box:focus{border-color:rgba(63,229,108,.4)}
.search-box::placeholder{color:var(--faint)}
.filt-btn{background:var(--s2);border-radius:var(--r);padding:6px 12px;color:var(--muted);font-size:11px;font-weight:600;cursor:pointer;transition:all .15s;white-space:nowrap;border:none}
.filt-btn:hover,.filt-btn.on{color:var(--primary);background:rgba(63,229,108,.1)}
.pg-bar{display:flex;gap:8px;margin-top:10px;align-items:center;font-size:11px;color:var(--muted)}
.pg-bar button{background:var(--s2);border:none;border-radius:var(--r);padding:4px 10px;color:var(--text);cursor:pointer;font-size:11px;font-family:'Inter',sans-serif}
.pg-bar button:disabled{opacity:.3;cursor:default}

/* ── Status codes ─────────────────────────────────────────── */
.status-ok{color:var(--primary);font-weight:700}
.status-redir{color:var(--warn);font-weight:700}
.status-err{color:var(--error);font-weight:700}
.depth-pip{display:inline-flex;gap:2px}
.dp{width:5px;height:5px;border-radius:50%;background:var(--s4)}
.dp.fill{background:var(--primary)}

/* ── Charts ───────────────────────────────────────────────── */
.charts-row{display:grid;grid-template-columns:1fr 1fr;gap:14px;margin-bottom:24px}
.chart-card{background:var(--s1);border-radius:var(--r2);padding:18px 22px}
.chart-title{font-size:10px;font-weight:600;color:var(--muted);margin-bottom:14px;text-transform:uppercase;letter-spacing:.08em}
.donut-wrap{display:flex;align-items:center;gap:20px}
.donut-legend{display:flex;flex-direction:column;gap:8px}
.dl-item{display:flex;align-items:center;gap:8px;font-size:12px}
.dl-dot{width:8px;height:8px;border-radius:50%;flex-shrink:0}
.dl-val{margin-left:auto;font-weight:700;min-width:32px;text-align:right}
.bar-chart{display:flex;flex-direction:column;gap:7px}
.bc-row{display:flex;align-items:center;gap:10px;font-size:11px}
.bc-label{width:120px;text-align:right;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex-shrink:0}
.bc-track{flex:1;height:14px;background:var(--s3);border-radius:3px;overflow:hidden}
.bc-fill{height:100%;border-radius:3px;min-width:3px;transition:width .5s}
.bc-num{width:32px;text-align:right;font-weight:600;color:var(--text)}

/* ── Platform badges ──────────────────────────────────────── */
.plat{display:inline-flex;align-items:center;padding:1px 6px;border-radius:4px;font-size:9px;font-weight:700;white-space:nowrap;letter-spacing:.02em}
.plat-m{background:rgba(142,215,147,.12);color:#8ed793}
.plat-d{background:rgba(63,229,108,.12);color:#3fe56c}
.plat-diff{background:rgba(255,183,174,.12);color:#ffb7ae}
.plat-both{display:none}

/* ── Misc ─────────────────────────────────────────────────── */
.empty{text-align:center;padding:36px;color:var(--faint);font-size:13px}
.hidden{display:none!important}
.pill{display:inline-block;padding:1px 7px;border-radius:8px;font-size:10px;font-weight:600}

/* ── Print ────────────────────────────────────────────────── */
@media print{
  aside{display:none}
  main{padding:0}
  .toolbar,.pg-bar{display:none}
  .cat-section-body{display:block!important}
  body{background:#fff;color:#111}
  .cover,.metric-card,.cat-card,.cat-section,.chart-card,.priority-item{background:#fff}
}
</style>
</head>
<body>
<div class="shell">

<!-- ── Sidebar ──────────────────────────────────────────── -->
<aside>
  <div class="brand">
    <div class="brand-logo">
      <div class="brand-dot">◉</div>
      SEO Observatory
    </div>
    <div class="brand-url">{{.SiteURL}}</div>
  </div>
  <nav>
    <div class="nav-label">Report</div>
    <a class="nav-link" href="#overview"><span class="icon">◈</span>Overview</a>
    <a class="nav-link" href="#categories"><span class="icon">◫</span>Categories</a>
    <a class="nav-link" href="#priority"><span class="icon">◎</span>Priority Issues<span class="nav-count err">{{len .TopIssues}}</span></a>
    <a class="nav-link" href="#all-issues"><span class="icon">≡</span>All Issues<span class="nav-count">{{len .AllIssues}}</span></a>
    <a class="nav-link" href="#pages"><span class="icon">◱</span>Pages<span class="nav-count">{{.PagesCrawled}}</span></a>
    <a class="nav-link" href="#checks"><span class="icon">✦</span>Checks Catalog<span class="nav-count">{{checkGroupCount .CheckGroups}}</span></a>
    <div class="nav-sep"></div>
    <div class="nav-label">By Severity</div>
    <a class="nav-link" href="#sec-errors"><span class="icon">●</span>Errors<span class="nav-count err">{{len .ErrorIssues}}</span></a>
    <a class="nav-link" href="#sec-warnings"><span class="icon">●</span>Warnings<span class="nav-count">{{len .WarnIssues}}</span></a>
    <a class="nav-link" href="#sec-notices"><span class="icon">●</span>Notices<span class="nav-count">{{len .NoticeIssues}}</span></a>
  </nav>
  <div class="nav-footer">
    Generated {{.CrawledAt.Format "Jan 02, 2006 · 15:04 UTC"}}<br>
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
            <circle id="scoreArc" cx="60" cy="60" r="54" fill="none" stroke="#3fe56c" stroke-width="10"
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
            <circle id="desktopArc" cx="60" cy="60" r="54" fill="none" stroke="#3fe56c" stroke-width="10"
              stroke-dasharray="339.3 339.3" stroke-linecap="round"
              style="transition:stroke-dasharray .8s ease"/>
          </svg>
          <div class="score-inner">
            <div class="score-num" id="desktopNum" style="font-size:22px">{{printf "%.0f" .DesktopHealthScore}}</div>
            <div class="score-lbl" style="font-size:9px">/ 100</div>
          </div>
        </div>
        <div style="font-size:13px;font-weight:800;color:#3fe56c">{{.DesktopGrade}}</div>
      </div>
      <!-- Mobile score -->
      <div class="score-block">
        <div class="score-block-label">📱 Mobile</div>
        <div class="score-ring">
          <svg width="80" height="80" viewBox="0 0 120 120">
            <circle cx="60" cy="60" r="54" fill="none" stroke="var(--s3)" stroke-width="10"/>
            <circle id="mobileArc" cx="60" cy="60" r="54" fill="none" stroke="#8ed793" stroke-width="10"
              stroke-dasharray="339.3 339.3" stroke-linecap="round"
              style="transition:stroke-dasharray .8s ease"/>
          </svg>
          <div class="score-inner">
            <div class="score-num" id="mobileNum" style="font-size:22px">{{printf "%.0f" .MobileHealthScore}}</div>
            <div class="score-lbl" style="font-size:9px">/ 100</div>
          </div>
        </div>
        <div style="font-size:13px;font-weight:800;color:#8ed793">{{.MobileGrade}}</div>
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
  <div class="metric-card ok"><div class="mc-label">Overall Score</div><div class="mc-value ok" id="scoreCard">{{printf "%.1f" .HealthScore}}</div><div class="mc-sub">Grade: {{.Grade}}</div></div>
  {{if .HasMobileData}}
  <div class="metric-card" style="border-left:3px solid #3fe56c"><div class="mc-label">🖥 Desktop Score</div><div class="mc-value" style="color:#3fe56c">{{printf "%.1f" .DesktopHealthScore}}</div><div class="mc-sub">Grade: {{.DesktopGrade}} · {{.DesktopStats.Errors}} err / {{.DesktopStats.Warnings}} warn</div></div>
  <div class="metric-card" style="border-left:3px solid #8ed793"><div class="mc-label">📱 Mobile Score</div><div class="mc-value" style="color:#8ed793">{{printf "%.1f" .MobileHealthScore}}</div><div class="mc-sub">Grade: {{.MobileGrade}} · {{.MobileStats.Errors}} err / {{.MobileStats.Warnings}} warn</div></div>
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
        <div class="dl-item"><div class="dl-dot" style="background:var(--error)"></div>Errors<div class="dl-val" style="color:var(--error)">{{len .ErrorIssues}}</div></div>
        <div class="dl-item"><div class="dl-dot" style="background:var(--warn)"></div>Warnings<div class="dl-val" style="color:var(--warn)">{{len .WarnIssues}}</div></div>
        <div class="dl-item"><div class="dl-dot" style="background:var(--notice)"></div>Notices<div class="dl-val" style="color:var(--notice)">{{len .NoticeIssues}}</div></div>
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
        {{if eq .Total 0}}<span style="font-size:11px;color:var(--primary)">✓ Clean</span>{{end}}
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
    <span style="width:1px;height:20px;background:rgba(60,74,60,.35);margin:0 4px"></span>
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
          <td>{{if gt (len $p.CheckResults) 0}}<span class="pill" style="background:var(--error-bg);color:var(--error)">{{len $p.CheckResults}}</span>{{else}}<span style="color:var(--primary)">✓</span>{{end}}</td>
          <td>
            <div class="depth-pip">
              {{range $d := iterate $p.Depth}}<span class="dp fill"></span>{{end}}
            </div>
            <span style="font-size:11px;color:var(--muted);margin-left:4px">{{$p.Depth}}</span>
          </td>
          <td>{{if $p.InSitemap}}<span style="color:var(--primary)">✓</span>{{else}}<span style="color:var(--faint)">—</span>{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  <div class="pg-bar" id="pagePg"></div>
</div>

<!-- Checks Catalog -->
<div id="checks" class="section">
  <div class="section-header">
    <h2>✦ Checks Catalog</h2>
    <span class="section-badge">{{checkGroupCount .CheckGroups}} checks · {{len .CheckGroups}} categories</span>
  </div>
  {{range .CheckGroups}}
  <div class="cat-section" style="margin-bottom:6px">
    <div class="cat-section-hdr" onclick="toggleCat(this)">
      <span style="font-size:15px">{{.Icon}}</span>
      <span style="font-weight:600">{{.Category}}</span>
      <span class="nav-count" style="margin-left:8px">{{len .Checks}}</span>
      <span class="arrow">▶</span>
    </div>
    <div class="cat-section-body" style="padding:12px 16px 10px;display:flex;flex-wrap:wrap;gap:6px">
      {{range .Checks}}
      <code style="font-size:10px;background:var(--s3);padding:3px 9px;border-radius:4px;color:var(--muted)">{{.ID}}</code>
      {{end}}
    </div>
  </div>
  {{end}}
</div>

</main>
</div>

<script>
// ── Score ring animation ──────────────────────────────────────
(function(){
  var circ = 339.3;
  var colors = {A:'#3fe56c',B:'#8ed793',C:'#ffb7ae',D:'#ffb4ab',F:'#ffb4ab'};

  function animateRing(arcId, score, grade) {
    var arc = document.getElementById(arcId);
    if(!arc) return;
    var col = colors[grade] || '#ffb4ab';
    arc.style.stroke = col;
    arc.style.strokeDasharray = (circ * score / 100) + ' ' + circ;
  }

  animateRing('scoreArc', {{printf "%.1f" .HealthScore}}, '{{.Grade}}');
  animateRing('desktopArc', {{printf "%.1f" .DesktopHealthScore}}, '{{.DesktopGrade}}');
  animateRing('mobileArc', {{printf "%.1f" .MobileHealthScore}}, '{{.MobileGrade}}');

  // Style overall grade pill
  var score = {{printf "%.1f" .HealthScore}};
  var grade = '{{.Grade}}';
  var col = colors[grade] || '#ffb4ab';
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
  if(total === 0){ ctx.fillStyle='#161c26'; ctx.beginPath(); ctx.arc(60,60,45,0,Math.PI*2); ctx.fill(); return; }
  var data = [{v:E,c:'#ffb4ab'},{v:W,c:'#ffb7ae'},{v:N,c:'#8ed793'}];
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
  ctx.fillStyle='#0e131e'; ctx.fill();
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
    var fillCol = cat.errors>0?'#ffb4ab':cat.warnings>0?'#ffb7ae':'#8ed793';
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
