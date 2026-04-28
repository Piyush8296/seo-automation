import { Fragment, useEffect, useMemo, useRef, useState } from 'react'
import {
  AlertCircle,
  AlertTriangle,
  Info,
  Search,
  ChevronDown,
  ChevronRight,
  ChevronLeft,
  X,
  ExternalLink,
} from 'lucide-react'
import { api } from '../lib/api'

const PAGE_SIZE = 25

const SEVERITY_META = {
  error:   { icon: AlertCircle,   color: '#ffb4ab', bg: 'rgba(147,0,10,0.2)',  label: 'Error'   },
  warning: { icon: AlertTriangle, color: '#ffb7ae', bg: 'rgba(118,37,31,0.2)', label: 'Warning' },
  notice:  { icon: Info,          color: '#8ed793', bg: 'rgba(2,83,30,0.2)',   label: 'Notice'  },
}

const PLATFORM_META = {
  desktop: { label: '🖥 Desktop', color: '#3fe56c',  bg: 'rgba(63,229,108,0.12)'  },
  mobile:  { label: '📱 Mobile',  color: '#8ed793',  bg: 'rgba(142,215,147,0.12)' },
  diff:    { label: '🔄 M↔D',     color: '#ffb7ae',  bg: 'rgba(255,183,174,0.12)' },
  both:    { label: '⊕ Both',     color: '#bbcbb8',  bg: 'rgba(187,203,184,0.08)' },
}

function flattenIssues(report) {
  if (!report) return []
  const out = []
  for (const r of report.site_checks ?? []) {
    out.push({ ...r, scope: 'site', pageURL: r.url || '(site-wide)' })
  }
  for (const page of report.pages ?? []) {
    for (const r of page.check_results ?? []) {
      out.push({ ...r, scope: 'page', pageURL: r.url || page.url })
    }
  }
  return out
}

function severityRank(s) {
  return s === 'error' ? 0 : s === 'warning' ? 1 : s === 'notice' ? 2 : 3
}

function SeverityBadge({ severity }) {
  const meta = SEVERITY_META[severity] ?? SEVERITY_META.notice
  const Icon = meta.icon
  return (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium"
      style={{ color: meta.color, background: meta.bg }}
    >
      <Icon size={12} />
      {meta.label}
    </span>
  )
}

function SearchableSelect({ value, onChange, options, placeholder }) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const ref = useRef(null)

  const selected = options.find((o) => o.value === value)
  const filtered = query
    ? options.filter((o) => o.label.toLowerCase().includes(query.toLowerCase()))
    : options

  useEffect(() => {
    if (!open) setQuery('')
  }, [open])

  useEffect(() => {
    const handler = (e) => { if (ref.current && !ref.current.contains(e.target)) setOpen(false) }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  return (
    <div ref={ref} className="relative min-w-[150px]">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="input text-sm w-full text-left flex items-center justify-between gap-2"
        style={{ cursor: 'pointer' }}
      >
        <span className={selected ? 'text-on-surface' : 'text-on-surface-variant'}>
          {selected ? selected.label : placeholder}
        </span>
        <ChevronDown size={12} className="shrink-0 text-on-surface-variant" />
      </button>
      {open && (
        <div
          className="absolute z-50 mt-1 w-full rounded-lg overflow-hidden shadow-lg"
          style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.5)', minWidth: '180px' }}
        >
          <div className="p-1.5" style={{ borderBottom: '1px solid rgba(60,74,60,0.3)' }}>
            <div className="relative">
              <Search size={11} className="absolute left-2 top-1/2 -translate-y-1/2 text-on-surface-variant" />
              <input
                autoFocus
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search…"
                className="w-full bg-surface-container-highest rounded px-2 py-1 pl-6 text-xs text-on-surface outline-none"
                style={{ border: '1px solid rgba(60,74,60,0.3)' }}
              />
            </div>
          </div>
          <div className="max-h-48 overflow-y-auto py-1">
            <button
              type="button"
              onClick={() => { onChange(''); setOpen(false) }}
              className="w-full text-left px-3 py-1.5 text-xs hover:bg-surface-bright transition-colors text-on-surface-variant"
            >
              {placeholder}
            </button>
            {filtered.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => { onChange(opt.value); setOpen(false) }}
                className="w-full text-left px-3 py-1.5 text-xs hover:bg-surface-bright transition-colors"
                style={{ color: opt.value === value ? '#3fe56c' : '#dde2f1' }}
              >
                {opt.label}
              </button>
            ))}
            {filtered.length === 0 && (
              <p className="px-3 py-2 text-xs text-on-surface-variant italic">No matches</p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function PlatformBadge({ platform }) {
  const key = (!platform || platform === 'both') ? 'both' : platform
  const meta = PLATFORM_META[key]
  if (!meta) return null
  return (
    <span
      className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium whitespace-nowrap"
      style={{ color: meta.color, background: meta.bg }}
    >
      {meta.label}
    </span>
  )
}

function ChecklistIDBadges({ ids }) {
  if (!ids?.length) return null
  return (
    <div className="mt-1 flex flex-wrap gap-1">
      {ids.map((id) => (
        <span
          key={id}
          className="font-mono rounded px-1.5 py-0.5"
          style={{ fontSize: '9px', color: '#9cc7ff', background: 'rgba(156,199,255,0.08)', border: '1px solid rgba(156,199,255,0.12)' }}
        >
          {id}
        </span>
      ))}
    </div>
  )
}

export default function IssueTable({ auditId }) {
  const [report, setReport] = useState(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)

  const [severity, setSeverity] = useState('')
  const [category, setCategory] = useState('')
  const [checkID, setCheckID] = useState('')
  const [platform, setPlatform] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [expanded, setExpanded] = useState(() => new Set())

  useEffect(() => {
    setLoading(true)
    api.getReportJSON(auditId)
      .then((r) => { setReport(r); setErr('') })
      .catch((e) => setErr(e.message || 'Failed to load report'))
      .finally(() => setLoading(false))
  }, [auditId])

  const allIssues = useMemo(() => flattenIssues(report), [report])

  const categories = useMemo(() => {
    const s = new Set()
    allIssues.forEach((i) => i.category && s.add(i.category))
    return [...s].sort()
  }, [allIssues])

  const checkIDs = useMemo(() => {
    const s = new Set()
    allIssues.forEach((i) => {
      if (i.id) s.add(i.id)
      for (const id of i.checklist_ids ?? []) s.add(id)
    })
    return [...s].sort()
  }, [allIssues])

  const hasPlatformData = useMemo(
    () => allIssues.some((i) => i.platform && i.platform !== 'both' && i.platform !== ''),
    [allIssues]
  )

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return allIssues
      .filter((i) => !severity || i.severity === severity)
      .filter((i) => !category || i.category === category)
      .filter((i) => !checkID || i.id === checkID || (i.checklist_ids ?? []).includes(checkID))
      .filter((i) => {
        if (!platform) return true
        const p = i.platform || 'both'
        if (platform === 'desktop') return p === 'desktop' || p === 'both' || p === ''
        if (platform === 'mobile')  return p === 'mobile'  || p === 'both' || p === ''
        if (platform === 'both')    return !p || p === 'both'
        return p === platform
      })
      .filter((i) =>
        !q ||
        (i.pageURL || '').toLowerCase().includes(q) ||
        (i.message || '').toLowerCase().includes(q) ||
        (i.id || '').toLowerCase().includes(q) ||
        (i.checklist_ids ?? []).some((id) => id.toLowerCase().includes(q))
      )
      .sort((a, b) => severityRank(a.severity) - severityRank(b.severity))
  }, [allIssues, severity, category, checkID, platform, search])

  useEffect(() => { setPage(1); setExpanded(new Set()) }, [severity, category, checkID, platform, search])

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE))
  const safePage = Math.min(page, totalPages)
  const pageStart = (safePage - 1) * PAGE_SIZE
  const pageItems = filtered.slice(pageStart, pageStart + PAGE_SIZE)

  const toggleRow = (idx) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      next.has(idx) ? next.delete(idx) : next.add(idx)
      return next
    })
  }

  const clearFilters = () => { setSeverity(''); setCategory(''); setCheckID(''); setPlatform(''); setSearch('') }
  const hasFilters = severity || category || checkID || platform || search

  if (loading) {
    return (
      <div className="card p-12 flex items-center justify-center">
        <span
          className="w-8 h-8 rounded-full animate-spin"
          style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }}
        />
      </div>
    )
  }

  if (err) {
    return (
      <div className="card p-8 text-center">
        <AlertCircle className="mx-auto mb-3" size={28} style={{ color: '#ffb4ab' }} />
        <p className="text-on-surface-variant">{err}</p>
      </div>
    )
  }

  return (
    <div className="card flex flex-col">
      {/* Filters — single row */}
      <div
        className="px-4 py-3 flex items-center gap-2 flex-wrap"
        style={{ borderBottom: '1px solid rgba(60,74,60,0.35)' }}
      >
        <div className="relative flex-1 min-w-[180px]">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search URL, message, or checklist ID…"
            className="input text-sm pl-8"
          />
        </div>
        <SearchableSelect
          value={severity}
          onChange={setSeverity}
          placeholder="All severities"
          options={[
            { value: 'error',   label: '● Error'   },
            { value: 'warning', label: '● Warning' },
            { value: 'notice',  label: '● Notice'  },
          ]}
        />
        <SearchableSelect
          value={category}
          onChange={setCategory}
          placeholder="All categories"
          options={categories.map((c) => ({ value: c, label: c }))}
        />
        <SearchableSelect
          value={checkID}
          onChange={setCheckID}
          placeholder="All checks"
          options={checkIDs.map((c) => ({ value: c, label: c }))}
        />
        <SearchableSelect
          value={platform}
          onChange={setPlatform}
          placeholder="All platforms"
          options={[
            { value: 'desktop', label: '🖥 Desktop' },
            { value: 'mobile',  label: '📱 Mobile'  },
            { value: 'diff',    label: '🔄 M↔D Diff' },
            { value: 'both',    label: '⊕ Both'      },
          ]}
        />
        {hasFilters && (
          <button onClick={clearFilters} className="btn-ghost" title="Clear filters">
            <X size={14} /> Clear
          </button>
        )}
      </div>

      {/* Result count */}
      <div
        className="px-4 py-2 text-xs text-on-surface-variant flex items-center justify-between"
        style={{ borderBottom: '1px solid rgba(60,74,60,0.35)' }}
      >
        <span>
          {filtered.length.toLocaleString()} {filtered.length === 1 ? 'issue' : 'issues'}
          {filtered.length !== allIssues.length && ` (filtered from ${allIssues.length.toLocaleString()})`}
        </span>
        <span>Page {safePage} of {totalPages}</span>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        {pageItems.length === 0 ? (
          <div className="p-12 text-center text-on-surface-variant text-sm">
            No issues match the current filters.
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead
              className="text-xs text-on-surface-variant uppercase tracking-wider"
              style={{ background: '#242a35' }}
            >
              <tr>
                <th className="w-8" />
                <th className="text-left font-medium px-3 py-2.5">Severity</th>
                <th className="text-left font-medium px-3 py-2.5">Check / Checklist</th>
                <th className="text-left font-medium px-3 py-2.5">Category</th>
                <th className="text-left font-medium px-3 py-2.5">Platform</th>
                <th className="text-left font-medium px-3 py-2.5">URL</th>
                <th className="text-left font-medium px-3 py-2.5">Message</th>
              </tr>
            </thead>
            <tbody>
              {pageItems.map((issue, idx) => {
                const rowIdx = pageStart + idx
                const isOpen = expanded.has(rowIdx)
                const hasDetails = !!issue.details
                return (
                  <Fragment key={rowIdx}>
                    <tr
                      className={`hover:bg-surface-bright transition-colors ${hasDetails ? 'cursor-pointer' : ''}`}
                      style={{ borderTop: '1px solid rgba(60,74,60,0.25)' }}
                      onClick={() => hasDetails && toggleRow(rowIdx)}
                    >
                      <td className="px-2 align-top pt-3 text-on-surface-variant">
                        {hasDetails ? (isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />) : null}
                      </td>
                      <td className="px-3 py-2.5 align-top whitespace-nowrap">
                        <SeverityBadge severity={issue.severity} />
                      </td>
                      <td className="px-3 py-2.5 align-top">
                        <code className="text-xs font-mono" style={{ color: '#3fe56c' }}>{issue.id}</code>
                        <ChecklistIDBadges ids={issue.checklist_ids} />
                      </td>
                      <td className="px-3 py-2.5 align-top text-xs text-on-surface-variant">{issue.category}</td>
                      <td className="px-3 py-2.5 align-top whitespace-nowrap">
                        <PlatformBadge platform={issue.platform} />
                      </td>
                      <td className="px-3 py-2.5 align-top max-w-[260px]">
                        {issue.scope === 'site' ? (
                          <span className="text-xs text-on-surface-variant italic">site-wide</span>
                        ) : (
                          <a
                            href={issue.pageURL}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-on-surface/70 hover:text-primary inline-flex items-center gap-1 truncate max-w-full transition-colors"
                            onClick={(e) => e.stopPropagation()}
                            title={issue.pageURL}
                          >
                            <span className="truncate">{issue.pageURL}</span>
                            <ExternalLink size={10} className="shrink-0" />
                          </a>
                        )}
                      </td>
                      <td className="px-3 py-2.5 align-top text-on-surface/80">{issue.message}</td>
                    </tr>
                    {isOpen && hasDetails && (
                      <tr style={{ background: '#1a202a', borderTop: '1px solid rgba(60,74,60,0.25)' }}>
                        <td />
                        <td colSpan={6} className="px-3 py-3">
                          <pre className="text-xs text-on-surface/70 whitespace-pre-wrap font-mono leading-relaxed">{issue.details}</pre>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                )
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div
          className="px-4 py-3 flex items-center justify-between"
          style={{ borderTop: '1px solid rgba(60,74,60,0.35)' }}
        >
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={safePage <= 1}
            className="btn-ghost disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <ChevronLeft size={14} /> Prev
          </button>
          <span className="text-xs text-on-surface-variant">
            Showing {pageStart + 1}–{Math.min(pageStart + PAGE_SIZE, filtered.length)} of {filtered.length}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={safePage >= totalPages}
            className="btn-ghost disabled:opacity-30 disabled:cursor-not-allowed"
          >
            Next <ChevronRight size={14} />
          </button>
        </div>
      )}
    </div>
  )
}
