import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement } from 'react'
import { useRoutes, useShortestPath, useReachableIslands, useCompareModes } from './useApi'
import * as client from '../api/client'

// Wrapper con QueryClient fresco para cada test
function makeWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: qc }, children)
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('useRoutes', () => {
  it('llama a routesApi.getAll y expone los datos', async () => {
    const routes = [{ id: 'r01', fromIsland: 'a', toIsland: 'b', distance: 100, travelHours: 4, danger: 1, weight: 100, bidirectional: true }]
    vi.spyOn(client.routesApi, 'getAll').mockResolvedValueOnce(routes)

    const { result } = renderHook(() => useRoutes(), { wrapper: makeWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(routes)
  })
})

describe('useShortestPath', () => {
  it('NO hace fetch cuando enabled=false', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest')

    renderHook(() => useShortestPath('a', 'b', 'fastest', false), { wrapper: makeWrapper() })

    await new Promise(r => setTimeout(r, 50))
    expect(spy).not.toHaveBeenCalled()
  })

  it('NO hace fetch cuando from está vacío', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest')

    renderHook(() => useShortestPath('', 'wano', 'fastest', true), { wrapper: makeWrapper() })

    await new Promise(r => setTimeout(r, 50))
    expect(spy).not.toHaveBeenCalled()
  })

  it('SÍ hace fetch cuando ambos IDs están definidos y enabled=true', async () => {
    const response = { from: 'a', to: 'b', mode: 'fastest' as const, found: true, totalCost: 500, hops: 2, totalDistance: 500, totalTime: 20, worstDanger: 2, bestDanger: 1, path: [] }
    vi.spyOn(client.routesApi, 'shortest').mockResolvedValueOnce(response)

    const { result } = renderHook(
      () => useShortestPath('windmill-village', 'wano', 'fastest', true),
      { wrapper: makeWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.found).toBe(true)
  })

  it('cambiar mode genera otra entrada de cache (fetch distinto)', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest').mockResolvedValue({
      from: 'a', to: 'b', mode: 'safest', found: true, totalCost: 3, hops: 1,
      totalDistance: 100, totalTime: 5, worstDanger: 3, bestDanger: 1, path: [],
    })

    const { rerender } = renderHook(
      ({ mode }) => useShortestPath('a', 'b', mode, true),
      { wrapper: makeWrapper(), initialProps: { mode: 'fastest' as 'fastest' | 'safest' } },
    )

    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1))

    rerender({ mode: 'safest' })
    await waitFor(() => expect(spy).toHaveBeenCalledTimes(2))
  })
})

describe('useReachableIslands', () => {
  it('NO hace fetch cuando maxCost = 0', async () => {
    const spy = vi.spyOn(client.routesApi, 'reachable')

    renderHook(() => useReachableIslands('loguetown', 0, true), { wrapper: makeWrapper() })

    await new Promise(r => setTimeout(r, 50))
    expect(spy).not.toHaveBeenCalled()
  })

  it('SÍ hace fetch con from y maxCost válidos', async () => {
    const response = { from: 'loguetown', maxCost: 1500, islands: [] }
    vi.spyOn(client.routesApi, 'reachable').mockResolvedValueOnce(response)

    const { result } = renderHook(
      () => useReachableIslands('loguetown', 1500, true),
      { wrapper: makeWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.from).toBe('loguetown')
  })
})

describe('useCompareModes', () => {
  function mkResp(mode: client.RouteMode, totalDistance: number) {
    return {
      from: 'a', to: 'b', mode, found: true, hops: 1, totalCost: totalDistance,
      totalDistance, totalTime: 5, worstDanger: 2, bestDanger: 1,
      path: [],
    } as client.ShortestPathResponse
  }

  it('NO hace fetch cuando enabled=false', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest')

    renderHook(() => useCompareModes('a', 'b', false), { wrapper: makeWrapper() })

    await new Promise(r => setTimeout(r, 50))
    expect(spy).not.toHaveBeenCalled()
  })

  it('NO hace fetch cuando from o to vacíos', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest')

    renderHook(() => useCompareModes('', 'b', true), { wrapper: makeWrapper() })

    await new Promise(r => setTimeout(r, 50))
    expect(spy).not.toHaveBeenCalled()
  })

  it('dispara los 4 modos en paralelo y retorna {mode, data, isFetching, isError}', async () => {
    const spy = vi.spyOn(client.routesApi, 'shortest').mockImplementation(
      async (_from, _to, mode) => mkResp(mode!, mode === 'fastest' ? 100 : 200),
    )

    const { result } = renderHook(
      () => useCompareModes('a', 'b', true),
      { wrapper: makeWrapper() },
    )

    await waitFor(() => expect(result.current.every(r => !r.isFetching)).toBe(true))

    expect(spy).toHaveBeenCalledTimes(4)
    expect(result.current.map(r => r.mode)).toEqual(['fastest', 'quickest', 'safest', 'riskiest'])
    expect(result.current.every(r => r.data?.found === true)).toBe(true)
    expect(result.current.every(r => r.isError === false)).toBe(true)
  })
})
