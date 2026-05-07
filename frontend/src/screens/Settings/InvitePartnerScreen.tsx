import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { createInviteLink, revokeInviteLink } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import styles from './InvitePartnerScreen.module.css'

function formatExpiry(expiresAt: string): string {
  const ms = new Date(expiresAt).getTime() - Date.now()
  if (ms <= 0) return 'Expired'
  const hours = Math.round(ms / 1000 / 60 / 60)
  if (hours < 1) return 'Expires in less than an hour'
  if (hours === 1) return 'Expires in 1 hour'
  if (hours < 24) return `Expires in ${hours} hours`
  const days = Math.round(hours / 24)
  return `Expires in ${days} day${days !== 1 ? 's' : ''}`
}

export default function InvitePartnerScreen() {
  const navigate = useNavigate()

  const [inviteUrl, setInviteUrl] = useState<string | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [expiresAt, setExpiresAt] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isRevoking, setIsRevoking] = useState(false)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function fetchLink() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await createInviteLink()
      setToken(res.token)
      setInviteUrl(res.invite_url)
      setExpiresAt(res.expires_at)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Could not generate invite link. Please try again.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchLink()
  }, [])

  async function handleCopy() {
    if (!inviteUrl) return
    await navigator.clipboard.writeText(inviteUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleRevoke() {
    if (!token) return
    setIsRevoking(true)
    setError(null)
    try {
      await revokeInviteLink(token)
      await fetchLink()
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Could not revoke the link. Please try again.')
      }
    } finally {
      setIsRevoking(false)
    }
  }

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <button className={styles.back} onClick={() => navigate('/settings')}>
          ‹
        </button>
        <h1 className={styles.title}>Invite partner</h1>
      </div>

      {isLoading ? (
        <div className={styles.loading}>Generating link…</div>
      ) : error ? (
        <div className={styles.card}>
          <div className={styles.error} role="alert">
            {error}
          </div>
          <button className={styles.button} onClick={fetchLink}>
            Try again
          </button>
        </div>
      ) : (
        <div className={styles.card}>
          <p className={styles.hint}>
            Share this link with your partner. They'll be able to join your family and start
            tracking sleep together.
          </p>

          <div className={styles.linkBox}>
            <input
              className={styles.linkInput}
              type="text"
              readOnly
              value={inviteUrl ?? ''}
              onFocus={(e) => e.target.select()}
            />
            <button className={styles.copyBtn} onClick={handleCopy}>
              {copied ? 'Copied!' : 'Copy link'}
            </button>
          </div>

          {expiresAt && <p className={styles.expiry}>{formatExpiry(expiresAt)}</p>}

          <button className={styles.revokeBtn} onClick={handleRevoke} disabled={isRevoking}>
            {isRevoking ? 'Revoking…' : 'Revoke and generate new link'}
          </button>
        </div>
      )}
    </div>
  )
}
