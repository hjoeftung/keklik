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
import WeekTab from './WeekTab'
import SummaryTab from './SummaryTab'
import styles from './TimelineScreen.module.css'

type Tab = 'today' | 'week' | 'summary'

function formatHeaderDate(): string {
  return new Date().toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' })
}

export default function TimelineScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''
  const babyName = family?.baby.name ?? 'Baby'

  const [activeTab, setActiveTab] = useState<Tab>('today')
  const [sessions, setSessions] = useState<SleepSession[]>([])
  const [stats, setStats] = useState<SleepStatsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [weekSessions, setWeekSessions] = useState<SleepSession[]>([])
  const [isLoadingWeek, setIsLoadingWeek] = useState(false)
  const [weekLoaded, setWeekLoaded] = useState(false)

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

  const loadWeek = useCallback(async () => {
    if (!babyId) return
    setIsLoadingWeek(true)
    try {
      const data = await getSleepHistory(babyId, '7d')
      setWeekSessions(data)
      setWeekLoaded(true)
    } catch {
      // show empty state in WeekTab
    } finally {
      setIsLoadingWeek(false)
    }
  }, [babyId])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (activeTab === 'week' && !weekLoaded) {
      loadWeek()
    }
  }, [activeTab, weekLoaded, loadWeek])

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <h1 className={styles.babyName}>{babyName}'s sleep</h1>
        <p className={styles.headerDate}>{formatHeaderDate()}</p>
        <div className={styles.tabs}>
          {(['today', 'week', 'summary'] as Tab[]).map(tab => (
            <button
              key={tab}
              className={`${styles.tabBtn} ${activeTab === tab ? styles.tabBtnActive : ''}`}
              onClick={() => setActiveTab(tab)}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {isLoading && <div className={styles.loading}>Loading…</div>}
      {!isLoading && error && <div className={styles.error}>{error}</div>}
      {!isLoading && !error && activeTab === 'today' && stats && (
        <TodayTab
          sessions={sessions}
          stats={stats}
          babyId={babyId}
          onRefresh={load}
        />
      )}
      {!isLoading && !error && activeTab === 'week' && (
        <WeekTab sessions={weekSessions} isLoading={isLoadingWeek} />
      )}
      {!isLoading && !error && activeTab === 'summary' && stats && (
        <SummaryTab summary={stats.summary ?? {}} />
      )}
    </div>
  )
}
