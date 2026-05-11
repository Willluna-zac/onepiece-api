import { describe, it, expect } from 'vitest'
import { formatHours } from './format'

describe('formatHours', () => {
  it('cero o negativo → "0 h"', () => {
    expect(formatHours(0)).toBe('0 h')
    expect(formatHours(-5)).toBe('0 h')
  })

  it('< 1h → minutos redondeados a múltiplos de 5', () => {
    expect(formatHours(0.5)).toBe('30 min')
    expect(formatHours(0.25)).toBe('15 min')
    expect(formatHours(0.1)).toBe('5 min') // 6 min redondea a 5
  })

  it('horas exactas < 24 → "Xh"', () => {
    expect(formatHours(1)).toBe('1 h')
    expect(formatHours(12)).toBe('12 h')
    expect(formatHours(23)).toBe('23 h')
  })

  it('horas con minutos < 24 → "Xh Ymin"', () => {
    expect(formatHours(1.5)).toBe('1 h 30 min')
    expect(formatHours(2.25)).toBe('2 h 15 min')
  })

  it('>= 24h exactas → "Xd"', () => {
    expect(formatHours(24)).toBe('1 d')
    expect(formatHours(48)).toBe('2 d')
    expect(formatHours(72)).toBe('3 d')
  })

  it('días con horas → "Xd Yh"', () => {
    expect(formatHours(25)).toBe('1 d 1 h')
    expect(formatHours(50)).toBe('2 d 2 h')
    expect(formatHours(99)).toBe('4 d 3 h')
  })

  it('valores no finitos', () => {
    expect(formatHours(NaN)).toBe('0 h')
    expect(formatHours(Infinity)).toBe('0 h')
  })
})
