import { createContext, useContext, type ReactNode } from 'react'
import { useTimeFormat, type TimeFormat } from '@/hooks/useTimeFormat'

interface TimeFormatContextValue {
  format: TimeFormat
  setFormat: (f: TimeFormat) => void
  use24h: boolean
}

const TimeFormatContext = createContext<TimeFormatContextValue | null>(null)

export function TimeFormatProvider({ children }: { children: ReactNode }) {
  const value = useTimeFormat()
  return <TimeFormatContext.Provider value={value}>{children}</TimeFormatContext.Provider>
}

export function useTimeFormatContext(): TimeFormatContextValue {
  const ctx = useContext(TimeFormatContext)
  if (!ctx) throw new Error('useTimeFormatContext must be used within TimeFormatProvider')
  return ctx
}
