function gradeColor(grade) {
  switch (grade) {
    case 'A': return 'text-emerald-400'
    case 'B': return 'text-green-400'
    case 'C': return 'text-yellow-400'
    case 'D': return 'text-orange-400'
    default:  return 'text-red-400'
  }
}

function gradeRing(grade) {
  switch (grade) {
    case 'A': return 'border-emerald-500'
    case 'B': return 'border-green-500'
    case 'C': return 'border-yellow-500'
    case 'D': return 'border-orange-500'
    default:  return 'border-red-500'
  }
}

export default function ScoreCard({ label, score, grade }) {
  return (
    <div className="card p-5 flex flex-col items-center gap-3">
      <div className={`w-20 h-20 rounded-full border-4 ${gradeRing(grade)} flex flex-col items-center justify-center`}>
        <span className={`text-2xl font-black ${gradeColor(grade)}`}>{grade}</span>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-gray-100">{score?.toFixed(1)}</div>
        <div className="text-xs text-gray-500 uppercase tracking-wider mt-0.5">{label}</div>
      </div>
    </div>
  )
}
