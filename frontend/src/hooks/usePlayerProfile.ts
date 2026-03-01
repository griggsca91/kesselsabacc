import { useCallback, useEffect, useState } from "react";
import type { PlayerProfile } from "../types";

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

interface UsePlayerProfileOptions {
  /** The user ID to fetch the profile for. If omitted, fetches the authenticated user's own profile. */
  userId?: string;
  /** JWT auth token for authenticated API calls. */
  token?: string | null;
}

interface UsePlayerProfileReturn {
  profile: PlayerProfile | null;
  isLoading: boolean;
  error: string | null;
  refresh: () => void;
}

export function usePlayerProfile({ userId, token }: UsePlayerProfileOptions): UsePlayerProfileReturn {
  const [profile, setProfile] = useState<PlayerProfile | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const headers: Record<string, string> = {};
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const url = userId
        ? `${API}/api/profile/${encodeURIComponent(userId)}`
        : `${API}/api/profile`;

      const res = await fetch(url, { headers });
      if (!res.ok) {
        const text = await res.text();
        setError(text || "Failed to load profile");
        return;
      }
      const data: PlayerProfile = await res.json();
      setProfile(data);
    } catch {
      setError("Unable to load profile. Please check your connection.");
    } finally {
      setIsLoading(false);
    }
  }, [userId, token]);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  return { profile, isLoading, error, refresh: fetchProfile };
}
