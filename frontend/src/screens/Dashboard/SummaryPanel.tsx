import { type DayStats, type RollingStats } from '@/api/endpoints'
import { formatDuration } from '@/utils/time'

interface SummaryPanelProps {
  today: DayStats
  rolling7d: RollingStats
  rolling14d: RollingStats
}

export default function SummaryPanel({ today, rolling7d, rolling14d }: SummaryPanelProps) {
  return (
    <section
      style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: '8px',
      }}
    >
      <div>
        <p>Today's sleep</p>
        <p>{formatDuration(today.total_sleep_seconds)}</p>
      </div>
      <div>
        <p>Today's active</p>
        <p>{formatDuration(today.total_active_seconds)}</p>
      </div>
      <div>
        <p>7-day avg sleep</p>
        <p>{formatDuration(rolling7d.avg_daily_sleep_seconds)}</p>
      </div>
      <div>
        <p>7-day avg active</p>
        <p>{formatDuration(rolling7d.avg_daily_active_seconds)}</p>
      </div>
      <div>
        <p>14-day avg sleep</p>
        <p>{formatDuration(rolling14d.avg_daily_sleep_seconds)}</p>
      </div>
      <div>
        <p>14-day avg active</p>
        <p>{formatDuration(rolling14d.avg_daily_active_seconds)}</p>
      </div>
    </section>
  )
}
