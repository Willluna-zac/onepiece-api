/**
 * Convierte un número de horas a un string legible.
 *
 * Reglas:
 *   - < 1h → "30 min" (redondeo a múltiplos de 5 minutos para legibilidad).
 *   - < 24h → "Xh" o "Xh Ym" si hay minutos significativos.
 *   - >= 24h → "Xd Yh" (días enteros + horas).
 *   - 0 o negativo → "0 h".
 */
export function formatHours(hours: number): string {
  if (!isFinite(hours) || hours <= 0) return '0 h'

  if (hours < 1) {
    const mins = Math.round((hours * 60) / 5) * 5
    return `${mins} min`
  }

  if (hours < 24) {
    const h = Math.floor(hours)
    const m = Math.round((hours - h) * 60)
    if (m === 0) return `${h} h`
    if (m === 60) return `${h + 1} h`
    return `${h} h ${m} min`
  }

  const days = Math.floor(hours / 24)
  const remH = Math.round(hours - days * 24)
  if (remH === 0) return `${days} d`
  if (remH === 24) return `${days + 1} d`
  return `${days} d ${remH} h`
}
