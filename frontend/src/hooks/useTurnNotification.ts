import { useEffect, useRef } from "react";
import type { GameState } from "../types";

const DEFAULT_TITLE = "Kessel Sabacc";
const YOUR_TURN_TITLE = "\u{1F3B4} Your Turn! - Sabacc";

/** Play a short sine-wave ding via the Web Audio API. */
function playDing() {
  try {
    const ctx = new AudioContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = "sine";
    osc.frequency.value = 800;

    gain.gain.setValueAtTime(0.3, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.15);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.15);

    // Clean up after sound finishes
    osc.onended = () => ctx.close();
  } catch {
    // AudioContext not available — silently ignore
  }
}

/** Show a browser notification if permitted and the tab is hidden. */
function showBrowserNotification() {
  if (
    typeof Notification !== "undefined" &&
    Notification.permission === "granted" &&
    document.hidden
  ) {
    new Notification("Kessel Sabacc", {
      body: "It's your turn!",
      icon: "/favicon.ico",
    });
  }
}

/**
 * Watches for turn changes and fires browser / audio / title notifications.
 *
 * Only reacts when the turn *becomes* yours (transition edge), not on every
 * render where it happens to be your turn.
 */
export function useTurnNotification(
  gameState: GameState | null,
  playerId: string,
  notificationsEnabled: boolean,
) {
  const prevTurnPlayerRef = useRef<string | null>(null);

  useEffect(() => {
    if (!gameState) {
      prevTurnPlayerRef.current = null;
      return;
    }

    const prevTurnPlayer = prevTurnPlayerRef.current;
    prevTurnPlayerRef.current = gameState.currentTurnPlayerId;

    // Detect the *transition* to our turn
    const justBecameMyTurn =
      gameState.phase === "turn" &&
      gameState.currentTurnPlayerId === playerId &&
      prevTurnPlayer !== playerId;

    if (justBecameMyTurn) {
      document.title = YOUR_TURN_TITLE;

      if (notificationsEnabled) {
        playDing();
        showBrowserNotification();
      }
    } else if (gameState.currentTurnPlayerId !== playerId) {
      document.title = DEFAULT_TITLE;
    }
  }, [gameState, playerId, notificationsEnabled]);

  // Restore the original title on unmount
  useEffect(() => {
    return () => {
      document.title = DEFAULT_TITLE;
    };
  }, []);
}
