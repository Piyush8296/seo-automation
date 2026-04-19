import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Play, Trash2, ExternalLink, ChevronDown } from 'lucide-react'
import StatusBadge from './StatusBadge'
import { api } from '../lib/api'

function accentColor(status) {
  switch (status) {
    case 'complete':  return '#3fe56c'
    case 'running':   return '#8ed793'
    case 'failed':    return '#ffb4ab'
    default:          return '#3c4a3c'
  }
}

function gradeColor(g) {
  const m = { A: '#3fe56c', B: '#8ed793', C: '#ffb7ae', D: '#ffb4ab' }
  return m[g] ?? '#ffb4ab'
}

function fmt(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

export default function AuditHistory({ audits, onDelete, onRerun, loading }) {
  const navigate = useNavigate()
  const [deleting, setDeleting] = useState(null)
  const [diffA, setDiffA] = useState(null)
  const [diffResult, setDiffResult] = useState(null)

  const handleDelete = async (id) => {
    setDeleting(id)
    try { await onDelete(id) } finally { setDeleting(null) }
  }

  const handleDiff = async (id) => {
    if (!diffA) { setDiffA(id); return }
    if (diffA === id) { setDiffA(null); return }
    try {
      const r = await api.diffAudits(diffA, id)
      setDiffResult(r)
      setDiffA(null)
    } catch {}
  }

  return (
    <div className="card flex flex-col overflow-hidden">
      {/* Header row */}
      <div className="flex items-center justify-between px-5 py-4 bg-surface-container">
        <h2 className="font-display font-bold text-on-surface">Audit History</h2>
        <span className="text-xs text-on-surface-variant">{audits.length} runs</span>
      </div>

      {diffA && (
        <div
          className="mx-5 mt-4 px-4 py-3 rounded-xl text-sm"
          style={{ background: 'rgba(142,215,147,0.08)', color: '#8ed793' }}
        >
          Select a second audit to compare…
          <button
            onClick={() => setDiffA(null)}
            className="ml-3 underline text-xs opacity-70 hover:opacity-100"
          >
            cancel
          </button>
        </div>
      )}

      {diffResult && (
        <div className="mx-5 mt-4 bg-surface-container rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-semibold text-on-surface">Diff Result</span>
            <button onClick={() => setDiffResult(null)} className="text-xs text-on-surface-variant hover:text-on-surface">close</button>
          </div>
          <div className="grid grid-cols-4 gap-3 text-center text-sm">
            {[
              { label: 'Score',    val: diffResult.score_delta,   positive: diffResult.score_delta > 0 },
              { label: 'Errors',   val: diffResult.error_delta,   positive: diffResult.error_delta < 0 },
              { label: 'Warnings', val: diffResult.warn_delta,    positive: diffResult.warn_delta < 0 },
              { label: 'Pages',    val: diffResult.page_delta,    positive: null },
            ].map(({ label, val, positive }) => (
              <div key={label} className="bg-surface-container-high rounded-lg p-3">
                <div
                  className="text-lg font-bold font-display"
                  style={{ color: positive === true ? '#3fe56c' : positive === false ? '#ffb4ab' : '#dde2f1' }}
                >
                  {val > 0 ? '+' : ''}{typeof val === 'number' ? val.toFixed(label === 'Score' ? 1 : 0) : val}
                </div>
                <div className="text-xs text-on-surface-variant">{label}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <span
            className="w-6 h-6 rounded-full animate-spin"
            style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }}
          />
        </div>
      ) : audits.length === 0 ? (
        <div className="py-16 text-center text-on-surface-variant text-sm">
          No audits yet. Start one above.
        </div>
      ) : (
        <div className="flex flex-col">
          {audits.map((a) => (
            <div
              key={a.id}
              className={`group relative px-5 py-4 hover:bg-surface-bright transition-colors cursor-pointer ${
                diffA && diffA !== a.id ? 'opacity-60' : ''
              }`}
              onClick={() => navigate(`/audit/${a.id}`)}
            >
              {/* Left accent bar */}
              <div
                className="absolute left-0 top-3 bottom-3 w-[3px] rounded-r-full"
                style={{ background: accentColor(a.status) }}
              />

              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 flex-wrap">
                    <StatusBadge status={a.status} />
                    {a.grade && (
                      <span
                        className="text-sm font-black font-display"
                        style={{ color: gradeColor(a.grade) }}
                      >
                        {a.grade}
                      </span>
                    )}
                    {a.health_score > 0 && (
                      <span className="text-xs text-on-surface-variant">{a.health_score.toFixed(1)}</span>
                    )}
                  </div>
                  <div className="mt-1.5 text-sm font-medium text-on-surface truncate" title={a.url}>
                    {a.url}
                  </div>
                  <div className="mt-1 flex items-center gap-3 text-xs text-on-surface-variant/60">
                    <span>{fmt(a.created_at)}</span>
                    {a.page_count > 0 && <span>{a.page_count} pages</span>}
                    {a.error_count > 0 && <span style={{ color: '#ffb4ab' }}>{a.error_count} errors</span>}
                    {a.warn_count > 0 && <span style={{ color: '#ffb7ae' }}>{a.warn_count} warnings</span>}
                  </div>
                </div>

                {/* Actions */}
                <div
                  className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <button onClick={() => onRerun(a)} className="btn-ghost text-xs" title="Re-run">
                    <Play size={13} />
                  </button>
                  <button
                    onClick={() => handleDiff(a.id)}
                    className={`btn-ghost text-xs ${diffA === a.id ? 'text-secondary' : ''}`}
                    title="Compare (select two audits)"
                    style={diffA === a.id ? { background: 'rgba(142,215,147,0.1)' } : {}}
                  >
                    <ChevronDown size={13} />
                    Diff
                  </button>
                  <a
                    href={api.reportURL(a.id)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn-ghost text-xs"
                    title="Open report"
                  >
                    <ExternalLink size={13} />
                  </a>
                  <button
                    onClick={() => handleDelete(a.id)}
                    className="btn-ghost text-xs"
                    disabled={deleting === a.id}
                    title="Delete"
                    style={{ color: '#ffb4ab' }}
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
