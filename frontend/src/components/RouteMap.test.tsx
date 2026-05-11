import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { RouteMap } from './RouteMap'
import type { Island, Route } from '../api/client'

const islands: Island[] = [
  { id: 'windmill-village', name: 'Windmill Village', region: 'East Blue', x: 500, y: 1800, logPoseHours: 0, description: '', notableCharacters: [] },
  { id: 'shells-town',      name: 'Shells Town',      region: 'East Blue', x: 700, y: 2200, logPoseHours: 0, description: '', notableCharacters: [] },
  { id: 'orange-town',      name: 'Orange Town',      region: 'East Blue', x: 900, y: 2100, logPoseHours: 0, description: '', notableCharacters: [] },
]

const routes: Route[] = [
  { id: 'r01', fromIsland: 'windmill-village', toIsland: 'shells-town', distance: 200, travelHours: 6, danger: 1, weight: 200, bidirectional: true },
  { id: 'r02', fromIsland: 'shells-town', toIsland: 'orange-town', distance: 180, travelHours: 5, danger: 3, weight: 288, bidirectional: false },
]

describe('RouteMap', () => {
  it('renderiza sin rutas — solo islas como círculos SVG', () => {
    const { container } = render(
      <RouteMap islands={islands} routes={[]} />
    )
    const circles = container.querySelectorAll('circle')
    // 3 islas × 1 círculo c/u (sin path activo no hay anillos)
    expect(circles.length).toBe(3)
  })

  it('con rutas, renderiza el número correcto de elementos <line>', () => {
    const { container } = render(
      <RouteMap islands={islands} routes={routes} />
    )
    // 2 rutas + 2 líneas de referencia (Grand Line, Red Line) + 5 líneas de leyenda = 9
    const lines = container.querySelectorAll('line')
    expect(lines.length).toBe(9)
  })

  it('con highlightPath, las aristas del path tienen stroke dorado', () => {
    const { container } = render(
      <RouteMap
        islands={islands}
        routes={routes}
        highlightPath={['windmill-village', 'shells-town']}
      />
    )
    const lines = Array.from(container.querySelectorAll('line'))
    const goldenLines = lines.filter(l => l.getAttribute('stroke') === '#f5c842')
    // 1 ruta resaltada
    expect(goldenLines.length).toBeGreaterThanOrEqual(1)
  })

  it('ruta solo-ida tiene strokeDasharray en el SVG', () => {
    const { container } = render(
      <RouteMap islands={islands} routes={routes} />
    )
    const lines = Array.from(container.querySelectorAll('line'))
    // r02 es solo-ida (bidirectional: false)
    const dashedLines = lines.filter(l => l.getAttribute('stroke-dasharray') === '6 3')
    expect(dashedLines.length).toBe(1)
  })

  it('tooltip <title> contiene los nombres de las islas', () => {
    const { container } = render(
      <RouteMap islands={islands} routes={routes} />
    )
    const titles = Array.from(container.querySelectorAll('title')).map(t => t.textContent)
    expect(titles.some(t => t?.includes('Windmill Village'))).toBe(true)
    expect(titles.some(t => t?.includes('Shells Town'))).toBe(true)
  })
})
