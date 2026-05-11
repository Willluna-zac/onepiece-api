import type { Island, Route } from '../api/client'
import { formatHours } from '../lib/format'

const W = 700
const H = 350

// Colores por nivel de peligro
const DANGER_COLOR: Record<number, string> = {
  1: '#3b82f6', // azul
  2: '#22c55e', // verde
  3: '#eab308', // amarillo
  4: '#f97316', // naranja
  5: '#ef4444', // rojo
}

function toSvgX(x: number) { return (x / 10000) * W }
function toSvgY(y: number) { return (y / 5000) * H }

interface RouteMapProps {
  islands: Island[]
  routes: Route[]
  highlightPath?: string[]
  nearestIsland?: Island
  onMapClick?: (x: number, y: number) => void
  /** Si true, anima un círculo dorado avanzando por highlightPath en ~5s. */
  animatePath?: boolean
}

export function RouteMap({ islands, routes, highlightPath = [], nearestIsland, onMapClick, animatePath = false }: RouteMapProps) {
  const islandMap = new Map(islands.map(i => [i.id, i]))

  // Set de pares de aristas del path resaltado para lookup rápido
  const pathEdges = new Set<string>()
  for (let i = 0; i < highlightPath.length - 1; i++) {
    pathEdges.add(`${highlightPath[i]}|${highlightPath[i + 1]}`)
    pathEdges.add(`${highlightPath[i + 1]}|${highlightPath[i]}`)
  }
  const pathSet = new Set(highlightPath)

  // Index de rutas por (from->to) para tooltips y peligro derivado por isla
  const routeByEdge = new Map<string, Route>()
  for (const r of routes) {
    routeByEdge.set(`${r.fromIsland}->${r.toIsland}`, r)
    if (r.bidirectional) routeByEdge.set(`${r.toIsland}->${r.fromIsland}`, r)
  }
  // Peligro derivado por isla = max(Danger) de aristas del path que la tocan
  const dangerByIsland = new Map<string, number>()
  for (let i = 0; i < highlightPath.length - 1; i++) {
    const r = routeByEdge.get(`${highlightPath[i]}->${highlightPath[i + 1]}`)
    if (!r) continue
    dangerByIsland.set(highlightPath[i],     Math.max(dangerByIsland.get(highlightPath[i])     ?? 0, r.danger))
    dangerByIsland.set(highlightPath[i + 1], Math.max(dangerByIsland.get(highlightPath[i + 1]) ?? 0, r.danger))
  }

  // Path SVG (M..L..) para la animación de la nave
  const pathDataString = (() => {
    if (!animatePath || highlightPath.length < 2) return ''
    const pts = highlightPath
      .map(id => islandMap.get(id))
      .filter((i): i is Island => Boolean(i))
      .map(i => `${toSvgX(i.x).toFixed(2)},${toSvgY(i.y).toFixed(2)}`)
    if (pts.length < 2) return ''
    return `M${pts[0]} L${pts.slice(1).join(' L')}`
  })()

  function handleClick(e: React.MouseEvent<SVGSVGElement>) {
    if (!onMapClick) return
    const rect = e.currentTarget.getBoundingClientRect()
    const px = ((e.clientX - rect.left) / rect.width) * 10000
    const py = ((e.clientY - rect.top) / rect.height) * 5000
    onMapClick(Math.round(px), Math.round(py))
  }

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className={`w-full rounded-xl border border-gold/20 ${onMapClick ? 'cursor-crosshair' : ''}`}
      style={{ background: 'linear-gradient(180deg, #07111b 0%, #0d2340 50%, #07111b 100%)' }}
      onClick={handleClick}
    >
      {/* Grand Line y Red Line */}
      <line x1={0} y1={H / 2} x2={W} y2={H / 2} stroke="#f5c842" strokeOpacity={0.15} strokeWidth={1} strokeDasharray="6 4" />
      <line x1={W / 2} y1={0} x2={W / 2} y2={H} stroke="#ef4444" strokeOpacity={0.15} strokeWidth={1} strokeDasharray="6 4" />

      {/* Labels */}
      <text x={8} y={14} fill="#60a5fa" fontSize={9} opacity={0.7}>East Blue</text>
      <text x={W / 2 + 6} y={14} fill="#f87171" fontSize={9} opacity={0.7}>New World</text>
      <text x={8} y={H - 6} fill="#34d399" fontSize={9} opacity={0.7}>South Blue</text>

      {/* Rutas — se dibujan antes que las islas para quedar debajo */}
      {routes.map(route => {
        const from = islandMap.get(route.fromIsland)
        const to = islandMap.get(route.toIsland)
        if (!from || !to) return null

        const isHighlighted = pathEdges.has(`${route.fromIsland}|${route.toIsland}`)
        const color = isHighlighted ? '#f5c842' : DANGER_COLOR[route.danger] ?? '#94a3b8'
        const opacity = isHighlighted ? 1 : highlightPath.length > 0 ? 0.15 : 0.4
        const width = isHighlighted ? 3 : 1
        const dashArray = route.bidirectional ? undefined : '6 3'

        return (
          <line
            key={route.id}
            x1={toSvgX(from.x)} y1={toSvgY(from.y)}
            x2={toSvgX(to.x)}   y2={toSvgY(to.y)}
            stroke={color}
            strokeOpacity={opacity}
            strokeWidth={width}
            strokeDasharray={dashArray}
          >
            <title>
              {from.name} → {to.name}
              {'\n'}📏 Distancia: {route.distance}
              {'\n'}⏱ Tiempo: {formatHours(route.travelHours)}
              {'\n'}☠ Peligro: {route.danger}/5
              {route.bidirectional ? '' : '\n↛ Unidireccional'}
            </title>
          </line>
        )
      })}

      {/* Islas */}
      {islands.map(isl => {
        const cx = toSvgX(isl.x)
        const cy = toSvgY(isl.y)
        const isNearest = nearestIsland?.id === isl.id
        const inPath = pathSet.has(isl.id)
        const r = isNearest || inPath ? 7 : 4
        const dangerLevel = dangerByIsland.get(isl.id)
        const ringColor = dangerLevel ? DANGER_COLOR[dangerLevel] ?? '#f5c842' : '#f5c842'

        return (
          <g key={isl.id}>
            {inPath && (
              <circle cx={cx} cy={cy} r={r + 4} fill="none" stroke={ringColor} strokeWidth={2} opacity={0.75} />
            )}
            <circle
              cx={cx} cy={cy} r={r}
              fill={inPath ? '#f5c842' : isNearest ? '#f5c842' : '#1e6091'}
              stroke={inPath || isNearest ? '#fbd96a' : '#2980b9'}
              strokeWidth={inPath || isNearest ? 2 : 1}
              opacity={0.9}
            >
              <title>
                {isl.name}
                {'\n'}🌍 Región: {isl.region}
                {'\n'}🧭 Log Pose: {isl.logPoseHours > 0 ? formatHours(isl.logPoseHours) : 'no requerido'}
                {dangerLevel ? `\n☠ Peligro derivado: ${dangerLevel}/5` : ''}
                {isl.description ? `\n${isl.description}` : ''}
              </title>
            </circle>
            <text
              x={cx + 7} y={cy + 4}
              fill={inPath || isNearest ? '#f5c842' : '#e8d5b0'}
              fontSize={inPath || isNearest ? 8 : 7}
              fontWeight={inPath || isNearest ? 'bold' : 'normal'}
              opacity={0.85}
            >
              {isl.name}
            </text>
          </g>
        )
      })}

      {/* Nave animada (MVP) — recorre el path en 5s lineales */}
      {animatePath && pathDataString && (
        <g pointerEvents="none">
          <circle r={5} fill="#fbd96a" stroke="#7c4a02" strokeWidth={1.5}>
            <animateMotion dur="5s" repeatCount="indefinite" path={pathDataString} rotate="auto" />
          </circle>
        </g>
      )}

      {/* Leyenda de peligro */}
      {[1, 2, 3, 4, 5].map((d, i) => (
        <g key={d} transform={`translate(${W - 90}, ${H - 68 + i * 12})`}>
          <line x1={0} y1={4} x2={14} y2={4} stroke={DANGER_COLOR[d]} strokeWidth={2} />
          <text x={18} y={8} fill="#e8d5b0" fontSize={7} opacity={0.7}>Peligro {d}</text>
        </g>
      ))}
    </svg>
  )
}
