import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import styles from './CaregiversScreen.module.css'

export default function CaregiversScreen() {
  const navigate = useNavigate()
  const { user, family } = useAuthContext()
  const members = family?.members ?? []

  return (
    <div className={styles.screen}>
      <div className={styles.header}>
        <button className={styles.back} onClick={() => navigate('/settings')}>‹</button>
        <h1 className={styles.title}>Caregivers</h1>
      </div>

      <div className={styles.list}>
        {members.map(member => (
          <div key={member.id} className={styles.item}>
            <div className={styles.avatar}>{member.name[0]?.toUpperCase() ?? '?'}</div>
            <span className={styles.name}>{member.name}</span>
            {member.account_id === user?.accountId && (
              <span className={styles.you}>You</span>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
