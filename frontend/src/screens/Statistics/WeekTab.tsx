import type { SleepSession, NightWindowInfo } from '@/api/endpoints'
import styles from './WeekTab.module.css'

interface Props {
  sessions: SleepSession[]
  nightWindow?: NightWindowInfo
  isLoading: boolean
}

const COL_H = 460     // px

const NIGHT_COLOR = '#5B7BB8'
const NAP_COLOR = '#E8B86E'

const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

function dateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function getLast7Days(): Date[] {
  const days: Date[] = []
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  for (let i = 6; i >= 0; i--) {
    days.push(new Date(today.getTime() - i * 24 * 3600 * 1000))
  }
  return days
}

interface Block {
  top: number
  height: number
  color: string
  borderRadius: number
}

function buildBlocks(
  daySessions: SleepSession[],
  dayMidnight: Date,
  winStart: number,
  winEnd: number,
  yFor: (h: number) => number,
): Block[] {
  return daySessions.flatMap(s => {
    if (!s.stopped_at) return []

    const startMs = new Date(s.started_at).getTime()
    const endMs = new Date(s.stopped_at).getTime()
    const midMs = dayMidnight.getTime()

    const startH = (startMs - midMs) / 3_600_000
    const endH = (endMs - midMs) / 3_600_000

    const clampedStart = Math.max(winStart, startH)
    const clampedEnd = Math.min(winEnd, endH)
    if (clampedEnd <= clampedStart) return []

    const top = yFor(clampedStart)
    const height = Math.max(2, yFor(clampedEnd) - top)
    const dur = endH - startH
    const isNight = s.classification === 'night'

    return [{
      top,
      height,
      color: isNight ? NIGHT_COLOR : NAP_COLOR,
      borderRadius: dur < 0.6 ? 1.5 : dur < 1.2 ? 2.5 : 4,
    }]
  })
}

export default function WeekTab({ sessions, nightWindow, isLoading }: Props) {
  const days = getLast7Days()

  const nightStartHour = nightWindow ? parseInt(nightWindow.end_hhmm.split(':')[0], 10) : 7
  const winStart = (nightStartHour - 1 + 24) % 24
  const winEnd = winStart + 24
  const hourTicks = [winStart, winStart + 6, winStart + 12, winStart + 18]
  const yFor = (h: number) => ((h - winStart) / (winEnd - winStart)) * COL_H

  if (isLoading) {
    return (
      <div className={styles.tab}>
        <div className={styles.legend}>
          <div className={styles.skeleton} style={{ width: 60, height: 14, borderRadius: 999 }} />
          <div className={styles.skeleton} style={{ width: 60, height: 14, borderRadius: 999 }} />
        </div>
        <div className={styles.headerRow}>
          <div className={styles.axisGutter} />
          {Array.from({ length: 7 }).map((_, i) => (
            <div key={i} className={`${styles.dayHeader} ${styles.skeleton}`} style={{ height: 28, borderRadius: 6 }} />
          ))}
        </div>
        <div className={styles.gridScroll} data-scrollable>
          <div className={styles.gridInner} style={{ height: COL_H }}>
            <div className={styles.axis} />
            <div className={styles.columnsArea}>
              {Array.from({ length: 7 }).map((_, i) => (
                <div key={i} className={styles.skeleton} style={{ flex: 1, height: COL_H, borderRadius: 8 }} />
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Build blocks per day
  const renderedBlocks = new Map<string, Block[]>()
  days.forEach(day => {
    const key = dateKey(day)
    const daySessions = sessions.filter(s => {
      if (!s.stopped_at || s.classification === 'active') return false
      return dateKey(new Date(s.started_at)) === key
    })
    renderedBlocks.set(key, buildBlocks(daySessions, day, winStart, winEnd, yFor))
  })

  return (
    <div className={styles.tab}>
      {/* Legend */}
      <div className={styles.legend}>
        <span className={styles.legendItem}>
          <span className={styles.swatch} style={{ background: NIGHT_COLOR }} />
          Night
        </span>
        <span className={styles.legendItem}>
          <span className={styles.swatch} style={{ background: NAP_COLOR }} />
          Nap
        </span>
      </div>

      {/* Column headers */}
      <div className={styles.headerRow}>
        <div className={styles.axisGutter} />
        {days.map(d => (
          <div key={dateKey(d)} className={styles.dayHeader}>
            <span className={styles.dayName}>{DAY_NAMES[d.getDay()]}</span>
            <span className={styles.dayNum}>{d.getDate()}</span>
          </div>
        ))}
      </div>

      {/* Grid */}
      <div className={styles.gridScroll} data-scrollable>
        <div className={styles.gridInner} style={{ height: COL_H }}>
          {/* Left axis */}
          <div className={styles.axis}>
            {hourTicks.map(h => (
              <div
                key={h}
                className={styles.axisTick}
                style={{ top: yFor(h) }}
              >
                {String(h % 24).padStart(2, '0')}
              </div>
            ))}
          </div>

          {/* Columns area */}
          <div className={styles.columnsArea}>
            {/* Hour grid lines */}
            {hourTicks.map(h => (
              <div
                key={h}
                className={styles.gridLine}
                style={{ top: yFor(h) }}
              />
            ))}

            {/* Day columns */}
            {days.map(d => {
              const key = dateKey(d)
              const blocks = renderedBlocks.get(key) ?? []
              return (
                <div key={key} className={styles.dayCol}>
                  {blocks.map((b, i) => (
                    <div
                      key={i}
                      className={styles.block}
                      style={{
                        top: b.top + 1,
                        height: b.height - 2,
                        background: b.color,
                        borderRadius: b.borderRadius,
                      }}
                    />
                  ))}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
