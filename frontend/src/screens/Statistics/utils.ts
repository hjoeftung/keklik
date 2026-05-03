import type { NightWindowInfo } from '@/api/endpoints'

export function computeDayWindow(
  date: Date,
  nightWindow?: NightWindowInfo,
): { windowStart: Date; windowEnd: Date } {
  const endHour = nightWindow ? parseInt(nightWindow.end_hhmm.split(':')[0], 10) : 7
  const viewStartHour = (endHour - 1 + 24) % 24
  const windowStart = new Date(date)
  windowStart.setHours(viewStartHour, 0, 0, 0)
  const windowEnd = new Date(windowStart.getTime() + 24 * 60 * 60 * 1000)
  return { windowStart, windowEnd }
}
