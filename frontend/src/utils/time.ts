// hhmm is "HH:MM" as returned by the backend (e.g. "22:00")
export function hhmmToDisplay(hhmm: string, use24h = false): string {
  const [h, m] = hhmm.split(':').map(Number)
  if (use24h) {
    return m === 0
      ? `${String(h).padStart(2, '0')}:00`
      : `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
  }
  const period = h < 12 ? 'AM' : 'PM'
  const hour12 = h % 12 === 0 ? 12 : h % 12
  return m === 0 ? `${hour12} ${period}` : `${hour12}:${String(m).padStart(2, '0')} ${period}`
}

export function formatDuration(seconds: number): string {
  if (seconds === 0) return '0 min'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h === 0) return `${m} min`
  if (m === 0) return `${h} h`
  return `${h} h ${m} min`
}

export function formatTime(isoString: string, use24h = false): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: !use24h,
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  }).format(new Date(isoString))
}

export function formatDate(isoString: string, familyTimezone: string): string {
  return new Intl.DateTimeFormat(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    timeZone: familyTimezone,
  }).format(new Date(isoString))
}

export function getLocalDayBoundaries(
  date: Date,
  familyTimezone: string,
): { start: Date; end: Date } {
  const fmt = new Intl.DateTimeFormat('en-CA', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: familyTimezone,
  })
  const localDateStr = fmt.format(date) // "YYYY-MM-DD"
  // Interpret that local midnight in the family timezone
  const startUtc = zonedMidnightToUtc(localDateStr, '00:00', familyTimezone)
  const endUtc = new Date(startUtc.getTime() + 24 * 60 * 60 * 1000)
  return { start: startUtc, end: endUtc }
}

function zonedMidnightToUtc(dateStr: string, time: string, tz: string): Date {
  // Use Intl to find the UTC instant that corresponds to midnight in `tz`.
  // Strategy: binary-search is overkill; instead construct a UTC candidate and
  // check what local date it produces, adjusting by the observed offset.
  const candidate = new Date(`${dateStr}T${time}:00Z`)
  // Get what local date the candidate maps to in the family timezone
  const localParts = new Intl.DateTimeFormat('en-CA', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
    timeZone: tz,
  }).formatToParts(candidate)

  const p = Object.fromEntries(localParts.map(({ type, value }) => [type, value]))
  const localCandidate = new Date(
    `${p.year}-${p.month}-${p.day}T${p.hour}:${p.minute}:${p.second}Z`,
  )
  // offset in ms between UTC and local representation
  const offsetMs = localCandidate.getTime() - candidate.getTime()
  return new Date(candidate.getTime() - offsetMs)
}
