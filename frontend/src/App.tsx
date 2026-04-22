import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/context/AuthContext'
import { RequireAuth } from '@/components/RequireAuth'
import { RequireNoFamily } from '@/components/RequireNoFamily'
import SignInScreen from '@/screens/SignIn/SignInScreen'
import OnboardingScreen from '@/screens/Onboarding/OnboardingScreen'
import InviteScreen from '@/screens/Invite/InviteScreen'
import DashboardScreen from '@/screens/Dashboard/DashboardScreen'
import TimelineScreen from '@/screens/Timeline/TimelineScreen'
import SettingsScreen from '@/screens/Settings/SettingsScreen'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<SignInScreen />} />
          <Route path="/dashboard" element={<RequireAuth><DashboardScreen /></RequireAuth>} />
          <Route path="/onboarding" element={<RequireNoFamily><OnboardingScreen /></RequireNoFamily>} />
          <Route path="/invite/:token" element={<InviteScreen />} />
          <Route path="/timeline" element={<RequireAuth><TimelineScreen /></RequireAuth>} />
          <Route path="/settings" element={<RequireAuth><SettingsScreen /></RequireAuth>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
