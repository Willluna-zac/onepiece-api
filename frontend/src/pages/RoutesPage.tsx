import { useEffect, useMemo, useState } from 'react'
import { useIslands, useShortestPath, useReachableIslands, useRoutes, useCompareModes } from '../hooks/useApi'
import { useLastRouteSearch } from '../hooks/useLastRouteSearch'
import type { RouteMode, Route } from '../api/client'
import { RouteMap } from '../components/RouteMap'
import { Loader } from '../components/Loader'
import { formatHours } from '../lib/format'

const REGION_COLORS: Record<string, string> = {
  'East Blue':  'bg-blue-900/40 text-blue-300',
  'North Blue': 'bg-cyan-900/40 text-cyan-300',
  'South Blue': 'bg-teal-900/40 text-teal-300',
  'Grand Line': 'bg-yellow-900/40 text-yellow-300',
  'New World':  'bg-red-900/40 text-red-300',
  'Sky Islands':'bg-purple-900/40 text-purple-300',
  'Red Line':   'bg-rose-900/40 text-rose-300',
}

// ── Configuración de modos de búsqueda ──────────────────────────────────────
type MetricKey = 'distance' | 'time' | 'worstDanger' | 'bestDanger'

type ModeConfig = {
  id: RouteMode
  label: string
  icon: string
  description: string
  /** Cuál de las 4 métricas globales se resalta como "primaria". */
  primaryMetric: MetricKey
}

const MODES: ModeConfig[] = [
  {
    id: 'fastest',
    label: 'Rápido',
    icon: '⚡',
    description: 'Minimiza la distancia total',
    primaryMetric: 'distance',
  },
  {
    id: 'quickest',
    label: 'Menos tiempo',
    icon: '⏱️',
    description: 'Minimiza el tiempo total (incluye espera de Log Pose)',
    primaryMetric: 'time',
  },
  {
    id: 'safest',
    label: 'Más segura',
    icon: '🛡️',
    description: 'Evita los tramos más peligrosos',
    primaryMetric: 'worstDanger',
  },
  {
    id: 'riskiest',
    label: 'Más peligrosa',
    icon: '☠️',
    description: 'Maximiza el peligro mínimo del viaje',
    primaryMetric: 'bestDanger',
  },
]

const METRIC_LABELS: Record<MetricKey, string> = {
  distance: 'Distancia total',
  time: 'Tiempo total',
  worstDanger: 'Peor tramo (peligro)',
  bestDanger: 'Mejor tramo (peligro)',
}

/** Color tailwind según nivel de peligro 1–5 (mismo índice que en RouteMap). */
const DANGER_COLORS: Record<number, string> = {
  1: 'text-blue-400 border-blue-400',
  2: 'text-green-400 border-green-400',
  3: 'text-yellow-400 border-yellow-400',
  4: 'text-orange-400 border-orange-400',
  5: 'text-red-400 border-red-400',
}

export default function RoutesPage() {
  const { data: islands, isLoading: islandsLoading } = useIslands()
  const { data: routes } = useRoutes()

  // ── Sección A: Ruta más corta ─────────────────────────────────────────────
  const { last, save: saveLastSearch } = useLastRouteSearch()
  // Restaurar (NO auto-buscar): rellenamos from/to/mode pero search queda en false
  const [from, setFrom] = useState(() => last?.from ?? '')
  const [to, setTo]     = useState(() => last?.to   ?? '')
  const [mode, setMode] = useState<RouteMode>(() => last?.mode ?? 'fastest')
  const [search, setSearch] = useState(false)
  const [showCompare, setShowCompare] = useState(false)
  const [animate, setAnimate]         = useState(false)

  const { data: pathResult, isFetching: pathFetching } = useShortestPath(from, to, mode, search)
  const compareResults = useCompareModes(from, to, search && showCompare)

  const activeMode = MODES.find(m => m.id === mode) ?? MODES[0]

  /** Peligro derivado de cada isla del path = max(Danger) de las aristas del camino que la tocan. */
  const dangerByIsland = useMemo(() => {
    const m = new Map<string, number>()
    if (!pathResult?.found || !routes || pathResult.path.length < 2) return m
    const ids = pathResult.path.map(s => s.islandId)
    const byEdge = new Map<string, Route>()
    for (const r of routes) {
      byEdge.set(`${r.fromIsland}->${r.toIsland}`, r)
      if (r.bidirectional) byEdge.set(`${r.toIsland}->${r.fromIsland}`, r)
    }
    for (let i = 0; i < ids.length - 1; i++) {
      const r = byEdge.get(`${ids[i]}->${ids[i + 1]}`)
      if (!r) continue
      m.set(ids[i], Math.max(m.get(ids[i]) ?? 0, r.danger))
      m.set(ids[i + 1], Math.max(m.get(ids[i + 1]) ?? 0, r.danger))
    }
    return m
  }, [pathResult, routes])

  function handleCalculate() {
    setSearch(true)
    if (from && to) saveLastSearch({ from, to, mode })
  }
  function handleFromChange(v: string) { setFrom(v); setSearch(false) }
  function handleToChange(v: string)   { setTo(v);   setSearch(false) }
  function handleModeChange(m: RouteMode) {
    setMode(m)
    setSearch(false) // forzar al usuario a recalcular con el nuevo modo
  }

  /** Navegación con flechas izq/der dentro del toggle de modos. */
  function handleModeKeyDown(e: React.KeyboardEvent<HTMLButtonElement>) {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
    e.preventDefault()
    const idx = MODES.findIndex(m => m.id === mode)
    const dir = e.key === 'ArrowRight' ? 1 : -1
    const next = MODES[(idx + dir + MODES.length) % MODES.length]
    handleModeChange(next.id)
    // Mover focus al nuevo botón
    requestAnimationFrame(() => {
      const btn = document.getElementById(`mode-btn-${next.id}`)
      btn?.focus()
    })
  }

  // ── Sección B: Islas alcanzables ──────────────────────────────────────────
  const [reachFrom, setReachFrom]   = useState('')
  const [maxCost, setMaxCost]       = useState(1500)
  const [debouncedMaxCost, setDebouncedMaxCost] = useState(maxCost)

  // Debounce del slider: evita refetch en cada pixel que se mueve
  useEffect(() => {
    const t = setTimeout(() => setDebouncedMaxCost(maxCost), 400)
    return () => clearTimeout(t)
  }, [maxCost])

  const { data: reachResult, isFetching: reachFetching } = useReachableIslands(
    reachFrom, debouncedMaxCost, reachFrom !== ''
  )

  if (islandsLoading) return <Loader />

  const islandOptions = islands ?? []

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 space-y-12">
      <div>
        <h1 className="font-pirate text-gold text-4xl mb-1">⚓ Rutas Marítimas</h1>
        <p className="text-straw/60">
          Navega el mundo de One Piece · {routes?.length ?? 0} rutas · {islands?.length ?? 0} islas
        </p>
      </div>

      {/* ── Sección A: Ruta más corta ─────────────────────────────────────── */}
      <section className="card space-y-6">
        <h2 className="font-pirate text-gold text-2xl">🗺️ Ruta más corta — Dijkstra</h2>

        {/* Toggle de modo de búsqueda */}
        <div>
          <label className="text-straw/60 text-xs uppercase tracking-wider mb-2 block">Modo de búsqueda</label>
          <div role="group" aria-label="Modo de búsqueda" className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            {MODES.map(m => {
              const active = m.id === mode
              return (
                <button
                  id={`mode-btn-${m.id}`}
                  key={m.id}
                  type="button"
                  onClick={() => handleModeChange(m.id)}
                  onKeyDown={handleModeKeyDown}
                  aria-pressed={active}
                  tabIndex={active ? 0 : -1}
                  title={m.description}
                  className={`px-3 py-2 rounded-lg border text-sm font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-gold/60 ${
                    active
                      ? 'bg-gold/20 border-gold text-gold'
                      : 'bg-navy border-gold/20 text-straw/70 hover:border-gold/50 hover:text-straw'
                  }`}
                >
                  <span className="mr-1">{m.icon}</span>{m.label}
                </button>
              )
            })}
          </div>
          <p className="text-straw/40 text-xs mt-2">
            {activeMode.description}
            <span className="text-straw/30 ml-2 hidden sm:inline">· ← / → para cambiar</span>
          </p>
        </div>

        {/* Selectores */}
        <div className="grid sm:grid-cols-2 gap-4">
          <div>
            <label className="text-straw/60 text-xs uppercase tracking-wider mb-1 block">Origen</label>
            <select
              value={from}
              onChange={e => handleFromChange(e.target.value)}
              className="w-full bg-navy border border-gold/20 focus:border-gold/60 rounded-lg px-3 py-2 text-straw outline-none transition-colors"
            >
              <option value="">— Selecciona isla —</option>
              {islandOptions.map(i => (
                <option key={i.id} value={i.id}>{i.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="text-straw/60 text-xs uppercase tracking-wider mb-1 block">Destino</label>
            <select
              value={to}
              onChange={e => handleToChange(e.target.value)}
              className="w-full bg-navy border border-gold/20 focus:border-gold/60 rounded-lg px-3 py-2 text-straw outline-none transition-colors"
            >
              <option value="">— Selecciona isla —</option>
              {islandOptions.map(i => (
                <option key={i.id} value={i.id}>{i.name}</option>
              ))}
            </select>
          </div>
        </div>

        <button
          onClick={handleCalculate}
          disabled={!from || !to || pathFetching}
          className="btn-primary disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {pathFetching ? '⏳ Calculando…' : '⚡ Calcular ruta'}
        </button>

        {/* Resultado */}
        {pathResult && (
          <div className="space-y-4" aria-live="polite">
            {/* Badge + acciones */}
            <div className="flex items-center gap-3 flex-wrap">
              {pathResult.found ? (
                <span className="badge bg-green-900/50 text-green-400 text-sm px-3 py-1">✅ Ruta encontrada</span>
              ) : (
                <span className="badge bg-red-900/50 text-red-400 text-sm px-3 py-1">❌ Sin ruta navegable</span>
              )}
              {pathResult.found && (
                <span className="text-straw/60 text-sm">Paradas: <span className="text-gold font-bold">{pathResult.hops}</span></span>
              )}
              {pathResult.found && (
                <div className="ml-auto flex gap-2">
                  <button
                    type="button"
                    onClick={() => setAnimate(v => !v)}
                    aria-pressed={animate}
                    className={`text-xs px-2 py-1 rounded border transition-colors ${
                      animate ? 'bg-gold/20 border-gold text-gold' : 'bg-navy border-gold/20 text-straw/70 hover:border-gold/50'
                    }`}
                  >
                    {animate ? '⏸ Pausar nave' : '▶ Animar nave'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowCompare(v => !v)}
                    aria-pressed={showCompare}
                    className={`text-xs px-2 py-1 rounded border transition-colors ${
                      showCompare ? 'bg-gold/20 border-gold text-gold' : 'bg-navy border-gold/20 text-straw/70 hover:border-gold/50'
                    }`}
                  >
                    {showCompare ? '✕ Ocultar comparativa' : '📊 Comparar modos'}
                  </button>
                </div>
              )}
            </div>

            {/* Las 4 métricas globales (siempre visibles, la del modo activo destacada) */}
            {pathResult.found && (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                {(['distance', 'time', 'worstDanger', 'bestDanger'] as MetricKey[]).map(key => {
                  const isPrimary = key === activeMode.primaryMetric
                  const value =
                    key === 'distance'    ? pathResult.totalDistance.toLocaleString() :
                    key === 'time'        ? formatHours(pathResult.totalTime) :
                    key === 'worstDanger' ? `${pathResult.worstDanger} / 5` :
                                            `${pathResult.bestDanger} / 5`
                  return (
                    <div
                      key={key}
                      className={`rounded-lg px-3 py-2 border ${
                        isPrimary
                          ? 'bg-gold/10 border-gold text-gold'
                          : 'bg-navy border-gold/20 text-straw/70'
                      }`}
                    >
                      <div className="text-[10px] uppercase tracking-wider opacity-70">{METRIC_LABELS[key]}</div>
                      <div className={`font-bold text-lg ${isPrimary ? 'text-gold' : 'text-straw'}`}>{value}</div>
                    </div>
                  )
                })}
              </div>
            )}

            {/* Mapa con path resaltado */}
            {pathResult.found && routes && islands && (
              <RouteMap
                islands={islands}
                routes={routes}
                highlightPath={pathResult.path.map(s => s.islandId)}
                animatePath={animate}
              />
            )}

            {/* Comparativa entre los 4 modos */}
            {pathResult.found && showCompare && (() => {
              // Detectar convergencia: todos los modos resueltos producen el mismo path
              const settled = compareResults.filter(r => !r.isFetching && r.data?.found)
              const allConverge =
                settled.length === compareResults.length &&
                settled.every(r =>
                  r.data!.path.map(s => s.islandId).join('>') ===
                  settled[0].data!.path.map(s => s.islandId).join('>')
                )
              return (
                <div className="space-y-2">
                  {allConverge && (
                    <div
                      role="note"
                      className="text-xs px-3 py-2 rounded border border-yellow-500/30 bg-yellow-500/10 text-yellow-200/90"
                    >
                      ⚠ Solo existe una ruta navegable entre estas islas — los 4 modos coinciden.
                      {' '}Probá pares más lejanos (ej. <span className="font-semibold">Windmill Village → Wano</span>) para ver diferencias.
                    </div>
                  )}
                  <div className="border border-gold/20 rounded-lg overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead className="bg-navy">
                        <tr className="text-straw/60 text-xs uppercase">
                          <th className="text-left px-3 py-2">Modo</th>
                          <th className="text-right px-3 py-2">Distancia</th>
                          <th className="text-right px-3 py-2">Tiempo</th>
                          <th className="text-right px-3 py-2">Peor ☠</th>
                          <th className="text-right px-3 py-2">Mejor ☠</th>
                          <th className="text-right px-3 py-2">Paradas</th>
                        </tr>
                      </thead>
                      <tbody>
                    {compareResults.map(r => {
                      const cfg = MODES.find(m => m.id === r.mode)!
                      const isActive = r.mode === mode
                      return (
                        <tr
                          key={r.mode}
                          className={`border-t border-gold/10 ${isActive ? 'bg-gold/5' : ''}`}
                        >
                          <td className={`px-3 py-2 font-semibold ${isActive ? 'text-gold' : 'text-straw'}`}>
                            <span className="mr-1">{cfg.icon}</span>{cfg.label}
                          </td>
                          {r.isFetching ? (
                            <td colSpan={5} className="px-3 py-2 text-straw/40 text-center text-xs">⏳ calculando…</td>
                          ) : r.isError || !r.data ? (
                            <td colSpan={5} className="px-3 py-2 text-red-400/70 text-center text-xs">— error —</td>
                          ) : !r.data.found ? (
                            <td colSpan={5} className="px-3 py-2 text-straw/40 text-center text-xs">sin ruta</td>
                          ) : (
                            <>
                              <td className="px-3 py-2 text-right text-straw/80">{r.data.totalDistance.toLocaleString()}</td>
                              <td className="px-3 py-2 text-right text-straw/80">{formatHours(r.data.totalTime)}</td>
                              <td className="px-3 py-2 text-right text-straw/80">{r.data.worstDanger}/5</td>
                              <td className="px-3 py-2 text-right text-straw/80">{r.data.bestDanger}/5</td>
                              <td className="px-3 py-2 text-right text-straw/80">{r.data.hops}</td>
                            </>
                          )}
                        </tr>
                      )
                    })}
                      </tbody>
                    </table>
                  </div>
                </div>
              )
            })()}

            {/* Timeline de paradas */}
            {pathResult.found && pathResult.path.length > 0 && (
              <div className="space-y-1 max-h-96 overflow-y-auto pr-1">
                {pathResult.path.map((step, i) => {
                  const isEndpoint = i === 0 || i === pathResult.path.length - 1
                  const dangerLevel = dangerByIsland.get(step.islandId)
                  const dangerClass = dangerLevel ? DANGER_COLORS[dangerLevel] ?? '' : ''
                  return (
                    <div key={step.islandId} className="flex items-start gap-3">
                      {/* Línea vertical + nodo coloreado por peligro */}
                      <div className="flex flex-col items-center flex-shrink-0 pt-1">
                        <div
                          className={`w-3 h-3 rounded-full border-2 ${
                            isEndpoint ? 'border-gold bg-gold' : `bg-navy ${dangerClass || 'border-gold/60'}`
                          }`}
                          title={dangerLevel ? `Peligro derivado: ${dangerLevel}/5` : undefined}
                        />
                        {i < pathResult.path.length - 1 && <div className="w-px h-7 bg-gold/20" />}
                      </div>
                      {/* Info */}
                      <div className="flex-1 py-0.5">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="text-straw text-sm font-semibold">{step.islandName}</span>
                          {!isEndpoint && dangerLevel && (
                            <span className={`text-[10px] px-1.5 py-0.5 rounded border ${dangerClass}`}>
                              ☠ {dangerLevel}/5
                            </span>
                          )}
                        </div>
                        {i > 0 && (
                          <div className="text-straw/40 text-xs flex flex-wrap gap-x-3 gap-y-0.5 mt-0.5">
                            <span>📏 {step.distanceSoFar.toLocaleString()}</span>
                            <span>⏱ {formatHours(step.timeSoFar)}</span>
                            <span>☠ peor {step.worstDangerSoFar}/5</span>
                            <span>· mejor {step.bestDangerSoFar}/5</span>
                          </div>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}
      </section>

      {/* ── Sección B: Islas alcanzables ──────────────────────────────────── */}
      <section className="card space-y-6">
        <h2 className="font-pirate text-gold text-2xl">🌊 Islas alcanzables</h2>

        <div className="grid sm:grid-cols-2 gap-4 items-end">
          <div>
            <label className="text-straw/60 text-xs uppercase tracking-wider mb-1 block">Origen</label>
            <select
              value={reachFrom}
              onChange={e => setReachFrom(e.target.value)}
              className="w-full bg-navy border border-gold/20 focus:border-gold/60 rounded-lg px-3 py-2 text-straw outline-none transition-colors"
            >
              <option value="">— Selecciona isla —</option>
              {islandOptions.map(i => (
                <option key={i.id} value={i.id}>{i.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="text-straw/60 text-xs uppercase tracking-wider mb-1 block">
              Presupuesto máximo: <span className="text-gold font-bold">{maxCost.toLocaleString()}</span>
            </label>
            <input
              type="range"
              min={100} max={10000} step={100}
              value={maxCost}
              onChange={e => setMaxCost(Number(e.target.value))}
              className="w-full accent-gold"
            />
            <div className="flex justify-between text-straw/30 text-xs mt-0.5">
              <span>100</span><span>10.000</span>
            </div>
          </div>
        </div>

        {reachResult && (
          <div className={`space-y-3 transition-opacity ${reachFetching ? 'opacity-50' : 'opacity-100'}`}>
            <p className="text-straw/60 text-sm">
              <span className="text-gold font-bold">{reachResult.islands.length}</span> isla{reachResult.islands.length !== 1 ? 's' : ''} alcanzables desde <span className="text-gold">{islandOptions.find(i => i.id === reachFrom)?.name ?? reachFrom}</span>
              {reachFetching && <span className="text-straw/40 ml-2 text-xs">actualizando…</span>}
            </p>

            {reachResult.islands.length === 0 ? (
              <p className="text-straw/40 text-sm py-4 text-center">Sin islas alcanzables con ese presupuesto</p>
            ) : (
              <div className="space-y-2">
                {reachResult.islands.map(island => {
                  const islandData = islandOptions.find(i => i.id === island.islandId)
                  const pct = Math.min(100, (island.cost / maxCost) * 100)
                  return (
                    <div key={island.islandId} className="bg-navy rounded-lg p-3">
                      <div className="flex items-center justify-between mb-1">
                        <div className="flex items-center gap-2">
                          <span className="text-straw text-sm font-semibold">{island.islandName}</span>
                          {islandData && (
                            <span className={`badge text-xs ${REGION_COLORS[islandData.region] ?? 'bg-navy text-straw/60'}`}>
                              {islandData.region}
                            </span>
                          )}
                        </div>
                        <span className="text-gold text-sm font-bold">{island.cost.toLocaleString()}</span>
                      </div>
                      {/* Barra de progreso */}
                      <div className="w-full bg-navy-light rounded-full h-1.5">
                        <div
                          className="bg-gold h-1.5 rounded-full transition-all duration-300"
                          style={{ width: `${pct}%` }}
                        />
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}

        {!reachFrom && (
          <p className="text-straw/30 text-sm text-center py-4">Selecciona una isla para ver las islas alcanzables</p>
        )}
      </section>
    </div>
  )
}
