import { useState, useEffect, useCallback } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { startSleep, stopSleep, editSleepSession, getSleepHistory, type SleepSession } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import ElapsedTimer from '@/components/ElapsedTimer'
import styles from './SleepScreen.module.css'

function toDatetimeLocal(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export default function SleepScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''

  const [session, setSession] = useState<SleepSession | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isToggling, setIsToggling] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [startInput, setStartInput] = useState('')
  const [endInput, setEndInput] = useState('')

  const isActive = session !== null && !session.stopped_at

  const loadSession = useCallback(async () => {
    if (!babyId) return
    try {
      const sessions = await getSleepHistory(babyId, '7d')
      const latest = sessions.length > 0 ? sessions[0] : null
      setSession(latest)
      if (latest) {
        setStartInput(toDatetimeLocal(latest.started_at))
        setEndInput(latest.stopped_at ? toDatetimeLocal(latest.stopped_at) : '')
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep status')
    } finally {
      setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => { loadSession() }, [loadSession])

  async function handleToggle() {
    if (!babyId || isToggling) return
    setIsToggling(true)
    setError(null)
    try {
      if (isActive && session) {
        const stopped = await stopSleep(babyId)
        const updated: SleepSession = {
          id: stopped.id,
          baby_id: babyId,
          started_at: stopped.started_at,
          stopped_at: stopped.stopped_at,
          classification: stopped.classification,
        }
        setSession(updated)
        setStartInput(toDatetimeLocal(stopped.started_at))
        setEndInput(toDatetimeLocal(stopped.stopped_at))
      } else {
        const started = await startSleep(babyId)
        const active: SleepSession = { id: started.id, baby_id: babyId, started_at: started.started_at }
        setSession(active)
        setStartInput(toDatetimeLocal(started.started_at))
        setEndInput('')
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update sleep')
    } finally {
      setIsToggling(false)
    }
  }

  async function handleStartBlur() {
    if (!session || !babyId || !startInput) return
    setError(null)
    try {
      const updated = await editSleepSession(babyId, session.id, {
        started_at: new Date(startInput).toISOString(),
      })
      setSession(updated)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update start time')
      setStartInput(toDatetimeLocal(session.started_at))
    }
  }

  async function handleEndBlur() {
    if (!session || !babyId || !endInput) return
    setError(null)
    try {
      const updated = await editSleepSession(babyId, session.id, {
        stopped_at: new Date(endInput).toISOString(),
      })
      setSession(updated)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update end time')
      if (session.stopped_at) setEndInput(toDatetimeLocal(session.stopped_at))
    }
  }

  if (isLoading) return <p>Loading…</p>

  return (
    <div className={styles.screen}>
      <button
        className={`${styles.button} ${isActive ? styles.stop : styles.start}`}
        onClick={handleToggle}
        disabled={isToggling}
      >
        {isToggling ? <span className={styles.spinner} /> : isActive ? 'Stop sleep' : 'Start sleep'}
      </button>

      {isActive && session && (
        <span className={styles.elapsed}>
          <ElapsedTimer startedAt={session.started_at} />
        </span>
      )}

      {session && (
        <div className={styles.fields}>
          <label className={styles.field}>
            <span className={styles.fieldName}>Start</span>
            <input
              type="datetime-local"
              className={styles.fieldInput}
              value={startInput}
              onChange={e => setStartInput(e.target.value)}
              onBlur={handleStartBlur}
            />
          </label>
          {!isActive && (
            <label className={styles.field}>
              <span className={styles.fieldName}>End</span>
              <input
                type="datetime-local"
                className={styles.fieldInput}
                value={endInput}
                onChange={e => setEndInput(e.target.value)}
                onBlur={handleEndBlur}
              />
            </label>
          )}
        </div>
      )}

      {error && <p className={styles.error} role="alert">{error}</p>}
    </div>
  )
}
