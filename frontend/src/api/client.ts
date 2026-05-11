// En dev, Vite proxea /api/* al backend (ver vite.config.ts) → usamos path relativo
// y así evitamos CORS. En prod, se puede setear VITE_API_URL al dominio del backend.
const BASE_URL = import.meta.env.VITE_API_URL ?? ''

/** Convierte una URL externa en una URL proxeada por el backend.
 *  Esto evita CORS y hotlink protection de Fandom wiki. */
export function proxyImage(url: string | undefined): string | undefined {
  if (!url) return undefined
  return `${BASE_URL}/api/proxy/image?url=${encodeURIComponent(url)}`
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, options)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  const json = await res.json()
  return json.data ?? json
}

// ─── Characters ──────────────────────────────────────────────────────────────

export interface DevilFruit {
  name: string
  type: string
  description: string
}
export interface HakiAbility {
  hakiType: string      // "Conqueror" | "Armament" | "Observation"
  proficiency: string   // "Basic" | "Advanced" | "Master"
  awakened: boolean
  notes?: string
}
export interface Ability {
  type: string
  notes?: string
}
export interface Character {
  id: string
  name: string
  alias: string
  species: string
  role: string
  firstAppearance: string
  imageUrl?: string
  devilFruit?: DevilFruit
  hakiAbilities?: HakiAbility[]
  abilities?: Ability[]
  notes?: string
}

export const charactersApi = {
  getAll: () => apiFetch<Character[]>('/api/characters'),
  getById: (id: string) => apiFetch<Character>(`/api/characters/${id}`),
  search: (name: string) => apiFetch<Character[]>(`/api/characters/search?name=${encodeURIComponent(name)}`),
  withDevilFruit: () => apiFetch<Character[]>('/api/characters/devil-fruits'),
}

// ─── Islands ─────────────────────────────────────────────────────────────────

export interface Island {
  id: string
  name: string
  region: string
  x: number
  y: number
  /** Tiempo (horas) que tarda el Log Pose en cargarse. 0 si la isla no usa Log Pose. */
  logPoseHours: number
  description: string
  notableCharacters: string[]
}

export const islandsApi = {
  getAll: () => apiFetch<Island[]>('/api/islands'),
  getById: (id: string) => apiFetch<Island>(`/api/islands/${id}`),
  nearest: (x: number, y: number) =>
    apiFetch<Island>(`/api/islands/nearest?x=${x}&y=${y}`),
  byRegion: (region: string) =>
    apiFetch<Island[]>(`/api/islands/region/${encodeURIComponent(region)}`),
}

// ─── Routes ──────────────────────────────────────────────────────────────────

export interface Route {
  id: string
  fromIsland: string
  toIsland: string
  /** Distancia del tramo (unidades del mapa). */
  distance: number
  /** Tiempo de navegación del tramo en horas. Independiente de distance. */
  travelHours: number
  /** Peligrosidad del tramo (1–5). */
  danger: number
  /** @deprecated Peso heredado (distance * multiplicador de peligro). Los modos modernos lo ignoran. */
  weight: number
  bidirectional: boolean
  notes?: string
}

export interface PathStep {
  islandId: string
  islandName: string
  /** Distancia acumulada hasta este paso. */
  distanceSoFar: number
  /** Tiempo acumulado (TravelHours + LogPose de islas intermedias). */
  timeSoFar: number
  /** Peor Danger de las aristas recorridas hasta aquí. */
  worstDangerSoFar: number
  /** Mejor Danger de las aristas recorridas hasta aquí. */
  bestDangerSoFar: number
  /** @deprecated Refleja la métrica del modo activo (use los 4 campos específicos). */
  costSoFar: number
}

/** Modo de búsqueda de ruta. Coincide con el enum del backend. */
export type RouteMode = 'fastest' | 'quickest' | 'safest' | 'riskiest'

export interface ShortestPathResponse {
  from: string
  to: string
  mode: RouteMode
  /** Suma total de distancias del camino. Siempre presente. */
  totalDistance: number
  /** Tiempo total del viaje en horas (TravelHours + LogPose intermedios). Siempre presente. */
  totalTime: number
  /** Peor Danger del camino. Siempre presente. */
  worstDanger: number
  /** Mejor Danger del camino. Siempre presente. */
  bestDanger: number
  /** @deprecated Refleja la métrica del modo activo (use los 4 campos específicos). */
  totalCost: number
  hops: number
  path: PathStep[]
  found: boolean
}

export interface ReachableIsland {
  islandId: string
  islandName: string
  cost: number
}

export interface ReachableResponse {
  from: string
  maxCost: number
  islands: ReachableIsland[]
}

export const routesApi = {
  getAll: () =>
    apiFetch<Route[]>('/api/routes'),
  fromIsland: (id: string) =>
    apiFetch<Route[]>(`/api/routes/from/${encodeURIComponent(id)}`),
  shortest: (from: string, to: string, mode: RouteMode = 'fastest') =>
    apiFetch<ShortestPathResponse>(
      `/api/routes/shortest?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}&mode=${mode}`,
    ),
  reachable: (from: string, maxCost: number) =>
    apiFetch<ReachableResponse>(`/api/routes/reachable?from=${encodeURIComponent(from)}&maxCost=${maxCost}`),
  stats: () => apiFetch<GraphStats>('/api/routes/stats'),
}

export interface GraphStats {
  totalIslands: number
  totalRoutes: number
  bidirectionalCount: number
  islandsWithLogPose: number
  connectedComponents: number
  largestComponent: number
  avgDistance: number
  avgTravelHours: number
  avgDanger: number
  /** Histograma fijo de 5 buckets: índice i = cantidad de rutas con Danger == i+1. */
  dangerHistogram: [number, number, number, number, number]
}
