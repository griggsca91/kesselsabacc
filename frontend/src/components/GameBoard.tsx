import { useRef } from "react";
import { motion } from "framer-motion";
import type { GameState, HandResult, ShiftToken } from "../types";
import { CardDisplay, CardBack } from "./CardDisplay";
import { AnimatedCard, FlipCard, ChipCounter } from "./AnimatedCard";

interface GameBoardProps {
  state: GameState;
  playerId: string;
  roomCode: string;
  onStartGame: () => void;
  onDraw: (suit: "sand" | "blood", token?: ShiftToken) => void;
  onStand: (token?: ShiftToken) => void;
  onNextRound: () => void;
}

const TOKEN_LABELS: Record<ShiftToken, string> = {
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

function HandRankBadge({ result }: { result: HandResult }) {
  if (result.rank === 0) return <span className="hand-rank-badge pure">Pure Sabacc</span>;
  if (result.rank === 1) return <span className="hand-rank-badge sabacc">Sabacc {result.value}/{result.value}</span>;
  return <span className="hand-rank-badge no-sabacc">No Sabacc (diff {result.value})</span>;
}

export function GameBoard({
  state,
  playerId,
  roomCode,
  onStartGame,
  onDraw,
  onStand,
  onNextRound,
}: GameBoardProps) {
  const me = state.players.find((p) => p.id === playerId);
  const isMyTurn = state.currentTurnPlayerId === playerId;
  const isReveal = state.phase === "reveal" || state.phase === "round_end" || state.phase === "game_over";

  // Track whether we already animated the deal for this round so the
  // stagger delay only fires once per round, not on every re-render.
  const dealtRoundRef = useRef<number>(-1);
  const isNewDeal = state.round !== dealtRoundRef.current &&
    (state.phase === "dealing" || state.phase === "turn");
  if (isNewDeal) {
    dealtRoundRef.current = state.round;
  }

  return (
    <div className="gameboard">

      {/* ── Header ── */}
      <header className="game-header">
        <h2 className="game-header-title">Kessel Sabacc</h2>
        <div className="game-meta">
          <span className="game-meta-item">
            Room <span className="room-code-badge">{roomCode}</span>
          </span>
          {state.phase !== "lobby" && (
            <>
              <span className="game-meta-item">Round <strong>{state.round}</strong></span>
              <span className="game-meta-item">Turn <strong>{state.turnInRound + 1}/3</strong></span>
            </>
          )}
        </div>
      </header>

      {/* ── Players bar ── */}
      <section className="players-bar">
        {state.players.map((p) => {
          const isActive = state.currentTurnPlayerId === p.id && state.phase === "turn";
          const isMe = p.id === playerId;
          return (
            <div
              key={p.id}
              className={[
                "player-panel",
                isMe ? "is-me" : "",
                isActive ? "active-turn" : "",
                p.eliminated ? "eliminated" : "",
              ].filter(Boolean).join(" ")}
            >
              {isActive && <div className="turn-indicator">Their Turn</div>}

              <div className="player-panel-header">
                <div className="player-name">
                  {p.name}
                  {p.isHost && <span className="player-host-star">★</span>}
                  {isMe && <span className="player-you-tag">you</span>}
                </div>
                {p.stood && <span className="stood-badge">stood</span>}
              </div>

              <div className="player-chips-row">
                <ChipCounter value={p.chips} />
                <span className="chip-label">chips</span>
                {p.invested > 0 && (
                  <span className="invested-badge">+{p.invested} in</span>
                )}
              </div>

              {isReveal && p.sandCard && p.bloodCard && (
                <div className="reveal-mini-cards">
                  <FlipCard card={p.sandCard} size="sm" isRevealed={isReveal} />
                  <FlipCard card={p.bloodCard} size="sm" isRevealed={isReveal} />
                </div>
              )}
            </div>
          );
        })}
      </section>

      {/* ── Lobby waiting ── */}
      {state.phase === "lobby" && (
        <section className="lobby-waiting">
          <div className="lobby-waiting-title">Waiting for players</div>
          <div className="player-count">{state.players.length} / 4</div>
          {me?.isHost ? (
            state.players.length >= 2 ? (
              <button className="btn-primary" onClick={onStartGame}>Start Game</button>
            ) : (
              <p>Share the room code — need at least 2 players</p>
            )
          ) : (
            <p>Waiting for the host to start the game…</p>
          )}
        </section>
      )}

      {/* ── Your hand ── */}
      {state.yourHand && state.phase !== "lobby" && (
        <section className="hand-section">
          <div className="hand-section-header">
            <div className="hand-section-title">Your Hand</div>
            {me && (
              <div className="hand-section-title">
                {me.chips} chips &nbsp;·&nbsp; {state.yourHand.tokens.length} tokens
              </div>
            )}
          </div>

          <div className="hand-cards-row">
            <div className="hand-card-slot">
              <span className="hand-card-suit-label sand">Sand</span>
              <AnimatedCard
                card={state.yourHand.sandCard}
                dealDelay={isNewDeal ? 0 : 0}
              />
            </div>
            <div className="hand-card-slot">
              <span className="hand-card-suit-label blood">Blood</span>
              <AnimatedCard
                card={state.yourHand.bloodCard}
                dealDelay={isNewDeal ? 200 : 0}
              />
            </div>
          </div>

          {/* Actions — my turn */}
          {state.phase === "turn" && isMyTurn && (
            <>
              <div className="actions-row">
                <button className="btn-draw-sand" onClick={() => onDraw("sand")}>
                  Draw Sand <span className="btn-chip-cost">−1 chip</span>
                </button>
                <button className="btn-draw-blood" onClick={() => onDraw("blood")}>
                  Draw Blood <span className="btn-chip-cost">−1 chip</span>
                </button>
                <button className="btn-stand" onClick={() => onStand()}>
                  Stand
                </button>
              </div>

              {state.yourHand.tokens.length > 0 && (
                <div className="tokens-section">
                  <div className="tokens-label">Shift Tokens</div>
                  <div className="tokens-row">
                    {state.yourHand.tokens.map((t) => (
                      <button key={t} className="btn-token" onClick={() => onStand(t)}>
                        {TOKEN_LABELS[t]}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </>
          )}

          {/* Waiting for opponent */}
          {state.phase === "turn" && !isMyTurn && (
            <div className="waiting-message">
              Waiting for{" "}
              <span className="waiting-name">
                {state.players.find((p) => p.id === state.currentTurnPlayerId)?.name ?? "opponent"}
              </span>
              …
            </div>
          )}
        </section>
      )}

      {/* ── Opponents' hidden cards (during play, not your hand) ── */}
      {state.phase === "turn" && (
        <section className="hand-section opponents-section">
          <div className="hand-section-header">
            <div className="hand-section-title">Opponents</div>
          </div>
          <div className="opponents-grid">
            {state.players
              .filter((p) => p.id !== playerId && !p.eliminated)
              .map((p, pIdx) => (
                <div key={p.id} className="opponent-slot">
                  <div className="opponent-cards">
                    <motion.div
                      initial={{ opacity: 0, y: 24 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.3, delay: isNewDeal ? 0.05 * pIdx : 0 }}
                    >
                      <CardBack />
                    </motion.div>
                    <motion.div
                      initial={{ opacity: 0, y: 24 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.3, delay: isNewDeal ? 0.05 * pIdx + 0.15 : 0 }}
                    >
                      <CardBack />
                    </motion.div>
                  </div>
                  <span className="opponent-name">{p.name}</span>
                </div>
              ))}
          </div>
        </section>
      )}

      {/* ── Round result ── */}
      {(state.phase === "round_end" || state.phase === "reveal") && state.lastResult && (
        <section className="round-result">
          <div className="round-result-header">
            <div className="round-result-title">Round {state.round} Result</div>
            <div className="round-winner">
              {state.lastResult.winnerIds.length > 1 ? "Tie" : (
                state.players.find((p) => state.lastResult!.winnerIds.includes(p.id))?.name
              )} wins
            </div>
          </div>

          <table className="result-table">
            <thead>
              <tr>
                <th>Player</th>
                <th>Hand</th>
                <th>Cards</th>
                <th>Change</th>
                <th>Chips</th>
              </tr>
            </thead>
            <tbody>
              {state.players.map((p) => {
                const hand = state.lastResult!.playerHands[p.id];
                const change = state.lastResult!.chipChanges[p.id] ?? 0;
                const isWinner = state.lastResult!.winnerIds.includes(p.id);
                return (
                  <tr key={p.id} className={isWinner ? "is-winner" : ""}>
                    <td>{p.name}{isWinner ? " ★" : ""}</td>
                    <td>{hand ? <HandRankBadge result={hand} /> : "—"}</td>
                    <td>
                      {hand && (
                        <div className="round-result-cards">
                          <CardDisplay card={hand.sandCard} size="sm" />
                          <CardDisplay card={hand.bloodCard} size="sm" />
                        </div>
                      )}
                    </td>
                    <td>
                      <span className={change > 0 ? "chip-gain" : change < 0 ? "chip-loss" : "chip-neutral"}>
                        {change > 0 ? `+${change}` : change}
                      </span>
                    </td>
                    <td>{p.chips}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>

          {state.phase === "round_end" && (
            <div className="round-result-actions">
              <button className="btn-primary" onClick={onNextRound}>Next Round</button>
            </div>
          )}
        </section>
      )}

      {/* ── Game over ── */}
      {state.phase === "game_over" && (
        <section className="game-over">
          <div className="game-over-label">Game Over</div>
          <h2 className="game-over-title">
            {state.players.find((p) => p.id === state.winnerId)?.name ?? "Unknown"}
          </h2>
          <p className="game-over-winner">wins the pot</p>
        </section>
      )}

      <div className="deck-info">
        <div className="deck-info-item">
          <span className="deck-dot sand" />
          Sand {state.sandRemaining}
        </div>
        <div className="deck-info-item">
          <span className="deck-dot blood" />
          Blood {state.bloodRemaining}
        </div>
      </div>

    </div>
  );
}
