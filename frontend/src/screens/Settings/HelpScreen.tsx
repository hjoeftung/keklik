import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import styles from './HelpScreen.module.css'

const HELP_TOPICS = [
  {
    icon: 'pillow',
    q: 'How does sleep tracking work?',
    a: 'Tap the pillow when your little one drifts off. Tap it again when they wake. Keklik figures out whether it counts as a nap or night sleep based on your night window.',
  },
  {
    icon: 'moon',
    q: 'What is the night window?',
    a: "It's the slice of the day Keklik treats as nighttime. Any sleep that starts inside it is counted toward night sleep — everything else is a nap.",
  },
  {
    icon: 'clock',
    q: 'I forgot to track a sleep — can I add it later?',
    a: 'Yes. From the Sleep tab, tap "Log past sleep" and pick a start and end time. The session shows up in your stats right away.',
  },
  {
    icon: 'edit',
    q: 'How do I edit or delete a session?',
    a: "Open the Today or Week tab in Stats, tap any session, and you'll find Edit and Delete buttons in the detail sheet.",
  },
  {
    icon: 'people',
    q: 'Can my partner track too?',
    a: 'Of course. Open Settings → Invite a partner and share the link. Logs sync between you in real time.',
  },
  {
    icon: 'shield',
    q: 'What about my data?',
    a: 'Your sleep logs live on our servers, encrypted in transit and at rest. Export or delete everything any time from Settings.',
  },
]

function BackChevron() {
  return (
    <svg width="11" height="18" viewBox="0 0 11 18" fill="none">
      <path d="M9 1L2 9l7 8" stroke="var(--kk-primary)" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function PlusIcon({ open }: { open: boolean }) {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 14 14"
      style={{ transition: 'transform .25s', transform: open ? 'rotate(45deg)' : 'rotate(0)' }}
    >
      <path d="M7 1v12M1 7h12" stroke="var(--kk-primary-deep)" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}

function TopicIcon({ kind }: { kind: string }) {
  const stroke = 'var(--kk-primary-deep)'
  if (kind === 'pillow') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
        <path d="M4 7c0-2 1.4-2.7 3-2.4 3-1 7-1 10 0 1.6-.3 3 .4 3 2.4.8 5-0 10-.8 10-.4 0-1.2-.1-2.2-.3-3 1-7 1-10 0-1 .2-1.8.3-2.2.3-.8 0-1.6-5-.8-10z" stroke={stroke} strokeWidth="1.6" />
      </svg>
    )
  }
  if (kind === 'moon') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
        <path d="M17 4a8 8 0 100 16 8 8 0 01-5-7 8 8 0 015-9z" stroke={stroke} strokeWidth="1.6" strokeLinejoin="round" />
      </svg>
    )
  }
  if (kind === 'people') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth="1.6" strokeLinecap="round">
        <circle cx="9" cy="9" r="3.2" />
        <path d="M3 19c0-3.5 2.7-6 6-6s6 2.5 6 6" />
        <circle cx="16.5" cy="8" r="2.4" />
        <path d="M14 13.6c.8-.4 1.6-.6 2.5-.6 2.7 0 4.5 2 4.5 5" />
      </svg>
    )
  }
  if (kind === 'clock') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth="1.6" strokeLinecap="round">
        <circle cx="12" cy="12" r="8.5" />
        <path d="M12 7v5l3.2 2" />
      </svg>
    )
  }
  if (kind === 'edit') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 17v3h3l10-10-3-3L4 17z" />
        <path d="M14 6l3 3" />
      </svg>
    )
  }
  if (kind === 'shield') {
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth="1.6" strokeLinejoin="round">
        <path d="M12 3l8 3v5c0 5-3.6 8.5-8 10-4.4-1.5-8-5-8-10V6l8-3z" />
        <path d="M9 12l2 2 4-4" strokeLinecap="round" />
      </svg>
    )
  }
  return null
}

function DecorativeBanner() {
  return (
    <div className={styles.banner}>
      <div className={styles.bannerCloud}>
        <svg width="84" height="50" viewBox="0 0 100 60">
          <path
            d="M 22 42 Q 8 42, 8 32 Q 8 22, 20 22 Q 22 12, 36 12 Q 48 8, 56 18 Q 70 16, 74 28 Q 88 28, 88 38 Q 88 50, 76 50 L 26 50 Q 18 50, 22 42 Z"
            fill="var(--kk-surface)"
            stroke="var(--kk-border)"
            strokeWidth="2"
          />
        </svg>
      </div>
      <div className={styles.bannerMoon}>
        <svg width="48" height="48" viewBox="0 0 60 60">
          <path d="M 42 12 A 22 22 0 1 0 42 48 A 16 16 0 1 1 42 12 Z" fill="var(--kk-nap)" />
        </svg>
      </div>
      <div className={styles.bannerStar1}>
        <svg width="11" height="11" viewBox="0 0 16 16" className={styles.twinkle} style={{ animationDelay: '0.2s' }}>
          <path d="M 8 1 L 9.4 6.6 L 15 8 L 9.4 9.4 L 8 15 L 6.6 9.4 L 1 8 L 6.6 6.6 Z" fill="var(--kk-nap)" />
        </svg>
      </div>
      <div className={styles.bannerStar2}>
        <svg width="8" height="8" viewBox="0 0 16 16" className={styles.twinkle} style={{ animationDelay: '0.9s' }}>
          <path d="M 8 1 L 9.4 6.6 L 15 8 L 9.4 9.4 L 8 15 L 6.6 9.4 L 1 8 L 6.6 6.6 Z" fill="var(--kk-nap)" />
        </svg>
      </div>
      <div className={styles.bannerStar3}>
        <svg width="9" height="9" viewBox="0 0 16 16" className={styles.twinkle} style={{ animationDelay: '1.4s' }}>
          <path d="M 8 1 L 9.4 6.6 L 15 8 L 9.4 9.4 L 8 15 L 6.6 9.4 L 1 8 L 6.6 6.6 Z" fill="var(--kk-nap)" />
        </svg>
      </div>
      <div className={styles.bannerPillow}>
        <svg width="170" height="136" viewBox="0 0 200 160">
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
      </div>
    </div>
  )
}

function FaqRow({ topic, open, onToggle, last }: { topic: (typeof HELP_TOPICS)[0]; open: boolean; onToggle: () => void; last: boolean }) {
  return (
    <div className={`${styles.faqRow} ${last ? styles.faqRowLast : ''}`}>
      <button className={styles.faqTrigger} onClick={onToggle}>
        <div className={styles.faqIcon}>
          <TopicIcon kind={topic.icon} />
        </div>
        <div className={styles.faqQuestion}>{topic.q}</div>
        <PlusIcon open={open} />
      </button>
      {open && (
        <div className={styles.faqAnswer}>
          {topic.a}
        </div>
      )}
    </div>
  )
}

export default function HelpScreen() {
  const navigate = useNavigate()
  const [openIndex, setOpenIndex] = useState<number | null>(null)

  function toggle(i: number) {
    setOpenIndex(openIndex === i ? null : i)
  }

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <button className={styles.back} onClick={() => navigate('/settings')}>
          <BackChevron />
        </button>
        <h1 className={styles.title}>Help</h1>
      </div>

      <div className={styles.scroll}>
        <DecorativeBanner />

        <div className={styles.greeting}>
          <div className={styles.greetingHeading}>
            Everything you might<br />want to know
          </div>
          <div className={styles.greetingSub}>
            Six common questions and a way<br />to reach us if anything's missing.
          </div>
        </div>

        <div className={styles.sectionLabel}>Frequently asked</div>
        <div className={styles.faqContainer}>
          {HELP_TOPICS.map((t, i) => (
            <FaqRow
              key={i}
              topic={t}
              open={openIndex === i}
              onToggle={() => toggle(i)}
              last={i === HELP_TOPICS.length - 1}
            />
          ))}
        </div>

        <div className={styles.contactCard}>
          <div className={styles.contactIcon}>
            <svg width="26" height="26" viewBox="0 0 26 26" fill="none" stroke="var(--kk-primary-deep)" strokeWidth="1.8" strokeLinejoin="round" strokeLinecap="round">
              <rect x="3" y="6" width="20" height="14" rx="2.5" />
              <path d="M4 7l9 7 9-7" />
            </svg>
          </div>
          <div className={styles.contactBody}>
            <div className={styles.contactTitle}>Get in touch</div>
            <div className={styles.contactSub}>
              <a href="mailto:mioqon@gmail.com" className={styles.contactEmail}>mioqon@gmail.com</a>
              {' · usually replies within a day'}
            </div>
          </div>
        </div>

        <div className={styles.version}>Keklik · v0.4.0</div>
      </div>
    </div>
  )
}
