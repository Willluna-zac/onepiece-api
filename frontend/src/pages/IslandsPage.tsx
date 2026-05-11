import { useState } from 'react'
import { useIslands, useNearestIsland, useRoutes } from '../hooks/useApi'
import { Loader, ErrorMsg } from '../components/Loader'
import { RouteMap } from '../components/RouteMap'
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


export default function IslandsPage() {
  const { data: islands, isLoading, error } = useIslands()
  const { data: routes } = useRoutes()
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
          <p className="text-straw/50 text-xs mb-2">
            🖱️ Haz clic en el mapa para encontrar la isla más cercana (Quadtree)
          </p>
          <RouteMap
            islands={islands}
            routes={routes ?? []}
            nearestIsland={nearest}
            onMapClick={(x: number, y: number) => setCoords({ x, y })}
          />
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
            {( isl.notableCharacters?.length ?? 0) > 0 && (
              <div className="flex flex-wrap gap-1">
                {isl.notableCharacters.map(n => (
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
