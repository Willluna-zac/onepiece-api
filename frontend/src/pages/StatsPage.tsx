import { useGraphStats } from '../hooks/useApi'
import { Loader, ErrorMsg } from '../components/Loader'

const DANGER_COLORS = ['#10b981', '#84cc16', '#facc15', '#f97316', '#ef4444']
const DANGER_LABELS = ['1 (calmo)', '2', '3', '4', '5 (mortal)']

export default function StatsPage() {
  const { data: stats, isLoading, error } = useGraphStats()

  if (isLoading) return <Loader />
  if (error) return <ErrorMsg message={String(error)} />
  if (!stats) return null

  const histMax = Math.max(...stats.dangerHistogram, 1)
  const componentsHealthy = stats.connectedComponents === 1

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="font-pirate text-gold text-4xl mb-2">📊 Análisis del grafo</h1>
      <p className="text-straw/60 mb-6">
        Métricas estructurales del grafo de rutas marítimas. Útil para entender la salud del seed
        y la diversidad de caminos navegables.
      </p>

      {/* ─── Cards principales ─── */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <StatCard label="Islas" value={stats.totalIslands} icon="🏝️" />
        <StatCard label="Rutas" value={stats.totalRoutes} icon="⚓" />
        <StatCard
          label="Componentes conexos"
          value={stats.connectedComponents}
          icon={componentsHealthy ? '✅' : '⚠️'}
          tone={componentsHealthy ? 'good' : 'warn'}
          hint={componentsHealthy ? 'todo conectado' : `mayor: ${stats.largestComponent} islas`}
        />
        <StatCard label="Bidireccionales" value={`${stats.bidirectionalCount}/${stats.totalRoutes}`} icon="↔️" />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8">
        <StatCard label="Distancia media" value={Math.round(stats.avgDistance)} icon="📏" hint="por tramo" />
        <StatCard label="Tiempo medio" value={`${stats.avgTravelHours.toFixed(1)}h`} icon="⏱️" hint="por tramo" />
        <StatCard label="Danger medio" value={stats.avgDanger.toFixed(2)} icon="🌊" hint="escala 1-5" />
      </div>

      {/* ─── Histograma de Danger ─── */}
      <div className="card mb-6">
        <h2 className="font-pirate text-gold text-2xl mb-1">Distribución de peligro</h2>
        <p className="text-straw/50 text-xs mb-4">
          Cuántas rutas existen para cada nivel de Danger. Una buena distribución hace que los modos
          <span className="text-gold"> safest</span> y <span className="text-gold">riskiest</span> diverjan
          de <span className="text-gold">fastest</span>.
        </p>
        <div className="space-y-2" role="list" aria-label="Histograma de Danger">
          {stats.dangerHistogram.map((count, i) => {
            const pct = (count / histMax) * 100
            return (
              <div key={i} className="flex items-center gap-3" role="listitem">
                <span className="text-straw/70 text-xs w-20 shrink-0">Danger {DANGER_LABELS[i]}</span>
                <div className="flex-1 bg-navy-light rounded h-6 overflow-hidden">
                  <div
                    className="h-full transition-all"
                    style={{ width: `${pct}%`, backgroundColor: DANGER_COLORS[i] }}
                  />
                </div>
                <span className="text-gold font-mono text-sm w-12 text-right">{count}</span>
              </div>
            )
          })}
        </div>
      </div>

      <div className="card">
        <h2 className="font-pirate text-gold text-2xl mb-2">Otros datos</h2>
        <ul className="text-straw/80 text-sm space-y-1">
          <li>• <span className="text-gold">{stats.islandsWithLogPose}</span> islas requieren Log Pose ({Math.round(stats.islandsWithLogPose / stats.totalIslands * 100)}%).</li>
          <li>• <span className="text-gold">{stats.bidirectionalCount}</span> rutas son bidireccionales — el resto solo se navega en un sentido.</li>
          <li>
            • Componentes conectados: <span className="text-gold">{stats.connectedComponents}</span>
            {!componentsHealthy && <> — el componente mayor cubre <span className="text-gold">{stats.largestComponent}</span> islas; el resto queda aislado.</>}
          </li>
        </ul>
      </div>
    </div>
  )
}

function StatCard({
  label,
  value,
  icon,
  hint,
  tone,
}: {
  label: string
  value: string | number
  icon: string
  hint?: string
  tone?: 'good' | 'warn'
}) {
  const border = tone === 'warn' ? 'border-yellow-500/40' : tone === 'good' ? 'border-emerald-500/40' : 'border-gold/20'
  return (
    <div className={`card !p-4 border ${border}`}>
      <div className="text-straw/60 text-xs mb-1 flex items-center gap-1">
        <span>{icon}</span> {label}
      </div>
      <div className="text-gold font-pirate text-3xl leading-tight">{value}</div>
      {hint && <div className="text-straw/40 text-xs mt-1">{hint}</div>}
    </div>
  )
}
