import ElapsedTimer from '@/components/ElapsedTimer'
import { type ActiveSessionSummary } from '@/api/endpoints'
import styles from './SleepControl.module.css'

interface SleepControlProps {
  activeSession: ActiveSessionSummary | null
  isLoading: boolean
  isToggling: boolean
  error: string | null
  onToggle: () => void
}

export default function SleepControl({
  activeSession,
  isLoading,
  isToggling,
  error,
  onToggle,
}: SleepControlProps) {
  if (isLoading) {
    return <p>Loading…</p>
  }

  const isSleeping = Boolean(activeSession)

  return (
    <div className={styles.control}>
      <button
        className={`${styles.button} ${isSleeping ? styles.stop : styles.start}`}
        onClick={onToggle}
        disabled={isToggling}
      >
        {isToggling && <span className={styles.spinner} />}
        {isToggling ? '' : isSleeping ? 'Stop sleep' : 'Start sleep'}
      </button>
      {isSleeping && activeSession && (
        <span className={styles.elapsed}>
          <ElapsedTimer startedAt={activeSession.started_at} />
        </span>
      )}
      {error && <p className={styles.error} role="alert">{error}</p>}
    </div>
  )
}
