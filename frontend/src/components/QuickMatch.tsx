import { useMatchmaking } from "../hooks/useMatchmaking";

interface QuickMatchProps {
  playerId: string;
  displayName?: string;
  token?: string | null;
  onMatched: (roomCode: string, playerName: string) => void;
}

export function QuickMatch({ playerId, displayName, token, onMatched }: QuickMatchProps) {
  const playerName = displayName?.trim() || "Player";

  const { isQueued, joinQueue, leaveQueue, queueTime } = useMatchmaking({
    playerId,
    playerName,
    token,
    onMatched: (roomCode) => onMatched(roomCode, playerName),
  });

  if (isQueued) {
    return (
      <div className="quick-match-panel quick-match-searching">
        <div className="quick-match-searching-inner">
          <span className="quick-match-dot" />
          <span className="quick-match-label">Searching for players...</span>
          <span className="quick-match-timer">{queueTime}s</span>
        </div>
        <button className="btn-ghost" onClick={leaveQueue}>
          Cancel
        </button>
      </div>
    );
  }

  return (
    <div className="quick-match-panel">
      <button className="btn-quick-match" onClick={joinQueue}>
        Quick Play
      </button>
      <p className="quick-match-hint">Auto-match with 2–4 players</p>
    </div>
  );
}
