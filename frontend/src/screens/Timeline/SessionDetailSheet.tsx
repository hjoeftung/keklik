import { useState, useEffect, useRef } from 'react'
import { editSleepSession, deleteSleepSession, type SleepSession } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './SessionDetailSheet.module.css'

interface Props {
  session: SleepSession
  babyId: string
  onClose: () => void
  onUpdated: (updated: SleepSession) => void
  onDeleted: () => void
}

type Mode = 'detail' | 'edit'

interface SleepSessionConflict {
  type?: string
  current_session?: SleepSession
  conflicting_session?: SleepSession
}

function toDatetimeLocal(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) return 'Today'
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return 'Yesterday'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function formatDur(seconds: number): string {
  if (seconds <= 0) return '0m'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

function durationSeconds(start: string, end: string): number | null {
  if (!start || !end) return null
  const s = new Date(start).getTime()
  const e = new Date(end).getTime()
  if (isNaN(s) || isNaN(e)) return null
  return Math.round((e - s) / 1000)
}

function formatDisplayTime(datetimeLocal: string): string {
  if (!datetimeLocal) return ''
  const d = new Date(datetimeLocal)
  const today = new Date()
  const timeStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  if (d.toDateString() === today.toDateString()) return `Today, ${timeStr}`
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return `Yesterday, ${timeStr}`
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + `, ${timeStr}`
}

function resetInputsFromSession(
  session: SleepSession,
  setStartInput: (value: string) => void,
  setEndInput: (value: string) => void,
): void {
  setStartInput(toDatetimeLocal(new Date(session.started_at)))
  setEndInput(session.stopped_at ? toDatetimeLocal(new Date(session.stopped_at)) : '')
}

function sleepSessionConflict(err: ApiError): SleepSessionConflict | null {
  const conflict = err.conflict
  if (!conflict || typeof conflict !== 'object') return null
  return conflict as SleepSessionConflict
}

const CLASSIFICATION_NOTE: Record<string, string> = {
  night: 'Night sleep — falls within your night window.',
  nap: 'Nap — falls outside the night window.',
}

export default function SessionDetailSheet({ session, babyId, onClose, onUpdated, onDeleted }: Props) {
  const [currentSession, setCurrentSession] = useState(session)
  const [mode, setMode] = useState<Mode>('detail')
  const [deleteConfirm, setDeleteConfirm] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const [startInput, setStartInput] = useState(() => toDatetimeLocal(new Date(currentSession.started_at)))
  const [endInput, setEndInput] = useState(() =>
    currentSession.stopped_at ? toDatetimeLocal(new Date(currentSession.stopped_at)) : '',
  )
  const [startTouched, setStartTouched] = useState(false)
  const [endTouched, setEndTouched] = useState(false)
  const [activePicker, setActivePicker] = useState<'start' | 'end' | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pickerRef = useRef<HTMLInputElement>(null)

  const durSecs = durationSeconds(startInput, endInput)
  const isEndBeforeStart = durSecs !== null && durSecs <= 0
  const isNight = currentSession.classification === 'night'
  const sessionDate = formatDate(currentSession.started_at)
  const displayTitle = isNight ? 'Night sleep' : 'Nap'

  useEffect(() => {
    setCurrentSession(session)
    resetInputsFromSession(session, setStartInput, setEndInput)
    setStartTouched(false)
    setEndTouched(false)
  }, [session])

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        if (activePicker) setActivePicker(null)
        else if (mode === 'edit') setMode('detail')
        else onClose()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [activePicker, mode, onClose])

  useEffect(() => {
    if (activePicker && pickerRef.current) {
      pickerRef.current.focus()
      pickerRef.current.showPicker?.()
    }
  }, [activePicker])

  async function handleDelete() {
    if (!deleteConfirm) {
      setDeleteConfirm(true)
      return
    }
    setIsDeleting(true)
    setError(null)
    try {
      await deleteSleepSession(babyId, currentSession.id, { version: currentSession.version })
      onDeleted()
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        const conflict = sleepSessionConflict(err)
        if (conflict?.type === 'stale_version' && conflict.current_session) {
          setCurrentSession(conflict.current_session)
          onUpdated(conflict.current_session)
          resetInputsFromSession(conflict.current_session, setStartInput, setEndInput)
          setError('This session changed. Review the latest times before saving.')
        } else {
          setError(err.message)
        }
      } else {
        setError(err instanceof ApiError ? err.message : 'Failed to delete.')
      }
      setIsDeleting(false)
      setDeleteConfirm(false)
    }
  }

  async function handleSave() {
    if (!startInput || !endInput || isEndBeforeStart || isSaving) return
    setError(null)
    setIsSaving(true)
    try {
      const updated = await editSleepSession(babyId, currentSession.id, {
        started_at: new Date(startInput).toISOString(),
        stopped_at: new Date(endInput).toISOString(),
        version: currentSession.version,
      })
      setCurrentSession(updated)
      onUpdated(updated)
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        const conflict = sleepSessionConflict(err)
        if (conflict?.type === 'stale_version' && conflict.current_session) {
          setCurrentSession(conflict.current_session)
          onUpdated(conflict.current_session)
          resetInputsFromSession(conflict.current_session, setStartInput, setEndInput)
          setStartTouched(false)
          setEndTouched(false)
          setError('This session changed. Review the latest times before saving.')
        } else if (conflict?.type === 'overlap' && conflict.conflicting_session) {
          const blocking = conflict.conflicting_session
          const end = blocking.stopped_at ? formatTime(blocking.stopped_at) : 'now'
          setError(`This time overlaps another session from ${formatTime(blocking.started_at)} to ${end}.`)
        } else {
          setError('This time overlaps another session.')
        }
      } else {
        setError(err instanceof ApiError ? err.message : 'Failed to save.')
      }
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className={styles.overlay} onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div className={styles.sheet} role="dialog" aria-modal="true">
        <div className={styles.handle} />

        {mode === 'detail' ? (
          <>
            <div className={styles.titleRow}>
              <div className={`${styles.typeIcon} ${isNight ? styles.typeIconNight : styles.typeIconNap}`}>
                {isNight ? (
                  <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
                    <path d="M17 11.5A7 7 0 0 1 8.5 3a7 7 0 1 0 8.5 8.5z"
                      fill="#fff" stroke="#fff" strokeWidth="0.5" strokeLinejoin="round" />
                  </svg>
                ) : (
                  <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
                    <circle cx="10" cy="10" r="4" fill="#fff" />
                    <g stroke="#fff" strokeWidth="1.8" strokeLinecap="round">
                      <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.1 4.1l1.4 1.4M14.5 14.5l1.4 1.4M4.1 15.9l1.4-1.4M14.5 5.5l1.4-1.4" />
                    </g>
                  </svg>
                )}
              </div>
              <div>
                <h2 className={styles.title}>{displayTitle}</h2>
                <p className={styles.subtitle}>{sessionDate}</p>
              </div>
            </div>

            <div className={styles.infoRows}>
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Started</span>
                <span className={styles.infoValue}>{formatTime(currentSession.started_at)}</span>
              </div>
              {currentSession.stopped_at && (
                <div className={styles.infoRow}>
                  <span className={styles.infoLabel}>Ended</span>
                  <span className={styles.infoValue}>{formatTime(currentSession.stopped_at)}</span>
                </div>
              )}
              {currentSession.duration_seconds != null && (
                <div className={styles.infoRow}>
                  <span className={styles.infoLabel}>Duration</span>
                  <span className={styles.infoValue}>{formatDur(currentSession.duration_seconds)}</span>
                </div>
              )}
            </div>

            {currentSession.classification && (
              <p className={styles.classNote}>
                {CLASSIFICATION_NOTE[currentSession.classification] ?? `Classified as ${currentSession.classification}.`}
              </p>
            )}

            {error && <p className={styles.error} role="alert">{error}</p>}

            <div className={styles.actions}>
              <button
                className={`${styles.deleteBtn} ${deleteConfirm ? styles.deleteBtnConfirm : ''}`}
                onClick={handleDelete}
                disabled={isDeleting}
              >
                {isDeleting ? <span className={styles.spinner} /> : deleteConfirm ? 'Confirm delete' : 'Delete'}
              </button>
              <button className={styles.editBtn} onClick={() => setMode('edit')}>
                Edit session
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 className={styles.title}>Edit session</h2>

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

            <div className={styles.timeRows}>
              <button
                type="button"
                className={`${styles.timeRow} ${startTouched ? styles.timeRowSet : styles.timeRowSuggested} ${activePicker === 'start' ? styles.timeRowActive : ''}`}
                onClick={() => setActivePicker(activePicker === 'start' ? null : 'start')}
              >
                <div className={styles.timeRowInner}>
                  <div>
                    <div className={styles.timeRowLabel}>STARTED</div>
                    <div className={styles.timeRowValue}>{formatDisplayTime(startInput)}</div>
                  </div>
                  <span className={styles.timeRowAction}>Change</span>
                </div>
              </button>

              <button
                type="button"
                className={`${styles.timeRow} ${endTouched ? styles.timeRowSet : styles.timeRowSuggested} ${isEndBeforeStart ? styles.timeRowError : ''} ${activePicker === 'end' ? styles.timeRowActive : ''}`}
                onClick={() => setActivePicker(activePicker === 'end' ? null : 'end')}
              >
                <div className={styles.timeRowInner}>
                  <div>
                    <div className={styles.timeRowLabel}>ENDED</div>
                    <div className={styles.timeRowValue}>{formatDisplayTime(endInput)}</div>
                  </div>
                  <span className={`${styles.timeRowAction} ${isEndBeforeStart ? styles.timeRowActionError : ''}`}>
                    Change
                  </span>
                </div>
              </button>

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
                ) : null}
              </div>
            </div>

            {error && <p className={styles.error} role="alert">{error}</p>}

            <div className={styles.actions}>
              <button className={styles.cancelBtn} onClick={() => { setMode('detail'); setError(null) }} disabled={isSaving}>
                Cancel
              </button>
              <button
                className={styles.saveBtn}
                onClick={handleSave}
                disabled={isSaving || isEndBeforeStart || !startInput || !endInput}
              >
                {isSaving ? <span className={styles.spinner} /> : 'Save changes'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
