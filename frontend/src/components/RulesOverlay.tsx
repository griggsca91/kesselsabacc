import { useState, useEffect, useCallback } from "react";

interface RulesOverlayProps {
  isOpen: boolean;
  onClose: () => void;
}

const SECTIONS = [
  {
    title: "How to Play",
    content: (
      <ul className="rules-list">
        <li>Each round, you receive 2 cards: one <strong>Sand</strong> and one <strong>Blood</strong>.</li>
        <li>On your turn, choose to <strong>Draw</strong> (replace a card; costs 1 chip) or <strong>Stand</strong> (keep your hand).</li>
        <li>Each round has <strong>3 turns</strong>, then all cards are revealed.</li>
        <li>The player with the <strong>lowest hand difference</strong> wins the round pot.</li>
      </ul>
    ),
  },
  {
    title: "Hand Rankings",
    content: (
      <div className="rules-rankings">
        <div className="rules-rank">
          <span className="rules-rank-badge pure">1. Pure Sabacc</span>
          <span className="rules-rank-desc">Both cards are Sylops (0 + 0) — best possible hand</span>
        </div>
        <div className="rules-rank">
          <span className="rules-rank-badge sabacc">2. Sabacc</span>
          <span className="rules-rank-desc">Same value on both cards (e.g. 3 + 3)</span>
        </div>
        <div className="rules-rank">
          <span className="rules-rank-badge no-sabacc">3. No Sabacc</span>
          <span className="rules-rank-desc">Different values — ranked by difference (lower is better)</span>
        </div>
      </div>
    ),
  },
  {
    title: "Card Types",
    content: (
      <ul className="rules-list">
        <li><strong>Value cards (1-6):</strong> Fixed face value.</li>
        <li><strong>Impostor:</strong> Random value 1-6 determined at reveal (dice roll).</li>
        <li><strong>Sylop:</strong> Value 0. Two Sylops make a Pure Sabacc.</li>
      </ul>
    ),
  },
  {
    title: "Shift Tokens",
    content: (
      <div className="rules-tokens-grid">
        <div className="rules-token">
          <strong>Free Draw</strong>
          <span>Draw a card without paying a chip.</span>
        </div>
        <div className="rules-token">
          <strong>Refund</strong>
          <span>Gain 2 chips back.</span>
        </div>
        <div className="rules-token">
          <strong>General Tariff</strong>
          <span>All other players pay 1 chip to the pot.</span>
        </div>
        <div className="rules-token">
          <strong>Markdown</strong>
          <span>Reduce your hand difference by 1.</span>
        </div>
        <div className="rules-token">
          <strong>Immunity</strong>
          <span>Protect yourself from other tokens this round.</span>
        </div>
        <div className="rules-token">
          <strong>Major Fraud</strong>
          <span>Swap one of your cards with a random one from the deck.</span>
        </div>
        <div className="rules-token">
          <strong>Cook the Books</strong>
          <span>Peek at the top card of either deck before drawing.</span>
        </div>
        <div className="rules-token">
          <strong>Direct Transaction</strong>
          <span>Steal a random card from another player.</span>
        </div>
        <div className="rules-token">
          <strong>Prime Sabacc</strong>
          <span>Your hand is treated as a Sabacc for this round.</span>
        </div>
      </div>
    ),
  },
  {
    title: "Scoring",
    content: (
      <ul className="rules-list">
        <li>Each player starts with <strong>10 chips</strong>.</li>
        <li>Drawing a card costs <strong>1 chip</strong> (added to the pot).</li>
        <li>The round winner takes the <strong>entire pot</strong>.</li>
        <li>A player with <strong>0 chips</strong> is eliminated.</li>
        <li>Last player standing wins the game.</li>
      </ul>
    ),
  },
];

export function RulesOverlay({ isOpen, onClose }: RulesOverlayProps) {
  const [openSection, setOpenSection] = useState(0);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    },
    [onClose],
  );

  useEffect(() => {
    if (!isOpen) return;
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, handleKeyDown]);

  if (!isOpen) return null;

  const toggleSection = (index: number) => {
    setOpenSection(openSection === index ? -1 : index);
  };

  return (
    <div className="rules-overlay-backdrop" onClick={onClose}>
      <div
        className="rules-overlay-content"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="rules-overlay-header">
          <h2 className="rules-overlay-title">Rules & Help</h2>
          <button
            className="rules-overlay-close"
            onClick={onClose}
            aria-label="Close rules"
          >
            {"\u2715"}
          </button>
        </div>

        <div className="rules-sections">
          {SECTIONS.map((section, i) => (
            <div
              key={section.title}
              className={`rules-section ${openSection === i ? "open" : ""}`}
            >
              <button
                className="rules-section-header"
                onClick={() => toggleSection(i)}
              >
                <span className="rules-section-arrow">
                  {openSection === i ? "\u25BE" : "\u25B8"}
                </span>
                <span className="rules-section-title">{section.title}</span>
              </button>
              {openSection === i && (
                <div className="rules-section-content">{section.content}</div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
