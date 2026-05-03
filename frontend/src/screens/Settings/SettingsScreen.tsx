import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import styles from './SettingsScreen.module.css'

export default function SettingsScreen() {
  const navigate = useNavigate()
  const { family, signOut } = useAuthContext()

  const baby = family?.baby
  const memberCount = family?.members.length ?? 0

  async function handleSignOut() {
    await signOut()
    navigate('/')
  }

  return (
    <div className={styles.screen}>
      <div className={styles.babyCard}>
        <div className={styles.avatar}>
          {baby?.name?.[0]?.toUpperCase() ?? '?'}
        </div>
        <div className={styles.babyInfo}>
          <p className={styles.babyName}>{baby?.name ?? '—'}</p>
          <p className={styles.babySub}>{memberCount} caregiver{memberCount !== 1 ? 's' : ''}</p>
        </div>
      </div>

      <div className={styles.section}>
        <p className={styles.sectionLabel}>Sleep</p>
        <div className={styles.group}>
          <div className={styles.row}>
            <span className={styles.rowLabel}>Night window</span>
            <span className={styles.rowChevron}>›</span>
          </div>
          <div className={styles.divider} />
          <div className={styles.row}>
            <span className={styles.rowLabel}>Day starts at</span>
            <span className={styles.rowChevron}>›</span>
          </div>
          <div className={styles.divider} />
          <div className={styles.row}>
            <span className={styles.rowLabel}>Time format</span>
            <span className={styles.rowChevron}>›</span>
          </div>
        </div>
      </div>

      <div className={styles.section}>
        <p className={styles.sectionLabel}>Family</p>
        <div className={styles.group}>
          <div className={styles.row}>
            <span className={styles.rowLabel}>Caregivers</span>
            <span className={styles.rowValue}>{memberCount}</span>
          </div>
          <div className={styles.divider} />
          <div className={styles.row}>
            <span className={styles.rowLabel}>Invite partner</span>
            <span className={styles.rowChevron}>›</span>
          </div>
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
