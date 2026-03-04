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
