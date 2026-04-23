import { useState, useEffect, useCallback } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { getDashboardSummary, startSleep, stopSleep, DashboardSummary } from '@/api/endpoints'
import { ApiError } from '@/api/client'

export default function SleepControl() {
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
      const data = await getDashboardSummary(babyId)
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

  if (isLoading) {
    return <p>Loading…</p>
  }

  const isSleeping = Boolean(summary?.active_session)

  return (
    <div>
      <button onClick={handleToggle} disabled={isToggling}>
        {isToggling ? 'Loading…' : isSleeping ? 'Stop sleep' : 'Start sleep'}
      </button>
      {error && <p role="alert">{error}</p>}
    </div>
  )
}
