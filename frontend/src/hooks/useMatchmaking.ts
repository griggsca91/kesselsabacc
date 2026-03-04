import { useCallback, useEffect, useRef, useState } from "react";

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

const POLL_INTERVAL_MS = 2000;

interface UseMatchmakingOptions {
  playerId: string;
  playerName: string;
  token?: string | null;
  onMatched: (roomCode: string) => void;
}

interface UseMatchmakingReturn {
  isQueued: boolean;
  joinQueue: () => Promise<void>;
  leaveQueue: () => Promise<void>;
  queueTime: number; // seconds since queued
}

export function useMatchmaking({
  playerId,
  playerName,
  token,
  onMatched,
}: UseMatchmakingOptions): UseMatchmakingReturn {
  const [isQueued, setIsQueued] = useState(false);
  const [queueTime, setQueueTime] = useState(0);

  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const isQueuedRef = useRef(false);

  const headers = useCallback((): Record<string, string> => {
    const h: Record<string, string> = { "Content-Type": "application/json" };
    if (token) {
      h["Authorization"] = `Bearer ${token}`;
    }
    return h;
  }, [token]);

  const stopPolling = useCallback(() => {
    if (pollRef.current !== null) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const startPolling = useCallback(() => {
    stopPolling();

    // Queue timer — increments every second
    timerRef.current = setInterval(() => {
      setQueueTime((t) => t + 1);
    }, 1000);

    // Poll for match result every 2 seconds
    pollRef.current = setInterval(async () => {
      if (!isQueuedRef.current) return;
      try {
        const url = `${API}/api/queue/result?playerId=${encodeURIComponent(playerId)}`;
        const res = await fetch(url, { headers: headers() });
        if (!res.ok) return;
        const data = await res.json();
        if (data.matched && data.roomCode) {
          stopPolling();
          isQueuedRef.current = false;
          setIsQueued(false);
          setQueueTime(0);
          onMatched(data.roomCode);
        }
      } catch {
        // ignore transient errors
      }
    }, POLL_INTERVAL_MS);
  }, [playerId, headers, stopPolling, onMatched]);

  const joinQueue = useCallback(async () => {
    if (isQueuedRef.current) return;
    try {
      const res = await fetch(`${API}/api/queue`, {
        method: "POST",
        headers: headers(),
        body: JSON.stringify({ playerId, playerName }),
      });
      if (!res.ok) return;
      const data = await res.json();
      if (data.queued) {
        isQueuedRef.current = true;
        setIsQueued(true);
        setQueueTime(0);
        startPolling();
      }
    } catch {
      // ignore transient errors
    }
  }, [playerId, playerName, headers, startPolling]);

  const leaveQueue = useCallback(async () => {
    if (!isQueuedRef.current) return;
    isQueuedRef.current = false;
    setIsQueued(false);
    setQueueTime(0);
    stopPolling();
    try {
      await fetch(`${API}/api/queue`, {
        method: "DELETE",
        headers: headers(),
        body: JSON.stringify({ playerId }),
      });
    } catch {
      // ignore transient errors
    }
  }, [playerId, headers, stopPolling]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  return { isQueued, joinQueue, leaveQueue, queueTime };
}
