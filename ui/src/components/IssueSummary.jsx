import { AlertCircle, AlertTriangle, Info } from 'lucide-react'

function Stat({ icon: Icon, label, count, color }) {
  return (
    <div className={`flex items-center gap-3 p-4 rounded-lg bg-gray-800/60 border border-gray-700/50`}>
      <div className={`p-2 rounded-lg ${color} bg-opacity-10`}>
        <Icon size={18} className={color.replace('bg-', 'text-').replace('/10', '')} />
      </div>
      <div>
        <div className="text-2xl font-bold text-gray-100">{count.toLocaleString()}</div>
        <div className="text-xs text-gray-500 uppercase tracking-wider">{label}</div>
      </div>
    </div>
  )
}

export default function IssueSummary({ errors, warnings, notices, pages }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <Stat icon={AlertCircle}   label="Errors"   count={errors}   color="bg-red-500" />
      <Stat icon={AlertTriangle} label="Warnings" count={warnings} color="bg-amber-500" />
      <Stat icon={Info}          label="Notices"  count={notices}  color="bg-blue-500" />
      <div className="flex items-center gap-3 p-4 rounded-lg bg-gray-800/60 border border-gray-700/50">
        <div className="p-2 rounded-lg bg-emerald-500 bg-opacity-10">
          <svg className="text-emerald-400" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </div>
        <div>
          <div className="text-2xl font-bold text-gray-100">{pages.toLocaleString()}</div>
          <div className="text-xs text-gray-500 uppercase tracking-wider">Pages</div>
        </div>
      </div>
    </div>
  )
}
