import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "../lib/api";
import { login as loginApi, getMe, refreshToken as refreshTokenApi } from "../lib/api";

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  error: string | null;
  loginUser: (email: string, password: string) => Promise<void>;
  fetchUser: () => Promise<void>;
  refreshTokens: () => Promise<boolean>;
  logout: () => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      loading: false,
      error: null,

      loginUser: async (email, password) => {
        set({ loading: true, error: null });
        try {
          const res = await loginApi(email, password);
          set({
            user: {
              id: res.id,
              email: res.email,
              role: res.role as "student" | "parent" | "educator",
              created_at: ""
            },
            accessToken: res.access_token,
            refreshToken: res.refresh_token,
            loading: false,
          });
        } catch (err: any) {
          set({ error: err.message || "Login failed", loading: false });
        }
      },

      fetchUser: async () => {
        set({ loading: true, error: null });
        try {
          const user = await getMe();
          set({ user, loading: false });
        } catch (err: any) {
          set({ error: err.message || "Not authenticated", loading: false });
        }
      },

      refreshTokens: async () => {
        const currentRefreshToken = get().refreshToken;
        if (!currentRefreshToken) return false;

        try {
          const res = await refreshTokenApi(currentRefreshToken);
          set({
            accessToken: res.access_token,
            refreshToken: res.refresh_token,
          });
          return true;
        } catch {
          set({ user: null, accessToken: null, refreshToken: null });
          return false;
        }
      },

      logout: () => {
        set({ user: null, accessToken: null, refreshToken: null, error: null });
      },

      setTokens: (accessToken, refreshToken) => {
        set({ accessToken, refreshToken });
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
);
