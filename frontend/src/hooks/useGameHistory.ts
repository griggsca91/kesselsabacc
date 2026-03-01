import { useCallback, useEffect, useState } from "react";
import type { GameHistoryEntry } from "../types";

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

interface UseGameHistoryOptions {
  playerId: string;
  /** JWT auth token. When provided, the server uses it to identify the player. */
  token?: string | null;
}

interface UseGameHistoryReturn {
  games: GameHistoryEntry[];
  isLoading: boolean;
  error: string | null;
  refresh: () => void;
}

export function useGameHistory({ playerId, token }: UseGameHistoryOptions): UseGameHistoryReturn {
  const [games, setGames] = useState<GameHistoryEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHistory = useCallback(async () => {
    if (!playerId) return;
    setIsLoading(true);
    setError(null);
    try {
      const headers: Record<string, string> = {};
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }
      const res = await fetch(
        `${API}/api/games?playerId=${encodeURIComponent(playerId)}`,
        { headers },
      );
      if (!res.ok) {
        const text = await res.text();
        setError(text || "Failed to load game history");
        return;
      }
      const data: GameHistoryEntry[] = await res.json();
      setGames(data);
    } catch {
      setError("Unable to load game history. Please check your connection.");
    } finally {
      setIsLoading(false);
    }
  }, [playerId, token]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return { games, isLoading, error, refresh: fetchHistory };
}
