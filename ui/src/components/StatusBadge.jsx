const config = {
  running:   { color: '#8ed793', bg: 'rgba(142,215,147,0.1)',  dot: true  },
  complete:  { color: '#3fe56c', bg: 'rgba(63,229,108,0.1)',   dot: false },
  failed:    { color: '#ffb4ab', bg: 'rgba(147,0,10,0.25)',    dot: false },
  cancelled: { color: '#bbcbb8', bg: 'rgba(60,74,60,0.3)',     dot: false },
}

export default function StatusBadge({ status }) {
  const c = config[status] ?? config.cancelled
  return (
    <span
      className="inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full"
      style={{ color: c.color, background: c.bg }}
    >
      {c.dot && (
        <span
          className="w-1.5 h-1.5 rounded-full mr-1.5 inline-block animate-pulse"
          style={{ background: c.color }}
        />
      )}
      {status}
    </span>
  )
}
