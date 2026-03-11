export async function initSession(gameType: string): Promise<{ session_token: string; seed: number; difficulty_level?: number }> {
  return request<{ session_token: string; seed: number; difficulty_level?: number }>(`/api/sessions/init`, {
    method: 'POST',
    body: JSON.stringify({ game_type: gameType, mode: 'solo' })
  });
}
const API_URL = process.env.NEXT_PUBLIC_API_URL;

export interface User {
  id: string;
  email: string;
  role: "student" | "parent" | "educator";
  created_at: string;
}

export interface APIError {
  code: string;
  message: string;
}

export interface WeeklySummary {
  attention_score: number | null;
  memory_score: number | null;
  engagement_frequency: number;
  coop_participation_rate: number | null;
  avg_reaction_time_ms: number | null;
  total_stars: number;
  total_xp: number;
  sessions_this_week: number;
  snapshot_date: string;
  student_nickname?: string;
  message?: string;
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  let data;
  try {
    data = await res.json();
  } catch (error) {
     throw { code: "UNKNOWN", message: "Malformed response from server." };
  }

  if (!res.ok) {
    // Backend always returns { code, message } for errors.
    throw data && data.message
      ? { code: data.code || "UNKNOWN", message: data.message }
      : { code: "UNKNOWN", message: res.statusText };
  }
  return data;
}

export async function login(email: string, password: string): Promise<void> {
  await request<void>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export function register(
  email: string,
  password: string,
  role: "student" | "parent" | "educator"
): Promise<User> {
  return request<User>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, role }),
  });
}

export function getMe(): Promise<User> {
  return request<User>("/api/auth/me");
}

export interface SessionPayload {
  zone: string;
  actions: Array<{
    type: string;
    timestamp: number;
    data?: Record<string, unknown>;
  }>;
  duration_ms: number;
}

export interface SessionResult {
  session_id?: string;
  score: number;
  accuracy: number;
  stars_earned: number;
  xp_earned: number;
  total_xp: number;
  unlocked_zones: string[];
  behavioral_metrics_count?: number;
}

export async function submitSession(submission: { session_token: string; actions: any[] }): Promise<any> {
  return request<any>("/api/sessions", {
    method: "POST",
    body: JSON.stringify(submission),
  });
}

export function getSessionManifest(token: string): Promise<any> {
  return request<any>(`/api/sessions/manifest?token=${encodeURIComponent(token)}`, {
    method: "GET"
  });
}

export function submitCoopSession(data: {
  game_type: "team_tower",
  mode: "cooperative",
  room_session_id: string,
  outcome: string,
  duration_ms: number
}): Promise<SessionResult> {
  return request<SessionResult>("/api/sessions/coop", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export interface Profile {
  id: string;
  nickname: string;
  avatar_id: number;
  total_stars: number;
  total_xp: number;
  play_mode: string;
  created_at: string;
}

export function getProfile(): Promise<Profile> {
  return request<Profile>("/api/profiles/me");
}

export function createProfile(data: {
  nickname: string;
  avatar_id: number;
  play_mode: string;
}): Promise<Profile> {
  return request<Profile>("/api/profiles", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function updateProfile(data: {
  nickname?: string;
  avatar_id?: number;
  play_mode?: string;
}): Promise<Profile> {
  return request<Profile>("/api/profiles/me", {
    method: "PATCH",
    body: JSON.stringify(data),
  });
}

export function getAnalyticsOverview(
  profileId: string
): Promise<WeeklySummary> {
  return request<WeeklySummary>(
    `/api/analytics/overview?profile_id=${encodeURIComponent(profileId)}`
  );
}

export function logout(): Promise<{ message: string }> {
  return request<{ message: string }>("/api/auth/logout", {
    method: "POST",
  });
}
