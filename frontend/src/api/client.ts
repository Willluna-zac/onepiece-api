const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

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
