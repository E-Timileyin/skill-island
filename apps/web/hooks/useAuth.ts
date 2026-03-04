"use client";

import { useEffect, useState } from "react";
import { getMe, type User } from "@/lib/api";

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getMe()
      .then(setUser)
      .catch((err) => {
        setError(err.message || "Not authenticated");
      })
      .finally(() => setLoading(false));
  }, []);

  return { user, loading, error };
}
