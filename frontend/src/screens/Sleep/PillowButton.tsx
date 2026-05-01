import styles from './PillowButton.module.css'

interface PillowButtonProps {
  label: string
  masked?: boolean
  isDisabled?: boolean
  onClick: () => void
}

export default function PillowButton({ label, masked, isDisabled, onClick }: PillowButtonProps) {
  return (
    <button
      className={styles.wrapper}
      onClick={onClick}
      disabled={isDisabled}
      aria-label={label}
    >
      <span className={styles.pillow}>
        <svg width="260" height="208" viewBox="0 0 200 160" className={styles.svg} aria-hidden="true">
          {/* Pillow body */}
          <path
            d="M 28 32
               Q 28 16, 48 18
               Q 100 9, 152 18
               Q 172 16, 172 32
               Q 180 80, 172 128
               Q 172 144, 152 142
               Q 100 151, 48 142
               Q 28 144, 28 128
               Q 20 80, 28 32 Z"
            fill="var(--kk-primary)"
            stroke="var(--kk-primary-deep)"
            strokeWidth="2"
            strokeOpacity="0.2"
          />
          {/* Top highlight */}
          <path
            d="M 32 32 Q 32 22, 50 24 Q 100 16, 150 24 Q 168 22, 168 32"
            fill="none"
            stroke="white"
            strokeWidth="3"
            strokeLinecap="round"
            strokeOpacity="0.4"
          />
          {/* Corner dimples */}
          <circle cx="42" cy="34" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
          <circle cx="158" cy="34" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
          <circle cx="42" cy="126" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />
          <circle cx="158" cy="126" r="2.2" fill="var(--kk-primary-deep)" opacity="0.35" />

          {masked ? (
            /* Sleep mask */
            <g>
              <line x1="44" y1="80" x2="60" y2="78" stroke="var(--kk-night)" strokeWidth="3" strokeLinecap="round" opacity="0.85" />
              <line x1="140" y1="78" x2="156" y2="80" stroke="var(--kk-night)" strokeWidth="3" strokeLinecap="round" opacity="0.85" />
              <path
                d="M60 72 Q80 62 100 66 Q120 62 140 72 Q142 82 126 86 Q113 90 100 88 Q87 90 74 86 Q58 82 60 72 Z"
                fill="var(--kk-night)"
                opacity="0.82"
              />
              <ellipse cx="82" cy="78" rx="12" ry="6" fill="none" stroke="rgba(255,255,255,0.2)" strokeWidth="1.5" />
              <ellipse cx="118" cy="78" rx="12" ry="6" fill="none" stroke="rgba(255,255,255,0.2)" strokeWidth="1.5" />
              <text
                x="100"
                y="114"
                textAnchor="middle"
                fontSize="18"
                fontFamily="Quicksand, system-ui, sans-serif"
                fontWeight="600"
                fill="white"
              >
                {label}
              </text>
            </g>
          ) : (
            <text
              x="100"
              y="92"
              textAnchor="middle"
              fontSize="22"
              fontFamily="Quicksand, system-ui, sans-serif"
              fontWeight="600"
              fill="white"
            >
              {label}
            </text>
          )}
        </svg>
      </span>
    </button>
  )
}
