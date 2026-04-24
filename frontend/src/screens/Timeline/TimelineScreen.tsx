import { useState, useEffect } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { getSleepHistory, type SleepSession, type SleepHistoryPeriod } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import TimelineChart from './TimelineChart'
import styles from './TimelineScreen.module.css'

const PERIODS: { value: SleepHistoryPeriod; label: string }[] = [
  { value: '7d', label: '7 days' },
  { value: '14d', label: '14 days' },
]

export default function TimelineScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [period, setPeriod] = useState<SleepHistoryPeriod>('7d')
  const [sessions, setSessions] = useState<SleepSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const familyTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone

  useEffect(() => {
    if (!babyId) return
    setIsLoading(true)
    setError(null)
    getSleepHistory(babyId, period)
      .then((data) => setSessions(data))
      .catch((err) =>
        setError(err instanceof ApiError ? err.message : 'Failed to load sleep history'),
      )
      .finally(() => setIsLoading(false))
  }, [babyId, period])

  return (
    <div className={styles.screen}>
      {/* Period toggle */}
      <div className={styles.periodToggle}>
        {PERIODS.map(({ value, label }) => (
          <button
            key={value}
            className={`${styles.periodButton} ${period === value ? styles.periodButtonActive : ''}`}
            onClick={() => setPeriod(value)}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Chart area */}
      <div className={styles.chartContainer}>
        {isLoading && <div className={styles.loading}>Loading…</div>}

        {!isLoading && error && <div className={styles.error}>{error}</div>}

        {!isLoading && !error && sessions.length === 0 && (
          <div className={styles.empty}>No sleep sessions recorded yet.</div>
        )}

        {!isLoading && !error && sessions.length > 0 && (
          <TimelineChart
            sessions={sessions}
            familyTimezone={familyTimezone}
            period={period === '7d' ? '7d' : '14d'}
          />
        )}
      </div>
    </div>
  )
}
