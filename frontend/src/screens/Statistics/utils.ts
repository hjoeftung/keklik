import type { NightWindowInfo } from '@/api/endpoints'

export function computeDayWindow(
  date: Date,
  nightWindow?: NightWindowInfo,
): { windowStart: Date; windowEnd: Date } {
  const [endHour, endMinute] = nightWindow
    ? nightWindow.end_hhmm.split(':').map(Number)
    : [7, 0]
  const windowStart = new Date(date)
  windowStart.setHours(endHour, endMinute, 0, 0)
  const windowEnd = new Date(windowStart.getTime() + 24 * 60 * 60 * 1000)
  return { windowStart, windowEnd }
}
