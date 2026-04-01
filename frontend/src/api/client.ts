const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

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
  type: string
  level: string
}
export interface Ability {
  name: string
  description: string
}
export interface Character {
  id: string
  name: string
  origin: string
  age: number
  bounty: number
  role: string
  crew: string
  status: string
  devilFruit?: DevilFruit
  hakiAbilities?: HakiAbility[]
  abilities?: Ability[]
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
  notable: string[]
}

export const islandsApi = {
  getAll: () => apiFetch<Island[]>('/api/islands'),
  getById: (id: string) => apiFetch<Island>(`/api/islands/${id}`),
  nearest: (x: number, y: number) =>
    apiFetch<Island>(`/api/islands/nearest?x=${x}&y=${y}`),
  byRegion: (region: string) =>
    apiFetch<Island[]>(`/api/islands/region/${encodeURIComponent(region)}`),
}
