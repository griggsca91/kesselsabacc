import { useCallback, useEffect, useRef, useState } from "react";
import type { CardSuit, GameState, ServerEnvelope, ShiftToken } from "../types";

const API = "http://localhost:8080";
const WS = "ws://localhost:8080";

interface UseGameReturn {
  gameState: GameState | null;
  error: string | null;
  connected: boolean;
  playerId: string;
  roomCode: string;
  createRoom: (name: string) => Promise<void>;
  joinRoom: (code: string, name: string) => Promise<void>;
  startGame: () => void;
  draw: (suit: CardSuit, token?: ShiftToken) => void;
  stand: (token?: ShiftToken) => void;
  nextRound: () => void;
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
  const [connected, setConnected] = useState(false);
  const [roomCode, setRoomCode] = useState("");
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback((code: string) => {
    const ws = new WebSocket(`${WS}/ws?playerId=${playerId}&roomCode=${code}`);
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onerror = () => setError("WebSocket connection error");

    ws.onmessage = (event) => {
      const envelope: ServerEnvelope = JSON.parse(event.data);
      if (envelope.type === "game_state") {
        setGameState(envelope.payload as GameState);
        setError(null);
      } else if (envelope.type === "error") {
        setError((envelope.payload as { message: string }).message);
      }
    };
  }, [playerId]);

  useEffect(() => {
    return () => wsRef.current?.close();
  }, []);

  const send = useCallback((type: string, payload: unknown) => {
    wsRef.current?.send(JSON.stringify({ type, payload }));
  }, []);

  const createRoom = useCallback(async (name: string) => {
    const res = await fetch(`${API}/rooms`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ playerId, playerName: name }),
    });
    if (!res.ok) throw new Error(await res.text());
    const { code } = await res.json();
    setRoomCode(code);
    connect(code);
  }, [playerId, connect]);

  const joinRoom = useCallback(async (code: string, name: string) => {
    const res = await fetch(`${API}/rooms/${code}/join`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ playerId, playerName: name }),
    });
    if (!res.ok) throw new Error(await res.text());
    setRoomCode(code.toUpperCase());
    connect(code.toUpperCase());
  }, [playerId, connect]);

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
    connected,
    playerId,
    roomCode,
    createRoom,
    joinRoom,
    startGame,
    draw,
    stand,
    nextRound,
  };
}
