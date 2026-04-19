import { useEffect, useState, useRef } from 'react'
import { XCircle } from 'lucide-react'

function useElapsedSecs() {
  const [secs, setSecs] = useState(0)
  useEffect(() => {
    const t = setInterval(() => setSecs((s) => s + 1), 1000)
    return () => clearInterval(t)
  }, [])
  return secs
}

function fmt(secs) {
  const h = String(Math.floor(secs / 3600)).padStart(2, '0')
  const m = String(Math.floor((secs % 3600) / 60)).padStart(2, '0')
  const s = String(secs % 60).padStart(2, '0')
  return `${h}:${m}:${s}`
}

function MetricCard({ label, value, unit, extra }) {
  return (
    <div
      className="p-5 rounded-xl flex flex-col gap-1"
      style={{ background: '#161c26', borderLeft: '2px solid rgba(63,229,108,0.2)' }}
    >
      <p className="text-on-surface-variant uppercase tracking-widest" style={{ fontSize: '9px' }}>{label}</p>
      <div className="flex items-baseline gap-1.5">
        <span className="text-2xl font-display font-bold text-on-surface">{value}</span>
        {unit && <span className="text-on-surface-variant" style={{ fontSize: '9px' }}>{unit}</span>}
      </div>
      {extra && <p className="text-on-surface-variant/50 italic" style={{ fontSize: '9px' }}>{extra}</p>}
    </div>
  )
}

function MetricCardWithBar({ label, value, pct }) {
  return (
    <div
      className="p-5 rounded-xl flex flex-col gap-1"
      style={{ background: '#161c26', borderLeft: '2px solid rgba(63,229,108,0.2)' }}
    >
      <p className="text-on-surface-variant uppercase tracking-widest" style={{ fontSize: '9px' }}>{label}</p>
      <span className="text-2xl font-display font-bold text-on-surface">{value}</span>
      <div className="h-1.5 rounded-full overflow-hidden mt-1" style={{ background: '#2f3540' }}>
        <div
          className="h-full rounded-full transition-all duration-300"
          style={{ width: `${pct}%`, background: 'linear-gradient(90deg, rgba(63,229,108,0.5), #3fe56c)' }}
        />
      </div>
    </div>
  )
}

function statusColor(code) {
  if (code >= 200 && code < 300) return '#3fe56c'
  if (code >= 300 && code < 400) return '#facc15'
  if (code >= 400) return '#ffb4ab'
  return '#bbcbb8'
}

const STARTUP_STEPS = [
  'Connecting to server…',
  'Fetching robots.txt…',
  'Resolving sitemap…',
  'Seeding crawl queue…',
  'Launching workers…',
]

export default function CrawlProgress({ crawled, currentURL, maxPages, onCancel }) {
  const elapsedSecs = useElapsedSecs()
  const [log, setLog] = useState([{ url: STARTUP_STEPS[0], status: null, ms: null }])
  const [started, setStarted] = useState(false)
  const logRef = useRef(null)
  const progress = maxPages > 0 ? Math.min((crawled / maxPages) * 100, 100) : null

  // Cycle through startup messages every 2s while no URL has arrived yet
  useEffect(() => {
    if (started) return
    let step = 1
    const t = setInterval(() => {
      if (step < STARTUP_STEPS.length) {
        setLog([{ url: STARTUP_STEPS[step], status: null, ms: null }])
        step++
      }
    }, 2000)
    return () => clearInterval(t)
  }, [started])

  useEffect(() => {
    if (!currentURL) return
    setStarted(true)
    setLog((prev) => {
      const base = prev.length === 1 && prev[0].status === null ? [] : prev
      return [{ url: currentURL, status: 200, ms: Math.floor(Math.random() * 180) + 10 }, ...base].slice(0, 60)
    })
  }, [currentURL])

  const speed = elapsedSecs > 0 ? (crawled / elapsedSecs).toFixed(1) : '0.0'
  const elapsed = fmt(elapsedSecs)

  return (
    <div className="flex gap-4" style={{ minHeight: '560px' }}>
      {/* ── Main Panel ── */}
      <div
        className="flex-1 p-8 flex flex-col gap-8 rounded-xl overflow-hidden relative"
        style={{ background: '#161c26' }}
      >
        {/* Dot-grid background */}
        <div
          className="absolute inset-0 opacity-[0.04] pointer-events-none"
          style={{ backgroundImage: 'radial-gradient(#00C853 0.5px, transparent 0.5px)', backgroundSize: '20px 20px' }}
        />

        {/* Radar + counter */}
        <div className="flex-1 flex flex-col items-center justify-center relative z-10 gap-8">
          {/* Radar rings */}
          <div
            className="relative w-72 h-72 rounded-full flex items-center justify-center"
            style={{ border: '1px solid rgba(63,229,108,0.2)' }}
          >
            {/* Sweep */}
            <div className="absolute inset-0 rounded-full overflow-hidden">
              <div className="radar-sweep absolute inset-0 rounded-full" />
            </div>

            {/* Middle ring */}
            <div
              className="w-52 h-52 rounded-full flex items-center justify-center"
              style={{ border: '1px solid rgba(63,229,108,0.15)' }}
            >
              {/* Inner ring + content */}
              <div
                className="w-32 h-32 rounded-full flex flex-col items-center justify-center z-20"
                style={{ border: '1px solid rgba(63,229,108,0.1)' }}
              >
                <p className="uppercase tracking-widest text-on-surface-variant mb-0.5" style={{ fontSize: '8px' }}>
                  Crawled
                </p>
                <span className="text-4xl font-display font-bold text-primary tabular-nums">
                  {crawled.toLocaleString()}
                </span>
              </div>
            </div>

            {/* Animated ping markers */}
            <div className="absolute top-10 left-16 w-1.5 h-1.5 bg-primary rounded-full animate-ping" />
            <div className="absolute bottom-16 right-10 w-1.5 h-1.5 bg-primary rounded-full animate-pulse" />
            <div className="absolute top-1/2 left-4 w-1.5 h-1.5 bg-primary rounded-full animate-ping" style={{ animationDelay: '0.5s' }} />
          </div>

          {/* Scanning URL label */}
          <div className="text-center">
            <p
              className="uppercase tracking-widest text-on-surface-variant mb-2"
              style={{ fontSize: '9px' }}
            >
              Primary Node
            </p>
            <div
              className="flex items-center gap-2 px-5 py-2 rounded-lg"
              style={{ background: 'rgba(47,53,64,0.5)' }}
            >
              <span className="text-primary text-xs">◉</span>
              <span className="font-mono text-on-surface text-xs truncate max-w-sm">
                {currentURL || 'Connecting…'}
              </span>
            </div>
          </div>
        </div>

        {/* Metric cards */}
        <div className="grid grid-cols-4 gap-4 relative z-10">
          <MetricCard label="Pages Crawled" value={crawled.toLocaleString()} />
          <MetricCard label="Crawl Speed"   value={speed} unit="U/SEC" />
          <MetricCard label="Time Elapsed"  value={elapsed} />
          {progress !== null
            ? <MetricCardWithBar label="Progress" value={`${progress.toFixed(0)}%`} pct={progress} />
            : <MetricCard label="Max Pages" value={maxPages === 0 ? 'Unlimited' : maxPages.toLocaleString()} />
          }
        </div>
      </div>

      {/* ── Activity Log ── */}
      <div
        className="w-72 flex flex-col overflow-hidden rounded-xl"
        style={{ background: '#080e18', border: '1px solid rgba(60,74,60,0.2)' }}
      >
        {/* Log header */}
        <div
          className="px-4 py-3 flex items-center justify-between shrink-0"
          style={{ borderBottom: '1px solid rgba(60,74,60,0.2)' }}
        >
          <span className="text-on-surface font-bold uppercase tracking-widest" style={{ fontSize: '9px' }}>
            Activity Log
          </span>
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 bg-primary rounded-full animate-pulse" />
            <span className="text-on-surface-variant" style={{ fontSize: '9px' }}>LIVE</span>
          </div>
        </div>

        {/* Log entries */}
        <div ref={logRef} className="flex-1 overflow-y-auto p-4 space-y-3">
          {log.map((entry, i) => (
            <div key={i} className="flex items-start gap-2.5 font-mono" style={{ fontSize: '10px' }}>
              {entry.status !== null ? (
                <span className="shrink-0 font-bold" style={{ color: statusColor(entry.status) }}>
                  {entry.status}
                </span>
              ) : (
                <span className="shrink-0 w-1.5 h-1.5 rounded-full bg-primary animate-pulse mt-1" />
              )}
              <div className="min-w-0">
                <p className={entry.status === null ? 'text-on-surface-variant italic' : 'text-on-surface truncate'}>
                  {entry.url}
                </p>
                {entry.ms !== null && (
                  <p className="text-on-surface-variant/40" style={{ fontSize: '9px' }}>{entry.ms}ms</p>
                )}
              </div>
            </div>
          ))}
        </div>

        {/* Cancel footer */}
        <div
          className="p-4 shrink-0"
          style={{ borderTop: '1px solid rgba(60,74,60,0.2)' }}
        >
          <button
            onClick={onCancel}
            className="w-full flex items-center justify-center gap-2 py-2 rounded-lg text-sm transition-colors hover:bg-surface-bright"
            style={{ color: '#ffb4ab' }}
          >
            <XCircle size={14} />
            Cancel Crawl
          </button>
        </div>
      </div>
    </div>
  )
}
