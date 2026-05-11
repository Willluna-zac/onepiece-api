import { describe, it, expect, vi, beforeEach } from 'vitest'
import { routesApi } from './client'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function makeResponse(data: unknown) {
  return Promise.resolve({
    ok: true,
    json: () => Promise.resolve(data),
  })
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('routesApi', () => {
  it('getAll llama a /api/routes', async () => {
    const routes = [{ id: 'r01', fromIsland: 'windmill-village', toIsland: 'shells-town' }]
    mockFetch.mockReturnValueOnce(makeResponse(routes))

    const result = await routesApi.getAll()

    expect(mockFetch).toHaveBeenCalledOnce()
    expect(mockFetch.mock.calls[0][0]).toContain('/api/routes')
    expect(result).toEqual(routes)
  })

  it('fromIsland construye la URL correcta con el ID', async () => {
    mockFetch.mockReturnValueOnce(makeResponse([]))

    await routesApi.fromIsland('loguetown')

    expect(mockFetch.mock.calls[0][0]).toContain('/api/routes/from/loguetown')
  })

  it('shortest construye query params from, to y mode correctamente', async () => {
    const response = { from: 'windmill-village', to: 'wano', mode: 'fastest' as const, found: true, totalCost: 9729, hops: 17, path: [] }
    mockFetch.mockReturnValueOnce(makeResponse(response))

    await routesApi.shortest('windmill-village', 'wano', 'fastest')

    const url: string = mockFetch.mock.calls[0][0]
    expect(url).toContain('/api/routes/shortest')
    expect(url).toContain('from=windmill-village')
    expect(url).toContain('to=wano')
    expect(url).toContain('mode=fastest')
  })

  it('shortest acepta mode safest, riskiest y quickest', async () => {
    mockFetch.mockReturnValue(makeResponse({ from: 'a', to: 'b', mode: 'safest', found: true, totalCost: 3, hops: 1, path: [] }))

    await routesApi.shortest('a', 'b', 'safest')
    expect(mockFetch.mock.calls[0][0]).toContain('mode=safest')

    await routesApi.shortest('a', 'b', 'riskiest')
    expect(mockFetch.mock.calls[1][0]).toContain('mode=riskiest')

    await routesApi.shortest('a', 'b', 'quickest')
    expect(mockFetch.mock.calls[2][0]).toContain('mode=quickest')
  })

  it('reachable construye query params from y maxCost correctamente', async () => {
    const response = { from: 'loguetown', maxCost: 1500, islands: [] }
    mockFetch.mockReturnValueOnce(makeResponse(response))

    await routesApi.reachable('loguetown', 1500)

    const url: string = mockFetch.mock.calls[0][0]
    expect(url).toContain('/api/routes/reachable')
    expect(url).toContain('from=loguetown')
    expect(url).toContain('maxCost=1500')
  })
})
