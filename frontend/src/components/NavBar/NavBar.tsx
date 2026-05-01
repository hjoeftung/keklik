import { NavLink } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import './NavBar.css'

function IconSleep({ active }: { active: boolean }) {
  return (
    <svg
      className="navbar__icon"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M5 9 Q5 5.5, 8 6 Q12 4.5, 16 6 Q19 5.5, 19 9 Q21 12, 19 15 Q19 17, 16 16.5 Q12 18, 8 16.5 Q5 17, 5 15 Q3 12, 5 9 Z"
        stroke="currentColor"
        strokeWidth="1.8"
        fill={active ? 'currentColor' : 'none'}
        fillOpacity={active ? 0.18 : 0}
        strokeLinejoin="round"
      />
    </svg>
  )
}

function IconStats({ active }: { active: boolean }) {
  return (
    <svg
      className="navbar__icon"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <rect x="4" y="13" width="4" height="7" rx="1.2" fill={active ? 'currentColor' : 'none'} fillOpacity={active ? 0.18 : 0} />
      <rect x="10" y="9" width="4" height="11" rx="1.2" fill={active ? 'currentColor' : 'none'} fillOpacity={active ? 0.18 : 0} />
      <rect x="16" y="5" width="4" height="15" rx="1.2" fill={active ? 'currentColor' : 'none'} fillOpacity={active ? 0.18 : 0} />
    </svg>
  )
}

function IconSettings({ active }: { active: boolean }) {
  return (
    <svg
      className="navbar__icon"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="3" fill={active ? 'currentColor' : 'none'} fillOpacity={active ? 0.18 : 0} />
      <path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4L7 17M17 7l1.4-1.4" />
    </svg>
  )
}

function IconSignOut() {
  return (
    <svg
      className="navbar__icon"
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M8 3H4a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h4"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path
        d="M13 6l4 4-4 4M17 10H8"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function NavBar() {
  const { signOut } = useAuthContext()

  function getLinkClass({ isActive }: { isActive: boolean }) {
    return isActive ? 'navbar__link navbar__link--active' : 'navbar__link'
  }

  return (
    <nav className="navbar" aria-label="Main navigation">
      <span className="navbar__brand">Keklik</span>

      <div className="navbar__links">
        <NavLink to="/sleep" className={getLinkClass} end>
          {({ isActive }) => (
            <>
              <IconSleep active={isActive} />
              <span className="navbar__label">Sleep</span>
            </>
          )}
        </NavLink>

        <NavLink to="/timeline" className={getLinkClass}>
          {({ isActive }) => (
            <>
              <IconStats active={isActive} />
              <span className="navbar__label">Stats</span>
            </>
          )}
        </NavLink>

        <NavLink to="/settings" className={getLinkClass}>
          {({ isActive }) => (
            <>
              <IconSettings active={isActive} />
              <span className="navbar__label">Settings</span>
            </>
          )}
        </NavLink>
      </div>

      <button
        type="button"
        className="navbar__signout"
        onClick={signOut}
        aria-label="Sign out"
      >
        <IconSignOut />
        <span className="navbar__label">Sign out</span>
      </button>
    </nav>
  )
}
