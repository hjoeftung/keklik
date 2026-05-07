import { useState, useEffect } from 'react'
import { formatDuration } from '@/utils/time'

interface ElapsedTimerProps {
  startedAt: string
}

export default function ElapsedTimer({ startedAt }: ElapsedTimerProps) {
  const [elapsed, setElapsed] = useState(() =>
    Math.floor((Date.now() - new Date(startedAt).getTime()) / 1000),
  )

  useEffect(() => {
    const id = setInterval(() => {
      setElapsed(Math.floor((Date.now() - new Date(startedAt).getTime()) / 1000))
    }, 1000)
    return () => clearInterval(id)
  }, [startedAt])

  return <span>{formatDuration(elapsed)}</span>
}
