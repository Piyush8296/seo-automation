import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Play, Trash2, ExternalLink, ChevronDown } from 'lucide-react'
import StatusBadge from './StatusBadge'
import { api } from '../lib/api'

function gradeColor(g) {
  const m = { A: 'text-emerald-400', B: 'text-green-400', C: 'text-yellow-400', D: 'text-orange-400' }
  return m[g] ?? 'text-red-400'
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
    <div className="card flex flex-col">
      <div className="flex items-center justify-between px-5 py-4 border-b border-gray-800">
        <h2 className="text-base font-bold text-gray-100">Audit History</h2>
        <span className="text-xs text-gray-600">{audits.length} runs</span>
      </div>

      {diffA && (
        <div className="mx-5 mt-4 px-4 py-3 bg-blue-500/10 border border-blue-500/20 rounded-lg text-sm text-blue-300">
          Select a second audit to compare with…
          <button onClick={() => setDiffA(null)} className="ml-3 text-blue-400 underline text-xs">cancel</button>
        </div>
      )}

      {diffResult && (
        <div className="mx-5 mt-4 card p-4 border-gray-700">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-semibold text-gray-200">Diff Result</span>
            <button onClick={() => setDiffResult(null)} className="text-xs text-gray-500 hover:text-gray-300">close</button>
          </div>
          <div className="grid grid-cols-4 gap-3 text-center text-sm">
            {[
              { label: 'Score', val: diffResult.score_delta, positive: diffResult.score_delta > 0 },
              { label: 'Errors', val: diffResult.error_delta, positive: diffResult.error_delta < 0 },
              { label: 'Warnings', val: diffResult.warn_delta, positive: diffResult.warn_delta < 0 },
              { label: 'Pages', val: diffResult.page_delta, positive: null },
            ].map(({ label, val, positive }) => (
              <div key={label} className="bg-gray-800/50 rounded-lg p-3">
                <div className={`text-lg font-bold ${positive === true ? 'text-emerald-400' : positive === false ? 'text-red-400' : 'text-gray-300'}`}>
                  {val > 0 ? '+' : ''}{typeof val === 'number' ? val.toFixed(label === 'Score' ? 1 : 0) : val}
                </div>
                <div className="text-xs text-gray-500">{label}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <span className="w-6 h-6 border-2 border-gray-700 border-t-emerald-500 rounded-full animate-spin" />
        </div>
      ) : audits.length === 0 ? (
        <div className="py-16 text-center text-gray-600 text-sm">No audits yet. Start one above.</div>
      ) : (
        <div className="divide-y divide-gray-800/70">
          {audits.map((a) => (
            <div
              key={a.id}
              className={`group px-5 py-4 hover:bg-gray-800/30 transition-colors cursor-pointer ${diffA && diffA !== a.id ? 'opacity-70' : ''}`}
              onClick={() => navigate(`/audit/${a.id}`)}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 flex-wrap">
                    <StatusBadge status={a.status} />
                    {a.grade && (
                      <span className={`text-sm font-black ${gradeColor(a.grade)}`}>{a.grade}</span>
                    )}
                    {a.health_score > 0 && (
                      <span className="text-xs text-gray-500">{a.health_score.toFixed(1)}</span>
                    )}
                  </div>
                  <div className="mt-1.5 text-sm font-medium text-gray-200 truncate" title={a.url}>
                    {a.url}
                  </div>
                  <div className="mt-1 flex items-center gap-3 text-xs text-gray-600">
                    <span>{fmt(a.created_at)}</span>
                    {a.page_count > 0 && <span>{a.page_count} pages</span>}
                    {a.error_count > 0 && <span className="text-red-500">{a.error_count} errors</span>}
                    {a.warn_count > 0 && <span className="text-amber-500">{a.warn_count} warnings</span>}
                  </div>
                </div>

                {/* Actions */}
                <div
                  className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <button
                    onClick={() => onRerun(a)}
                    className="btn-ghost text-xs"
                    title="Re-run with same config"
                  >
                    <Play size={13} />
                  </button>
                  <button
                    onClick={() => handleDiff(a.id)}
                    className={`btn-ghost text-xs ${diffA === a.id ? 'bg-blue-500/20 text-blue-400' : ''}`}
                    title="Compare (select two audits)"
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
                    className="btn-ghost text-red-400 hover:text-red-300 hover:bg-red-500/10 text-xs"
                    disabled={deleting === a.id}
                    title="Delete"
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
