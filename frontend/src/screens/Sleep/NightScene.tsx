import styles from './NightScene.module.css'

const STARS = [
  { cx: 18, cy: 12, r: 1.2, delay: 0 },
  { cx: 42, cy: 7, r: 1.5, delay: 0.8 },
  { cx: 65, cy: 15, r: 1.0, delay: 1.6 },
  { cx: 80, cy: 5, r: 1.3, delay: 0.4 },
  { cx: 92, cy: 20, r: 1.1, delay: 2.1 },
  { cx: 10, cy: 30, r: 1.0, delay: 1.2 },
  { cx: 55, cy: 28, r: 1.4, delay: 0.2 },
  { cx: 73, cy: 35, r: 0.9, delay: 1.9 },
  { cx: 88, cy: 40, r: 1.2, delay: 0.7 },
  { cx: 30, cy: 45, r: 1.0, delay: 2.4 },
  { cx: 5, cy: 55, r: 1.3, delay: 0.9 },
  { cx: 48, cy: 52, r: 1.1, delay: 1.5 },
  { cx: 95, cy: 58, r: 1.0, delay: 0.3 },
  { cx: 22, cy: 68, r: 1.2, delay: 1.8 },
  { cx: 70, cy: 62, r: 0.8, delay: 2.7 },
  { cx: 85, cy: 72, r: 1.1, delay: 1.1 },
  { cx: 38, cy: 78, r: 1.3, delay: 0.6 },
  { cx: 12, cy: 82, r: 0.9, delay: 2.2 },
  { cx: 60, cy: 80, r: 1.0, delay: 1.4 },
  { cx: 92, cy: 85, r: 1.2, delay: 0.5 },
]

export default function NightScene() {
  return (
    <div className={styles.scene} aria-hidden="true">
      {/* Star field */}
      <svg className={styles.starfield} viewBox="0 0 100 100" preserveAspectRatio="xMidYMid slice">
        {STARS.map((s, i) => (
          <circle
            key={i}
            cx={s.cx}
            cy={s.cy}
            r={s.r}
            fill="white"
            style={{ animationDelay: `${s.delay}s` }}
            className={styles.star}
          />
        ))}
      </svg>

      {/* Moon — upper right */}
      <div className={styles.moonWrap}>
        <svg viewBox="0 0 60 60" className={styles.moonSvg}>
          {/* Crescent: full circle minus offset circle */}
          <defs>
            <mask id="crescentMask">
              <rect width="60" height="60" fill="white" />
              <circle cx="38" cy="22" r="24" fill="black" />
            </mask>
          </defs>
          <circle cx="28" cy="30" r="22" fill="#F5E6B8" mask="url(#crescentMask)" />
        </svg>
      </div>

      {/* Floating Z's */}
      <div className={styles.zees}>
        <span className={`${styles.zee} ${styles.z1}`}>z</span>
        <span className={`${styles.zee} ${styles.z2}`}>z</span>
        <span className={`${styles.zee} ${styles.z3}`}>Z</span>
      </div>
    </div>
  )
}
