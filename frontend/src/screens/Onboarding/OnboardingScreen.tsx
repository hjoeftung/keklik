import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { createFamily, createSleepProfile } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import { BirthdayPicker, TimePicker } from '@/components/DrumPicker'
import styles from './OnboardingScreen.module.css'

const DEFAULT_TZ = Intl.DateTimeFormat().resolvedOptions().timeZone

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

function BunMascot() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="104" height="104" viewBox="0 0 120 120">
      <g fill="#FBE4D1" stroke="#C97B4D" strokeWidth="1.4" strokeOpacity="0.3">
        <path d="M 58 36 C 56 28, 50 24, 48 20 Q 54 26, 60 28 Q 60 32, 58 36 Z" />
        <path d="M 64 35 C 70 28, 78 28, 80 22 Q 76 30, 72 32 Q 70 36, 64 35 Z" />
        <path d="M 50 38 C 46 34, 40 36, 38 32 Q 44 36, 48 36 Q 50 37, 50 38 Z" />
      </g>
      <path
        d="M 60 36 C 38 36, 24 56, 24 78 C 24 100, 40 116, 60 116 C 80 116, 96 100, 96 78 C 96 56, 82 36, 60 36 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="2"
        strokeOpacity="0.22"
      />
      <path
        d="M 28 74 Q 18 70, 16 64 Q 20 70, 24 72 Q 18 74, 20 78 Q 24 76, 28 76 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="1.4"
        strokeOpacity="0.25"
      />
      <path
        d="M 92 74 Q 102 70, 104 64 Q 100 70, 96 72 Q 102 74, 100 78 Q 96 76, 92 76 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="1.4"
        strokeOpacity="0.25"
      />
      <ellipse cx="40" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
      <ellipse cx="78" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
      <ellipse cx="49" cy="69" rx="3.2" ry="3.6" fill="#2E2A33" />
      <ellipse cx="71" cy="69" rx="3.2" ry="3.6" fill="#2E2A33" />
      <circle cx="50" cy="67.5" r="1" fill="#fff" />
      <circle cx="72" cy="67.5" r="1" fill="#fff" />
      <path
        d="M 56 78 L 60 84 L 64 78 Q 60 76, 56 78 Z"
        fill="#E8B86E"
        stroke="#C97B4D"
        strokeWidth="1.2"
        strokeOpacity="0.4"
        strokeLinejoin="round"
      />
    </svg>
  )
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

      await createSleepProfile(family.baby_id, {
        timezone: DEFAULT_TZ,
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
              onClick={() => setBirthdayOpen((o) => !o)}
            >
              <span>
                {birthday.toLocaleDateString('en-US', {
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </span>
              <span className={styles.displayRowAction}>{birthdayOpen ? 'Done' : 'Edit'}</span>
            </button>
            {birthdayOpen && <BirthdayPicker initialDate={birthday} onChange={setBirthday} />}
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Night starts</label>
            <button
              type="button"
              className={styles.displayRow}
              onClick={() => setNightStartOpen((o) => !o)}
            >
              <span>{formatTime(nightStart)}</span>
              <span className={styles.displayRowAction}>{nightStartOpen ? 'Done' : 'Edit'}</span>
            </button>
            {nightStartOpen && <TimePicker value={nightStart} onChange={setNightStart} />}
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Night ends</label>
            <button
              type="button"
              className={styles.displayRow}
              onClick={() => setNightEndOpen((o) => !o)}
            >
              <span>{formatTime(nightEnd)}</span>
              <span className={styles.displayRowAction}>{nightEndOpen ? 'Done' : 'Edit'}</span>
            </button>
            {nightEndOpen && <TimePicker value={nightEnd} onChange={setNightEnd} />}
          </div>
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
