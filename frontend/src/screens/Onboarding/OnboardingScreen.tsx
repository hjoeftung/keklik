import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { createFamily, createSleepProfile } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './OnboardingScreen.module.css'

const DEFAULT_TZ = Intl.DateTimeFormat().resolvedOptions().timeZone
const ALL_TIMEZONES = Intl.supportedValuesOf('timeZone')

function parseTime(value: string): { hour: number; minute: number } {
  const [h, m] = value.split(':').map(Number)
  return { hour: h, minute: m }
}

export default function OnboardingScreen() {
  const navigate = useNavigate()
  const { refreshFamily } = useAuthContext()

  const [yourName, setYourName] = useState('')
  const [babyName, setBabyName] = useState('')
  const [timezone, setTimezone] = useState(DEFAULT_TZ)
  const [nightStart, setNightStart] = useState('22:00')
  const [nightEnd, setNightEnd] = useState('08:00')

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      const family = await createFamily({ baby_name: babyName, creator_name: yourName })

      const start = parseTime(nightStart)
      const end = parseTime(nightEnd)

      await createSleepProfile(family.baby_id, {
        timezone,
        night_window: {
          start_hour: start.hour,
          start_minute: start.minute,
          end_hour: end.hour,
          end_minute: end.minute,
        },
      })

      await refreshFamily()
      navigate('/sleep')
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Something went wrong. Please try again.')
      }
      setIsSubmitting(false)
    }
  }

  return (
    <div className={styles.screen}>
      <form className={styles.card} onSubmit={handleSubmit} noValidate>
        <div>
          <h1 className={styles.title}>Set up your family</h1>
          <p className={styles.subtitle}>Just a few details to get started.</p>
        </div>

        {error && (
          <div className={styles.error} role="alert">
            {error}
          </div>
        )}

        <div className={styles.field}>
          <label className={styles.label} htmlFor="yourName">Your name</label>
          <input
            id="yourName"
            className={styles.input}
            type="text"
            value={yourName}
            onChange={e => setYourName(e.target.value)}
            required
            autoComplete="name"
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="babyName">Baby's name</label>
          <input
            id="babyName"
            className={styles.input}
            type="text"
            value={babyName}
            onChange={e => setBabyName(e.target.value)}
            required
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="timezone">Timezone</label>
          <input
            id="timezone"
            className={styles.input}
            list="tz-list"
            value={timezone}
            onChange={e => setTimezone(e.target.value)}
            required
          />
          <datalist id="tz-list">
            {ALL_TIMEZONES.map(tz => (
              <option key={tz} value={tz} />
            ))}
          </datalist>
        </div>

        <div className={styles.row}>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="nightStart">Night starts</label>
            <input
              id="nightStart"
              className={styles.input}
              type="time"
              value={nightStart}
              onChange={e => setNightStart(e.target.value)}
              required
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label} htmlFor="nightEnd">Night ends</label>
            <input
              id="nightEnd"
              className={styles.input}
              type="time"
              value={nightEnd}
              onChange={e => setNightEnd(e.target.value)}
              required
            />
          </div>
        </div>

        <button className={styles.button} type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating…' : 'Create family'}
        </button>
      </form>
    </div>
  )
}
