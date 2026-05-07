import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { getSleepStats, setNightWindow } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './NightWindowScreen.module.css'

// backend returns "HH:MM" which is already the time input format
function hhmmToTimeInput(hhmm: string): string {
  return hhmm
}

function parseTimeInput(value: string): { hour: number; minute: number } {
  const [h, m] = value.split(':').map(Number)
  return { hour: h, minute: m }
}

export default function NightWindowScreen() {
  const navigate = useNavigate()
  const { family } = useAuthContext()
  const babyId = family?.baby.id

  const [nightStart, setNightStart] = useState('22:00')
  const [nightEnd, setNightEnd] = useState('06:00')
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!babyId) return
    getSleepStats(babyId)
      .then((stats) => {
        if (stats.night_window) {
          setNightStart(hhmmToTimeInput(stats.night_window.start_hhmm))
          setNightEnd(hhmmToTimeInput(stats.night_window.end_hhmm))
        }
      })
      .catch(() => {})
      .finally(() => setIsLoading(false))
  }, [babyId])

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    if (!babyId) return
    setError(null)
    setIsSaving(true)

    const start = parseTimeInput(nightStart)
    const end = parseTimeInput(nightEnd)

    try {
      await setNightWindow(babyId, {
        start_hour: start.hour,
        start_minute: start.minute,
        end_hour: end.hour,
        end_minute: end.minute,
        effective_from: new Date().toISOString(),
      })
      navigate('/settings')
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Something went wrong. Please try again.')
      }
      setIsSaving(false)
    }
  }

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <button className={styles.back} onClick={() => navigate('/settings')}>
          ‹
        </button>
        <h1 className={styles.title}>Night window</h1>
      </div>

      {isLoading ? (
        <div className={styles.loading}>Loading…</div>
      ) : (
        <form className={styles.card} onSubmit={handleSave} noValidate>
          <p className={styles.hint}>
            Sleeps that start during the night window are classified as night sleep.
          </p>

          {error && (
            <div className={styles.error} role="alert">
              {error}
            </div>
          )}

          <div className={styles.fields}>
            <div className={styles.field}>
              <label className={styles.label} htmlFor="nightStart">
                Night starts
              </label>
              <input
                id="nightStart"
                className={styles.input}
                type="time"
                value={nightStart}
                onChange={(e) => setNightStart(e.target.value)}
                required
              />
            </div>
            <div className={styles.field}>
              <label className={styles.label} htmlFor="nightEnd">
                Night ends
              </label>
              <input
                id="nightEnd"
                className={styles.input}
                type="time"
                value={nightEnd}
                onChange={(e) => setNightEnd(e.target.value)}
                required
              />
            </div>
          </div>

          <button className={styles.button} type="submit" disabled={isSaving}>
            {isSaving ? 'Saving…' : 'Save'}
          </button>
        </form>
      )}
    </div>
  )
}
