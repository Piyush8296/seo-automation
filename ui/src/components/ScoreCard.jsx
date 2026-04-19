function gradeColor(grade) {
  switch (grade) {
    case 'A': return '#3fe56c'
    case 'B': return '#8ed793'
    case 'C': return '#ffb7ae'
    default:  return '#ffb4ab'
  }
}

function RingGauge({ grade, score }) {
  const r = 38
  const circ = 2 * Math.PI * r
  const pct = Math.min((score ?? 0) / 100, 1)
  const color = gradeColor(grade)

  return (
    <svg width="110" height="110" viewBox="0 0 110 110" style={{ overflow: 'visible' }}>
      {/* Track */}
      <circle cx="55" cy="55" r={r} fill="none" stroke="#2f3540" strokeWidth="8" />
      {/* Progress arc */}
      <circle
        cx="55" cy="55" r={r}
        fill="none"
        stroke={color}
        strokeWidth="8"
        strokeLinecap="round"
        strokeDasharray={`${circ * pct} ${circ}`}
        transform="rotate(-90 55 55)"
        style={{ transition: 'stroke-dasharray 0.6s ease' }}
      />
      {/* Grade letter */}
      <text
        x="55" y="50"
        textAnchor="middle"
        fontSize="22"
        fontWeight="700"
        fill={color}
        fontFamily="'Space Grotesk', sans-serif"
      >
        {grade ?? '—'}
      </text>
      {/* Score number */}
      <text
        x="55" y="68"
        textAnchor="middle"
        fontSize="11"
        fill="#bbcbb8"
        fontFamily="'Inter', sans-serif"
      >
        {score != null ? score.toFixed(0) : '—'}
      </text>
    </svg>
  )
}

export default function ScoreCard({ label, score, grade }) {
  return (
    <div className="bg-surface-container rounded-xl p-5 flex flex-col items-center gap-2">
      <RingGauge grade={grade} score={score} />
      <div className="text-xs text-on-surface-variant uppercase tracking-wider font-medium">{label}</div>
    </div>
  )
}
