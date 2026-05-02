import { useState, useEffect, useRef } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { useAppData } from '@/context/AppDataContext'
import TodayTab from './TodayTab'
import WeekTab from './WeekTab'
import SummaryTab from './SummaryTab'
import styles from './StatisticsScreen.module.css'

type Tab = 'today' | 'week' | 'summary'

const PULL_THRESHOLD = 72

function formatHeaderDate(): string {
  return new Date().toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' })
}

function localDateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export default function StatsScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''
  const babyName = family?.baby.name ?? 'Baby'
  const { sessions7d, stats, isLoading, error, refresh } = useAppData()

  const [activeTab, setActiveTab] = useState<Tab>('today')
  const [pullDisplay, setPullDisplay] = useState(0)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const screenRef = useRef<HTMLDivElement>(null)
  const refreshRef = useRef(refresh)
  const pullDeltaRef = useRef(0)

  useEffect(() => { refreshRef.current = refresh }, [refresh])

  const todaySessions = sessions7d.filter(
    s => localDateKey(new Date(s.started_at)) === localDateKey(new Date())
  )

  useEffect(() => {
    const el = screenRef.current
    if (!el) return
    let startY = 0

    function onStart(e: TouchEvent) {
      startY = e.touches[0].clientY
    }

    function onMove(e: TouchEvent) {
      if (!el) return
      const delta = e.touches[0].clientY - startY
      const scroller = el.querySelector('[data-scrollable]') as HTMLElement | null
      if (delta > 0 && (!scroller || scroller.scrollTop <= 0)) {
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
      {!isLoading && error && <div className={styles.error}>{error}</div>}
      {!error && activeTab === 'today' && (
        <TodayTab
          sessions={todaySessions}
          stats={stats}
          babyId={babyId}
          onRefresh={refresh}
          isLoading={isLoading}
        />
      )}
      {!error && activeTab === 'week' && (
        <WeekTab sessions={sessions7d} nightWindow={stats?.night_window} isLoading={isLoading} />
      )}
      {!error && activeTab === 'summary' && (
        <SummaryTab summary={stats?.summary ?? {}} isLoading={isLoading} />
      )}
    </div>
  )
}
