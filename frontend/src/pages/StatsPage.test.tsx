import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { createElement } from 'react'
import type { UseQueryResult } from '@tanstack/react-query'
import StatsPage from './StatsPage'
import * as apiHooks from '../hooks/useApi'
import type { GraphStats } from '../api/client'

function mockQuery<T>(data: T | undefined, extra: Record<string, unknown> = {}): UseQueryResult<T> {
  return { data, isLoading: false, isFetching: false, error: null, ...extra } as UseQueryResult<T>
}

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return createElement(QueryClientProvider, { client: qc },
    createElement(MemoryRouter, null, children)
  )
}

const baseStats: GraphStats = {
  totalIslands: 32,
  totalRoutes: 48,
  bidirectionalCount: 41,
  islandsWithLogPose: 19,
  connectedComponents: 1,
  largestComponent: 32,
  avgDistance: 478.4,
  avgTravelHours: 29.9,
  avgDanger: 3.02,
  dangerHistogram: [5, 12, 14, 11, 6],
}

describe('StatsPage', () => {
  it('muestra los totales y el histograma de Danger', () => {
    vi.spyOn(apiHooks, 'useGraphStats').mockReturnValue(mockQuery(baseStats))
    render(createElement(StatsPage), { wrapper })

    expect(screen.getByRole('heading', { name: /análisis del grafo/i })).toBeInTheDocument()
    expect(screen.getByText('32')).toBeInTheDocument() // totalIslands
    expect(screen.getByText('48')).toBeInTheDocument() // totalRoutes
    // Histograma: 5 filas
    const list = screen.getByRole('list', { name: /histograma de danger/i })
    expect(list.querySelectorAll('[role="listitem"]').length).toBe(5)
    // Cada bucket muestra el conteo
    expect(screen.getByText('14')).toBeInTheDocument()
  })

  it('marca el componente como saludable cuando connectedComponents === 1', () => {
    vi.spyOn(apiHooks, 'useGraphStats').mockReturnValue(mockQuery(baseStats))
    render(createElement(StatsPage), { wrapper })
    expect(screen.getByText(/todo conectado/i)).toBeInTheDocument()
  })

  it('marca warning si hay más de un componente conexo', () => {
    vi.spyOn(apiHooks, 'useGraphStats').mockReturnValue(
      mockQuery({ ...baseStats, connectedComponents: 2, largestComponent: 23 }),
    )
    render(createElement(StatsPage), { wrapper })
    expect(screen.getByText(/mayor: 23 islas/i)).toBeInTheDocument()
  })
})
