import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { createElement } from 'react'
import type { UseQueryResult } from '@tanstack/react-query'
import RoutesPage from './RoutesPage'
import * as apiHooks from '../hooks/useApi'
import type { Island, Route, ShortestPathResponse, ReachableResponse } from '../api/client'

// Helper: crea un resultado de useQuery parcial para mocks
function mockQuery<T>(data: T | undefined, extra: Record<string, unknown> = {}): UseQueryResult<T> {
  return { data, isLoading: false, isFetching: false, error: null, ...extra } as UseQueryResult<T>
}

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return createElement(QueryClientProvider, { client: qc },
    createElement(MemoryRouter, null, children)
  )
}

const mockIslands = [
  { id: 'windmill-village', name: 'Windmill Village', region: 'East Blue', x: 500, y: 1800, logPoseHours: 0, description: '', notableCharacters: [] },
  { id: 'wano',             name: 'Wano Country',     region: 'New World',  x: 7200, y: 2700, logPoseHours: 96, description: '', notableCharacters: [] },
]

const mockRoutes = [
  { id: 'r01', fromIsland: 'windmill-village', toIsland: 'wano', distance: 500, travelHours: 27.78, danger: 3, weight: 800, bidirectional: true },
]

beforeEach(() => {
  vi.restoreAllMocks()
  vi.spyOn(apiHooks, 'useIslands').mockReturnValue(mockQuery<Island[]>(mockIslands))
  vi.spyOn(apiHooks, 'useRoutes').mockReturnValue(mockQuery<Route[]>(mockRoutes))
  vi.spyOn(apiHooks, 'useShortestPath').mockReturnValue(mockQuery<ShortestPathResponse>(undefined))
  vi.spyOn(apiHooks, 'useReachableIslands').mockReturnValue(mockQuery<ReachableResponse>(undefined))
  vi.spyOn(apiHooks, 'useCompareModes').mockReturnValue([])
})

describe('RoutesPage', () => {
  it('renderiza ambas secciones', () => {
    render(<RoutesPage />, { wrapper })

    expect(screen.getByRole('heading', { name: /Ruta más corta/i })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Islas alcanzables/i })).toBeInTheDocument()
  })

  it('botón "Calcular ruta" está disabled cuando no hay selects elegidos', () => {
    render(<RoutesPage />, { wrapper })

    const btn = screen.getByRole('button', { name: /calcular ruta/i })
    expect(btn).toBeDisabled()
  })

  it('slider de maxCost tiene valor inicial 1500', () => {
    render(<RoutesPage />, { wrapper })

    const slider = screen.getByRole('slider')
    expect(slider).toHaveValue('1500')
  })

  it('con resultado found=true, muestra distancia total y hops', () => {
    const pathData: ShortestPathResponse = {
      from: 'windmill-village', to: 'wano', mode: 'fastest',
      found: true, totalCost: 9729, hops: 17,
      totalDistance: 9729, totalTime: 540, worstDanger: 4, bestDanger: 1,
      path: [
        { islandId: 'windmill-village', islandName: 'Windmill Village', costSoFar: 0,    distanceSoFar: 0,    timeSoFar: 0,   worstDangerSoFar: 0, bestDangerSoFar: 0 },
        { islandId: 'wano',             islandName: 'Wano Country',     costSoFar: 9729, distanceSoFar: 9729, timeSoFar: 540, worstDangerSoFar: 4, bestDangerSoFar: 1 },
      ],
    }
    vi.spyOn(apiHooks, 'useShortestPath').mockReturnValue(mockQuery(pathData))

    render(<RoutesPage />, { wrapper })

    expect(screen.getByText(/Ruta encontrada/i)).toBeInTheDocument()
    // Las 4 métricas globales deben verse siempre, sin importar el modo
    expect(screen.getAllByText(/Distancia total/i).length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText(/Tiempo total/i)).toBeInTheDocument()
    expect(screen.getByText(/Peor tramo/i)).toBeInTheDocument()
    expect(screen.getByText(/Mejor tramo/i)).toBeInTheDocument()
    // Los nombres aparecen en selects y en el timeline
    expect(screen.getAllByText('Windmill Village').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Wano Country').length).toBeGreaterThanOrEqual(1)
  })

  it('con found=false, muestra badge de Sin ruta', () => {
    const pathData: ShortestPathResponse = {
      from: 'a', to: 'b', mode: 'fastest', found: false,
      totalCost: 0, hops: 0, totalDistance: 0, totalTime: 0, worstDanger: 0, bestDanger: 0,
      path: [],
    }
    vi.spyOn(apiHooks, 'useShortestPath').mockReturnValue(mockQuery(pathData))

    render(<RoutesPage />, { wrapper })

    expect(screen.getByText(/Sin ruta navegable/i)).toBeInTheDocument()
  })

  it('renderiza el toggle de modo con los 4 botones (fastest activo por default)', () => {
    render(<RoutesPage />, { wrapper })

    const fastest  = screen.getByRole('button', { name: /Rápido/i })
    const quickest = screen.getByRole('button', { name: /Menos tiempo/i })
    const safest   = screen.getByRole('button', { name: /Más segura/i })
    const riskiest = screen.getByRole('button', { name: /Más peligrosa/i })

    expect(fastest).toHaveAttribute('aria-pressed', 'true')
    expect(quickest).toHaveAttribute('aria-pressed', 'false')
    expect(safest).toHaveAttribute('aria-pressed', 'false')
    expect(riskiest).toHaveAttribute('aria-pressed', 'false')
  })

  it('al togglear "Comparar modos" muestra la tabla con 4 filas (una por modo)', async () => {
    const user = userEvent.setup()

    const pathData: ShortestPathResponse = {
      from: 'windmill-village', to: 'wano', mode: 'fastest',
      found: true, totalCost: 100, hops: 1,
      totalDistance: 100, totalTime: 5, worstDanger: 1, bestDanger: 1,
      path: [
        { islandId: 'windmill-village', islandName: 'Windmill Village', costSoFar: 0,   distanceSoFar: 0,   timeSoFar: 0, worstDangerSoFar: 0, bestDangerSoFar: 0 },
        { islandId: 'wano',             islandName: 'Wano Country',     costSoFar: 100, distanceSoFar: 100, timeSoFar: 5, worstDangerSoFar: 1, bestDangerSoFar: 1 },
      ],
    }
    vi.spyOn(apiHooks, 'useShortestPath').mockReturnValue(mockQuery(pathData))
    // Mock de useCompareModes: 4 filas resueltas
    vi.spyOn(apiHooks, 'useCompareModes').mockReturnValue([
      { mode: 'fastest',  data: { ...pathData, mode: 'fastest'  }, isFetching: false, isError: false },
      { mode: 'quickest', data: { ...pathData, mode: 'quickest' }, isFetching: false, isError: false },
      { mode: 'safest',   data: { ...pathData, mode: 'safest'   }, isFetching: false, isError: false },
      { mode: 'riskiest', data: { ...pathData, mode: 'riskiest' }, isFetching: false, isError: false },
    ])

    render(<RoutesPage />, { wrapper })

    // No visible inicialmente
    expect(screen.queryByRole('table')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Comparar modos/i }))

    const table = await screen.findByRole('table')
    expect(table).toBeInTheDocument()
    // Las 4 filas + header
    expect(table.querySelectorAll('tbody tr').length).toBe(4)
    // Y dado que los 4 paths son idénticos → debe verse el badge de convergencia
    expect(screen.getByRole('note')).toHaveTextContent(/Solo existe una ruta navegable/i)
  })
})
