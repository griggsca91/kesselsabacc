import { useEffect, useRef } from "react";
import type { GameState } from "../types";
import type { SoundEngine } from "./useSoundEngine";

export function useGameSounds(
  gameState: GameState | null,
  playerId: string,
  sound: SoundEngine,
) {
  const prevRef = useRef<GameState | null>(null);

  useEffect(() => {
    const prev = prevRef.current;
    prevRef.current = gameState;

    if (!gameState || !prev) return;

    if (prev.phase !== gameState.phase) {
      if (gameState.phase === "dealing") {
        sound.playCardShuffle();
      }
      if (
        (gameState.phase === "round_end" || gameState.phase === "game_over") &&
        gameState.lastResult
      ) {
        const winnerIds = gameState.lastResult.winnerIds;
        if (winnerIds.includes(playerId)) {
          sound.playWinFanfare();
        } else {
          sound.playLoseStinger();
        }
      }
      if (gameState.phase === "reveal") {
        sound.playCardFlip();
      }
    }

    if (
      gameState.phase === "turn" &&
      gameState.currentTurnPlayerId === playerId &&
      prev.currentTurnPlayerId !== playerId
    ) {
      sound.playButtonClick();
    }

    if (gameState.yourHand && prev.yourHand) {
      const sandChanged =
        gameState.yourHand.sandCard.id !== prev.yourHand.sandCard.id;
      const bloodChanged =
        gameState.yourHand.bloodCard.id !== prev.yourHand.bloodCard.id;
      if (sandChanged || bloodChanged) {
        sound.playCardDraw();
      }
    }

    const me = gameState.players.find((p) => p.id === playerId);
    const prevMe = prev.players.find((p) => p.id === playerId);
    if (me && prevMe && me.chips !== prevMe.chips) {
      sound.playChipClink();
    }
  }, [gameState, playerId, sound]);
}
