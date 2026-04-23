import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'

const INVITE_TOKEN_KEY = 'keklik_invite_token'

export default function InviteScreen() {
  const { token } = useParams<{ token: string }>()
  const { user, isLoading } = useAuthContext()
  const navigate = useNavigate()

  useEffect(() => {
    if (isLoading) return
    if (!user) {
      if (token) {
        sessionStorage.setItem(INVITE_TOKEN_KEY, token)
      }
      navigate('/', { replace: true })
    }
  }, [isLoading, user, token, navigate])

  if (isLoading || !user) {
    return null
  }

  return <div>Accept Invite — token: {token}</div>
}
