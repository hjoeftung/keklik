import { api } from './client'

// --- Auth ---

export function logout(): Promise<void> {
  return api.post('/auth/logout')
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
  version: number
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
  version: number
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
  version: number
}

export interface EditSleepSessionRequest {
  started_at?: string
  stopped_at?: string
  version: number
}

export function editSleepSession(
  babyId: string,
  sessionId: string,
  req: EditSleepSessionRequest,
): Promise<SleepSession> {
  return api.patch(`/babies/${babyId}/sleep-sessions/${sessionId}`, req)
}

export interface DeleteSleepSessionRequest {
  version: number
}

export function deleteSleepSession(
  babyId: string,
  sessionId: string,
  req: DeleteSleepSessionRequest,
): Promise<void> {
  return api.delete(`/babies/${babyId}/sleep-sessions/${sessionId}`, req)
}

export interface LogPastSleepRequest {
  started_at: string
  stopped_at: string
}

export function logPastSleep(
  babyId: string,
  req: LogPastSleepRequest,
): Promise<SleepSession> {
  return api.post(`/babies/${babyId}/sleep-sessions`, req)
}

export type SleepHistoryPeriod = 'today' | '7d' | '14d'

export function getSleepHistory(
  babyId: string,
  period: SleepHistoryPeriod = '7d',
): Promise<SleepSession[]> {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
  return api.get(`/babies/${babyId}/sleep-sessions?period=${period}&timezone=${encodeURIComponent(tz)}`)
}

// --- Sleep stats ---

export interface TodayStats {
  total_sleep_seconds: number
  total_nap_seconds: number
  total_active_seconds: number
}

export interface PeriodAverage {
  avg_sleep_seconds: number
  avg_nap_seconds: number
  avg_active_seconds: number
}

export interface NightWindowInfo {
  start_hhmm: string
  end_hhmm: string
}

export interface SleepStatsResponse {
  today: TodayStats
  summary: Record<string, PeriodAverage>
  night_window?: NightWindowInfo
}

export function getSleepStats(babyId: string): Promise<SleepStatsResponse> {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
  return api.get(`/babies/${babyId}/sleep-stats?timezone=${encodeURIComponent(tz)}`)
}

// --- Sleep profiles ---

export interface NightWindowRequest {
  start_hour: number
  start_minute: number
  end_hour: number
  end_minute: number
}

export interface CreateSleepProfileRequest {
  timezone: string
  night_window: NightWindowRequest
}

export function createSleepProfile(
  babyId: string,
  req: CreateSleepProfileRequest,
): Promise<void> {
  return api.post(`/babies/${babyId}/sleep-profiles`, req)
}
