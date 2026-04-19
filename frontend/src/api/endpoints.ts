import { api } from './client'

// --- Auth ---

export interface AuthSessionResponse {
  token: string
  account_id: string
}

export function getGoogleOAuthStartUrl(): string {
  return `${import.meta.env.VITE_API_BASE_URL ?? ''}/auth/google/start`
}

// --- Family ---

export interface CreateFamilyRequest {
  baby_name: string
  creator_name: string
}

export interface CreateFamilyResponse {
  family_id: string
  member_id: string
  baby_id: string
}

export function createFamily(req: CreateFamilyRequest): Promise<CreateFamilyResponse> {
  return api.post('/families', req)
}

export interface FamilyMember {
  id: string
  name: string
}

export interface FamilyBaby {
  id: string
  name: string
}

export interface GetFamilyResponse {
  family_id: string
  members: FamilyMember[]
  baby: FamilyBaby
}

export function getFamily(): Promise<GetFamilyResponse> {
  return api.get('/family')
}

// --- Invite links ---

export interface CreateInviteLinkResponse {
  token: string
  invite_url: string
  expires_at: string
}

export function createInviteLink(): Promise<CreateInviteLinkResponse> {
  return api.post('/families/invite-links')
}

export interface JoinFamilyRequest {
  token: string
  member_name: string
}

export interface JoinFamilyResponse {
  family_id: string
  member_id: string
}

export function joinFamilyByInviteLink(req: JoinFamilyRequest): Promise<JoinFamilyResponse> {
  return api.post('/families/join-by-invite-link', req)
}

// --- Sleep sessions ---

export interface StartSleepRequest {
  started_at?: string
}

export interface StartSleepResponse {
  id: string
  started_at: string
}

export function startSleep(babyId: string, req?: StartSleepRequest): Promise<StartSleepResponse> {
  return api.post(`/babies/${babyId}/sleep-sessions/active`, req)
}

export interface StopSleepRequest {
  stopped_at?: string
}

export interface StopSleepResponse {
  id: string
  started_at: string
  stopped_at: string
  classification: string
}

export function stopSleep(babyId: string, req?: StopSleepRequest): Promise<StopSleepResponse> {
  return api.delete(`/babies/${babyId}/sleep-sessions/active`, req)
}

export interface SleepSession {
  id: string
  baby_id: string
  started_at: string
  stopped_at?: string
  classification?: string
  duration_seconds?: number
}

export interface EditSleepSessionRequest {
  started_at?: string
  stopped_at?: string
}

export function editSleepSession(
  babyId: string,
  sessionId: string,
  req: EditSleepSessionRequest,
): Promise<SleepSession> {
  return api.patch(`/babies/${babyId}/sleep-sessions/${sessionId}`, req)
}

export function deleteSleepSession(babyId: string, sessionId: string): Promise<void> {
  return api.delete(`/babies/${babyId}/sleep-sessions/${sessionId}`)
}

export type SleepHistoryPeriod = 'today' | '7d' | '14d'

export function getSleepHistory(
  babyId: string,
  period: SleepHistoryPeriod = '7d',
): Promise<SleepSession[]> {
  return api.get(`/babies/${babyId}/sleep-sessions?period=${period}`)
}

// --- Dashboard summary ---

export interface ActiveSessionSummary {
  id: string
  started_at: string
}

export interface SinceLastSummary {
  since_sleep_start_seconds: number | null
  since_awakening_seconds: number | null
}

export interface DayStats {
  total_sleep_seconds: number
  total_active_seconds: number
}

export interface RollingStats {
  avg_daily_sleep_seconds: number
  avg_daily_active_seconds: number
}

export interface DashboardSummary {
  active_session: ActiveSessionSummary | null
  since_last: SinceLastSummary | null
  today: DayStats
  rolling_7d: RollingStats
  rolling_14d: RollingStats
}

export function getDashboardSummary(babyId: string): Promise<DashboardSummary> {
  return api.get(`/babies/${babyId}/sleep-sessions/summary`)
}
