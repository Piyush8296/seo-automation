import { AlertCircle, AlertTriangle, Info, Globe } from 'lucide-react'

function Stat({ icon: Icon, label, count, color, bg }) {
  return (
    <div className="rounded-xl p-4 flex items-center gap-3" style={{ background: bg }}>
      <div
        className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
        style={{ background: color + '28' }}
      >
        <Icon size={16} style={{ color }} />
      </div>
      <div>
        <div className="text-2xl font-display font-bold text-on-surface">{count.toLocaleString()}</div>
        <div className="text-xs text-on-surface-variant uppercase tracking-wider mt-0.5">{label}</div>
      </div>
    </div>
  )
}

export default function IssueSummary({ errors, warnings, notices, pages }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <Stat icon={AlertCircle}   label="Errors"   count={errors}   color="#ffb4ab" bg="rgba(147,0,10,0.25)"  />
      <Stat icon={AlertTriangle} label="Warnings" count={warnings} color="#ffb7ae" bg="rgba(118,37,31,0.25)" />
      <Stat icon={Info}          label="Notices"  count={notices}  color="#8ed793" bg="rgba(2,83,30,0.25)"   />
      <Stat icon={Globe}         label="Pages"    count={pages}    color="#3fe56c" bg="rgba(0,50,18,0.3)"    />
    </div>
  )
}
