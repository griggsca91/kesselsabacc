import type { ShiftToken } from "../types";

const TOKEN_NAMES: Record<ShiftToken, string> = {
  free_draw: "Free Draw",
  refund: "Refund 2",
  general_tariff: "General Tariff",
  markdown: "Markdown",
  immunity: "Immunity",
  major_fraud: "Major Fraud",
  cook_the_books: "Cook the Books",
  direct_transaction: "Direct Transaction",
  prime_sabacc: "Prime Sabacc",
};

const TOKEN_DESCRIPTIONS: Record<ShiftToken, string> = {
  free_draw: "Draw a card without paying a chip",
  refund: "Get 2 chips back from the pot",
  general_tariff: "All opponents pay 1 chip to the pot",
  markdown: "Reduce your investment by 2",
  immunity: "Protect yourself from elimination this round",
  major_fraud: "Swap one of your cards with a random opponent's card",
  cook_the_books: "Look at the top card of either deck before drawing",
  direct_transaction: "Steal 2 chips from any opponent",
  prime_sabacc: "Automatically win this round with Pure Sabacc",
};

interface TokenTooltipProps {
  token: ShiftToken;
  isUsed?: boolean;
}

export function TokenTooltip({ token, isUsed }: TokenTooltipProps) {
  return (
    <div className="token-tooltip">
      <div className="token-tooltip-arrow" />
      <div className="token-tooltip-name">{TOKEN_NAMES[token]}</div>
      <div className="token-tooltip-desc">{TOKEN_DESCRIPTIONS[token]}</div>
      {isUsed && <div className="token-tooltip-used">Already used</div>}
      {!isUsed && (
        <div className="token-tooltip-hint">Click during your turn to use</div>
      )}
    </div>
  );
}

export { TOKEN_NAMES, TOKEN_DESCRIPTIONS };
