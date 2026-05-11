import { useEffect, useState } from 'react'
import type { RouteMode } from '../api/client'

const KEY = 'onepiece.lastRouteSearch.v1'

export interface LastRouteSearch {
  from: string
  to: string
  mode: RouteMode
}

const VALID_MODES: RouteMode[] = ['fastest', 'quickest', 'safest', 'riskiest']

function isValid(v: unknown): v is LastRouteSearch {
  if (!v || typeof v !== 'object') return false
  const o = v as Record<string, unknown>
  return typeof o.from === 'string'
    && typeof o.to === 'string'
    && typeof o.mode === 'string'
    && (VALID_MODES as string[]).includes(o.mode)
}

/**
 * Persiste la última búsqueda de ruta en localStorage.
 * Modo "solo restaurar": el componente decide cuándo reanudar; el hook
 * NO dispara queries automáticamente.
 */
export function useLastRouteSearch() {
  const [last, setLast] = useState<LastRouteSearch | null>(() => {
    try {
      const raw = localStorage.getItem(KEY)
      if (!raw) return null
      const parsed = JSON.parse(raw)
      return isValid(parsed) ? parsed : null
    } catch {
      return null
    }
  })

  function save(value: LastRouteSearch) {
    setLast(value)
    try { localStorage.setItem(KEY, JSON.stringify(value)) } catch { /* quota / private mode */ }
  }

  function clear() {
    setLast(null)
    try { localStorage.removeItem(KEY) } catch { /* noop */ }
  }

  // Sincroniza si otra pestaña actualiza el storage
  useEffect(() => {
    function onStorage(e: StorageEvent) {
      if (e.key !== KEY) return
      if (!e.newValue) { setLast(null); return }
      try {
        const parsed = JSON.parse(e.newValue)
        if (isValid(parsed)) setLast(parsed)
      } catch { /* ignore */ }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  return { last, save, clear }
}
