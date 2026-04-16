const styles = {
  running:   'bg-blue-500/15 text-blue-400 border-blue-500/30',
  complete:  'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  failed:    'bg-red-500/15 text-red-400 border-red-500/30',
  cancelled: 'bg-gray-500/15 text-gray-400 border-gray-500/30',
}

const dots = {
  running: <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse mr-1.5 inline-block" />,
}

export default function StatusBadge({ status }) {
  return (
    <span className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${styles[status] ?? styles.cancelled}`}>
      {dots[status]}
      {status}
    </span>
  )
}
