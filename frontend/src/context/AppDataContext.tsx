import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from 'react'
import { useAuthContext } from './AuthContext'
import {
  getSleepHistory,
  getSleepStats,
  type SleepSession,
  type SleepStatsResponse,
} from '@/api/endpoints'
import { ApiError } from '@/api/client'

interface AppDataContextValue {
  sessions7d: SleepSession[]
  stats: SleepStatsResponse | null
  isLoading: boolean
  error: string | null
  refresh: (silent?: boolean) => Promise<void>
  updateSessions7d: (updater: (prev: SleepSession[]) => SleepSession[]) => void
}

const AppDataContext = createContext<AppDataContextValue | null>(null)

export function AppDataProvider({ children }: { children: ReactNode }) {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [sessions7d, setSessions7d] = useState<SleepSession[]>([])
  const [stats, setStats] = useState<SleepStatsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async (silent = false) => {
    if (!babyId) return
    if (!silent) setIsLoading(true)
    setError(null)
    try {
      const [sessionData, statsData] = await Promise.all([
        getSleepHistory(babyId, '7d'),
        getSleepStats(babyId),
      ])
      setSessions7d(prev => JSON.stringify(prev) === JSON.stringify(sessionData) ? prev : sessionData)
      setStats(prev => JSON.stringify(prev) === JSON.stringify(statsData) ? prev : statsData)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep data')
    } finally {
      if (!silent) setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    refresh()
  }, [refresh])

  useEffect(() => {
    const lastRefresh = { at: 0 }

    const handleVisibilityChange = () => {
      if (document.visibilityState !== 'visible') return
      const now = Date.now()
      if (now - lastRefresh.at < 30000) return
      lastRefresh.at = now
      refresh(true)
    }

    const handleFocus = () => {
      const now = Date.now()
      if (now - lastRefresh.at < 30000) return
      lastRefresh.at = now
      refresh(true)
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('focus', handleFocus)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      window.removeEventListener('focus', handleFocus)
    }
  }, [refresh])

  return (
    <AppDataContext.Provider value={{ sessions7d, stats, isLoading, error, refresh, updateSessions7d: setSessions7d }}>
      {children}
    </AppDataContext.Provider>
  )
}

export function useAppData(): AppDataContextValue {
  const ctx = useContext(AppDataContext)
  if (!ctx) throw new Error('useAppData must be used within AppDataProvider')
  return ctx
}
