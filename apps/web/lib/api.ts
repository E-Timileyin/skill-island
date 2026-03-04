const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

  if (!res.ok) {
    const error: APIError = await res.json().catch(() => ({
      code: "UNKNOWN",
      message: res.statusText,
    }));
    throw error;
  }

  return res.json();
}

export function login(email: string, password: string): Promise<User> {
  return request<User>("/api/auth/login", {
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
  session_id: string;
  stars: number;
  xp_earned: number;
}

export function submitSession(
  payload: SessionPayload
): Promise<SessionResult> {
  return request<SessionResult>("/api/sessions", {
    method: "POST",
    body: JSON.stringify(payload),
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
