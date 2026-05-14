import { useState, useEffect, useRef } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { useAppData } from '@/context/AppDataContext'
import type { NightWindowInfo } from '@/api/endpoints'
import TodayTab from './TodayTab'
import WeekTab from './WeekTab'
import SummaryTab from './SummaryTab'
import { computeDayWindow } from './utils'
import styles from './StatisticsScreen.module.css'

type Tab = 'today' | 'week' | 'summary'

const PULL_THRESHOLD = 72

function formatSelectedDate(d: Date): string {
  return d.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' })
}

function sessionsForDate(
  sessions: ReturnType<typeof useAppData>['sessions7d'],
  date: Date,
  nightWindow?: NightWindowInfo,
) {
  const { windowStart, windowEnd } = computeDayWindow(date, nightWindow)
  return sessions.filter((s) => {
    if (!s.stopped_at) return false
    return new Date(s.started_at) < windowEnd && new Date(s.stopped_at) >= windowStart
  })
}

export default function StatsScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''
  const babyName = family?.baby.name ?? 'Baby'
  const { sessions7d, stats, isLoading, error, refresh } = useAppData()

  const [activeTab, setActiveTab] = useState<Tab>('today')
  const [pullDisplay, setPullDisplay] = useState(0)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [selectedDate, setSelectedDate] = useState<Date>(() => {
    const d = new Date()
    return new Date(d.getFullYear(), d.getMonth(), d.getDate())
  })

  const screenRef = useRef<HTMLDivElement>(null)
  const refreshRef = useRef(refresh)
  const pullDeltaRef = useRef(0)

  useEffect(() => {
    refreshRef.current = refresh
  }, [refresh])

  const filteredSessions = sessionsForDate(sessions7d, selectedDate, stats?.night_window)

  useEffect(() => {
    if (!stats?.days.length) return
    const selectedDateKey = `${selectedDate.getFullYear()}-${String(selectedDate.getMonth() + 1).padStart(2, '0')}-${String(selectedDate.getDate()).padStart(2, '0')}`
    if (stats.days.some((day) => day.date === selectedDateKey)) return
    const [year, month, day] = stats.days[0].date.split('-').map(Number)
    setSelectedDate(new Date(year, month - 1, day))
  }, [selectedDate, stats?.days])

  useEffect(() => {
    const el = screenRef.current
    if (!el) return
    let startY = 0
    let startedInScrollable = false

    function onStart(e: TouchEvent) {
      startY = e.touches[0].clientY
      // If the touch started inside a scrollable child, don't hijack the gesture
      let node = e.target as Element | null
      startedInScrollable = false
      while (node && node !== el) {
        if (node.scrollHeight > node.clientHeight) {
          startedInScrollable = true
          break
        }
        node = node.parentElement
      }
    }

    function onMove(e: TouchEvent) {
      if (!el || startedInScrollable) return
      const delta = e.touches[0].clientY - startY
      if (delta > 0 && el.scrollTop <= 0) {
        pullDeltaRef.current = Math.min(PULL_THRESHOLD * 1.5, delta * 0.4)
        setPullDisplay(pullDeltaRef.current)
        e.preventDefault()
      }
    }

    async function onEnd() {
      const d = pullDeltaRef.current
      pullDeltaRef.current = 0
      setPullDisplay(0)
      if (d >= PULL_THRESHOLD) {
        setIsRefreshing(true)
        await refreshRef.current()
        setIsRefreshing(false)
      }
    }

    el.addEventListener('touchstart', onStart, { passive: true })
    el.addEventListener('touchmove', onMove, { passive: false })
    el.addEventListener('touchend', onEnd, { passive: true })
    return () => {
      el.removeEventListener('touchstart', onStart)
      el.removeEventListener('touchmove', onMove)
      el.removeEventListener('touchend', onEnd)
    }
  }, [])

  return (
    <div className={styles.screen} ref={screenRef}>
      {pullDisplay > 8 && (
        <div className={styles.pullIndicator} style={{ height: pullDisplay }}>
          <div
            className={`${styles.pullSpinner} ${pullDisplay >= PULL_THRESHOLD || isRefreshing ? styles.pullSpinnerActive : ''}`}
          />
        </div>
      )}
      <div className={styles.header}>
        <h1 className={styles.babyName}>{babyName}'s sleep</h1>
        <p className={styles.headerDate}>
          {formatSelectedDate(activeTab === 'today' ? selectedDate : new Date())}
        </p>
        <div className={styles.tabs}>
          {(['today', 'week', 'summary'] as Tab[]).map((tab) => (
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
      {!isLoading && error && <div className={styles.error}>{error}</div>}
      {!error && activeTab === 'today' && (
        <TodayTab
          sessions={filteredSessions}
          stats={stats}
          babyId={babyId}
          onRefresh={refresh}
          isLoading={isLoading}
          selectedDate={selectedDate}
          onDateChange={setSelectedDate}
        />
      )}
      {!error && activeTab === 'week' && (
        <WeekTab
          sessions={sessions7d}
          nightWindow={stats?.night_window}
          isLoading={isLoading}
          babyId={babyId}
          onRefresh={refresh}
        />
      )}
      {!error && activeTab === 'summary' && (
        <SummaryTab summary={stats?.summary ?? {}} isLoading={isLoading} />
      )}
    </div>
  )
}
