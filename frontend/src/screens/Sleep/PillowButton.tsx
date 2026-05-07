import styles from './PillowButton.module.css'

interface PillowButtonProps {
  label: string
  masked?: boolean
  isDisabled?: boolean
  onClick: () => void
}

export default function PillowButton({ label, masked, isDisabled, onClick }: PillowButtonProps) {
  return (
    <button className={styles.wrapper} onClick={onClick} disabled={isDisabled} aria-label={label}>
      <span className={styles.pillow}>
        <svg
          width="260"
          height="208"
          viewBox="0 0 200 160"
          className={styles.svg}
          aria-hidden="true"
          overflow="visible"
        >
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
            <>
              {/* Mascot sleeping in upper-right corner of pillow */}
              <g transform="translate(85, -40) scale(1.2)">
                {/* Hair tufts */}
                <g fill="#FBE4D1" stroke="#C97B4D" strokeWidth="1.4" strokeOpacity="0.3">
                  <path d="M 58 36 C 56 28, 50 24, 48 20 Q 54 26, 60 28 Q 60 32, 58 36 Z" />
                  <path d="M 64 35 C 70 28, 78 28, 80 22 Q 76 30, 72 32 Q 70 36, 64 35 Z" />
                  <path d="M 50 38 C 46 34, 40 36, 38 32 Q 44 36, 48 36 Q 50 37, 50 38 Z" />
                </g>
                {/* Head */}
                <path
                  d="M 60 36 C 38 36, 24 56, 24 78 C 24 100, 40 116, 60 116 C 80 116, 96 100, 96 78 C 96 56, 82 36, 60 36 Z"
                  fill="#FBE4D1"
                  stroke="#E59B6A"
                  strokeWidth="2"
                  strokeOpacity="0.22"
                />
                {/* Left ear */}
                <path
                  d="M 28 74 Q 18 70, 16 64 Q 20 70, 24 72 Q 18 74, 20 78 Q 24 76, 28 76 Z"
                  fill="#FBE4D1"
                  stroke="#E59B6A"
                  strokeWidth="1.4"
                  strokeOpacity="0.25"
                />
                {/* Right ear */}
                <path
                  d="M 92 74 Q 102 70, 104 64 Q 100 70, 96 72 Q 102 74, 100 78 Q 96 76, 92 76 Z"
                  fill="#FBE4D1"
                  stroke="#E59B6A"
                  strokeWidth="1.4"
                  strokeOpacity="0.25"
                />
                {/* Rosy cheeks */}
                <ellipse cx="40" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
                <ellipse cx="78" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
                {/* Closed sleeping eyes */}
                <path
                  d="M 46 69 Q 49 65 52 69"
                  fill="none"
                  stroke="#2E2A33"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
                <path
                  d="M 68 69 Q 71 65 74 69"
                  fill="none"
                  stroke="#2E2A33"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
                {/* Nose */}
                <path
                  d="M 56 78 L 60 84 L 64 78 Q 60 76, 56 78 Z"
                  fill="#E8B86E"
                  stroke="#C97B4D"
                  strokeWidth="1.2"
                  strokeOpacity="0.4"
                  strokeLinejoin="round"
                />
              </g>
              {/* Floating Z's — originate from the top of the mascot's head (SVG coords ~175,5) */}
              <text x="175" y="5" fontSize="22" className={`${styles.zee} ${styles.z1}`}>
                Z
              </text>
              <text x="175" y="5" fontSize="16" className={`${styles.zee} ${styles.z2}`}>
                z
              </text>
              <text x="175" y="5" fontSize="12" className={`${styles.zee} ${styles.z3}`}>
                z
              </text>
            </>
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
