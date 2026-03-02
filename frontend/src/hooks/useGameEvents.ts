import { useEffect, useRef } from "react";
import type { GameState } from "../types";
import type { ToastType } from "./useToast";

/**
 * Watches for game-state transitions and fires toast notifications.
 *
 * Uses a ref to track the previous state so every event fires exactly once.
 */
export function useGameEvents(
  gameState: GameState | null,
  playerId: string,
  addToast: (message: string, type?: ToastType) => void,
) {
  const prevRef = useRef<GameState | null>(null);

  useEffect(() => {
    const prev = prevRef.current;
    prevRef.current = gameState;

    // Nothing to compare on first mount or when there is no state
    if (!gameState || !prev) return;

    // Helper to look up player name
    const nameOf = (id: string): string =>
      gameState.players.find((p) => p.id === id)?.name ?? "Unknown";

    // ── Turn change ──
    if (
      gameState.phase === "turn" &&
      gameState.currentTurnPlayerId !== prev.currentTurnPlayerId
    ) {
      if (gameState.currentTurnPlayerId === playerId) {
        addToast("It's your turn!", "warning");
      } else {
        addToast(`${nameOf(gameState.currentTurnPlayerId)} is up`, "info");
      }
    }

    // ── Player stood ──
    for (const player of gameState.players) {
      const prevPlayer = prev.players.find((p) => p.id === player.id);
      if (player.stood && prevPlayer && !prevPlayer.stood) {
        if (player.id === playerId) {
          addToast("You stood", "info");
        } else {
          addToast(`${player.name} stood`, "info");
        }
      }
    }

    // ── Phase transitions ──
    if (prev.phase !== gameState.phase) {
      // Reveal phase
      if (gameState.phase === "reveal") {
        addToast("Cards revealed!", "info");
      }

      // Round end – announce winner
      if (gameState.phase === "round_end" && gameState.lastResult) {
        const winnerIds = gameState.lastResult.winnerIds;
        if (winnerIds.length > 1) {
          addToast("Round ends in a tie!", "success");
        } else if (winnerIds.length === 1) {
          const winnerName = nameOf(winnerIds[0]);
          if (winnerIds[0] === playerId) {
            addToast("You win the round!", "success");
          } else {
            addToast(`${winnerName} wins the round!`, "success");
          }
        }
      }

      // Game over
      if (gameState.phase === "game_over" && gameState.winnerId) {
        if (gameState.winnerId === playerId) {
          addToast("You win the game!", "success");
        } else {
          addToast(`${nameOf(gameState.winnerId)} wins the game!`, "success");
        }
      }
    }
  }, [gameState, playerId, addToast]);
}
