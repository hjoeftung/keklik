import { useState, useEffect, useCallback } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import {
  getSleepHistory,
  getSleepStats,
  type SleepSession,
  type SleepStatsResponse,
} from '@/api/endpoints'
import { ApiError } from '@/api/client'
import TodayTab from './TodayTab'
import styles from './TimelineScreen.module.css'

export default function TimelineScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [sessions, setSessions] = useState<SleepSession[]>([])
  const [stats, setStats] = useState<SleepStatsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!babyId) return
    setIsLoading(true)
    setError(null)
    try {
      const [sessionData, statsData] = await Promise.all([
        getSleepHistory(babyId, 'today'),
        getSleepStats(babyId),
      ])
      setSessions(sessionData)
      setStats(statsData)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep data')
    } finally {
      setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className={styles.screen}>
      {isLoading && <div className={styles.loading}>Loading…</div>}
      {!isLoading && error && <div className={styles.error}>{error}</div>}
      {!isLoading && !error && stats && (
        <TodayTab
          sessions={sessions}
          stats={stats}
          babyId={babyId}
          onRefresh={load}
        />
      )}
    </div>
  )
}
