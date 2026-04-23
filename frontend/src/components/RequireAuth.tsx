import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'

export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, isLoading } = useAuthContext()

  if (isLoading) return null
  if (!user) return <Navigate to="/" replace />

  return <>{children}</>
}
