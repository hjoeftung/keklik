import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate, Link } from 'react-router-dom'
import { setAccountId, ApiError } from '@/api/client'
import { getFamily } from '@/api/endpoints'
import styles from './AuthCallbackScreen.module.css'

const INVITE_TOKEN_KEY = 'keklik_invite_token'

export default function AuthCallbackScreen() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  const accountId = searchParams.get('account_id')

  const hasMissingParams = !accountId

  useEffect(() => {
    if (hasMissingParams) return

    async function processCallback() {
      setAccountId(accountId!)

      const savedInviteToken = sessionStorage.getItem(INVITE_TOKEN_KEY)
      if (savedInviteToken) {
        sessionStorage.removeItem(INVITE_TOKEN_KEY)
        navigate(`/invite/${savedInviteToken}`, { replace: true })
        return
      }

      try {
        await getFamily()
        navigate('/sleep', { replace: true })
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) {
          navigate('/onboarding', { replace: true })
        } else {
          setError('Something went wrong. Please try again.')
        }
      }
    }

    processCallback()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  if (hasMissingParams) {
    return (
      <div className={styles.screen}>
        <div className={styles.card}>
          <p className={styles.errorTitle}>Sign-in failed</p>
          <p className={styles.message}>Missing account information from the sign-in response.</p>
          <Link to="/" className={styles.link}>
            Back to sign in
          </Link>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.screen}>
        <div className={styles.card}>
          <p className={styles.errorTitle}>Sign-in failed</p>
          <p className={styles.message}>{error}</p>
          <Link to="/" className={styles.link}>
            Back to sign in
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.screen}>
      <div className={styles.card}>
        <span className={styles.spinner} aria-hidden="true" />
        <p className={styles.message}>Signing you in…</p>
      </div>
    </div>
  )
}
