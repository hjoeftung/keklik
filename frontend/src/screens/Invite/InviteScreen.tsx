import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { joinFamilyByInviteLink } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './InviteScreen.module.css'

const INVITE_TOKEN_KEY = 'keklik_invite_token'

export default function InviteScreen() {
  const { token } = useParams<{ token: string }>()
  const { user, isLoading, refreshFamily } = useAuthContext()
  const navigate = useNavigate()

  const [memberName, setMemberName] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expired, setExpired] = useState(false)
  const [invalidToken, setInvalidToken] = useState(false)

  useEffect(() => {
    if (isLoading) return
    if (!user) {
      if (token) sessionStorage.setItem(INVITE_TOKEN_KEY, token)
      navigate('/', { replace: true })
    }
  }, [isLoading, user, token, navigate])

  if (isLoading || !user) return null

  if (expired) {
    return (
      <div className={styles.screen}>
        <div className={styles.card}>
          <p className={styles.errorTitle}>Invite link expired</p>
          <p className={styles.message}>
            This invite link is no longer valid. Ask the person who invited you to send a new one.
          </p>
          <Link to="/" className={styles.link}>
            Back to home
          </Link>
        </div>
      </div>
    )
  }

  if (invalidToken) {
    return (
      <div className={styles.screen}>
        <div className={styles.card}>
          <p className={styles.errorTitle}>Invalid invite link</p>
          <p className={styles.message}>
            This invite link doesn't exist. Check the link and try again.
          </p>
          <Link to="/" className={styles.link}>
            Back to home
          </Link>
        </div>
      </div>
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!token) return
    setError(null)
    setIsSubmitting(true)

    try {
      await joinFamilyByInviteLink({ token, member_name: memberName })
      await refreshFamily()
      navigate('/', { replace: true })
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'invalid_invite_link') {
          setExpired(true)
          return
        }
        if (err.code === 'not_found') {
          setInvalidToken(true)
          return
        }
        if (err.code === 'conflict') {
          await refreshFamily()
          navigate('/', { replace: true })
          return
        }
        setError(err.message)
      } else {
        setError('Something went wrong. Please try again.')
      }
      setIsSubmitting(false)
    }
  }

  return (
    <div className={styles.screen}>
      <form className={styles.card} onSubmit={handleSubmit} noValidate>
        <div>
          <h1 className={styles.heading}>You've been invited to join a family</h1>
          <p className={styles.sub}>Enter your name to get started.</p>
        </div>

        {error && (
          <div className={styles.error} role="alert">
            {error}
          </div>
        )}

        <div className={styles.field}>
          <label className={styles.label} htmlFor="memberName">
            Your name
          </label>
          <input
            id="memberName"
            className={styles.input}
            type="text"
            value={memberName}
            onChange={(e) => setMemberName(e.target.value)}
            required
            autoComplete="name"
            autoFocus
          />
        </div>

        <button
          className={styles.button}
          type="submit"
          disabled={isSubmitting || !memberName.trim()}
        >
          {isSubmitting ? 'Joining…' : 'Join family'}
        </button>
      </form>
    </div>
  )
}
