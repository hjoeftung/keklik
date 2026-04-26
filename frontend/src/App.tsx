import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/context/AuthContext'
import { RequireAuth } from '@/components/RequireAuth'
import { RequireNoFamily } from '@/components/RequireNoFamily'
import { AuthenticatedLayout } from '@/components/NavBar/AuthenticatedLayout'
import SignInScreen from '@/screens/SignIn/SignInScreen'
import AuthCallbackScreen from '@/screens/AuthCallback/AuthCallbackScreen'
import OnboardingScreen from '@/screens/Onboarding/OnboardingScreen'
import InviteScreen from '@/screens/Invite/InviteScreen'
import SleepScreen from '@/screens/Sleep/SleepScreen'
import TimelineScreen from '@/screens/Timeline/TimelineScreen'
import SettingsScreen from '@/screens/Settings/SettingsScreen'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<SignInScreen />} />
          <Route path="/auth/callback" element={<AuthCallbackScreen />} />
          <Route path="/sleep" element={<RequireAuth><AuthenticatedLayout><SleepScreen /></AuthenticatedLayout></RequireAuth>} />
          <Route path="/onboarding" element={<RequireNoFamily><OnboardingScreen /></RequireNoFamily>} />
          <Route path="/invite/:token" element={<InviteScreen />} />
          <Route path="/timeline" element={<RequireAuth><AuthenticatedLayout><TimelineScreen /></AuthenticatedLayout></RequireAuth>} />
          <Route path="/settings" element={<RequireAuth><AuthenticatedLayout><SettingsScreen /></AuthenticatedLayout></RequireAuth>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
