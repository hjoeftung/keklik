import { useState, useRef, useEffect } from 'react'
import type { SleepSession, SleepStatsResponse } from '@/api/endpoints'
import SessionDetailSheet from './SessionDetailSheet'
import DatePickerStrip from './DatePickerStrip'
import styles from './TodayTab.module.css'

interface Props {
  sessions: SleepSession[]
  stats: SleepStatsResponse | null
  babyId: string
  onRefresh: () => void
  isLoading: boolean
  selectedDate: Date
  onDateChange: (d: Date) => void
}

const PX_PER_MIN = 0.9

function formatDur(seconds: number): string {
  if (seconds <= 0) return '0h'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

interface HourTick {
  label: string
  offsetMin: number
}

interface GapPill {
  label: string
  midOffsetMin: number
}

interface SessionBar {
  session: SleepSession
  top: number
  height: number
  color: string
  borderRadius: number
  showLabel: boolean
  showDuration: boolean
}

function isToday(d: Date): boolean {
  const now = new Date()
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate()
}

export default function TodayTab({ sessions, stats, babyId, onRefresh, isLoading, selectedDate, onDateChange }: Props) {
  const selectedIsToday = isToday(selectedDate)

  const [selectedSession, setSelectedSession] = useState<SleepSession | null>(null)
  const [localSessions, setLocalSessions] = useState(sessions)
  const diaryRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLocalSessions(sessions)
  }, [sessions])

  const completedSessions = localSessions
    .filter(s => s.stopped_at != null)
    .sort((a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime())

  let windowStart = new Date(selectedDate)
  windowStart.setHours(7, 0, 0, 0)
  let windowEnd = new Date(windowStart.getTime() + 24 * 60 * 60 * 1000)
  if (stats?.night_window) {
    const startHour = parseInt(stats.night_window.end_hhmm.split(':')[0], 10)
    const viewStartHour = (startHour - 1 + 24) % 24
    windowStart = new Date(selectedDate)
    windowStart.setHours(viewStartHour, 0, 0, 0)
    windowEnd = new Date(windowStart.getTime() + 24 * 60 * 60 * 1000)
  }
  const totalMinutes = (windowEnd.getTime() - windowStart.getTime()) / 60000
  const totalHeight = totalMinutes * PX_PER_MIN

  useEffect(() => {
    if (!diaryRef.current || isLoading) return
    const now = new Date()
    const nowOffsetMin = (now.getTime() - windowStart.getTime()) / 60000
    const scrollTarget = nowOffsetMin > 0 && nowOffsetMin < totalMinutes
      ? nowOffsetMin * PX_PER_MIN - 240
      : 0
    diaryRef.current.scrollTop = Math.max(0, scrollTarget)
  }, [isLoading]) // eslint-disable-line react-hooks/exhaustive-deps

  const hourTicks: HourTick[] = []
  {
    let t = new Date(windowStart)
    t.setMinutes(0, 0, 0)
    t = new Date(t.getTime() + 3_600_000)
    while (t <= windowEnd) {
      const offsetMin = (t.getTime() - windowStart.getTime()) / 60000
      if (offsetMin >= 0 && offsetMin <= totalMinutes) {
        hourTicks.push({
          label: `${String(t.getHours()).padStart(2, '0')}:00`,
          offsetMin,
        })
      }
      t = new Date(t.getTime() + 3_600_000)
    }
  }

  const gapPills: GapPill[] = []
  for (let i = 0; i < completedSessions.length - 1; i++) {
    const curr = completedSessions[i]
    const next = completedSessions[i + 1]
    if (!curr.stopped_at) continue
    const gapStartMs = new Date(curr.stopped_at).getTime()
    const gapEndMs = new Date(next.started_at).getTime()
    const gapSec = (gapEndMs - gapStartMs) / 1000
    if (gapSec < 3600) continue
    const gapStartOffsetMin = (gapStartMs - windowStart.getTime()) / 60000
    const gapEndOffsetMin = (gapEndMs - windowStart.getTime()) / 60000
    gapPills.push({
      label: `✦ awake · ${formatDur(gapSec)}`,
      midOffsetMin: (gapStartOffsetMin + gapEndOffsetMin) / 2,
    })
  }

  const now = new Date()
  const nowOffsetMin = (now.getTime() - windowStart.getTime()) / 60000
  const showNowLine = nowOffsetMin > 0 && nowOffsetMin < totalMinutes

  let nightWindowStartOffsetMin: number | null = null
  if (stats?.night_window) {
    const [swH, swM] = stats.night_window.start_hhmm.split(':').map(Number)
    const startDate = new Date(selectedDate)
    startDate.setHours(swH, swM, 0, 0)
    const offset = (startDate.getTime() - windowStart.getTime()) / 60000
    if (offset > 0 && offset < totalMinutes) nightWindowStartOffsetMin = offset
  }

  const sessionBars: SessionBar[] = completedSessions.flatMap(session => {
    const startMs = new Date(session.started_at).getTime()
    const endMs = new Date(session.stopped_at!).getTime()
    const offsetMin = (startMs - windowStart.getTime()) / 60000
    const durMin = (endMs - startMs) / 60000

    if (offsetMin + durMin < 0 || offsetMin > totalMinutes) return []

    const clampedOffsetMin = Math.max(0, offsetMin)
    const clampedDurMin = Math.min(durMin, totalMinutes - clampedOffsetMin) - Math.max(0, -offsetMin)
    const top = clampedOffsetMin * PX_PER_MIN
    const height = Math.max(6, clampedDurMin * PX_PER_MIN)
    const isNight = session.classification === 'night'
    const color = isNight ? '#5B7BB8' : '#E8B86E'
    const borderRadius = Math.min(12, Math.max(3, height * 0.18))

    return [{
      session,
      top,
      height,
      color,
      borderRadius,
      showLabel: height >= 22,
      showDuration: height >= 22,
    }]
  })

  function handleSessionUpdated(updated: SleepSession) {
    setLocalSessions(prev => prev.map(s => s.id === updated.id ? updated : s))
    setSelectedSession(null)
    onRefresh()
  }

  function handleSessionDeleted() {
    if (selectedSession) {
      setLocalSessions(prev => prev.filter(s => s.id !== selectedSession.id))
    }
    setSelectedSession(null)
    onRefresh()
  }

  if (isLoading) {
    return (
      <div className={styles.tab}>
        <div className={styles.summaryRow}>
          <div className={`${styles.statPill} ${styles.skeleton}`} style={{ height: 64 }} />
          <div className={`${styles.statPill} ${styles.skeleton}`} style={{ height: 64 }} />
          <div className={`${styles.statPill} ${styles.skeleton}`} style={{ height: 64 }} />
        </div>
        <div className={styles.skeleton} style={{ flex: 1, margin: '0 16px 24px', borderRadius: 16 }} />
      </div>
    )
  }

  const totalSleepSec = selectedIsToday
    ? (stats?.today.total_sleep_seconds ?? 0)
    : localSessions.filter(s => s.classification === 'night').reduce((a, s) => a + (s.duration_seconds ?? 0), 0)
  const totalNapSec = selectedIsToday
    ? (stats?.today.total_nap_seconds ?? 0)
    : localSessions.filter(s => s.classification === 'nap').reduce((a, s) => a + (s.duration_seconds ?? 0), 0)
  const totalActiveSec = selectedIsToday
    ? (stats?.today.total_active_seconds ?? 0)
    : localSessions.filter(s => s.classification === 'active').reduce((a, s) => a + (s.duration_seconds ?? 0), 0)

  return (
    <div className={styles.tab}>
      <DatePickerStrip selectedDate={selectedDate} onChange={onDateChange} />

      {/* Summary pills */}
      <div className={styles.summaryRow}>
        <div className={`${styles.statPill} ${styles.statPillNight}`}>
          <span className={styles.statLabel}>Sleep</span>
          <span className={`${styles.statValue} ${styles.statNight}`}>
            {formatDur(totalSleepSec)}
          </span>
        </div>
        <div className={`${styles.statPill} ${styles.statPillNap}`}>
          <span className={styles.statLabel}>Naps</span>
          <span className={`${styles.statValue} ${styles.statNap}`}>
            {formatDur(totalNapSec)}
          </span>
        </div>
        <div className={`${styles.statPill} ${styles.statPillActive}`}>
          <span className={styles.statLabel}>Active</span>
          <span className={`${styles.statValue} ${styles.statActive}`}>
            {formatDur(totalActiveSec)}
          </span>
        </div>
      </div>

      {/* Empty state */}
      {completedSessions.length === 0 && (
        <div className={styles.emptyState}>
          No sleep sessions recorded
        </div>
      )}

      {/* Diary */}
      <div
        className={`${styles.diaryScroll} ${completedSessions.length === 0 ? styles.diaryScrollHidden : ''}`}
        ref={diaryRef}
        data-scrollable
      >
        <div className={styles.diaryInner} style={{ height: totalHeight }}>

          {/* Hour label column */}
          <div className={styles.hourLabels}>
            {hourTicks.map(tick => (
              <div
                key={tick.label}
                className={styles.hourLabel}
                style={{ top: tick.offsetMin * PX_PER_MIN }}
              >
                {tick.label}
              </div>
            ))}
          </div>

          <div className={styles.sessionCol}>

            {/* Grid lines */}
            {hourTicks.map(tick => (
              <div
                key={tick.label}
                className={styles.gridLine}
                style={{ top: tick.offsetMin * PX_PER_MIN }}
              />
            ))}

            {/* Night window end marker */}
            {nightWindowStartOffsetMin !== null && (
              <div
                className={styles.nightWindowLine}
                style={{ top: nightWindowStartOffsetMin * PX_PER_MIN }}
              />
            )}

            {/* Current time indicator */}
            {showNowLine && (
              <div
                className={styles.nowLine}
                style={{ top: nowOffsetMin * PX_PER_MIN }}
              />
            )}

            {/* Session bars */}
            {sessionBars.map(({ session, top, height, color, borderRadius, showLabel, showDuration }) => (
              <button
                key={session.id}
                className={styles.sessionBar}
                style={{ top, height, background: color, borderRadius }}
                onClick={() => setSelectedSession(session)}
                aria-label={`${session.classification === 'night' ? 'Night sleep' : 'Nap'}, ${formatDur(session.duration_seconds ?? 0)}`}
              >
                {(showLabel || showDuration) && (
                  <div className={styles.barContent}>
                    {showLabel && (
                      <span className={styles.barLabel}>
                        {session.classification === 'night' ? 'Night sleep' : 'Nap'}
                      </span>
                    )}
                    {showDuration && (
                      <span className={styles.barDuration}>{formatDur(session.duration_seconds ?? 0)}</span>
                    )}
                  </div>
                )}
              </button>
            ))}

            {/* Awake gap pills */}
            {gapPills.map((pill, i) => (
              <div
                key={i}
                className={styles.awakePill}
                style={{ top: pill.midOffsetMin * PX_PER_MIN }}
              >
                {pill.label}
              </div>
            ))}

          </div>
        </div>
      </div>

      {/* Session detail sheet */}
      {selectedSession && (
        <SessionDetailSheet
          session={selectedSession}
          babyId={babyId}
          onClose={() => setSelectedSession(null)}
          onUpdated={handleSessionUpdated}
          onDeleted={handleSessionDeleted}
        />
      )}
    </div>
  )
}
