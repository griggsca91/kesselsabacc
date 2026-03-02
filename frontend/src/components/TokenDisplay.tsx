import { useState } from "react";
import type { ShiftToken } from "../types";
import { TokenTooltip, TOKEN_NAMES } from "./TokenTooltip";

type TokenCategory = "economy" | "attack" | "defense" | "special";

const TOKEN_ABBREVS: Record<ShiftToken, string> = {
  free_draw: "FD",
  refund: "RF",
  general_tariff: "GT",
  markdown: "MD",
  immunity: "IM",
  major_fraud: "MF",
  cook_the_books: "CB",
  direct_transaction: "DT",
  prime_sabacc: "PS",
};

const TOKEN_CATEGORIES: Record<ShiftToken, TokenCategory> = {
  free_draw: "economy",
  refund: "economy",
  markdown: "economy",
  general_tariff: "attack",
  direct_transaction: "attack",
  major_fraud: "attack",
  immunity: "defense",
  cook_the_books: "special",
  prime_sabacc: "special",
};

interface TokenDisplayProps {
  tokens: ShiftToken[];
  usedTokens?: ShiftToken[];
  onUseToken?: (token: ShiftToken) => void;
  selectedToken?: ShiftToken | null;
  isMyTurn: boolean;
  compact?: boolean;
}

export function TokenDisplay({
  tokens,
  usedTokens = [],
  onUseToken,
  selectedToken,
  isMyTurn,
  compact,
}: TokenDisplayProps) {
  const [hoveredToken, setHoveredToken] = useState<ShiftToken | null>(null);

  if (compact) {
    return (
      <span className="token-count-badge" title={`${tokens.length} token${tokens.length !== 1 ? "s" : ""}`}>
        {tokens.length} {tokens.length === 1 ? "token" : "tokens"}
      </span>
    );
  }

  if (tokens.length === 0) {
    return (
      <div className="tokens-section">
        <div className="tokens-label">Shift Tokens</div>
        <div className="tokens-empty">No tokens available</div>
      </div>
    );
  }

  return (
    <div className="tokens-section">
      <div className="tokens-label">Shift Tokens</div>
      <div className="tokens-row">
        {tokens.map((token) => {
          const isUsed = usedTokens.includes(token);
          const isSelected = selectedToken === token;
          const category = TOKEN_CATEGORIES[token];
          const canClick = isMyTurn && !isUsed && onUseToken;

          return (
            <div
              key={token}
              className="token-chip-wrapper"
              onMouseEnter={() => setHoveredToken(token)}
              onMouseLeave={() => setHoveredToken(null)}
            >
              <button
                className={[
                  "token-chip",
                  `token-${category}`,
                  isUsed ? "used" : "",
                  isSelected ? "selected" : "",
                  canClick ? "interactive" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
                onClick={() => {
                  if (canClick) {
                    onUseToken(token);
                  }
                }}
                disabled={isUsed || !isMyTurn}
                title={TOKEN_NAMES[token]}
              >
                <span className="token-chip-abbrev">{TOKEN_ABBREVS[token]}</span>
                <span className="token-chip-name">{TOKEN_NAMES[token]}</span>
              </button>
              {hoveredToken === token && (
                <TokenTooltip token={token} isUsed={isUsed} />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
