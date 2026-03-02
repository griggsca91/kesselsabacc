import { motion } from "framer-motion";
import type { Phase } from "../types";

interface PhaseOverlayProps {
  phase: Phase;
  round: number;
  winnerName?: string;
}

const FADE_IN = 0.3;

function overlayContent(phase: Phase, round: number, winnerName?: string) {
  switch (phase) {
    case "turn":
      return (
        <>
          <div className="phase-overlay-text">Round {round}</div>
        </>
      );
    case "reveal":
      return (
        <>
          <div className="phase-overlay-text phase-overlay-text--reveal">
            Reveal!
          </div>
        </>
      );
    case "round_end":
      return (
        <>
          <div className="phase-overlay-text">Round Complete</div>
        </>
      );
    case "game_over":
      return (
        <div className="phase-overlay-gameover">
          <div className="phase-overlay-text phase-overlay-text--gameover">
            Game Over
          </div>
          {winnerName && (
            <div className="phase-overlay-subtitle">
              {winnerName} wins!
            </div>
          )}
        </div>
      );
    default:
      return null;
  }
}

export function PhaseOverlay({ phase, round, winnerName }: PhaseOverlayProps) {
  return (
    <motion.div
      className="phase-overlay"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: FADE_IN }}
    >
      <motion.div
        className="phase-overlay-content"
        initial={{ opacity: 0, scale: 0.85 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 1.05 }}
        transition={{ duration: FADE_IN + 0.1, ease: "easeOut" }}
      >
        {overlayContent(phase, round, winnerName)}
      </motion.div>
    </motion.div>
  );
}
