import { useState, useEffect } from 'react'
import { logPastSleep } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import DrumPicker from '@/components/DrumPicker'
import styles from './LogPastSleepSheet.module.css'

interface LogPastSleepSheetProps {
  babyId: string
  onSaved: () => void
  onClose: () => void
}

function formatDisplayTime(d: Date): string {
  const today = new Date()
  const timeStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  if (d.toDateString() === today.toDateString()) return `Today, ${timeStr}`
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + `, ${timeStr}`
}

function durationSeconds(start: Date, end: Date): number {
  return Math.round((end.getTime() - start.getTime()) / 1000)
}

function formatDur(secs: number): string {
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

export default function LogPastSleepSheet({ babyId, onSaved, onClose }: LogPastSleepSheetProps) {
  const defaultEnd = new Date()
  const defaultStart = new Date(defaultEnd.getTime() - 60 * 60 * 1000)

  const [startDate, setStartDate] = useState<Date>(defaultStart)
  const [endDate, setEndDate] = useState<Date>(defaultEnd)
  const [activePicker, setActivePicker] = useState<'start' | 'end' | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const durSecs = durationSeconds(startDate, endDate)
  const isEndBeforeStart = durSecs <= 0

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        if (activePicker) setActivePicker(null)
        else onClose()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose, activePicker])

  async function handleSave() {
    if (isEndBeforeStart || isSaving) return
    setError(null)
    setIsSaving(true)
    try {
      await logPastSleep(babyId, {
        started_at: startDate.toISOString(),
        stopped_at: endDate.toISOString(),
      })
      onSaved()
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError('This session overlaps an existing one.')
      } else {
        setError(err instanceof ApiError ? err.message : 'Failed to save.')
      }
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div
      className={styles.overlay}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className={styles.sheet} role="dialog" aria-modal="true" aria-label="Log past sleep">
        <div className={styles.handle} />

        <h2 className={styles.title}>Log past sleep</h2>

        {/* Inline drum picker — shown when a field is active */}
        {activePicker && (
          <DrumPicker
            initialDate={activePicker === 'start' ? startDate : endDate}
            onChange={(d) => (activePicker === 'start' ? setStartDate(d) : setEndDate(d))}
          />
        )}

        {/* Time rows */}
        <div className={styles.timeRows}>
          <button
            type="button"
            className={`${styles.timeRow} ${styles.timeRowSet} ${activePicker === 'start' ? styles.timeRowActive : ''}`}
            onClick={() => setActivePicker(activePicker === 'start' ? null : 'start')}
          >
            <div className={styles.timeRowInner}>
              <div>
                <div className={styles.timeRowLabel}>STARTED</div>
                <div className={styles.timeRowValue}>{formatDisplayTime(startDate)}</div>
              </div>
              <span className={styles.timeRowAction}>
                {activePicker === 'start' ? 'Done ✓' : 'Change'}
              </span>
            </div>
          </button>

          <button
            type="button"
            className={`${styles.timeRow} ${styles.timeRowSet} ${isEndBeforeStart ? styles.timeRowError : ''} ${activePicker === 'end' ? styles.timeRowActive : ''}`}
            onClick={() => setActivePicker(activePicker === 'end' ? null : 'end')}
          >
            <div className={styles.timeRowInner}>
              <div>
                <div className={styles.timeRowLabel}>ENDED</div>
                <div className={styles.timeRowValue}>{formatDisplayTime(endDate)}</div>
              </div>
              <span
                className={`${styles.timeRowAction} ${isEndBeforeStart ? styles.timeRowActionError : ''}`}
              >
                {activePicker === 'end' ? 'Done ✓' : 'Change'}
              </span>
            </div>
          </button>

          {/* Duration row */}
          <div className={styles.durationRow}>
            {isEndBeforeStart ? (
              <>
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  strokeWidth="2"
                  strokeLinecap="round"
                  stroke="#D4806E"
                >
                  <circle cx="8" cy="8" r="6.5" />
                  <path d="M8 5v3M8 11h.01" />
                </svg>
                <span className={styles.durationError}>End must be after start</span>
              </>
            ) : (
              <>
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  stroke="#86B6A6"
                  strokeWidth="2"
                  strokeLinecap="round"
                >
                  <path d="M3 8l3 3 7-7" />
                </svg>
                <span className={styles.durationText}>
                  Duration <span className={styles.durationValue}>{formatDur(durSecs)}</span>
                </span>
              </>
            )}
          </div>
        </div>

        {error && (
          <p className={styles.error} role="alert">
            {error}
          </p>
        )}

        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onClose} disabled={isSaving}>
            Cancel
          </button>
          <button
            className={styles.saveBtn}
            onClick={handleSave}
            disabled={isSaving || isEndBeforeStart}
          >
            {isSaving ? <span className={styles.spinner} /> : 'Save sleep'}
          </button>
        </div>
      </div>
    </div>
  )
}
