import ElapsedTimer from '@/components/ElapsedTimer'
import { type ActiveSessionSummary } from '@/api/endpoints'

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
    <div>
      <button onClick={onToggle} disabled={isToggling}>
        {isToggling ? 'Loading…' : isSleeping ? 'Stop sleep' : 'Start sleep'}
      </button>
      {isSleeping && activeSession && (
        <ElapsedTimer startedAt={activeSession.started_at} />
      )}
      {error && <p role="alert">{error}</p>}
    </div>
  )
}
