import { useState, useEffect, useRef, useMemo } from 'react'
import styles from './DrumPicker.module.css'

const ITEM_H = 44
const VISIBLE_ROWS = 5
const PAD_ITEMS = Math.floor(VISIBLE_ROWS / 2)

function pad2(n: number): string {
  return String(n).padStart(2, '0')
}

function buildDayItems(maxDaysBack: number): { date: Date; label: string }[] {
  const items: { date: Date; label: string }[] = []
  const today = new Date()
  for (let i = 0; i <= maxDaysBack; i++) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const label =
      i === 0
        ? 'Today'
        : i === 1
          ? 'Yesterday'
          : d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
    items.push({ date: d, label })
  }
  return items
}

function buildHours(): string[] {
  return Array.from({ length: 24 }, (_, i) => pad2(i))
}

function buildMinutes(): string[] {
  return Array.from({ length: 60 }, (_, i) => pad2(i))
}

interface DrumColumnProps {
  items: string[]
  value: string
  onChange: (v: string) => void
  width: number
  label: string
}

function DrumColumn({ items, value, onChange, width, label }: DrumColumnProps) {
  const ref = useRef<HTMLDivElement>(null)
  const ignoreScroll = useRef(false)
  const scrollTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const startY = useRef(0)
  const startScrollTop = useRef(0)

  const padded = [
    ...Array<null>(PAD_ITEMS).fill(null),
    ...items,
    ...Array<null>(PAD_ITEMS).fill(null),
  ]

  const selectedIdx = items.indexOf(value)

  useEffect(() => {
    if (!ref.current) return
    ignoreScroll.current = true
    ref.current.scrollTop = selectedIdx * ITEM_H
    setTimeout(() => {
      ignoreScroll.current = false
    }, 50)
  }, [value, selectedIdx])

  useEffect(() => {
    return () => {
      if (scrollTimeout.current) clearTimeout(scrollTimeout.current)
    }
  }, [])

  function snapToNearest() {
    if (!ref.current) return
    const raw = ref.current.scrollTop
    const idx = Math.round(raw / ITEM_H)
    const clamped = Math.max(0, Math.min(idx, items.length - 1))
    ref.current.scrollTop = clamped * ITEM_H
    if (items[clamped] !== value) onChange(items[clamped])
  }

  function handleScroll() {
    if (ignoreScroll.current) return
    if (scrollTimeout.current) clearTimeout(scrollTimeout.current)
    scrollTimeout.current = setTimeout(snapToNearest, 120)
  }

  function onTouchStart(e: React.TouchEvent) {
    startY.current = e.touches[0].clientY
    startScrollTop.current = ref.current!.scrollTop
  }

  function onTouchMove(e: React.TouchEvent) {
    const dy = startY.current - e.touches[0].clientY
    ref.current!.scrollTop = startScrollTop.current + dy
  }

  function onTouchEnd() {
    snapToNearest()
  }

  return (
    <div className={styles.columnWrapper}>
      {label && <div className={styles.columnLabel}>{label}</div>}
      <div className={styles.columnOuter} style={{ width }}>
        <div className={styles.fadeTop} />
        <div className={styles.fadeBottom} />
        <div className={styles.highlight} />
        <div
          ref={ref}
          className={styles.scroll}
          onScroll={handleScroll}
          onTouchStart={onTouchStart}
          onTouchMove={onTouchMove}
          onTouchEnd={onTouchEnd}
        >
          {padded.map((item, i) => {
            const isSelected = item !== null && item === value
            return (
              <div
                key={i}
                className={isSelected ? styles.itemSelected : styles.item}
                style={item === null ? { cursor: 'default' } : undefined}
                onClick={() => {
                  if (item === null || !ref.current) return
                  ignoreScroll.current = true
                  ref.current.scrollTop = items.indexOf(item) * ITEM_H
                  setTimeout(() => {
                    ignoreScroll.current = false
                  }, 50)
                  onChange(item)
                }}
              >
                {item !== null ? item : ''}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

export interface DrumPickerProps {
  initialDate: Date
  onChange: (d: Date) => void
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

function buildYears(): string[] {
  const currentYear = new Date().getFullYear()
  return Array.from({ length: 6 }, (_, i) => String(currentYear - i))
}

function buildDays(): string[] {
  return Array.from({ length: 31 }, (_, i) => pad2(i + 1))
}

export interface BirthdayPickerProps {
  initialDate: Date
  onChange: (d: Date) => void
}

export function BirthdayPicker({ initialDate, onChange }: BirthdayPickerProps) {
  const years = useMemo(buildYears, [])
  const days = useMemo(buildDays, [])

  const [month, setMonth] = useState(() => MONTHS[initialDate.getMonth()])
  const [day, setDay] = useState(() => pad2(initialDate.getDate()))
  const [year, setYear] = useState(() => String(initialDate.getFullYear()))

  const onChangeRef = useRef(onChange)
  useEffect(() => {
    onChangeRef.current = onChange
  })

  useEffect(() => {
    const m = MONTHS.indexOf(month)
    const result = new Date(parseInt(year, 10), m, parseInt(day, 10))
    onChangeRef.current(result)
  }, [month, day, year])

  return (
    <div className={styles.wrapper}>
      <div className={styles.columns}>
        <DrumColumn items={MONTHS} value={month} onChange={setMonth} width={72} label="MONTH" />
        <DrumColumn items={days} value={day} onChange={setDay} width={60} label="DAY" />
        <DrumColumn items={years} value={year} onChange={setYear} width={76} label="YEAR" />
      </div>
    </div>
  )
}

export interface TimePickerProps {
  value: string
  onChange: (v: string) => void
}

export function TimePicker({ value, onChange }: TimePickerProps) {
  const hourItems = useMemo(buildHours, [])
  const minItems = useMemo(buildMinutes, [])

  const [hour, setHour] = useState(() => value.split(':')[0])
  const [minute, setMinute] = useState(() => value.split(':')[1])

  const onChangeRef = useRef(onChange)
  useEffect(() => {
    onChangeRef.current = onChange
  })

  useEffect(() => {
    onChangeRef.current(`${hour}:${minute}`)
  }, [hour, minute])

  return (
    <div className={styles.wrapper}>
      <div className={styles.columns}>
        <DrumColumn items={hourItems} value={hour} onChange={setHour} width={64} label="HOUR" />
        <div className={styles.colon}>:</div>
        <DrumColumn items={minItems} value={minute} onChange={setMinute} width={64} label="MIN" />
      </div>
    </div>
  )
}

export default function DrumPicker({ initialDate, onChange }: DrumPickerProps) {
  const dayItems = useMemo(() => buildDayItems(30), [])
  const hourItems = useMemo(buildHours, [])
  const minItems = useMemo(buildMinutes, [])
  const dayLabels = useMemo(() => dayItems.map((d) => d.label), [dayItems])

  const [day, setDay] = useState(() => {
    const today = new Date()
    const yest = new Date(today)
    yest.setDate(today.getDate() - 1)
    if (initialDate.toDateString() === today.toDateString()) return 'Today'
    if (initialDate.toDateString() === yest.toDateString()) return 'Yesterday'
    return initialDate.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    })
  })
  const [hour, setHour] = useState(() => pad2(initialDate.getHours()))
  const [minute, setMinute] = useState(() => pad2(initialDate.getMinutes()))

  const onChangeRef = useRef(onChange)
  useEffect(() => {
    onChangeRef.current = onChange
  })

  useEffect(() => {
    const dayObj = dayItems.find((d) => d.label === day)
    if (!dayObj) return
    const result = new Date(dayObj.date)
    result.setHours(parseInt(hour, 10), parseInt(minute, 10), 0, 0)
    onChangeRef.current(result)
  }, [day, hour, minute, dayItems])

  return (
    <div className={styles.wrapper}>
      <div className={styles.columns}>
        <DrumColumn items={dayLabels} value={day} onChange={setDay} width={130} label="DAY" />
        <div className={styles.dotDivider}>·</div>
        <DrumColumn items={hourItems} value={hour} onChange={setHour} width={64} label="HOUR" />
        <div className={styles.colon}>:</div>
        <DrumColumn items={minItems} value={minute} onChange={setMinute} width={64} label="MIN" />
      </div>
    </div>
  )
}
