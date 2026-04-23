import { type ReactNode } from 'react'
import { NavBar } from './NavBar'
import './AuthenticatedLayout.css'

interface AuthenticatedLayoutProps {
  children: ReactNode
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  return (
    <div className="auth-layout">
      <NavBar />
      <main className="auth-layout__content">{children}</main>
    </div>
  )
}
