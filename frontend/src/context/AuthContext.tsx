import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { getAccountId, clearAccountId, ApiError } from '@/api/client'
import { logout, getFamily, type GetFamilyResponse } from '@/api/endpoints'

export interface User {
  accountId: string
}

interface AuthContextValue {
  user: User | null
  family: GetFamilyResponse | null
  isLoading: boolean
  signOut: () => Promise<void>
  refreshFamily: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [family, setFamily] = useState<GetFamilyResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function init() {
      const accountId = getAccountId()
      if (!accountId) {
        setIsLoading(false)
        return
      }
      setUser({ accountId })
      try {
        const familyData = await getFamily()
        setFamily(familyData)
      } catch (err) {
        if (!(err instanceof ApiError) || err.status !== 404) {
          // 401 is handled by client.ts (page redirects); 404 means no family yet
        }
      }
      setIsLoading(false)
    }
    init()
  }, [])

  async function refreshFamily() {
    const data = await getFamily()
    setFamily(data)
  }

  async function signOut() {
    try {
      await logout()
    } catch {
      // proceed regardless of API failure
    }
    clearAccountId()
    setUser(null)
    setFamily(null)
  }

  return (
    <AuthContext.Provider value={{ user, family, isLoading, signOut, refreshFamily }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuthContext must be used within AuthProvider')
  return ctx
}
