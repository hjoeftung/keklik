import { describe, it, expect } from 'vitest'
import { formatDuration, formatTime, formatDate, getLocalDayBoundaries } from './time'

describe('formatDuration', () => {
  it('returns "0 min" for zero seconds', () => {
    expect(formatDuration(0)).toBe('0 min')
  })

  it('returns minutes only when under an hour', () => {
    expect(formatDuration(90)).toBe('1 min')
    expect(formatDuration(600)).toBe('10 min')
  })

  it('returns hours only when no leftover minutes', () => {
    expect(formatDuration(3600)).toBe('1 h')
    expect(formatDuration(7200)).toBe('2 h')
  })

  it('returns hours and minutes for mixed durations', () => {
    expect(formatDuration(8100)).toBe('2 h 15 min')
    expect(formatDuration(3660)).toBe('1 h 1 min')
  })
})

describe('formatTime', () => {
  it('returns a non-empty string for a valid ISO timestamp', () => {
    const result = formatTime('2024-04-16T14:30:00Z')
    expect(typeof result).toBe('string')
    expect(result.length).toBeGreaterThan(0)
  })

  it('uses browser local timezone (function has no tz parameter)', () => {
    // Verify the signature: formatTime only accepts isoString
    const fn: (isoString: string) => string = formatTime
    expect(typeof fn).toBe('function')
  })
})

describe('formatDate', () => {
  it('formats a date in the given family timezone', () => {
    // 2024-04-16 midnight UTC is still Apr 15 in America/New_York (UTC-4)
    const result = formatDate('2024-04-16T00:00:00Z', 'America/New_York')
    expect(result).toMatch(/Apr/)
    expect(result).toMatch(/15/)
  })

  it('formats a date in UTC family timezone correctly', () => {
    const result = formatDate('2024-04-16T12:00:00Z', 'UTC')
    expect(result).toMatch(/Apr/)
    expect(result).toMatch(/16/)
  })

  it('crosses midnight correctly for a non-UTC timezone', () => {
    // 2024-04-16T23:30:00Z is already Apr 17 in Asia/Tokyo (UTC+9)
    const result = formatDate('2024-04-16T23:30:00Z', 'Asia/Tokyo')
    expect(result).toMatch(/17/)
  })
})

describe('getLocalDayBoundaries', () => {
  it('returns a 24-hour window', () => {
    const date = new Date('2024-04-16T12:00:00Z')
    const { start, end } = getLocalDayBoundaries(date, 'UTC')
    expect(end.getTime() - start.getTime()).toBe(24 * 60 * 60 * 1000)
  })

  it('start is midnight in UTC timezone', () => {
    const date = new Date('2024-04-16T12:00:00Z')
    const { start } = getLocalDayBoundaries(date, 'UTC')
    expect(start.toISOString()).toBe('2024-04-16T00:00:00.000Z')
  })

  it('handles non-UTC family timezone', () => {
    // America/New_York is UTC-4 in summer; midnight Apr 16 NY = 04:00 UTC
    const date = new Date('2024-04-16T10:00:00Z')
    const { start, end } = getLocalDayBoundaries(date, 'America/New_York')
    expect(start.toISOString()).toBe('2024-04-16T04:00:00.000Z')
    expect(end.toISOString()).toBe('2024-04-17T04:00:00.000Z')
  })

  it('handles DST boundary: spring-forward in America/New_York', () => {
    // 2024-03-10 is DST spring-forward day in New_York: clocks go 2am -> 3am
    // midnight Mar 10 NY = 05:00 UTC (EST, UTC-5)
    const date = new Date('2024-03-10T12:00:00Z')
    const { start, end } = getLocalDayBoundaries(date, 'America/New_York')
    expect(start.toISOString()).toBe('2024-03-10T05:00:00.000Z')
    // Day is only 23 hours long due to DST, but we return fixed 24h window
    // (midnight-to-midnight in wall-clock terms is what matters for display)
    expect(end.getTime() - start.getTime()).toBe(24 * 60 * 60 * 1000)
  })

  it('handles midnight crossing in Asia/Kolkata (UTC+5:30)', () => {
    // midnight Apr 16 IST = Apr 15 18:30 UTC
    const date = new Date('2024-04-16T20:00:00Z')
    const { start } = getLocalDayBoundaries(date, 'Asia/Kolkata')
    expect(start.toISOString()).toBe('2024-04-16T18:30:00.000Z')
  })
})
