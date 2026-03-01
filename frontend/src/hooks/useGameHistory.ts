import { useCallback, useEffect, useState } from "react";
import type { GameHistoryEntry } from "../types";

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

interface UseGameHistoryReturn {
  games: GameHistoryEntry[];
  isLoading: boolean;
  error: string | null;
  refresh: () => void;
}

export function useGameHistory(playerId: string): UseGameHistoryReturn {
  const [games, setGames] = useState<GameHistoryEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHistory = useCallback(async () => {
    if (!playerId) return;
    setIsLoading(true);
    setError(null);
    try {
      const res = await fetch(
        `${API}/api/games?playerId=${encodeURIComponent(playerId)}`
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
  }, [playerId]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return { games, isLoading, error, refresh: fetchHistory };
}
