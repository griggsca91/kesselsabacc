import { useGameHistory } from "../hooks/useGameHistory";

interface GameHistoryProps {
  playerId: string;
  /** JWT auth token for authenticated history lookups. */
  token?: string | null;
  onClose: () => void;
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function GameHistory({ playerId, token, onClose }: GameHistoryProps) {
  const { games, isLoading, error, refresh } = useGameHistory({ playerId, token });

  const myResult = (game: (typeof games)[number]) => {
    const me = game.players.find((p) => p.userId === playerId);
    if (!me) return null;
    return me;
  };

  return (
    <div className="game-history">
      <div className="game-history-header">
        <div className="lobby-card-title">Game History</div>
        <div className="game-history-actions">
          <button className="btn-ghost btn-sm" onClick={refresh} disabled={isLoading}>
            {isLoading ? "Loading..." : "Refresh"}
          </button>
          <button className="btn-ghost btn-sm" onClick={onClose}>
            Close
          </button>
        </div>
      </div>

      {error && <p className="error">{error}</p>}

      {!isLoading && !error && games.length === 0 && (
        <p className="game-history-empty">No games played yet</p>
      )}

      {games.length > 0 && (
        <div className="game-history-list">
          {games.map((game) => {
            const me = myResult(game);
            const didWin = me?.isWinner ?? false;
            return (
              <div
                key={game.gameId}
                className={`game-history-entry ${didWin ? "won" : "lost"}`}
              >
                <div className="game-history-entry-top">
                  <span className="game-history-date">
                    {formatDate(game.finishedAt)}
                  </span>
                  <span className="room-code-badge">{game.roomCode}</span>
                </div>
                <div className="game-history-entry-details">
                  <span className="game-history-rounds">
                    {game.rounds} {game.rounds === 1 ? "round" : "rounds"}
                  </span>
                  <span className="game-history-winner">
                    Winner: <strong>{game.winnerName}</strong>
                  </span>
                  {me && (
                    <span
                      className={`game-history-result ${
                        didWin ? "chip-gain" : "chip-loss"
                      }`}
                    >
                      {didWin ? "Won" : "Lost"} &mdash; {me.finalChips} chips
                    </span>
                  )}
                </div>
                <div className="game-history-players">
                  {game.players.map((p) => (
                    <span
                      key={p.userId}
                      className={`game-history-player ${
                        p.isWinner ? "is-winner" : ""
                      } ${p.userId === playerId ? "is-me" : ""}`}
                    >
                      {p.displayName} ({p.finalChips})
                    </span>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
