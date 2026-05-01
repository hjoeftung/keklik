import { useState } from 'react'
import type { PeriodAverage } from '@/api/endpoints'
import styles from './SummaryTab.module.css'

type Period = '7d' | '14d' | '30d' | '90d'
const PERIODS: Period[] = ['7d', '14d', '30d', '90d']

interface Props {
  summary: Record<string, PeriodAverage>
}

function formatDur(seconds: number): string {
  if (seconds <= 0) return '0h'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

function formatDateRange(period: Period): string {
  const days = parseInt(period)
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const end = new Date(today.getTime() - 86_400_000)
  const start = new Date(today.getTime() - days * 86_400_000)
  const fmtLong = (d: Date) =>
    d.toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' })
  const fmtShort = (d: Date) =>
    d.toLocaleDateString('en-GB', { day: 'numeric', month: 'long' })
  if (start.getFullYear() === end.getFullYear()) {
    return `${fmtShort(start)} – ${fmtLong(end)}`
  }
  return `${fmtLong(start)} – ${fmtLong(end)}`
}

function MoonIcon() {
  return (
    <svg width="30" height="30" viewBox="0 0 30 30" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path
        d="M22 16.5C16.2 16.5 11.5 11.8 11.5 6c0-1.4.25-2.74.7-4C6.95 3.36 3 8.23 3 14c0 7.18 5.82 13 13 13 5.77 0 10.64-3.95 12-9.2-1.26.45-2.6.7-4 .7z"
        fill="white"
      />
    </svg>
  )
}

function SunIcon() {
  return (
    <svg width="30" height="30" viewBox="0 0 30 30" fill="white" xmlns="http://www.w3.org/2000/svg">
      <circle cx="15" cy="15" r="6" />
      <g stroke="white" strokeWidth="2.2" strokeLinecap="round">
        <line x1="15" y1="4" x2="15" y2="7" />
        <line x1="15" y1="23" x2="15" y2="26" />
        <line x1="4" y1="15" x2="7" y2="15" />
        <line x1="23" y1="15" x2="26" y2="15" />
        <line x1="7.05" y1="7.05" x2="9.17" y2="9.17" />
        <line x1="20.83" y1="20.83" x2="22.95" y2="22.95" />
        <line x1="7.05" y1="22.95" x2="9.17" y2="20.83" />
        <line x1="20.83" y1="9.17" x2="22.95" y2="7.05" />
      </g>
    </svg>
  )
}

function SparkIcon() {
  return (
    <svg width="30" height="30" viewBox="0 0 30 30" fill="white" xmlns="http://www.w3.org/2000/svg">
      <path d="M15 5l2.5 7.5L25 15l-7.5 2.5L15 25l-2.5-7.5L5 15l7.5-2.5z" />
    </svg>
  )
}

export default function SummaryTab({ summary }: Props) {
  const [period, setPeriod] = useState<Period>('7d')
  const data = summary[period]

  const rows: Array<{
    label: string
    sub: string
    value: number | undefined
    cardClass: string
    circleClass: string
    Icon: () => JSX.Element
  }> = [
    {
      label: 'Avg sleep',
      sub: 'per night',
      value: data?.avg_sleep_seconds,
      cardClass: styles.cardNight,
      circleClass: styles.iconCircleNight,
      Icon: MoonIcon,
    },
    {
      label: 'Avg nap',
      sub: 'per day',
      value: data?.avg_nap_seconds,
      cardClass: styles.cardNap,
      circleClass: styles.iconCircleNap,
      Icon: SunIcon,
    },
    {
      label: 'Avg active',
      sub: 'per day',
      value: data?.avg_active_seconds,
      cardClass: styles.cardActive,
      circleClass: styles.iconCircleActive,
      Icon: SparkIcon,
    },
  ]

  return (
    <div className={styles.tab}>
      <div className={styles.periodControl}>
        {PERIODS.map(p => (
          <button
            key={p}
            className={`${styles.periodBtn} ${p === period ? styles.periodBtnActive : ''}`}
            onClick={() => setPeriod(p)}
          >
            {p}
          </button>
        ))}
      </div>

      <div className={styles.dateRange}>{formatDateRange(period)}</div>

      <div className={styles.cards}>
        {rows.map(({ label, sub, value, cardClass, circleClass, Icon }) => (
          <div key={label} className={`${styles.card} ${cardClass}`}>
            <div className={`${styles.iconCircle} ${circleClass}`}>
              <Icon />
            </div>
            <div className={styles.cardBody}>
              <div className={styles.cardLabel}>{label}</div>
              <div className={styles.cardValue}>
                {value != null ? formatDur(value) : '—'}
              </div>
              <div className={styles.cardSub}>{sub}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
