import { useState, useEffect, useCallback } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { getDashboardSummary, startSleep, stopSleep, type DashboardSummary } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import SleepControl from './SleepControl'
import SinceLastPanel from './SinceLastPanel'
import SummaryPanel from './SummaryPanel'

export default function DashboardScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isToggling, setIsToggling] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchSummary = useCallback(async () => {
    if (!babyId) return
    setError(null)
    try {
      const data = await getDashboardSummary(babyId, Intl.DateTimeFormat().resolvedOptions().timeZone)
      setSummary(data)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep status')
    } finally {
      setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => {
    fetchSummary()
  }, [fetchSummary])

  async function handleToggle() {
    if (!babyId || isToggling) return
    setIsToggling(true)
    setError(null)
    try {
      if (summary?.active_session) {
        await stopSleep(babyId)
      } else {
        await startSleep(babyId)
      }
      await fetchSummary()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update sleep state')
    } finally {
      setIsToggling(false)
    }
  }

  return (
    <div>
      <SleepControl
        activeSession={summary?.active_session ?? null}
        isLoading={isLoading}
        isToggling={isToggling}
        error={error}
        onToggle={handleToggle}
      />
      <SinceLastPanel
        timeSinceSleepStart={summary?.time_since_sleep_start_seconds ?? null}
        timeSinceAwakening={summary?.time_since_awakening_seconds ?? null}
      />
      {summary && (
        <SummaryPanel
          today={summary.today}
          rolling7d={summary.rolling_7d}
          rolling14d={summary.rolling_14d}
        />
      )}
    </div>
  )
}
