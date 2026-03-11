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

// Global access token getter - will be set by auth store
let getAccessToken: (() => string | null) | null = null;
let refreshTokensCallback: (() => Promise<boolean>) | null = null;

export function setAuthCallbacks(
  getAccessTokenFn: () => string | null,
  refreshTokensFn: () => Promise<boolean>
) {
  getAccessToken = getAccessTokenFn;
  refreshTokensCallback = refreshTokensFn;
}

interface RequestOptions extends RequestInit {
  _isRetry?: boolean;
  _skipAuth?: boolean;
}

async function request<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { _isRetry, _skipAuth, ...fetchOptions } = options;
  const accessToken = getAccessToken?.();
  
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(fetchOptions.headers as Record<string, string>),
  };
  
  if (accessToken && !_skipAuth) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...fetchOptions,
    headers,
  });

  let data;
  try {
    data = await res.json();
  } catch (error) {
     throw { code: "UNKNOWN", message: "Malformed response from server." };
  }

  if (!res.ok) {
    // If 401 and we have a refresh callback, try to refresh and retry once
    if (res.status === 401 && refreshTokensCallback && !_isRetry) {
      const refreshed = await refreshTokensCallback();
      if (refreshed) {
        const newToken = getAccessToken?.();
        if (newToken) {
          return request<T>(path, {
            ...options,
            _isRetry: true,
          });
        }
      }
    }
    // Backend always returns { code, message } for errors.
    throw data && data.message
      ? { code: data.code || "UNKNOWN", message: data.message }
      : { code: "UNKNOWN", message: res.statusText };
  }
  return data;
}

// Login function - returns both tokens (no auth header needed)
export async function login(email: string, password: string): Promise<{ id: string; email: string; role: string; access_token: string; refresh_token: string }> {
  return await request<{ id: string; email: string; role: string; access_token: string; refresh_token: string }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
    _skipAuth: true,
  });
}

// Refresh token function (no auth header needed - uses refresh token in body)
export async function refreshToken(refreshTokenValue: string): Promise<{ access_token: string; refresh_token: string }> {
  return await request<{ access_token: string; refresh_token: string }>("/api/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshTokenValue }),
    _skipAuth: true,
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

export async function getMe(): Promise<User | null> {
  try {
    const data = await request<any>("/api/auth/me");
    // Handle both `{...fields}` or `{ user: { ...fields } }`
    if (data && typeof data === "object") {
      if ("role" in data) {
        return data as User;
      }
      if ("user" in data && typeof data.user === "object" && "role" in data.user) {
        return data.user as User;
      }
    }
    return null;
  } catch (e) {
    return null;
  }
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
