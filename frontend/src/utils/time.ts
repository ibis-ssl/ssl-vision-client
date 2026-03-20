/**
 * Formats a nanosecond timestamp as "m:ss".
 */
export function formatTimeNs(ns: number): string {
  const sec = Math.max(0, Math.floor(ns / 1e9))
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}
