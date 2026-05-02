import styles from './DatePickerStrip.module.css'

interface Props {
  selectedDate: Date
  onChange: (date: Date) => void
}

function sameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
}

export default function DatePickerStrip({ selectedDate, onChange }: Props) {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  const days: Date[] = []
  for (let i = 6; i >= 0; i--) {
    days.push(new Date(today.getFullYear(), today.getMonth(), today.getDate() - i))
  }

  return (
    <div className={styles.strip}>
      {days.map(day => {
        const isToday = sameDay(day, today)
        const isSelected = sameDay(day, selectedDate)
        const chipClass = [
          styles.chip,
          isSelected ? styles.chipSelected :
          isToday ? styles.chipTodayUnselected :
          styles.chipDefault,
        ].filter(Boolean).join(' ')

        return (
          <button
            key={day.getTime()}
            className={chipClass}
            onClick={() => onChange(day)}
          >
            <span className={`${styles.dayNum} kk-num`}>{day.getDate()}</span>
            <span className={styles.weekday}>
              {day.toLocaleDateString([], { weekday: 'short' })}
            </span>
          </button>
        )
      })}
    </div>
  )
}
