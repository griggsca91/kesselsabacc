import { useEffect, useRef, useState } from "react";
import type { Phase } from "../types";

/** Phases that should NOT trigger an overlay banner. */
const SKIP_OVERLAY: ReadonlySet<Phase> = new Set(["lobby", "dealing"]);

/** Duration (ms) of the overlay per phase. */
const OVERLAY_DURATION: Partial<Record<Phase, number>> = {
  turn: 1500,
  reveal: 1500,
  round_end: 1500,
  game_over: 2500,
};

const DEFAULT_DURATION = 1500;

export interface PhaseTransitionState {
  showOverlay: boolean;
  overlayPhase: Phase;
  overlayRound: number;
}

export function usePhaseTransition(
  phase: Phase,
  round: number,
): PhaseTransitionState {
  const prevPhase = useRef(phase);
  const [showOverlay, setShowOverlay] = useState(false);
  const [overlayPhase, setOverlayPhase] = useState<Phase>(phase);
  const [overlayRound, setOverlayRound] = useState(round);

  useEffect(() => {
    const prev = prevPhase.current;
    prevPhase.current = phase;

    // No change, or transitioning to a skipped phase
    if (phase === prev || SKIP_OVERLAY.has(phase)) return;

    setOverlayPhase(phase);
    setOverlayRound(round);
    setShowOverlay(true);

    const duration = OVERLAY_DURATION[phase] ?? DEFAULT_DURATION;
    const timer = setTimeout(() => setShowOverlay(false), duration);

    return () => clearTimeout(timer);
  }, [phase, round]);

  return { showOverlay, overlayPhase, overlayRound };
}
