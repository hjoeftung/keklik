import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import { useTimeFormatContext } from '@/context/TimeFormatContext'
import { getSleepStats } from '@/api/endpoints'
import { hhmmToDisplay } from '@/utils/time'
import styles from './SettingsScreen.module.css'

export default function SettingsScreen() {
  const navigate = useNavigate()
  const { family, signOut } = useAuthContext()
  const { format, setFormat, use24h } = useTimeFormatContext()

  const baby = family?.baby
  const memberCount = family?.members.length ?? 0

  const [nightWindowLabel, setNightWindowLabel] = useState<string | null>(null)

  useEffect(() => {
    if (!baby?.id) return
    getSleepStats(baby.id)
      .then((stats) => {
        if (stats.night_window) {
          const start = hhmmToDisplay(stats.night_window.start_hhmm, use24h)
          const end = hhmmToDisplay(stats.night_window.end_hhmm, use24h)
          setNightWindowLabel(`${start} – ${end}`)
        }
      })
      .catch(() => {})
  }, [baby?.id, use24h])

  async function handleSignOut() {
    await signOut()
    navigate('/')
  }

  return (
    <div className={styles.screen}>
      <div className={styles.babyCard}>
        <div className={styles.avatar}>{baby?.name?.[0]?.toUpperCase() ?? '?'}</div>
        <div className={styles.babyInfo}>
          <p className={styles.babyName}>{baby?.name ?? '—'}</p>
          <p className={styles.babySub}>
            {memberCount} caregiver{memberCount !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      <div className={styles.section}>
        <p className={styles.sectionLabel}>Sleep</p>
        <div className={styles.group}>
          <button className={styles.row} onClick={() => navigate('/settings/night-window')}>
            <span className={styles.rowLabel}>Night window</span>
            <span className={styles.rowRight}>
              {nightWindowLabel && <span className={styles.rowValue}>{nightWindowLabel}</span>}
              <span className={styles.rowChevron}>›</span>
            </span>
          </button>
          <div className={styles.divider} />
          <div className={styles.row}>
            <span className={styles.rowLabel}>Time format</span>
            <div className={styles.formatToggle}>
              <button
                className={`${styles.formatBtn} ${format === '12h' ? styles.formatBtnActive : ''}`}
                onClick={() => setFormat('12h')}
              >
                12h
              </button>
              <button
                className={`${styles.formatBtn} ${format === '24h' ? styles.formatBtnActive : ''}`}
                onClick={() => setFormat('24h')}
              >
                24h
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className={styles.section}>
        <p className={styles.sectionLabel}>Family</p>
        <div className={styles.group}>
          <button className={styles.row} onClick={() => navigate('/settings/caregivers')}>
            <span className={styles.rowLabel}>Caregivers</span>
            <span className={styles.rowRight}>
              <span className={styles.rowValue}>{memberCount}</span>
              <span className={styles.rowChevron}>›</span>
            </span>
          </button>
          <div className={styles.divider} />
          <button className={styles.row} onClick={() => navigate('/settings/invite')}>
            <span className={styles.rowLabel}>Invite partner</span>
            <span className={styles.rowChevron}>›</span>
          </button>
        </div>
      </div>

      <div className={styles.section}>
        <p className={styles.sectionLabel}>About</p>
        <div className={styles.group}>
          <div className={styles.row}>
            <span className={styles.rowLabel}>Help</span>
            <span className={styles.rowChevron}>›</span>
          </div>
          <div className={styles.divider} />
          <button className={styles.signOutRow} onClick={handleSignOut}>
            Sign out
          </button>
        </div>
      </div>

      <p className={styles.version}>Keklik · v0.1</p>
    </div>
  )
}
