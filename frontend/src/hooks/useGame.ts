import { useCallback, useEffect, useRef, useState } from "react";
import type {
  CardSuit,
  ConnectionStatus,
  GameState,
  ServerEnvelope,
  ShiftToken,
} from "../types";

const API = import.meta.env.VITE_API_URL || "http://localhost:8080";
const WS = import.meta.env.VITE_WS_URL || "ws://localhost:8080";

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_RECONNECT_DELAY_MS = 1000;
const MAX_RECONNECT_DELAY_MS = 16000;

interface UseGameReturn {
  gameState: GameState | null;
  error: string | null;
  connectionStatus: ConnectionStatus;
  playerId: string;
  roomCode: string;
  createRoom: (name: string) => Promise<void>;
  joinRoom: (code: string, name: string) => Promise<void>;
  startGame: () => void;
  draw: (suit: CardSuit, token?: ShiftToken) => void;
  stand: (token?: ShiftToken) => void;
  nextRound: () => void;
  reconnect: () => void;
}

function getOrCreatePlayerId(): string {
  let id = localStorage.getItem("sabacc_player_id");
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem("sabacc_player_id", id);
  }
  return id;
}

export function useGame(): UseGameReturn {
  const playerId = getOrCreatePlayerId();
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>("disconnected");
  const [roomCode, setRoomCode] = useState("");

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptRef = useRef(0);
  const manualDisconnectRef = useRef(false);
  const roomCodeRef = useRef("");

  // Keep roomCodeRef in sync so the reconnect closure always has the latest value
  useEffect(() => {
    roomCodeRef.current = roomCode;
  }, [roomCode]);

  const clearReconnectTimer = useCallback(() => {
    if (reconnectTimerRef.current !== null) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  }, []);

  const scheduleReconnect = useCallback(
    (code: string) => {
      const attempt = reconnectAttemptRef.current;
      if (attempt >= MAX_RECONNECT_ATTEMPTS) {
        setConnectionStatus("disconnected");
        setError(
          "Unable to reconnect after multiple attempts. Click Retry to try again."
        );
        return;
      }

      setConnectionStatus("reconnecting");
      const delay = Math.min(
        BASE_RECONNECT_DELAY_MS * Math.pow(2, attempt),
        MAX_RECONNECT_DELAY_MS
      );

      reconnectTimerRef.current = setTimeout(() => {
        reconnectAttemptRef.current += 1;
        connectWs(code);
      }, delay);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  );

  const connectWs = useCallback(
    (code: string) => {
      // Close any existing connection
      if (wsRef.current) {
        // Prevent the onclose handler from triggering reconnect for this close
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }

      const ws = new WebSocket(
        `${WS}/ws?playerId=${playerId}&roomCode=${code}`
      );
      wsRef.current = ws;

      ws.onopen = () => {
        setConnectionStatus("connected");
        setError(null);
        reconnectAttemptRef.current = 0;
        clearReconnectTimer();
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (manualDisconnectRef.current) {
          setConnectionStatus("disconnected");
          return;
        }
        // Auto-reconnect
        scheduleReconnect(code);
      };

      ws.onerror = () => {
        // onerror is always followed by onclose, so reconnect logic lives there.
        // We only set a transient error message here.
        setError("WebSocket connection error");
      };

      ws.onmessage = (event) => {
        const envelope: ServerEnvelope = JSON.parse(event.data);
        if (envelope.type === "game_state") {
          setGameState(envelope.payload as GameState);
          setError(null);
        } else if (envelope.type === "error") {
          setError((envelope.payload as { message: string }).message);
        }
      };
    },
    [playerId, clearReconnectTimer, scheduleReconnect]
  );

  // Manual reconnect (e.g. from a "Retry" button)
  const reconnect = useCallback(() => {
    const code = roomCodeRef.current;
    if (!code) return;
    reconnectAttemptRef.current = 0;
    manualDisconnectRef.current = false;
    setError(null);
    setConnectionStatus("reconnecting");
    connectWs(code);
  }, [connectWs]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      manualDisconnectRef.current = true;
      clearReconnectTimer();
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [clearReconnectTimer]);

  const send = useCallback((type: string, payload: unknown) => {
    wsRef.current?.send(JSON.stringify({ type, payload }));
  }, []);

  const createRoom = useCallback(
    async (name: string) => {
      const res = await fetch(`${API}/rooms`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ playerId, playerName: name }),
      });
      if (!res.ok) throw new Error(await res.text());
      const { code } = await res.json();
      setRoomCode(code);
      roomCodeRef.current = code;
      manualDisconnectRef.current = false;
      connectWs(code);
    },
    [playerId, connectWs]
  );

  const joinRoom = useCallback(
    async (code: string, name: string) => {
      const res = await fetch(`${API}/rooms/${code}/join`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ playerId, playerName: name }),
      });
      if (!res.ok) throw new Error(await res.text());
      const upper = code.toUpperCase();
      setRoomCode(upper);
      roomCodeRef.current = upper;
      manualDisconnectRef.current = false;
      connectWs(upper);
    },
    [playerId, connectWs]
  );

  const startGame = useCallback(() => send("start_game", {}), [send]);

  const draw = useCallback(
    (suit: CardSuit, token?: ShiftToken) =>
      send("draw", { suit, ...(token ? { tokenUsed: token } : {}) }),
    [send]
  );

  const stand = useCallback(
    (token?: ShiftToken) =>
      send("stand", { ...(token ? { tokenUsed: token } : {}) }),
    [send]
  );

  const nextRound = useCallback(() => send("next_round", {}), [send]);

  return {
    gameState,
    error,
    connectionStatus,
    playerId,
    roomCode,
    createRoom,
    joinRoom,
    startGame,
    draw,
    stand,
    nextRound,
    reconnect,
  };
}
