import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'

export function RequireNoFamily({ children }: { children: ReactNode }) {
  const { user, family, isLoading } = useAuthContext()

  if (isLoading) return null
  if (!user) return <Navigate to="/" replace />
  if (family) return <Navigate to="/sleep" replace />

  return <>{children}</>
}
