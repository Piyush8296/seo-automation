import { useEffect, useState } from 'react'
import { XCircle } from 'lucide-react'

function useElapsed(running) {
  const [secs, setSecs] = useState(0)
  useEffect(() => {
    if (!running) return
    setSecs(0)
    const t = setInterval(() => setSecs((s) => s + 1), 1000)
    return () => clearInterval(t)
  }, [running])
  const m = String(Math.floor(secs / 60)).padStart(2, '0')
  const s = String(secs % 60).padStart(2, '0')
  return `${m}:${s}`
}

export default function CrawlProgress({ crawled, currentURL, maxPages, onCancel }) {
  const elapsed = useElapsed(true)
  const progress = maxPages > 0 ? Math.min((crawled / maxPages) * 100, 100) : null

  return (
    <div className="card p-8 flex flex-col items-center gap-8">
      {/* Radar animation */}
      <div className="relative w-40 h-40 flex items-center justify-center">
        {/* Outer rings */}
        <div className="absolute inset-0 rounded-full border border-emerald-500/10" />
        <div className="absolute inset-4 rounded-full border border-emerald-500/15" />
        <div className="absolute inset-8 rounded-full border border-emerald-500/20" />
        {/* Sweep */}
        <div className="absolute inset-0 rounded-full overflow-hidden">
          <div className="radar-sweep w-1/2 h-full origin-right"
               style={{ background: 'conic-gradient(from 0deg, transparent 70%, rgba(16,185,129,0.35) 100%)' }} />
        </div>
        {/* Ping rings */}
        <div className="absolute inset-8 rounded-full bg-emerald-500/10 animate-ping" />
        {/* Center dot */}
        <div className="w-3 h-3 rounded-full bg-emerald-500 shadow-[0_0_12px_rgba(16,185,129,0.8)]" />
      </div>

      {/* Counter */}
      <div className="text-center">
        <div className="text-5xl font-black text-gray-100 tabular-nums">{crawled.toLocaleString()}</div>
        <div className="text-sm text-gray-500 mt-1 uppercase tracking-widest">pages crawled</div>
        <div className="text-xs text-gray-600 mt-2 font-mono">{elapsed}</div>
      </div>

      {/* Progress bar — only when max-pages is set */}
      {progress !== null && (
        <div className="w-full max-w-sm">
          <div className="flex justify-between text-xs text-gray-500 mb-1.5">
            <span>{crawled} / {maxPages} pages</span>
            <span>{progress.toFixed(0)}%</span>
          </div>
          <div className="h-1.5 bg-gray-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-emerald-500 rounded-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Current URL ticker */}
      {currentURL && (
        <div className="w-full max-w-md overflow-hidden bg-gray-800/50 rounded-lg px-4 py-2.5 border border-gray-700/50">
          <div className="text-xs text-gray-500 mb-1">Currently crawling</div>
          <div className="overflow-hidden">
            <div className="ticker text-xs font-mono text-emerald-400">{currentURL}</div>
          </div>
        </div>
      )}

      {/* Cancel */}
      <button onClick={onCancel} className="btn-ghost text-red-400 hover:text-red-300 hover:bg-red-500/10">
        <XCircle size={15} />
        Cancel crawl
      </button>
    </div>
  )
}
