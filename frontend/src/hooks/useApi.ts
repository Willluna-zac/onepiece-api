import { useQuery } from '@tanstack/react-query'
import { charactersApi, islandsApi } from '../api/client'

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
