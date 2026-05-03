import { useState, useCallback } from 'react'

export type TimeFormat = '12h' | '24h'

const KEY = 'keklik_time_format'

function readStoredFormat(): TimeFormat {
  const v = localStorage.getItem(KEY)
  return v === '24h' ? '24h' : '12h'
}

export function useTimeFormat() {
  const [format, setFormatState] = useState<TimeFormat>(readStoredFormat)

  const setFormat = useCallback((f: TimeFormat) => {
    localStorage.setItem(KEY, f)
    setFormatState(f)
  }, [])

  return { format, setFormat, use24h: format === '24h' }
}
