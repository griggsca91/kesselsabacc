import { useRef, useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import type { Card } from "../types";
import { CardDisplay, CardBack } from "./CardDisplay";

/* ─── Deal + Draw-swap animation ─── */
interface AnimatedCardProps {
  card: Card;
  size?: "normal" | "sm";
  /** Stagger delay for deal-in animation (ms). E.g. 0 for first card, 200 for second. */
  dealDelay?: number;
}

/**
 * Wraps a CardDisplay with framer-motion AnimatePresence so that:
 *  - On mount it slides up + fades in (deal animation)
 *  - When `card.id` changes the old card exits (slide out / fade) and
 *    the new one enters (slide in / fade) — draw swap animation
 */
export function AnimatedCard({ card, size = "normal", dealDelay = 0 }: AnimatedCardProps) {
  return (
    <AnimatePresence mode="popLayout">
      <motion.div
        key={card.id}
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, x: -40, scale: 0.92 }}
        transition={{
          duration: 0.35,
          delay: dealDelay / 1000,
          ease: "easeOut",
        }}
      >
        <CardDisplay card={card} size={size} />
      </motion.div>
    </AnimatePresence>
  );
}

/* ─── Flip animation (card back → card face) ─── */
interface FlipCardProps {
  card: Card;
  size?: "normal" | "sm";
  /** When true the card face is shown; when false the back is shown. */
  isRevealed: boolean;
}

/**
 * A 3D-flip card. Starts showing the back, then flips to reveal the face
 * when `isRevealed` becomes true.
 */
export function FlipCard({ card, size = "normal", isRevealed }: FlipCardProps) {
  const hasFlipped = useRef(false);
  const [showFace, setShowFace] = useState(isRevealed);

  useEffect(() => {
    if (isRevealed && !hasFlipped.current) {
      hasFlipped.current = true;
      setShowFace(true);
    }
  }, [isRevealed]);

  return (
    <div className={`flip-card-wrapper${size === "sm" ? " flip-sm" : ""}`}>
      <motion.div
        className="flip-inner"
        initial={false}
        animate={{ rotateY: showFace ? 180 : 0 }}
        transition={{ duration: 0.45, ease: "easeInOut" }}
      >
        {/* Back face (visible at rotateY 0) */}
        <div className="flip-face flip-back">
          <CardBack size={size} />
        </div>
        {/* Front face (visible at rotateY 180) */}
        <div className="flip-face flip-front">
          <CardDisplay card={card} size={size} />
        </div>
      </motion.div>
    </div>
  );
}

/* ─── Chip counter with animated number ─── */
interface ChipCounterProps {
  value: number;
}

/**
 * Animated chip counter — when the value changes the number "ticks" to the
 * new value over ~300ms using a simple requestAnimationFrame counting approach.
 */
export function ChipCounter({ value }: ChipCounterProps) {
  const [display, setDisplay] = useState(value);
  const prevRef = useRef(value);
  const rafRef = useRef<number>(0);

  useEffect(() => {
    const from = prevRef.current;
    const to = value;
    prevRef.current = value;

    if (from === to) return;

    const duration = 300; // ms
    const start = performance.now();

    function tick(now: number) {
      const elapsed = now - start;
      const progress = Math.min(elapsed / duration, 1);
      // ease-out quad
      const eased = 1 - (1 - progress) * (1 - progress);
      setDisplay(Math.round(from + (to - from) * eased));
      if (progress < 1) {
        rafRef.current = requestAnimationFrame(tick);
      }
    }

    rafRef.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafRef.current);
  }, [value]);

  return <span className="chip-count chip-count-animated">{display}</span>;
}
