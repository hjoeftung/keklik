import styles from './SignOutDialog.module.css'
import BunMascot from './BunMascot'

interface Props {
  babyName?: string
  onConfirm: () => void
  onCancel: () => void
}

function MoonDecor({ size = 36 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 52 52" aria-hidden="true">
      <defs>
        <mask id="kkMoonMask">
          <rect width="52" height="52" fill="white" />
          <circle cx="36" cy="18" r="20" fill="black" />
        </mask>
      </defs>
      <circle cx="24" cy="26" r="20" fill="#F5E6B8" mask="url(#kkMoonMask)" />
    </svg>
  )
}

function StarDecor({ size = 14, delay = 0, color = '#E8B86E' }: { size?: number; delay?: number; color?: string }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 14 14"
      aria-hidden="true"
      style={{ animation: `kk-twinkle 2.4s ease-in-out ${delay}s infinite` }}
    >
      <circle cx="7" cy="7" r="2" fill={color} />
      <line x1="7" y1="1" x2="7" y2="4" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
      <line x1="7" y1="10" x2="7" y2="13" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
      <line x1="1" y1="7" x2="4" y2="7" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
      <line x1="10" y1="7" x2="13" y2="7" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  )
}

function CloudDecor({ size = 56 }: { size?: number }) {
  return (
    <svg width={size} height={size * 0.6} viewBox="0 0 80 48" aria-hidden="true">
      <ellipse cx="40" cy="34" rx="34" ry="14" fill="#FBE4D1" />
      <ellipse cx="28" cy="28" rx="18" ry="16" fill="#FBE4D1" />
      <ellipse cx="52" cy="26" rx="14" ry="12" fill="#FBE4D1" />
      <ellipse cx="40" cy="34" rx="34" ry="14" fill="white" opacity="0.4" />
    </svg>
  )
}

export function SignOutDialog({ babyName, onConfirm, onCancel }: Props) {
  const name = babyName ?? 'your baby'

  return (
    <div className={styles.scrim} onClick={onCancel}>
      <div className={styles.card} onClick={(e) => e.stopPropagation()}>
        <div className={styles.band} />
        <div className={styles.cloud}><CloudDecor size={56} /></div>
        <div className={styles.moon}><MoonDecor size={36} /></div>
        <div className={styles.star1}><StarDecor size={10} delay={0.4} /></div>
        <div className={styles.star2}><StarDecor size={8} delay={0} color="#A38FC4" /></div>
        <div className={styles.star3}><StarDecor size={9} delay={1.1} /></div>

        <div className={styles.mascotWrap}>
          <BunMascot size={108} sleepy />
        </div>

        <h2 className={styles.heading}>See you soon?</h2>
        <p className={styles.body}>
          We'll keep {name}'s sleep log safe until you're back.
        </p>

        <div className={styles.actions}>
          <button className={styles.primaryBtn} onClick={onCancel}>
            Stay signed in
          </button>
          <button className={styles.dangerBtn} onClick={onConfirm}>
            Sign out
          </button>
        </div>
      </div>
    </div>
  )
}
