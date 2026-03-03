import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

interface DiceRollProps {
  /** The final resolved value the impostor lands on. */
  finalValue: number;
  /** Duration of the cycling phase in ms (default 800). */
  duration?: number;
  /** Called when the roll animation finishes. */
  onComplete?: () => void;
}

/**
 * Dice roll animation for impostor cards.
 * Shows cycling numbers (1-6) that settle on the resolved value.
 */
export function DiceRoll({ finalValue, duration = 800, onComplete }: DiceRollProps) {
  const [display, setDisplay] = useState(1);
  const [settled, setSettled] = useState(false);

  useEffect(() => {
    let frame: number;
    const start = performance.now();
    const interval = 60;
    let lastChange = 0;

    function tick(now: number) {
      const elapsed = now - start;

      if (elapsed >= duration) {
        setDisplay(finalValue);
        setSettled(true);
        onComplete?.();
        return;
      }

      if (now - lastChange > interval) {
        setDisplay((prev) => {
          let next = prev;
          while (next === prev) {
            next = Math.floor(Math.random() * 6) + 1;
          }
          return next;
        });
        lastChange = now;
      }

      frame = requestAnimationFrame(tick);
    }

    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [finalValue, duration, onComplete]);

  return (
    <div className="dice-roll">
      <AnimatePresence mode="wait">
        <motion.span
          key={`${display}-${settled}`}
          className={`dice-roll-value${settled ? " dice-settled" : ""}`}
          initial={{ opacity: 0, scale: 0.5, y: -8 }}
          animate={{
            opacity: 1,
            scale: settled ? [1.3, 1] : 1,
            y: 0,
          }}
          exit={{ opacity: 0, scale: 0.5, y: 8 }}
          transition={{
            duration: settled ? 0.35 : 0.06,
            ease: settled ? "backOut" : "linear",
          }}
        >
          {display}
        </motion.span>
      </AnimatePresence>
    </div>
  );
}
