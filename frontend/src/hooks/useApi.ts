import { useQuery, useQueries } from '@tanstack/react-query'
import { charactersApi, islandsApi, routesApi, type RouteMode, type ShortestPathResponse } from '../api/client'

export const useCharacters = () =>
  useQuery({ queryKey: ['characters'], queryFn: charactersApi.getAll, staleTime: 60_000 })

export const useCharacter = (id: string) =>
  useQuery({ queryKey: ['characters', id], queryFn: () => charactersApi.getById(id), enabled: !!id })

export const useSearch = (name: string) =>
  useQuery({
    queryKey: ['characters', 'search', name],
    queryFn: () => charactersApi.search(name),
    enabled: name.length > 1,
    staleTime: 30_000,
  })

export const useDevilFruitChars = () =>
  useQuery({ queryKey: ['characters', 'devilfruits'], queryFn: charactersApi.withDevilFruit })

export const useIslands = () =>
  useQuery({ queryKey: ['islands'], queryFn: islandsApi.getAll, staleTime: 300_000 })

export const useNearestIsland = (x: number, y: number, enabled: boolean) =>
  useQuery({
    queryKey: ['islands', 'nearest', x, y],
    queryFn: () => islandsApi.nearest(x, y),
    enabled,
  })

// ─── Routes ──────────────────────────────────────────────────────────────────

export const useRoutes = () =>
  useQuery({ queryKey: ['routes'], queryFn: routesApi.getAll, staleTime: 300_000 })

export const useRoutesFromIsland = (islandID: string, enabled: boolean) =>
  useQuery({
    queryKey: ['routes', 'from', islandID],
    queryFn: () => routesApi.fromIsland(islandID),
    enabled: enabled && islandID !== '',
  })

export const useShortestPath = (from: string, to: string, mode: RouteMode, enabled: boolean) =>
  useQuery({
    queryKey: ['routes', 'shortest', from, to, mode],
    queryFn: () => routesApi.shortest(from, to, mode),
    enabled: enabled && from !== '' && to !== '',
  })

/**
 * Resuelve los 4 modos en paralelo para comparar resultados.
 * Reusa la misma queryKey que useShortestPath, así el hit en cache
 * del modo activo es gratis.
 */
const ALL_MODES: RouteMode[] = ['fastest', 'quickest', 'safest', 'riskiest']

export type CompareModesResult = {
  mode: RouteMode
  data: ShortestPathResponse | undefined
  isFetching: boolean
  isError: boolean
}

export function useCompareModes(from: string, to: string, enabled: boolean): CompareModesResult[] {
  const results = useQueries({
    queries: ALL_MODES.map(m => ({
      queryKey: ['routes', 'shortest', from, to, m] as const,
      queryFn: () => routesApi.shortest(from, to, m),
      enabled: enabled && from !== '' && to !== '',
    })),
  })
  return results.map((r, i) => ({
    mode: ALL_MODES[i],
    data: r.data,
    isFetching: r.isFetching,
    isError: r.isError,
  }))
}

export const useReachableIslands = (from: string, maxCost: number, enabled: boolean) =>
  useQuery({
    queryKey: ['routes', 'reachable', from, maxCost],
    queryFn: () => routesApi.reachable(from, maxCost),
    enabled: enabled && from !== '' && maxCost > 0,
  })

export const useGraphStats = () =>
  useQuery({
    queryKey: ['routes', 'stats'],
    queryFn: routesApi.stats,
    staleTime: 60_000,
  })
