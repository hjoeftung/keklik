import { useState, useEffect, useRef, useCallback } from 'react'
import type { SleepSession } from '@/api/endpoints'
import { getLocalDayBoundaries, formatDate, formatTime, formatDuration } from '@/utils/time'
import styles from './TimelineScreen.module.css'

interface TimelineChartProps {
  sessions: SleepSession[]
  familyTimezone: string
  period: '7d' | '14d'
}

interface Block {
  session: SleepSession
  left: number  // percentage 0-100
  width: number // percentage 0-100
  isActive: boolean
}

interface DayRow {
  dayDate: Date
  dayStart: Date
  dayEnd: Date
  label: string
  blocks: Block[]
}

interface OverlayState {
  session: SleepSession
  x: number
  y: number
  elapsedSeconds: number
}

function buildDayRows(
  sessions: SleepSession[],
  period: '7d' | '14d',
  familyTimezone: string,
  now: Date,
): DayRow[] {
  const numDays = period === '7d' ? 7 : 14

  // Build day dates (newest first)
  const days: Date[] = []
  for (let i = 0; i < numDays; i++) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    days.push(d)
  }

  return days.map((dayDate) => {
    const { start: dayStart, end: dayEnd } = getLocalDayBoundaries(dayDate, familyTimezone)
    const label = formatDate(dayStart.toISOString(), familyTimezone)
    const blocks: Block[] = []

    for (const session of sessions) {
      const sessionStart = new Date(session.started_at)
      const isActive = !session.stopped_at
      const sessionEnd = isActive ? now : new Date(session.stopped_at!)

      // Skip if session doesn't overlap this day
      if (sessionEnd <= dayStart || sessionStart >= dayEnd) continue

      const blockStart = sessionStart < dayStart ? dayStart : sessionStart
      const blockEnd = sessionEnd > dayEnd ? dayEnd : sessionEnd

      const dayMs = dayEnd.getTime() - dayStart.getTime()
      const left = ((blockStart.getTime() - dayStart.getTime()) / dayMs) * 100
      const width = ((blockEnd.getTime() - blockStart.getTime()) / dayMs) * 100

      blocks.push({ session, left, width, isActive })
    }

    return { dayDate, dayStart, dayEnd, label, blocks }
  })
}

const AXIS_TICKS = [
  { pct: 0, label: '00' },
  { pct: 25, label: '06' },
  { pct: 50, label: '12' },
  { pct: 75, label: '18' },
  { pct: 100, label: '24' },
]

export default function TimelineChart({ sessions, familyTimezone, period }: TimelineChartProps) {
  const [now, setNow] = useState(() => new Date())
  const [overlay, setOverlay] = useState<OverlayState | null>(null)
  const overlayRef = useRef<HTMLDivElement>(null)

  const hasActiveSession = sessions.some((s) => !s.stopped_at)

  // Update `now` every second only when there is an active session
  useEffect(() => {
    if (!hasActiveSession) return
    const id = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(id)
  }, [hasActiveSession])

  // Dismiss overlay on Escape
  useEffect(() => {
    if (!overlay) return
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOverlay(null)
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [overlay])

  // Dismiss overlay on outside click
  const handleOutsideClick = useCallback(
    (e: React.MouseEvent) => {
      if (!overlay) return
      if (overlayRef.current && !overlayRef.current.contains(e.target as Node)) {
        setOverlay(null)
      }
    },
    [overlay],
  )

  function handleBlockClick(
    e: React.MouseEvent,
    session: SleepSession,
  ) {
    e.stopPropagation()
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
    const isActive = !session.stopped_at
    const elapsedSeconds = isActive
      ? Math.floor((now.getTime() - new Date(session.started_at).getTime()) / 1000)
      : (session.duration_seconds ?? 0)

    // Position overlay below / above the block
    let x = rect.left + rect.width / 2 - 90 // center horizontally (~180px wide)
    let y = rect.bottom + 8
    // Clamp to viewport
    if (x < 8) x = 8
    if (x + 260 > window.innerWidth) x = window.innerWidth - 268

    setOverlay({ session, x, y, elapsedSeconds })
  }

  const dayRows = buildDayRows(sessions, period, familyTimezone, now)

  return (
    <div onClick={handleOutsideClick}>
      {/* Legend */}
      <div className={styles.legend}>
        <span className={styles.legendItem}>
          <span className={styles.legendSwatch} style={{ background: '#3b82f6' }} />
          Night sleep
        </span>
        <span className={styles.legendItem}>
          <span className={styles.legendSwatch} style={{ background: '#f59e0b' }} />
          Nap
        </span>
      </div>

      {/* Axis ticks row */}
      <div className={styles.axisRow} style={{ height: 18 }}>
        {AXIS_TICKS.map(({ pct, label }) => (
          <span
            key={pct}
            className={styles.axisTick}
            style={{ left: `${pct}%` }}
          >
            {label}
          </span>
        ))}
      </div>

      {/* Day rows */}
      {dayRows.map((row) => (
        <div key={row.dayStart.toISOString()} className={styles.dayRow}>
          <span className={styles.dayLabel}>{row.label}</span>
          <div className={styles.dayTrack}>
            {row.blocks.map((block) => {
              const isNight = block.session.classification === 'night'
              const blockClass = [
                styles.block,
                isNight ? styles.blockNight : styles.blockNap,
                block.isActive ? styles.blockActive : '',
              ]
                .filter(Boolean)
                .join(' ')
              return (
                <div
                  key={`${block.session.id}-${row.dayStart.toISOString()}`}
                  className={blockClass}
                  style={{
                    left: `${block.left}%`,
                    width: `${Math.max(block.width, 0.5)}%`,
                  }}
                  onClick={(e) => handleBlockClick(e, block.session)}
                  title={isNight ? 'Night sleep' : 'Nap'}
                />
              )
            })}
          </div>
        </div>
      ))}

      {/* Session detail overlay */}
      {overlay && (
        <div
          ref={overlayRef}
          className={styles.overlay}
          style={{ left: overlay.x, top: overlay.y }}
          onClick={(e) => e.stopPropagation()}
        >
          <div className={styles.overlayRow}>
            <span className={styles.overlayLabel}>Start</span>
            <span className={styles.overlayValue}>
              {formatTime(overlay.session.started_at)}
            </span>
          </div>
          <div className={styles.overlayRow}>
            <span className={styles.overlayLabel}>End</span>
            <span className={styles.overlayValue}>
              {overlay.session.stopped_at
                ? formatTime(overlay.session.stopped_at)
                : 'ongoing'}
            </span>
          </div>
          <div className={styles.overlayRow}>
            <span className={styles.overlayLabel}>Duration</span>
            <span className={styles.overlayValue}>
              {formatDuration(
                overlay.session.stopped_at
                  ? (overlay.session.duration_seconds ?? overlay.elapsedSeconds)
                  : overlay.elapsedSeconds,
              )}
            </span>
          </div>
          <div className={styles.overlayRow}>
            <span className={styles.overlayLabel}>Type</span>
            <span className={styles.overlayValue}>
              {overlay.session.classification === 'night' ? 'Night sleep' : 'Nap'}
            </span>
          </div>
        </div>
      )}
    </div>
  )
}
