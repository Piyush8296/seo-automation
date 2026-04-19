import { Fragment, useEffect, useMemo, useState } from 'react'
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

function FilterSelect({ value, onChange, options, placeholder }) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="input text-sm min-w-[140px]"
    >
      <option value="">{placeholder}</option>
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>{opt.label}</option>
      ))}
    </select>
  )
}

export default function IssueTable({ auditId }) {
  const [report, setReport] = useState(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)

  const [severity, setSeverity] = useState('')
  const [category, setCategory] = useState('')
  const [checkID, setCheckID] = useState('')
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
    allIssues.forEach((i) => i.id && s.add(i.id))
    return [...s].sort()
  }, [allIssues])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return allIssues
      .filter((i) => !severity || i.severity === severity)
      .filter((i) => !category || i.category === category)
      .filter((i) => !checkID || i.id === checkID)
      .filter((i) => !q || (i.pageURL || '').toLowerCase().includes(q) || (i.message || '').toLowerCase().includes(q))
      .sort((a, b) => severityRank(a.severity) - severityRank(b.severity))
  }, [allIssues, severity, category, checkID, search])

  useEffect(() => { setPage(1); setExpanded(new Set()) }, [severity, category, checkID, search])

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

  const clearFilters = () => { setSeverity(''); setCategory(''); setCheckID(''); setSearch('') }
  const hasFilters = severity || category || checkID || search

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
      {/* Filters */}
      <div
        className="p-4 flex flex-wrap items-center gap-2"
        style={{ borderBottom: '1px solid rgba(60,74,60,0.35)' }}
      >
        <div className="relative flex-1 min-w-[220px]">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search URL or message…"
            className="input text-sm pl-9"
          />
        </div>
        <FilterSelect
          value={severity}
          onChange={setSeverity}
          placeholder="All severities"
          options={[
            { value: 'error',   label: 'Errors' },
            { value: 'warning', label: 'Warnings' },
            { value: 'notice',  label: 'Notices' },
          ]}
        />
        <FilterSelect
          value={category}
          onChange={setCategory}
          placeholder="All categories"
          options={categories.map((c) => ({ value: c, label: c }))}
        />
        <FilterSelect
          value={checkID}
          onChange={setCheckID}
          placeholder="All checks"
          options={checkIDs.map((c) => ({ value: c, label: c }))}
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
                <th className="text-left font-medium px-3 py-2.5">Check</th>
                <th className="text-left font-medium px-3 py-2.5">Category</th>
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
                      </td>
                      <td className="px-3 py-2.5 align-top text-xs text-on-surface-variant">{issue.category}</td>
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
                        <td colSpan={5} className="px-3 py-3">
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
