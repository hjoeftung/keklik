import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/context/AuthContext'
import { AppDataProvider } from '@/context/AppDataContext'
import { RequireAuth } from '@/components/RequireAuth'
import { RequireNoFamily } from '@/components/RequireNoFamily'
import { AuthenticatedLayout } from '@/components/NavBar/AuthenticatedLayout'
import SignInScreen from '@/screens/SignIn/SignInScreen'
import AuthCallbackScreen from '@/screens/AuthCallback/AuthCallbackScreen'
import OnboardingScreen from '@/screens/Onboarding/OnboardingScreen'
import InviteScreen from '@/screens/Invite/InviteScreen'
import SleepScreen from '@/screens/Sleep/SleepScreen'
import StatisticsScreen from '@/screens/Statistics/StatisticsScreen'
import SettingsScreen from '@/screens/Settings/SettingsScreen'
import NightWindowScreen from '@/screens/Settings/NightWindowScreen'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppDataProvider>
          <Routes>
            <Route path="/" element={<SignInScreen />} />
            <Route path="/auth/callback" element={<AuthCallbackScreen />} />
            <Route path="/sleep" element={<RequireAuth><AuthenticatedLayout><SleepScreen /></AuthenticatedLayout></RequireAuth>} />
            <Route path="/onboarding" element={<RequireNoFamily><OnboardingScreen /></RequireNoFamily>} />
            <Route path="/invite/:token" element={<InviteScreen />} />
            <Route path="/statistics" element={<RequireAuth><AuthenticatedLayout><StatisticsScreen /></AuthenticatedLayout></RequireAuth>} />
            <Route path="/settings" element={<RequireAuth><AuthenticatedLayout><SettingsScreen /></AuthenticatedLayout></RequireAuth>} />
            <Route path="/settings/night-window" element={<RequireAuth><AuthenticatedLayout><NightWindowScreen /></AuthenticatedLayout></RequireAuth>} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AppDataProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
