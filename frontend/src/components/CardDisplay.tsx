import type { Card } from "../types";

interface CardDisplayProps {
  card: Card;
  size?: "normal" | "sm";
}

const suitSymbol: Record<string, string> = { sand: "◆", blood: "♦" };
const suitLabel: Record<string, string> = { sand: "Sand", blood: "Blood" };

function centerValue(card: Card): string {
  if (card.kind === "sylop") return "◈";
  if (card.kind === "impostor") return "?";
  return String(card.value);
}

function centerLabel(card: Card): string {
  if (card.kind === "sylop") return "Sylop";
  if (card.kind === "impostor") return "Impostor";
  return suitLabel[card.suit];
}

function cornerValue(card: Card): string {
  if (card.kind === "sylop") return "◈";
  if (card.kind === "impostor") return "?";
  return String(card.value);
}

export function CardDisplay({ card, size = "normal" }: CardDisplayProps) {
  const sizeClass = size === "sm" ? " card-sm" : "";
  const suitClass = ` suit-${card.suit}`;
  const kindClass = ` kind-${card.kind}`;

  return (
    <div className={`card${suitClass}${kindClass}${sizeClass}`}>
      {/* Inner pattern overlay for texture */}
      <div className="card-pattern" />

      {/* Sylop shimmer overlay */}
      {card.kind === "sylop" && <div className="card-sylop-shimmer" />}

      {/* Impostor mask overlay */}
      {card.kind === "impostor" && <div className="card-impostor-overlay" />}

      <div className="card-corner">
        <span className="card-corner-value">{cornerValue(card)}</span>
        <span className="card-corner-suit">{suitSymbol[card.suit]}</span>
      </div>

      <div className="card-center">
        <span className="card-center-value">{centerValue(card)}</span>
        <span className="card-center-label">{centerLabel(card)}</span>
      </div>

      <div className="card-corner card-corner-br">
        <span className="card-corner-value">{cornerValue(card)}</span>
        <span className="card-corner-suit">{suitSymbol[card.suit]}</span>
      </div>
    </div>
  );
}

export function CardBack({ size = "normal" }: { size?: "normal" | "sm" }) {
  const sizeClass = size === "sm" ? " card-back-sm" : "";
  return (
    <div className={`card-back${sizeClass}`}>
      <div className="card-back-pattern">
        <div className="card-back-diamond" />
        <span className="card-back-icon">✦</span>
        <div className="card-back-diamond card-back-diamond-2" />
      </div>
    </div>
  );
}
