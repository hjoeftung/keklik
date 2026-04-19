import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import SignInScreen from '@/screens/SignIn/SignInScreen'
import OnboardingScreen from '@/screens/Onboarding/OnboardingScreen'
import InviteScreen from '@/screens/Invite/InviteScreen'
import DashboardScreen from '@/screens/Dashboard/DashboardScreen'
import TimelineScreen from '@/screens/Timeline/TimelineScreen'
import SettingsScreen from '@/screens/Settings/SettingsScreen'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<DashboardScreen />} />
        <Route path="/signin" element={<SignInScreen />} />
        <Route path="/onboarding" element={<OnboardingScreen />} />
        <Route path="/invite/:token" element={<InviteScreen />} />
        <Route path="/timeline" element={<TimelineScreen />} />
        <Route path="/settings" element={<SettingsScreen />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
