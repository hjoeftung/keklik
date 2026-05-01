import { useState, useEffect, useRef } from 'react'
import { logPastSleep } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './LogPastSleepSheet.module.css'

interface LogPastSleepSheetProps {
  babyId: string
  onSaved: () => void
  onClose: () => void
}

function toDatetimeLocal(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatDisplayTime(datetimeLocal: string): string {
  if (!datetimeLocal) return ''
  const d = new Date(datetimeLocal)
  const today = new Date()
  const isToday = d.toDateString() === today.toDateString()
  const timeStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  return isToday
    ? `Today, ${timeStr}`
    : d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + `, ${timeStr}`
}

function durationSeconds(start: string, end: string): number | null {
  if (!start || !end) return null
  const s = new Date(start).getTime()
  const e = new Date(end).getTime()
  if (isNaN(s) || isNaN(e)) return null
  return Math.round((e - s) / 1000)
}

function formatDur(secs: number): string {
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

type SleepType = 'night' | 'nap'

export default function LogPastSleepSheet({ babyId, onSaved, onClose }: LogPastSleepSheetProps) {
  const defaultEnd = new Date()
  const defaultStart = new Date(defaultEnd.getTime() - 60 * 60 * 1000)

  const [startInput, setStartInput] = useState(toDatetimeLocal(defaultStart))
  const [endInput, setEndInput] = useState(toDatetimeLocal(defaultEnd))
  const [startTouched, setStartTouched] = useState(false)
  const [endTouched, setEndTouched] = useState(false)
  const [activePicker, setActivePicker] = useState<'start' | 'end' | null>(null)
  const [sleepType, setSleepType] = useState<SleepType>('nap')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pickerRef = useRef<HTMLInputElement>(null)

  const durSecs = durationSeconds(startInput, endInput)
  const isEndBeforeStart = durSecs !== null && durSecs <= 0

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

  useEffect(() => {
    if (activePicker && pickerRef.current) {
      pickerRef.current.focus()
      pickerRef.current.showPicker?.()
    }
  }, [activePicker])

  async function handleSave() {
    if (!startInput || !endInput || isEndBeforeStart || isSaving) return
    setError(null)
    setIsSaving(true)
    try {
      await logPastSleep(babyId, {
        started_at: new Date(startInput).toISOString(),
        stopped_at: new Date(endInput).toISOString(),
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
    <div className={styles.overlay} onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div className={styles.sheet} role="dialog" aria-modal="true" aria-label="Log past sleep">
        <div className={styles.handle} />

        <h2 className={styles.title}>Log past sleep</h2>
        <p className={styles.subtitle}>
          {startTouched || endTouched ? "For sleep that wasn’t tracked live" : "We’ve suggested times — adjust if needed"}
        </p>

        {/* Sleep type toggle */}
        <div className={styles.typeToggle}>
          <button
            className={`${styles.typeBtn} ${sleepType === 'night' ? styles.typeBtnNight : styles.typeBtnOff}`}
            onClick={() => setSleepType('night')}
          >
            <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
              <path d="M17 11.5A7 7 0 0 1 8.5 3a7 7 0 1 0 8.5 8.5z"
                fill={sleepType === 'night' ? '#fff' : 'var(--kk-night)'}
                stroke={sleepType === 'night' ? '#fff' : 'var(--kk-night)'}
                strokeWidth="0.5" strokeLinejoin="round" />
            </svg>
            Night sleep
          </button>
          <button
            className={`${styles.typeBtn} ${sleepType === 'nap' ? styles.typeBtnNap : styles.typeBtnOff}`}
            onClick={() => setSleepType('nap')}
          >
            <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
              <circle cx="10" cy="10" r="4" fill={sleepType === 'nap' ? '#fff' : 'var(--kk-nap)'} />
              <g stroke={sleepType === 'nap' ? '#fff' : 'var(--kk-nap)'} strokeWidth="1.8" strokeLinecap="round">
                <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.1 4.1l1.4 1.4M14.5 14.5l1.4 1.4M4.1 15.9l1.4-1.4M14.5 5.5l1.4-1.4" />
              </g>
            </svg>
            Nap
          </button>
        </div>

        {/* Inline picker — shown above the time rows when a field is active */}
        {activePicker && (
          <div className={styles.pickerPanel}>
            <div className={styles.pickerPanelLabel}>
              {activePicker === 'start' ? 'SET START TIME' : 'SET END TIME'}
            </div>
            <input
              ref={pickerRef}
              type="datetime-local"
              className={styles.pickerInput}
              value={activePicker === 'start' ? startInput : endInput}
              onChange={e => {
                if (activePicker === 'start') { setStartInput(e.target.value); setStartTouched(true) }
                else { setEndInput(e.target.value); setEndTouched(true) }
              }}
            />
            <button className={styles.pickerDoneBtn} onClick={() => setActivePicker(null)}>Done</button>
          </div>
        )}

        {/* Time rows */}
        <div className={styles.timeRows}>
          <button
            type="button"
            className={`${styles.timeRow} ${startTouched ? styles.timeRowSet : styles.timeRowSuggested} ${activePicker === 'start' ? styles.timeRowActive : ''}`}
            onClick={() => setActivePicker(activePicker === 'start' ? null : 'start')}
          >
            <div className={styles.timeRowInner}>
              <div>
                <div className={styles.timeRowLabel}>
                  {startTouched ? 'STARTED' : 'STARTED · SUGGESTED'}
                </div>
                <div className={styles.timeRowValue}>{formatDisplayTime(startInput)}</div>
              </div>
              <span className={styles.timeRowAction}>{startTouched ? 'Change' : 'Tap to set'}</span>
            </div>
          </button>

          <button
            type="button"
            className={`${styles.timeRow} ${endTouched ? styles.timeRowSet : styles.timeRowSuggested} ${isEndBeforeStart ? styles.timeRowError : ''} ${activePicker === 'end' ? styles.timeRowActive : ''}`}
            onClick={() => setActivePicker(activePicker === 'end' ? null : 'end')}
          >
            <div className={styles.timeRowInner}>
              <div>
                <div className={styles.timeRowLabel}>
                  {endTouched ? 'ENDED' : 'ENDED · NOW'}
                </div>
                <div className={styles.timeRowValue}>{formatDisplayTime(endInput)}</div>
              </div>
              <span className={`${styles.timeRowAction} ${isEndBeforeStart ? styles.timeRowActionError : ''}`}>
                {endTouched ? 'Change' : 'Tap to set'}
              </span>
            </div>
          </button>

          {/* Duration row */}
          <div className={styles.durationRow}>
            {isEndBeforeStart ? (
              <>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" strokeWidth="2" strokeLinecap="round" stroke="#D4806E">
                  <circle cx="8" cy="8" r="6.5" /><path d="M8 5v3M8 11h.01" />
                </svg>
                <span className={styles.durationError}>End must be after start</span>
              </>
            ) : durSecs !== null ? (
              <>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#86B6A6" strokeWidth="2" strokeLinecap="round">
                  <path d="M3 8l3 3 7-7" />
                </svg>
                <span className={styles.durationText}>
                  Duration <span className={styles.durationValue}>{formatDur(durSecs)}</span>
                </span>
              </>
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="var(--kk-ink-muted)" strokeWidth="2" strokeLinecap="round">
                  <circle cx="8" cy="8" r="6" /><path d="M8 5v3l2 1.5" />
                </svg>
                <span className={styles.durationMuted}>Duration appears once both times are set</span>
              </>
            )}
          </div>
        </div>

        {error && <p className={styles.error} role="alert">{error}</p>}

        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onClose} disabled={isSaving}>
            Cancel
          </button>
          <button
            className={styles.saveBtn}
            onClick={handleSave}
            disabled={isSaving || isEndBeforeStart || !startInput || !endInput}
          >
            {isSaving ? <span className={styles.spinner} /> : 'Save sleep'}
          </button>
        </div>
      </div>
    </div>
  )
}
