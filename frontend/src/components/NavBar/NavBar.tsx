import { NavLink } from 'react-router-dom'
import { useAuthContext } from '@/context/AuthContext'
import './NavBar.css'

function IconHome() {
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
        d="M3 9.5L10 3l7 6.5V17a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V9.5Z"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path
        d="M7.5 18v-5h5v5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function IconTimeline() {
  return (
    <svg
      className="navbar__icon"
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      aria-hidden="true"
    >
      <rect x="3" y="13" width="3" height="4" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
      <rect x="8.5" y="9" width="3" height="8" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
      <rect x="14" y="5" width="3" height="12" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  )
}

function IconSettings() {
  return (
    <svg
      className="navbar__icon"
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="10" cy="10" r="2.5" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M10 2v1.5M10 16.5V18M2 10h1.5M16.5 10H18M4.1 4.1l1.06 1.06M14.84 14.84l1.06 1.06M15.9 4.1l-1.06 1.06M5.16 14.84l-1.06 1.06"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
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
        <NavLink to="/dashboard" className={getLinkClass} end>
          <IconHome />
          <span className="navbar__label">Dashboard</span>
        </NavLink>

        <NavLink to="/timeline" className={getLinkClass}>
          <IconTimeline />
          <span className="navbar__label">Timeline</span>
        </NavLink>

        <NavLink to="/settings" className={getLinkClass}>
          <IconSettings />
          <span className="navbar__label">Settings</span>
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
