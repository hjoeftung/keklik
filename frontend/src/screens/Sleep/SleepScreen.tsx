import { useState, useEffect, useCallback, useRef } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import {
  startSleep,
  stopSleep,
  editSleepSession,
  getSleepHistory,
  type SleepSession,
} from '@/api/endpoints'
import { ApiError } from '@/api/client'
import PillowButton from './PillowButton'
import LogPastSleepSheet from './LogPastSleepSheet'
import styles from './SleepScreen.module.css'

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
}

function toDatetimeLocal(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
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

function elapsedSecs(from: string): number {
  return Math.max(0, Math.floor((Date.now() - new Date(from).getTime()) / 1000))
}

function fmtHM(secs: number): { h: number; m: number } {
  return { h: Math.floor(secs / 3600), m: Math.floor((secs % 3600) / 60) }
}

function todayTotalSecs(sessions: SleepSession[]): number {
  const midnight = new Date()
  midnight.setHours(0, 0, 0, 0)
  return sessions
    .filter(s => s.stopped_at && new Date(s.started_at) >= midnight)
    .reduce((acc, s) => acc + (s.duration_seconds ?? 0), 0)
}

export default function SleepScreen() {
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''
  const babyName = family?.baby.name ?? 'Baby'

  const [sessions, setSessions] = useState<SleepSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isToggling, setIsToggling] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)
  const [showLogPast, setShowLogPast] = useState(false)
  const [showEditStart, setShowEditStart] = useState(false)
  const [editStartInput, setEditStartInput] = useState('')
  const [showStartPicker, setShowStartPicker] = useState(false)
  const [isSavingStart, setIsSavingStart] = useState(false)
  const pickerRef = useRef<HTMLInputElement>(null)

  const session = sessions.find(s => !s.stopped_at) ?? null
  const isActive = session !== null
  const lastCompleted = sessions.find(s => !!s.stopped_at) ?? null

  const loadSessions = useCallback(async () => {
    if (!babyId) return
    try {
      const all = await getSleepHistory(babyId, '7d')
      setSessions(all)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load sleep status')
    } finally {
      setIsLoading(false)
    }
  }, [babyId])

  useEffect(() => { loadSessions() }, [loadSessions])

  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 60_000)
    return () => clearInterval(id)
  }, [])

  // tick is only used to force re-render for elapsed timers — suppress lint warning
  void tick

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
          version: stopped.version,
        }
        setSessions(prev => [updated, ...prev.filter(s => s.id !== updated.id)])
      } else {
        const started = await startSleep(babyId)
        const active: SleepSession = { id: started.id, baby_id: babyId, started_at: started.started_at, version: started.version }
        setSessions(prev => [active, ...prev])
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update sleep')
    } finally {
      setIsToggling(false)
    }
  }

  function openEditStart() {
    if (!session) return
    setEditStartInput(toDatetimeLocal(new Date(session.started_at)))
    setShowStartPicker(false)
    setShowEditStart(true)
  }

  useEffect(() => {
    if (showStartPicker && pickerRef.current) {
      pickerRef.current.focus()
      pickerRef.current.showPicker?.()
    }
  }, [showStartPicker])

  async function handleSaveStartTime() {
    if (!session || !babyId || !editStartInput || isSavingStart) return
    setIsSavingStart(true)
    setError(null)
    try {
      const updated = await editSleepSession(babyId, session.id, {
        started_at: new Date(editStartInput).toISOString(),
        version: session.version,
      })
      setSessions(prev => [updated, ...prev.filter(s => s.id !== updated.id)])
      setShowEditStart(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update start time')
    } finally {
      setIsSavingStart(false)
    }
  }

  if (isLoading) return <div className={styles.loading}>Loading…</div>

  const dateStr = new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })

  // ── Asleep / Soft state ──────────────────────────────────────────────────
  if (isActive && session) {
    const sleeping = fmtHM(elapsedSecs(session.started_at))
    const startedAtStr = fmtTime(session.started_at)

    return (
      <div className={styles.screen}>
        <div className={styles.header}>
          <div>
            <div className={styles.greeting}>{babyName}</div>
            <div className={styles.dateStr}>{dateStr}</div>
          </div>
          <div className={styles.avatar}>{babyName[0]}</div>
        </div>

        <div className={styles.durationBlock}>
          <div className={styles.durationLabel}>Sleeping for</div>
          <div className={styles.durationValue}>
            {sleeping.h}<span className={styles.durationUnit}>h </span>
            {String(sleeping.m).padStart(2, '0')}<span className={styles.durationUnit}>m</span>
          </div>
          <div className={styles.startedAtSoft}>Started at {startedAtStr}</div>
        </div>

        {/* Pillow with floating Z's positioned relative to it */}
        <div className={styles.pillowWrap}>
          <div className={styles.pillowZeeContainer}>
            <span className={`${styles.zee} ${styles.z1}`}>z</span>
            <span className={`${styles.zee} ${styles.z2}`}>z</span>
            <span className={`${styles.zee} ${styles.z3}`}>Z</span>
            <PillowButton
              label="Tap to wake"
              masked
              onClick={handleToggle}
              isDisabled={isToggling}
            />
          </div>
          <div className={styles.wakeHint}>Tap pillow to wake · shhh</div>
        </div>

        <button className={styles.editStartPillSoft} onClick={openEditStart}>
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M3 11l1.5-4.5L9 2l3 3-4.5 4.5L3 11zM2 13h12" />
          </svg>
          Edit start time · {startedAtStr}
        </button>

        {error && <p className={styles.error} role="alert">{error}</p>}

        {showEditStart && (
          <div className={styles.editSheetOverlay} onClick={e => { if (e.target === e.currentTarget) setShowEditStart(false) }}>
            <div className={styles.editSheet} role="dialog" aria-modal="true" aria-label="Edit start time">
              <div className={styles.editSheetHandle} />

              <div className={styles.editSheetHeader}>
                <div className={styles.editSheetIcon}>
                  <svg width="22" height="22" viewBox="0 0 20 20" fill="none" aria-hidden="true">
                    <path d="M17 11.5A7 7 0 0 1 8.5 3a7 7 0 1 0 8.5 8.5z" fill="#fff" stroke="#fff" strokeWidth="0.5" strokeLinejoin="round" />
                  </svg>
                </div>
                <div>
                  <div className={styles.editSheetTitle}>Sleep in progress</div>
                  <div className={styles.editSheetSubtitle}>Today</div>
                </div>
              </div>

              {showStartPicker && (
                <div className={styles.editSheetPickerPanel}>
                  <div className={styles.editSheetPickerLabel}>SET START TIME</div>
                  <input
                    ref={pickerRef}
                    type="datetime-local"
                    className={styles.editSheetPickerInput}
                    value={editStartInput}
                    onChange={e => setEditStartInput(e.target.value)}
                  />
                  <button className={styles.editSheetPickerDone} onClick={() => setShowStartPicker(false)}>Done</button>
                </div>
              )}

              <div className={styles.editSheetTimeRows}>
                <button
                  type="button"
                  className={`${styles.editSheetTimeRow} ${showStartPicker ? styles.editSheetTimeRowActive : ''}`}
                  onClick={() => setShowStartPicker(p => !p)}
                >
                  <div className={styles.editSheetTimeRowInner}>
                    <div>
                      <div className={styles.editSheetTimeRowLabel}>STARTED</div>
                      <div className={styles.editSheetTimeRowValue}>{formatDisplayTime(editStartInput)}</div>
                    </div>
                    <span className={styles.editSheetTimeRowAction}>Change</span>
                  </div>
                </button>
              </div>

              {error && <p className={styles.error} role="alert">{error}</p>}

              <div className={styles.editSheetActions}>
                <button
                  className={styles.editSheetCancelBtn}
                  onClick={() => setShowEditStart(false)}
                  disabled={isSavingStart}
                >
                  Cancel
                </button>
                <button
                  className={styles.editSheetSaveBtn}
                  onClick={handleSaveStartTime}
                  disabled={isSavingStart || !editStartInput}
                >
                  {isSavingStart ? <span className={styles.spinner} /> : 'Save changes'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }

  // ── Awake / Minimal state ────────────────────────────────────────────────
  const awakeFrom = lastCompleted?.stopped_at
  const awake = fmtHM(awakeFrom ? elapsedSecs(awakeFrom) : 0)
  const todayHM = fmtHM(todayTotalSecs(sessions))
  const todaySecs = todayTotalSecs(sessions)

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <div>
          <div className={styles.greeting}>{babyName}</div>
          <div className={styles.dateStr}>{dateStr}</div>
        </div>
        <div className={styles.avatar}>{babyName[0]}</div>
      </div>

      <div className={styles.centerContent}>
        <div className={styles.durationBlock}>
          <div className={styles.durationLabel}>Active for</div>
          <div className={styles.durationValue}>
            {awake.h}<span className={styles.durationUnit}>h </span>
            {String(awake.m).padStart(2, '0')}<span className={styles.durationUnit}>m</span>
          </div>
        </div>

        <div className={styles.pillowWrap}>
          <PillowButton
            label="Tap to sleep"
            onClick={handleToggle}
            isDisabled={isToggling}
          />
          {lastCompleted && (
            <div className={styles.lastSleepEnd}>
              Last sleep ended {fmtTime(lastCompleted.stopped_at!)}
            </div>
          )}
        </div>
      </div>

      <div className={styles.quickActions}>
        <button className={styles.quickCard} onClick={() => setShowLogPast(true)}>
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--kk-primary-deep)" strokeWidth="1.8" strokeLinecap="round" aria-hidden="true">
            <circle cx="10" cy="10" r="7" /><path d="M10 6v4l2.5 2" />
          </svg>
          <div>
            <div className={styles.quickCardLabel}>Log past</div>
            <div className={styles.quickCardValue}>sleep</div>
          </div>
        </button>
        <div className={styles.quickCard}>
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--kk-primary-deep)" strokeWidth="1.8" strokeLinecap="round" aria-hidden="true">
            <rect x="3" y="4" width="14" height="13" rx="2" /><path d="M7 2v3M13 2v3M3 8h14" />
          </svg>
          <div>
            <div className={styles.quickCardLabel}>Today</div>
            <div className={styles.quickCardValue}>
              {todaySecs > 0 ? `${todayHM.h}h ${todayHM.m}m sleep` : 'No sleep yet'}
            </div>
          </div>
        </div>
      </div>

      {error && <p className={styles.error} role="alert">{error}</p>}

      {showLogPast && (
        <LogPastSleepSheet
          babyId={babyId}
          onSaved={() => { setShowLogPast(false); loadSessions() }}
          onClose={() => setShowLogPast(false)}
        />
      )}
    </div>
  )
}
