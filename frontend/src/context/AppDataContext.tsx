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
  refresh: () => Promise<void>
}

const AppDataContext = createContext<AppDataContextValue | null>(null)

export function AppDataProvider({ children }: { children: ReactNode }) {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [sessions7d, setSessions7d] = useState<SleepSession[]>([])
  const [stats, setStats] = useState<SleepStatsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    if (!babyId) return
    setIsLoading(true)
    setError(null)
    try {
      const [sessionData, statsData] = await Promise.all([
        getSleepHistory(babyId, '7d'),
        getSleepStats(babyId),
      ])
      setSessions7d(sessionData)
      setStats(statsData)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep data')
    } finally {
      setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => {
    refresh()
  }, [refresh])

  return (
    <AppDataContext.Provider value={{ sessions7d, stats, isLoading, error, refresh }}>
      {children}
    </AppDataContext.Provider>
  )
}

export function useAppData(): AppDataContextValue {
  const ctx = useContext(AppDataContext)
  if (!ctx) throw new Error('useAppData must be used within AppDataProvider')
  return ctx
}
