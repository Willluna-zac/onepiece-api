import { useState } from 'react'
import { useIslands, useNearestIsland } from '../hooks/useApi'
import { Loader, ErrorMsg } from '../components/Loader'
import type { Island } from '../api/client'

const REGION_COLORS: Record<string, string> = {
  'East Blue':  'border-blue-500 bg-blue-900/20',
  'West Blue':  'border-indigo-500 bg-indigo-900/20',
  'North Blue': 'border-cyan-500 bg-cyan-900/20',
  'South Blue': 'border-teal-500 bg-teal-900/20',
  'Grand Line': 'border-yellow-500 bg-yellow-900/20',
  'New World':  'border-red-500 bg-red-900/20',
}

const REGION_DOT: Record<string, string> = {
  'East Blue':  'bg-blue-400',
  'West Blue':  'bg-indigo-400',
  'North Blue': 'bg-cyan-400',
  'South Blue': 'bg-teal-400',
  'Grand Line': 'bg-yellow-400',
  'New World':  'bg-red-400',
}

// Simple SVG world map — islands plotted by (x/10000 * W, y/5000 * H)
function WorldMap({ islands, nearest, onClick }: {
  islands: Island[]
  nearest?: Island
  onClick: (x: number, y: number) => void
}) {
  const W = 700, H = 350

  function handleClick(e: React.MouseEvent<SVGSVGElement>) {
    const rect = e.currentTarget.getBoundingClientRect()
    const px = ((e.clientX - rect.left) / rect.width) * 10000
    const py = ((e.clientY - rect.top) / rect.height) * 5000
    onClick(Math.round(px), Math.round(py))
  }

  return (
    <div className="relative">
      <p className="text-straw/50 text-xs mb-2">
        🖱️ Haz clic en el mapa para encontrar la isla más cercana (Quadtree)
      </p>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="w-full rounded-xl border border-gold/20 cursor-crosshair"
        style={{ background: 'linear-gradient(180deg, #07111b 0%, #0d2340 50%, #07111b 100%)' }}
        onClick={handleClick}
      >
        {/* Grand Line horizontal line */}
        <line x1={0} y1={H / 2} x2={W} y2={H / 2} stroke="#f5c842" strokeOpacity={0.15} strokeWidth={1} strokeDasharray="6 4" />
        {/* Red Line vertical barrier */}
        <line x1={W / 2} y1={0} x2={W / 2} y2={H} stroke="#ef4444" strokeOpacity={0.15} strokeWidth={1} strokeDasharray="6 4" />

        {/* Labels */}
        <text x={8} y={14} fill="#60a5fa" fontSize={9} opacity={0.7}>East Blue</text>
        <text x={W / 2 + 6} y={14} fill="#f87171" fontSize={9} opacity={0.7}>New World</text>
        <text x={8} y={H - 6} fill="#34d399" fontSize={9} opacity={0.7}>South Blue</text>

        {/* Islands */}
        {islands.map(isl => {
          const cx = (isl.x / 10000) * W
          const cy = (isl.y / 5000) * H
          const isNearest = nearest?.id === isl.id
          return (
            <g key={isl.id}>
              <circle
                cx={cx} cy={cy}
                r={isNearest ? 8 : 5}
                fill={isNearest ? '#f5c842' : '#1e6091'}
                stroke={isNearest ? '#fbd96a' : '#2980b9'}
                strokeWidth={isNearest ? 2 : 1}
                opacity={0.9}
              />
              <text
                x={cx + 7} y={cy + 4}
                fill={isNearest ? '#f5c842' : '#e8d5b0'}
                fontSize={isNearest ? 8 : 7}
                fontWeight={isNearest ? 'bold' : 'normal'}
                opacity={0.85}
              >
                {isl.name}
              </text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

export default function IslandsPage() {
  const { data: islands, isLoading, error } = useIslands()
  const [coords, setCoords] = useState<{ x: number; y: number } | null>(null)
  const [activeRegion, setActiveRegion] = useState<string | null>(null)

  const { data: nearest } = useNearestIsland(
    coords?.x ?? 0,
    coords?.y ?? 0,
    coords !== null,
  )

  const regions = [...new Set(islands?.map(i => i.region) ?? [])]
  const filtered = activeRegion ? islands?.filter(i => i.region === activeRegion) : islands

  if (isLoading) return <Loader />
  if (error) return <ErrorMsg message={String(error)} />

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="font-pirate text-gold text-4xl mb-2">🗺️ Grand Line — Mapa de Islas</h1>
      <p className="text-straw/60 mb-6">
        {islands?.length} islas indexadas · Búsqueda espacial con <span className="text-gold font-semibold">Quadtree</span>
      </p>

      {/* World Map */}
      {islands && (
        <div className="card mb-8">
          <WorldMap islands={islands} nearest={nearest} onClick={(x, y) => setCoords({ x, y })} />
          {coords && (
            <div className="mt-3 flex items-center gap-3 text-sm">
              <span className="text-straw/50">Punto seleccionado: ({coords.x}, {coords.y})</span>
              {nearest && (
                <span className="text-gold font-semibold">
                  📍 Más cercana: <span className="underline">{nearest.name}</span>
                  <span className="text-straw/50 font-normal ml-1">({nearest.region})</span>
                </span>
              )}
              <button onClick={() => { setCoords(null) }} className="ml-auto text-straw/40 hover:text-straw text-xs">✕ Limpiar</button>
            </div>
          )}
        </div>
      )}

      {/* Region filter */}
      <div className="flex flex-wrap gap-2 mb-6">
        <button
          onClick={() => setActiveRegion(null)}
          className={`badge text-xs px-3 py-1 cursor-pointer ${!activeRegion ? 'bg-gold text-navy' : 'bg-navy-light text-straw/60 hover:text-straw'}`}
        >
          Todas
        </button>
        {regions.map(r => (
          <button
            key={r}
            onClick={() => setActiveRegion(r === activeRegion ? null : r)}
            className={`badge text-xs px-3 py-1 cursor-pointer ${activeRegion === r ? 'bg-gold text-navy' : 'bg-navy-light text-straw/60 hover:text-straw'}`}
          >
            {r}
          </button>
        ))}
      </div>

      {/* Island cards */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered?.map(isl => (
          <div
            key={isl.id}
            className={`card border-l-4 ${REGION_COLORS[isl.region] ?? 'border-gold/30'} ${nearest?.id === isl.id ? 'ring-2 ring-gold' : ''}`}
          >
            <div className="flex items-center gap-2 mb-1">
              <span className={`w-2 h-2 rounded-full flex-shrink-0 ${REGION_DOT[isl.region] ?? 'bg-gold'}`} />
              <h3 className="font-pirate text-gold text-xl leading-tight">{isl.name}</h3>
              {nearest?.id === isl.id && <span className="badge bg-gold text-navy text-xs ml-auto">📍 Más cercana</span>}
            </div>
            <p className="text-straw/50 text-xs mb-2">{isl.region} · ({isl.x}, {isl.y})</p>
            <p className="text-straw/70 text-sm mb-3">{isl.description}</p>
            {isl.notable.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {isl.notable.map(n => (
                  <span key={n} className="badge bg-navy text-straw/60 text-xs">{n}</span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
