import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useLastRouteSearch } from './useLastRouteSearch'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('useLastRouteSearch', () => {
  it('arranca en null si no hay nada guardado', () => {
    const { result } = renderHook(() => useLastRouteSearch())
    expect(result.current.last).toBeNull()
  })

  it('save() persiste en localStorage y actualiza estado', () => {
    const { result } = renderHook(() => useLastRouteSearch())

    act(() => {
      result.current.save({ from: 'a', to: 'b', mode: 'quickest' })
    })

    expect(result.current.last).toEqual({ from: 'a', to: 'b', mode: 'quickest' })
    expect(JSON.parse(localStorage.getItem('onepiece.lastRouteSearch.v1') ?? 'null')).toEqual({
      from: 'a', to: 'b', mode: 'quickest',
    })
  })

  it('restaura desde localStorage al montar', () => {
    localStorage.setItem('onepiece.lastRouteSearch.v1', JSON.stringify({
      from: 'windmill-village', to: 'wano', mode: 'safest',
    }))
    const { result } = renderHook(() => useLastRouteSearch())
    expect(result.current.last).toEqual({ from: 'windmill-village', to: 'wano', mode: 'safest' })
  })

  it('ignora payload corrupto', () => {
    localStorage.setItem('onepiece.lastRouteSearch.v1', '{garbage')
    const { result } = renderHook(() => useLastRouteSearch())
    expect(result.current.last).toBeNull()
  })

  it('ignora mode inválido', () => {
    localStorage.setItem('onepiece.lastRouteSearch.v1', JSON.stringify({
      from: 'a', to: 'b', mode: 'bogus',
    }))
    const { result } = renderHook(() => useLastRouteSearch())
    expect(result.current.last).toBeNull()
  })

  it('clear() borra del storage y resetea estado', () => {
    const { result } = renderHook(() => useLastRouteSearch())
    act(() => { result.current.save({ from: 'a', to: 'b', mode: 'fastest' }) })
    act(() => { result.current.clear() })

    expect(result.current.last).toBeNull()
    expect(localStorage.getItem('onepiece.lastRouteSearch.v1')).toBeNull()
  })
})
