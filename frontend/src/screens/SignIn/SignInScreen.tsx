import { useState } from 'react'
import { getGoogleOAuthStartUrl } from '@/api/endpoints'
import styles from './SignInScreen.module.css'

function PillowHero() {
  return (
    <svg width="200" height="160" viewBox="0 0 200 160" aria-hidden="true" overflow="visible">
      <path
        d="M 28 32 Q 28 16, 48 18 Q 100 9, 152 18 Q 172 16, 172 32 Q 180 80, 172 128 Q 172 144, 152 142 Q 100 151, 48 142 Q 28 144, 28 128 Q 20 80, 28 32 Z"
        fill="var(--kk-primary)"
        stroke="var(--kk-primary-deep)"
        strokeWidth="2"
        strokeOpacity="0.2"
      />
      <path
        d="M 32 32 Q 32 22, 50 24 Q 100 16, 150 24 Q 168 22, 168 32"
        fill="none"
        stroke="white"
        strokeWidth="3"
        strokeLinecap="round"
        strokeOpacity="0.4"
      />
      <circle cx="42" cy="34" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
      <circle cx="158" cy="34" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
      <circle cx="42" cy="126" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
      <circle cx="158" cy="126" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
    </svg>
  )
}

function CloudDecor({ size }: { size: number }) {
  return (
    <svg
      width={size}
      height={size * 0.6}
      viewBox="0 0 80 48"
      aria-hidden="true"
      style={{ display: 'block' }}
    >
      <ellipse cx="40" cy="34" rx="34" ry="14" fill="var(--kk-primary-soft)" />
      <ellipse cx="28" cy="28" rx="18" ry="16" fill="var(--kk-primary-soft)" />
      <ellipse cx="52" cy="26" rx="14" ry="12" fill="var(--kk-primary-soft)" />
      <ellipse cx="40" cy="34" rx="34" ry="14" fill="white" opacity="0.4" />
    </svg>
  )
}

function MoonDecor({ size }: { size: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 52 52" aria-hidden="true">
      <defs>
        <mask id="wsMoonMask">
          <rect width="52" height="52" fill="white" />
          <circle cx="36" cy="18" r="20" fill="black" />
        </mask>
      </defs>
      <circle cx="24" cy="26" r="20" fill="#F5E6B8" mask="url(#wsMoonMask)" />
    </svg>
  )
}

function StarDecor({ size, delay }: { size: number; delay: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 14 14"
      aria-hidden="true"
      style={{ animation: `twinkle 2.4s ease-in-out ${delay}s infinite` }}
    >
      <circle cx="7" cy="7" r="2" fill="var(--kk-nap)" />
      <line x1="7" y1="1" x2="7" y2="4" stroke="var(--kk-nap)" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="7" y1="10" x2="7" y2="13" stroke="var(--kk-nap)" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="1" y1="7" x2="4" y2="7" stroke="var(--kk-nap)" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="10" y1="7" x2="13" y2="7" stroke="var(--kk-nap)" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  )
}

export default function SignInScreen() {
  const [isSigningIn, setIsSigningIn] = useState(false)

  function handleSignIn() {
    setIsSigningIn(true)
    window.location.href = getGoogleOAuthStartUrl()
  }

  return (
    <div className={styles.screen}>
      <div className={styles.inner}>
        <div className={styles.hero}>
          <div className={styles.cloud}>
            <CloudDecor size={80} />
          </div>
          <div className={styles.moon}>
            <MoonDecor size={52} />
          </div>
          <div className={styles.starA}>
            <StarDecor size={14} delay={0} />
          </div>
          <div className={styles.starB}>
            <StarDecor size={10} delay={0.6} />
          </div>
          <div className={styles.pillowWrap}>
            <PillowHero />
          </div>
        </div>

        <h1 className={styles.title}>
          A cozy little
          <br />
          sleep tracker
        </h1>
        <p className={styles.subtitle}>
          Tap a pillow when baby naps.
          <br />
          Watch the day fill up with rest.
        </p>

        <div className={styles.bottom}>
          <button className={styles.button} onClick={handleSignIn} disabled={isSigningIn}>
            {isSigningIn ? (
              <>
                <span className={styles.spinner} aria-hidden="true" />
                Signing in…
              </>
            ) : (
              'Continue with Google'
            )}
          </button>
          <div className={styles.dots}>
            <span className={styles.dotActive} />
            <span className={styles.dotInactive} />
          </div>
        </div>
      </div>
    </div>
  )
}
