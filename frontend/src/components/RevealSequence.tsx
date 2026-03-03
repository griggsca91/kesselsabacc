import { useEffect, useState, useCallback, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import type { PlayerView, RoundResult, HandResult } from "../types";
import { CardDisplay } from "./CardDisplay";
import { DiceRoll } from "./DiceRoll";
import { WinnerCelebration } from "./WinnerCelebration";

interface RevealSequenceProps {
  players: PlayerView[];
  playerId: string;
  lastResult: RoundResult;
  round: number;
  onComplete: () => void;
}

/** Hand rank display text */
function rankLabel(result: HandResult): string {
  if (result.rank === 0) return "Pure Sabacc";
  if (result.rank === 1) return `Sabacc ${result.value}/${result.value}`;
  return `No Sabacc (diff ${result.value})`;
}

function rankClass(result: HandResult): string {
  if (result.rank === 0) return "pure";
  if (result.rank === 1) return "sabacc";
  return "no-sabacc";
}

type RevealStage =
  | { type: "title" }
  | { type: "player"; playerIndex: number; cardStep: "sand" | "blood" | "done" }
  | { type: "winner" };

/**
 * Dramatic reveal sequence that shows results one player at a time.
 *
 * Sequence:
 * 1. "Reveal!" title fades in (0.8s)
 * 2. For each player:
 *    a. Player name slides in
 *    b. Sand card flips (0.6s), then blood card flips (0.6s)
 *    c. If impostor card, dice roll animation
 *    d. Hand rank badge fades in
 * 3. Winner celebration with golden glow and sparkles
 */
export function RevealSequence({
  players,
  playerId,
  lastResult,
  round,
  onComplete,
}: RevealSequenceProps) {
  const [stage, setStage] = useState<RevealStage>({ type: "title" });
  const [revealedPlayers, setRevealedPlayers] = useState<number[]>([]);
  const [showRank, setShowRank] = useState<Record<number, boolean>>({});
  const timeoutsRef = useRef<ReturnType<typeof setTimeout>[]>([]);
  const skippedRef = useRef(false);

  // Players that have hand data (non-eliminated players with results)
  const activePlayers = players.filter((p) => lastResult.playerHands[p.id]);

  const clearTimeouts = useCallback(() => {
    timeoutsRef.current.forEach(clearTimeout);
    timeoutsRef.current = [];
  }, []);

  const addTimeout = useCallback((fn: () => void, ms: number) => {
    const id = setTimeout(fn, ms);
    timeoutsRef.current.push(id);
    return id;
  }, []);

  // Skip handler — immediately reveal everything
  const handleSkip = useCallback(() => {
    if (skippedRef.current) return;
    skippedRef.current = true;
    clearTimeouts();
    onComplete();
  }, [clearTimeouts, onComplete]);

  // Main sequencing effect — start with title, then advance to first player
  useEffect(() => {
    if (skippedRef.current) return;

    addTimeout(() => {
      if (skippedRef.current) return;
      if (activePlayers.length > 0) {
        setStage({ type: "player", playerIndex: 0, cardStep: "sand" });
      } else {
        setStage({ type: "winner" });
      }
    }, 800);

    return clearTimeouts;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [round]);

  // Advance card steps within a player
  useEffect(() => {
    if (skippedRef.current) return;
    if (stage.type !== "player") return;

    const { playerIndex, cardStep } = stage;

    if (cardStep === "sand") {
      addTimeout(() => {
        if (skippedRef.current) return;
        setStage({ type: "player", playerIndex, cardStep: "blood" });
      }, 600);
    } else if (cardStep === "blood") {
      addTimeout(() => {
        if (skippedRef.current) return;
        setShowRank((prev) => ({ ...prev, [playerIndex]: true }));
      }, 600);

      addTimeout(() => {
        if (skippedRef.current) return;
        setRevealedPlayers((prev) => [...prev, playerIndex]);

        const nextIndex = playerIndex + 1;
        if (nextIndex < activePlayers.length) {
          setStage({ type: "player", playerIndex: nextIndex, cardStep: "sand" });
        } else {
          addTimeout(() => {
            if (skippedRef.current) return;
            setStage({ type: "winner" });
          }, 500);
        }
      }, 1200);
    }

    return clearTimeouts;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stage]);

  // Auto-complete after winner celebration
  useEffect(() => {
    if (stage.type !== "winner" || skippedRef.current) return;
    addTimeout(() => {
      if (skippedRef.current) return;
      onComplete();
    }, 3000);
    return clearTimeouts;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stage]);

  const currentPlayerIdx = stage.type === "player" ? stage.playerIndex : -1;
  const currentCardStep = stage.type === "player" ? stage.cardStep : null;

  const winnerNames = lastResult.winnerIds
    .map((id) => players.find((p) => p.id === id)?.name ?? "Unknown")
    .join(" & ");

  const winnerChipChange = lastResult.winnerIds.length > 0
    ? lastResult.chipChanges[lastResult.winnerIds[0]] ?? 0
    : 0;

  return (
    <section className="reveal-sequence">
      <button className="btn-skip-reveal" onClick={handleSkip}>
        Skip
      </button>

      <AnimatePresence>
        {stage.type === "title" && (
          <motion.div
            className="reveal-title"
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.5, ease: "backOut" }}
          >
            Round {round} — Reveal!
          </motion.div>
        )}
      </AnimatePresence>

      <div className="reveal-players">
        {activePlayers.map((player, idx) => {
          const hand = lastResult.playerHands[player.id];
          if (!hand) return null;

          const isCurrentOrRevealed = idx <= currentPlayerIdx || revealedPlayers.includes(idx);
          const isBeingRevealed = idx === currentPlayerIdx;
          const showSand = isBeingRevealed
            ? (currentCardStep === "sand" || currentCardStep === "blood" || currentCardStep === "done")
            : revealedPlayers.includes(idx);
          const showBlood = isBeingRevealed
            ? (currentCardStep === "blood" || currentCardStep === "done")
            : revealedPlayers.includes(idx);
          const sandIsImpostor = hand.sandCard.kind === "impostor";
          const bloodIsImpostor = hand.bloodCard.kind === "impostor";
          const isWinner = lastResult.winnerIds.includes(player.id);
          const chipChange = lastResult.chipChanges[player.id] ?? 0;
          const isMe = player.id === playerId;

          if (!isCurrentOrRevealed && stage.type !== "winner") return null;

          return (
            <motion.div
              key={player.id}
              className={[
                "reveal-player-row",
                isWinner && stage.type === "winner" ? "reveal-winner-glow" : "",
                isMe ? "reveal-is-me" : "",
              ].filter(Boolean).join(" ")}
              initial={{ opacity: 0, x: -24 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.4, ease: "easeOut" }}
            >
              <div className="reveal-player-name">
                {player.name}
                {isMe && <span className="player-you-tag">you</span>}
              </div>

              <div className="reveal-cards-row">
                <div className="reveal-card-slot">
                  {showSand ? (
                    <motion.div
                      className="reveal-card-flip"
                      initial={{ rotateY: 90, opacity: 0 }}
                      animate={{ rotateY: 0, opacity: 1 }}
                      transition={{ duration: 0.45, ease: "easeOut" }}
                    >
                      <CardDisplay card={hand.sandCard} size="sm" />
                      {sandIsImpostor && isBeingRevealed && !revealedPlayers.includes(idx) && (
                        <div className="reveal-dice-overlay">
                          <DiceRoll finalValue={hand.sandCard.value} duration={800} />
                        </div>
                      )}
                    </motion.div>
                  ) : (
                    <div className="reveal-card-placeholder" />
                  )}
                </div>

                <div className="reveal-card-slot">
                  {showBlood ? (
                    <motion.div
                      className="reveal-card-flip"
                      initial={{ rotateY: 90, opacity: 0 }}
                      animate={{ rotateY: 0, opacity: 1 }}
                      transition={{ duration: 0.45, ease: "easeOut" }}
                    >
                      <CardDisplay card={hand.bloodCard} size="sm" />
                      {bloodIsImpostor && isBeingRevealed && !revealedPlayers.includes(idx) && (
                        <div className="reveal-dice-overlay">
                          <DiceRoll finalValue={hand.bloodCard.value} duration={800} />
                        </div>
                      )}
                    </motion.div>
                  ) : (
                    <div className="reveal-card-placeholder" />
                  )}
                </div>
              </div>

              {(showRank[idx] || revealedPlayers.includes(idx) || stage.type === "winner") && hand && (
                <motion.div
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <span className={`hand-rank-badge ${rankClass(hand)}`}>
                    {rankLabel(hand)}
                  </span>
                </motion.div>
              )}

              {stage.type === "winner" && (
                <motion.span
                  className={`reveal-chip-change ${chipChange > 0 ? "chip-gain" : chipChange < 0 ? "chip-loss" : "chip-neutral"}`}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ duration: 0.3, delay: 0.2 }}
                >
                  {chipChange > 0 ? `+${chipChange}` : chipChange}
                </motion.span>
              )}
            </motion.div>
          );
        })}
      </div>

      <AnimatePresence>
        {stage.type === "winner" && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.4 }}
          >
            <WinnerCelebration
              winnerName={winnerNames}
              chipChange={winnerChipChange}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  );
}
