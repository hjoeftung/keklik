import { useAuthContext } from '@/context/AuthContext'

export function useAuth() {
  const { user, isLoading, signOut } = useAuthContext()
  return { user, isLoading, signOut }
}
