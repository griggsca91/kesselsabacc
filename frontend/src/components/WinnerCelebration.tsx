import { motion } from "framer-motion";

interface WinnerCelebrationProps {
  winnerName: string;
  chipChange: number;
}

/**
 * Winner highlight with golden glow and CSS-only sparkle particles.
 */
export function WinnerCelebration({ winnerName, chipChange }: WinnerCelebrationProps) {
  return (
    <motion.div
      className="winner-celebration"
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.5, ease: "backOut" }}
    >
      {/* Sparkle particles */}
      <div className="winner-sparkles">
        {Array.from({ length: 12 }).map((_, i) => (
          <span
            key={i}
            className="winner-sparkle"
            style={{
              "--sparkle-delay": `${i * 0.15}s`,
              "--sparkle-x": `${Math.cos((i * Math.PI * 2) / 12) * 50 + 50}%`,
              "--sparkle-y": `${Math.sin((i * Math.PI * 2) / 12) * 50 + 50}%`,
            } as React.CSSProperties}
          />
        ))}
      </div>

      <motion.div
        className="winner-text"
        initial={{ opacity: 0, scale: 0.3 }}
        animate={{ opacity: 1, scale: [1.4, 1] }}
        transition={{ duration: 0.5, delay: 0.15, ease: "backOut" }}
      >
        Winner!
      </motion.div>

      <motion.div
        className="winner-name"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, delay: 0.35 }}
      >
        {winnerName}
      </motion.div>

      {chipChange !== 0 && (
        <motion.div
          className={`winner-chips ${chipChange > 0 ? "chip-gain" : "chip-loss"}`}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.5 }}
        >
          {chipChange > 0 ? `+${chipChange}` : chipChange} chips
        </motion.div>
      )}
    </motion.div>
  );
}
