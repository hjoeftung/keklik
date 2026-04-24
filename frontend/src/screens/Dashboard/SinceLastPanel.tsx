import { formatDuration } from '@/utils/time'

interface SinceLastPanelProps {
  timeSinceSleepStart: number | null
  timeSinceAwakening: number | null
}

export default function SinceLastPanel({ timeSinceSleepStart, timeSinceAwakening }: SinceLastPanelProps) {
  const sleepStart =
    timeSinceSleepStart != null
      ? formatDuration(timeSinceSleepStart)
      : 'No data yet'

  const awakening =
    timeSinceAwakening != null
      ? formatDuration(timeSinceAwakening)
      : 'No data yet'

  return (
    <section>
      <p>Since last sleep started: {sleepStart}</p>
      <p>Since last awakening: {awakening}</p>
    </section>
  )
}
