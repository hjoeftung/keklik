import { useState } from 'react'
import { getGoogleOAuthStartUrl } from '@/api/endpoints'
import styles from './SignInScreen.module.css'

export default function SignInScreen() {
  const [isSigningIn, setIsSigningIn] = useState(false)

  function handleSignIn() {
    setIsSigningIn(true)
    window.location.href = getGoogleOAuthStartUrl()
  }

  return (
    <div className={styles.screen}>
      <div className={styles.card}>
        <h1 className={styles.title}>Keklik</h1>
        <p className={styles.subtitle}>Track your baby's sleep</p>
        <button
          className={styles.button}
          onClick={handleSignIn}
          disabled={isSigningIn}
        >
          {isSigningIn ? (
            <>
              <span className={styles.spinner} aria-hidden="true" />
              Signing in…
            </>
          ) : (
            'Sign in with Google'
          )}
        </button>
      </div>
    </div>
  )
}
