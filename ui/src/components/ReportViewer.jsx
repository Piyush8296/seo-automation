import { useState } from 'react'
import { Download, Maximize2, Minimize2, ExternalLink } from 'lucide-react'
import { api } from '../lib/api'

export default function ReportViewer({ auditId }) {
  const [expanded, setExpanded] = useState(false)
  const reportURL = api.reportURL(auditId)

  const handleDownload = () => {
    const a = document.createElement('a')
    a.href = reportURL
    a.download = `seo-report-${auditId}.html`
    a.click()
  }

  return (
    <div className={`card flex flex-col ${expanded ? 'fixed inset-4 z-50' : 'h-[700px]'}`}>
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800 shrink-0">
        <span className="text-sm font-medium text-gray-300">Interactive Report</span>
        <div className="flex items-center gap-1">
          <a
            href={reportURL}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-ghost"
            title="Open in new tab"
          >
            <ExternalLink size={14} />
            Open
          </a>
          <button onClick={handleDownload} className="btn-ghost" title="Download HTML">
            <Download size={14} />
            Download
          </button>
          <button
            onClick={() => setExpanded((e) => !e)}
            className="btn-ghost"
            title={expanded ? 'Exit fullscreen' : 'Fullscreen'}
          >
            {expanded ? <Minimize2 size={14} /> : <Maximize2 size={14} />}
          </button>
        </div>
      </div>

      {/* iframe */}
      <iframe
        src={reportURL}
        className="flex-1 w-full bg-white rounded-b-xl"
        title="SEO Audit Report"
        sandbox="allow-scripts allow-same-origin"
      />
    </div>
  )
}
