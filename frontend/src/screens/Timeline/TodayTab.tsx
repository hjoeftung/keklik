import { useState, useRef, useEffect } from 'react'
import type { SleepSession, SleepStatsResponse } from '@/api/endpoints'
import SessionDetailSheet from './SessionDetailSheet'
import styles from './TodayTab.module.css'

interface Props {
  sessions: SleepSession[]
  stats: SleepStatsResponse
  babyId: string
  onRefresh: () => void
}

const PX_PER_MIN = 1.2

function formatDur(seconds: number): string {
  if (seconds <= 0) return '0h'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
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
  showTimes: boolean
  showDuration: boolean
}

export default function TodayTab({ sessions, stats, babyId, onRefresh }: Props) {
  const [selectedSession, setSelectedSession] = useState<SleepSession | null>(null)
  const [localSessions, setLocalSessions] = useState(sessions)
  const diaryRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLocalSessions(sessions)
  }, [sessions])

  const completedSessions = localSessions
    .filter(s => s.stopped_at != null)
    .sort((a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime())

  // Derive window from sessions, padding 1h each side rounded to the hour.
  // Fall back to ±8h / +2h around now when there are no sessions.
  let windowStart: Date
  let windowEnd: Date
  if (completedSessions.length === 0) {
    const now = new Date()
    windowStart = new Date(now.getTime() - 8 * 60 * 60 * 1000)
    windowStart.setMinutes(0, 0, 0)
    windowEnd = new Date(now.getTime() + 2 * 60 * 60 * 1000)
    windowEnd.setMinutes(0, 0, 0)
    windowEnd = new Date(windowEnd.getTime() + 60 * 60 * 1000)
  } else {
    const earliest = new Date(completedSessions[0].started_at)
    const latest = new Date(completedSessions[completedSessions.length - 1].stopped_at!)
    windowStart = new Date(earliest)
    windowStart.setMinutes(0, 0, 0)
    windowStart = new Date(windowStart.getTime() - 60 * 60 * 1000)
    windowEnd = new Date(latest)
    windowEnd.setMinutes(0, 0, 0)
    windowEnd = new Date(windowEnd.getTime() + 2 * 60 * 60 * 1000)
  }
  const totalMinutes = (windowEnd.getTime() - windowStart.getTime()) / 60000
  const totalHeight = totalMinutes * PX_PER_MIN

  // Scroll to show current time or near end of first session on mount
  useEffect(() => {
    if (!diaryRef.current) return
    const now = new Date()
    const nowOffsetMin = (now.getTime() - windowStart.getTime()) / 60000
    const scrollTarget = nowOffsetMin > 0 && nowOffsetMin < totalMinutes
      ? nowOffsetMin * PX_PER_MIN - 240
      : 0
    diaryRef.current.scrollTop = Math.max(0, scrollTarget)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Hour ticks at each full local hour within the window
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

  // Awake gap pills between consecutive sessions
  const gapPills: GapPill[] = []
  for (let i = 0; i < completedSessions.length - 1; i++) {
    const curr = completedSessions[i]
    const next = completedSessions[i + 1]
    if (!curr.stopped_at) continue
    const gapStartMs = new Date(curr.stopped_at).getTime()
    const gapEndMs = new Date(next.started_at).getTime()
    const gapSec = (gapEndMs - gapStartMs) / 1000
    if (gapSec < 60) continue
    const gapStartOffsetMin = (gapStartMs - windowStart.getTime()) / 60000
    const gapEndOffsetMin = (gapEndMs - windowStart.getTime()) / 60000
    gapPills.push({
      label: `awake · ${formatDur(gapSec)}`,
      midOffsetMin: (gapStartOffsetMin + gapEndOffsetMin) / 2,
    })
  }

  // Current time indicator
  const now = new Date()
  const nowOffsetMin = (now.getTime() - windowStart.getTime()) / 60000
  const showNowLine = nowOffsetMin > 0 && nowOffsetMin < totalMinutes

  // Build session bar descriptors
  const sessionBars: SessionBar[] = completedSessions.flatMap(session => {
    const startMs = new Date(session.started_at).getTime()
    const endMs = new Date(session.stopped_at!).getTime()
    const offsetMin = (startMs - windowStart.getTime()) / 60000
    const durMin = (endMs - startMs) / 60000

    // Skip sessions entirely outside the window
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
      showTimes: height >= 60,
      showDuration: height >= 30,
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

  return (
    <div className={styles.tab}>
      {/* Summary pills */}
      <div className={styles.summaryRow}>
        <div className={styles.statPill}>
          <span className={styles.statLabel}>Total Sleep</span>
          <span className={`${styles.statValue} ${styles.statNight}`}>
            {formatDur(stats.today.total_sleep_seconds)}
          </span>
        </div>
        <div className={styles.statPill}>
          <span className={styles.statLabel}>Total Nap</span>
          <span className={`${styles.statValue} ${styles.statNap}`}>
            {formatDur(stats.today.total_nap_seconds)}
          </span>
        </div>
        <div className={styles.statPill}>
          <span className={styles.statLabel}>Active</span>
          <span className={`${styles.statValue} ${styles.statActive}`}>
            {formatDur(stats.today.total_active_seconds)}
          </span>
        </div>
      </div>

      {/* Empty state (shown in place of diary when no sessions) */}
      {completedSessions.length === 0 && (
        <div className={styles.emptyState}>
          No sleep sessions recorded today
        </div>
      )}

      {/* Diary */}
      <div className={`${styles.diaryScroll} ${completedSessions.length === 0 ? styles.diaryScrollHidden : ''}`} ref={diaryRef}>
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

          {/* Timeline column */}
          <div className={styles.timelineCol}>

            {/* Grid lines */}
            {hourTicks.map(tick => (
              <div
                key={tick.label}
                className={styles.gridLine}
                style={{ top: tick.offsetMin * PX_PER_MIN }}
              />
            ))}

            {/* Current time indicator */}
            {showNowLine && (
              <div
                className={styles.nowLine}
                style={{ top: nowOffsetMin * PX_PER_MIN }}
              />
            )}

            {/* Session bars */}
            {sessionBars.map(({ session, top, height, color, borderRadius, showTimes, showDuration }) => (
              <button
                key={session.id}
                className={styles.sessionBar}
                style={{ top, height, background: color, borderRadius }}
                onClick={() => setSelectedSession(session)}
                aria-label={`${session.classification ?? 'Sleep'} session, ${formatDur(session.duration_seconds ?? 0)}`}
              >
                {showTimes && (
                  <span className={styles.barTime}>{formatTime(session.started_at)}</span>
                )}
                {showDuration && (
                  <span className={styles.barDuration}>{formatDur(session.duration_seconds ?? 0)}</span>
                )}
                {showTimes && session.stopped_at && (
                  <span className={styles.barTime}>{formatTime(session.stopped_at)}</span>
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
