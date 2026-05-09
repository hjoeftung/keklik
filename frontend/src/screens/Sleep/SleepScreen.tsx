import { useState, useEffect } from 'react'
import { useAuthContext } from '@/context/AuthContext'
import { useTimeFormatContext } from '@/context/TimeFormatContext'
import { useAppData } from '@/context/AppDataContext'
import {
  startSleep,
  stopSleep,
  editSleepSession,
  type SleepSession,
} from '@/api/endpoints'
import { ApiError } from '@/api/client'
import PillowButton from './PillowButton'
import LogPastSleepSheet from './LogPastSleepSheet'
import DrumPicker from '@/components/DrumPicker'
import styles from './SleepScreen.module.css'

function makeFmtTime(use24h: boolean) {
  return (iso: string) =>
    new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: !use24h })
}

function makeFormatDisplayTime(use24h: boolean) {
  return (d: Date): string => {
    const today = new Date()
    const timeStr = d.toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      hour12: !use24h,
    })
    if (d.toDateString() === today.toDateString()) return `Today, ${timeStr}`
    const yesterday = new Date(today)
    yesterday.setDate(today.getDate() - 1)
    if (d.toDateString() === yesterday.toDateString()) return `Yesterday, ${timeStr}`
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + `, ${timeStr}`
  }
}

function elapsedSecs(from: string): number {
  return Math.max(0, Math.floor((Date.now() - new Date(from).getTime()) / 1000))
}

function fmtHM(secs: number): { h: number; m: number } {
  return { h: Math.floor(secs / 3600), m: Math.floor((secs % 3600) / 60) }
}

export default function SleepScreen() {
  const { use24h } = useTimeFormatContext()
  const fmtTime = makeFmtTime(use24h)
  const formatDisplayTime = makeFormatDisplayTime(use24h)
  const { family } = useAuthContext()
  const babyId = family?.baby.id ?? ''
  const babyName = family?.baby.name ?? 'Baby'
  const { sessions7d: sessions, isLoading, refresh, updateSessions7d: setSessions } = useAppData()

  const [isToggling, setIsToggling] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)
  const [showLogPast, setShowLogPast] = useState(false)
  const [showEditStart, setShowEditStart] = useState(false)
  const [editStartDate, setEditStartDate] = useState<Date>(new Date())
  const [isSavingStart, setIsSavingStart] = useState(false)

  const session = sessions.find((s) => !s.stopped_at) ?? null
  const isActive = session !== null
  const lastCompleted = sessions.find((s) => !!s.stopped_at) ?? null

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 60_000)
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
        setSessions((prev) => [updated, ...prev.filter((s) => s.id !== updated.id)])
      } else {
        const started = await startSleep(babyId)
        const active: SleepSession = {
          id: started.id,
          baby_id: babyId,
          started_at: started.started_at,
          version: started.version,
        }
        setSessions((prev) => [active, ...prev])
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update sleep')
    } finally {
      setIsToggling(false)
    }
  }

  function openEditStart() {
    if (!session) return
    setEditStartDate(new Date(session.started_at))
    setShowEditStart(true)
  }

  async function handleSaveStartTime() {
    if (!session || !babyId || isSavingStart) return
    setIsSavingStart(true)
    setError(null)
    try {
      const updated = await editSleepSession(babyId, session.id, {
        started_at: editStartDate.toISOString(),
        version: session.version,
      })
      setSessions((prev) => [updated, ...prev.filter((s) => s.id !== updated.id)])
      setShowEditStart(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update start time')
    } finally {
      setIsSavingStart(false)
    }
  }

  if (isLoading) return <div className={styles.loading}>Loading…</div>

  // ── Empty state: brand new family, no sessions yet ───────────────────────
  if (sessions.length === 0) {
    return (
      <div className={styles.screen}>
        <div className={styles.header}>
          <div className={styles.greeting}>{babyName}</div>
        </div>

        <div className={styles.durationBlock}>
          <div className={styles.emptyHeadline}>hasn't slept yet</div>
          <div className={styles.emptySubtitle}>
            Tap the pillow when they doze off. We'll keep track from there.
          </div>
        </div>

        <div className={styles.pillowWrap}>
          <PillowButton label="Start sleep" mascot="awake" onClick={handleToggle} isDisabled={isToggling} />
          <div className={styles.tapToWakeHint}>Tap to start sleep</div>
        </div>

        <button className={styles.editStartPillSoft} onClick={() => setShowLogPast(true)}>
          <svg
            width="14"
            height="14"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <circle cx="10" cy="10" r="7" />
            <path d="M10 6v4l2.5 2" />
          </svg>
          Log past sleep
        </button>

        {error && (
          <p className={styles.error} role="alert">
            {error}
          </p>
        )}

        {showLogPast && (
          <LogPastSleepSheet
            babyId={babyId}
            onSaved={() => {
              setShowLogPast(false)
              refresh()
            }}
            onClose={() => setShowLogPast(false)}
          />
        )}
      </div>
    )
  }

  // ── Asleep / Soft state ──────────────────────────────────────────────────
  if (isActive && session) {
    const sleeping = fmtHM(elapsedSecs(session.started_at))
    const startedAtStr = fmtTime(session.started_at)

    return (
      <div className={styles.screen}>
        <div className={styles.header}>
          <div className={styles.greeting}>{babyName}</div>
        </div>

        <div className={styles.durationBlock}>
          <div className={styles.durationLabel}>Sleeping for</div>
          <div className={styles.durationValue}>
            {sleeping.h}
            <span className={styles.durationUnit}>h </span>
            {String(sleeping.m).padStart(2, '0')}
            <span className={styles.durationUnit}>m</span>
          </div>
          <div className={styles.startedAtSoft}>Started at {startedAtStr}</div>
        </div>

        <div className={styles.pillowWrap}>
          <PillowButton label="Tap to wake" mascot="sleeping" onClick={handleToggle} isDisabled={isToggling} />
          <div className={styles.tapToWakeHint}>Tap to wake · shhh</div>
        </div>

        <button className={styles.editStartPillSoft} onClick={openEditStart}>
          <svg
            width="14"
            height="14"
            viewBox="0 0 16 16"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M3 11l1.5-4.5L9 2l3 3-4.5 4.5L3 11zM2 13h12" />
          </svg>
          Edit start time · {startedAtStr}
        </button>

        {error && (
          <p className={styles.error} role="alert">
            {error}
          </p>
        )}

        {showEditStart && (
          <div
            className={styles.editSheetOverlay}
            onClick={(e) => {
              if (e.target === e.currentTarget) setShowEditStart(false)
            }}
          >
            <div
              className={styles.editSheet}
              role="dialog"
              aria-modal="true"
              aria-label="Edit start time"
            >
              <div className={styles.editSheetHandle} />

              <div className={styles.editSheetHeader}>
                <div>
                  <div className={styles.editSheetTitle}>Sleep in progress</div>
                </div>
              </div>

              <DrumPicker initialDate={editStartDate} onChange={setEditStartDate} />

              <div className={styles.editSheetTimeRows}>
                <div className={styles.editSheetTimeRow}>
                  <div className={styles.editSheetTimeRowInner}>
                    <div>
                      <div className={styles.editSheetTimeRowLabel}>STARTED</div>
                      <div className={styles.editSheetTimeRowValue}>
                        {formatDisplayTime(editStartDate)}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {error && (
                <p className={styles.error} role="alert">
                  {error}
                </p>
              )}

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
                  disabled={isSavingStart}
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

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <div className={styles.greeting}>{babyName}</div>
      </div>

      <div className={styles.durationBlock}>
        <div className={styles.durationLabel}>Active for</div>
        <div className={styles.durationValue}>
          {awake.h}
          <span className={styles.durationUnit}>h </span>
          {String(awake.m).padStart(2, '0')}
          <span className={styles.durationUnit}>m</span>
        </div>
        {lastCompleted && (
          <div className={styles.startedAtSoft}>
            Awake since {fmtTime(lastCompleted.stopped_at!)}
          </div>
        )}
      </div>

      <div className={styles.pillowWrap}>
        <PillowButton label="Tap to sleep" mascot="awake" onClick={handleToggle} isDisabled={isToggling} />
        <div className={styles.tapToWakeHint}>Tap to start sleep</div>
      </div>

      <button className={styles.editStartPillSoft} onClick={() => setShowLogPast(true)}>
        <svg
          width="14"
          height="14"
          viewBox="0 0 20 20"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          aria-hidden="true"
        >
          <circle cx="10" cy="10" r="7" />
          <path d="M10 6v4l2.5 2" />
        </svg>
        Log past sleep
      </button>

      {error && (
        <p className={styles.error} role="alert">
          {error}
        </p>
      )}

      {showLogPast && (
        <LogPastSleepSheet
          babyId={babyId}
          onSaved={() => {
            setShowLogPast(false)
            refresh()
          }}
          onClose={() => setShowLogPast(false)}
        />
      )}
    </div>
  )
}
