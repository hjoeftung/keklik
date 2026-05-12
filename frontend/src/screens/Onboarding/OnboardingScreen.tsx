import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { createFamily, setNightWindow } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import { BirthdayPicker, TimePicker } from '@/components/DrumPicker'
import { BottomSheet } from '@/components/BottomSheet'
import BunMascot from '@/components/BunMascot'
import styles from './OnboardingScreen.module.css'


function parseTime(value: string): { hour: number; minute: number } {
  const [h, m] = value.split(':').map(Number)
  return { hour: h, minute: m }
}

function formatTime(value: string): string {
  const [h, m] = value.split(':').map(Number)
  const period = h >= 12 ? 'PM' : 'AM'
  const h12 = h % 12 || 12
  return `${h12}:${String(m).padStart(2, '0')} ${period}`
}


const DEFAULT_BIRTHDAY = (() => {
  const d = new Date()
  d.setMonth(d.getMonth() - 6)
  return d
})()

export default function OnboardingScreen() {
  const navigate = useNavigate()
  const { refreshFamily } = useAuthContext()

  const [yourName, setYourName] = useState('')
  const [babyName, setBabyName] = useState('')
  const [birthday, setBirthday] = useState<Date>(DEFAULT_BIRTHDAY)
  const [birthdayOpen, setBirthdayOpen] = useState(false)
  const [nightStart, setNightStart] = useState('19:00')
  const [nightStartOpen, setNightStartOpen] = useState(false)
  const [nightEnd, setNightEnd] = useState('08:00')
  const [nightEndOpen, setNightEndOpen] = useState(false)

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.SyntheticEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      const family = await createFamily({ baby_name: babyName, creator_name: yourName })

      const start = parseTime(nightStart)
      const end = parseTime(nightEnd)

      await setNightWindow(family.baby_id, {
        start_hour: start.hour,
        start_minute: start.minute,
        end_hour: end.hour,
        end_minute: end.minute,
        effective_from: new Date(0).toISOString(),
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
      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <div className={styles.header}>
          <div>
            <div className={styles.stepLabel}>Step 2 of 2</div>
            <h1 className={styles.title}>
              Tell us about your
              <br />
              little one
            </h1>
          </div>
          <div className={styles.mascot}>
            <BunMascot />
          </div>
        </div>

        {error && (
          <div className={styles.error} role="alert">
            {error}
          </div>
        )}

        <div className={styles.fields}>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="yourName">
              Your name
            </label>
            <input
              id="yourName"
              className={styles.input}
              type="text"
              value={yourName}
              onChange={(e) => setYourName(e.target.value)}
              required
              autoComplete="name"
              autoFocus
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label} htmlFor="babyName">
              Baby's name
            </label>
            <input
              id="babyName"
              className={styles.input}
              type="text"
              value={babyName}
              onChange={(e) => setBabyName(e.target.value)}
              required
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Birthday</label>
            <button
              type="button"
              className={styles.displayRow}
              onClick={() => setBirthdayOpen(true)}
            >
              <span>
                {birthday.toLocaleDateString('en-US', {
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </span>
              <span className={styles.displayRowAction}>Edit</span>
            </button>
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Night starts</label>
            <button
              type="button"
              className={styles.displayRow}
              onClick={() => setNightStartOpen(true)}
            >
              <span>{formatTime(nightStart)}</span>
              <span className={styles.displayRowAction}>Edit</span>
            </button>
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Night ends</label>
            <button
              type="button"
              className={styles.displayRow}
              onClick={() => setNightEndOpen(true)}
            >
              <span>{formatTime(nightEnd)}</span>
              <span className={styles.displayRowAction}>Edit</span>
            </button>
          </div>

          {birthdayOpen && (
            <BottomSheet title="Birthday" onClose={() => setBirthdayOpen(false)}>
              <BirthdayPicker initialDate={birthday} onChange={setBirthday} />
            </BottomSheet>
          )}
          {nightStartOpen && (
            <BottomSheet title="Night starts" onClose={() => setNightStartOpen(false)}>
              <TimePicker value={nightStart} onChange={setNightStart} />
            </BottomSheet>
          )}
          {nightEndOpen && (
            <BottomSheet title="Night ends" onClose={() => setNightEndOpen(false)}>
              <TimePicker value={nightEnd} onChange={setNightEnd} />
            </BottomSheet>
          )}
        </div>

        <button
          className={styles.button}
          type="submit"
          disabled={isSubmitting || !yourName.trim() || !babyName.trim()}
        >
          {isSubmitting ? 'Setting up…' : 'Start tracking'}
        </button>

        <div className={styles.dots}>
          <span className={styles.dotInactive} />
          <span className={styles.dotActive} />
        </div>
      </form>
    </div>
  )
}
