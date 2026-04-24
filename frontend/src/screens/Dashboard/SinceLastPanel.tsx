import { type SinceLastSummary } from '@/api/endpoints'
import { formatDuration } from '@/utils/time'

interface SinceLastPanelProps {
  sinceLast: SinceLastSummary | null
}

export default function SinceLastPanel({ sinceLast }: SinceLastPanelProps) {
  const sleepStart =
    sinceLast?.since_sleep_start_seconds != null
      ? formatDuration(sinceLast.since_sleep_start_seconds)
      : 'No data yet'

  const awakening =
    sinceLast?.since_awakening_seconds != null
      ? formatDuration(sinceLast.since_awakening_seconds)
      : 'No data yet'

  return (
    <section>
      <p>Since last sleep started: {sleepStart}</p>
      <p>Since last awakening: {awakening}</p>
    </section>
  )
}
