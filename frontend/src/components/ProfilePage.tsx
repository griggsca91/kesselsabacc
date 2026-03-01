import { usePlayerProfile } from "../hooks/usePlayerProfile";
import { GameHistory } from "./GameHistory";

interface ProfilePageProps {
  /** The user ID whose profile to display. */
  userId: string;
  /** Whether this is the current user's own profile. */
  isOwnProfile: boolean;
  /** JWT auth token for authenticated API calls. */
  token?: string | null;
  /** Callback to navigate back to the lobby. */
  onBack: () => void;
  /** Callback to view another player's profile. */
  onViewProfile?: (userId: string) => void;
}

function formatHandRank(rank: string): string {
  switch (rank) {
    case "pure_sabacc":
      return "Pure Sabacc";
    case "sabacc":
      return "Sabacc";
    case "no_sabacc":
      return "No Sabacc";
    default:
      return rank;
  }
}

function handRankClass(rank: string): string {
  switch (rank) {
    case "pure_sabacc":
      return "pure";
    case "sabacc":
      return "sabacc";
    default:
      return "no-sabacc";
  }
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, {
    month: "long",
    year: "numeric",
  });
}

function formatWinRate(rate: number): string {
  return `${(rate * 100).toFixed(1)}%`;
}

export function ProfilePage({
  userId,
  isOwnProfile,
  token,
  onBack,
}: ProfilePageProps) {
  const { profile, isLoading, error, refresh } = usePlayerProfile({
    userId: isOwnProfile ? undefined : userId,
    token,
  });

  if (isLoading) {
    return (
      <div className="lobby">
        <div className="lobby-brand">
          <div className="lobby-eyebrow">Star Wars</div>
          <h1>Kessel Sabacc</h1>
          <div className="lobby-divider" />
          <p className="lobby-subtitle">Loading profile...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="lobby">
        <div className="lobby-brand">
          <div className="lobby-eyebrow">Star Wars</div>
          <h1>Kessel Sabacc</h1>
          <div className="lobby-divider" />
        </div>
        <div className="lobby-card">
          <p className="error">{error}</p>
          <button className="btn-ghost" onClick={onBack}>
            Back to Lobby
          </button>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="lobby">
        <div className="lobby-brand">
          <div className="lobby-eyebrow">Star Wars</div>
          <h1>Kessel Sabacc</h1>
          <div className="lobby-divider" />
        </div>
        <div className="lobby-card">
          <p className="error">Player not found</p>
          <button className="btn-ghost" onClick={onBack}>
            Back to Lobby
          </button>
        </div>
      </div>
    );
  }

  const { user, stats } = profile;

  return (
    <div className="lobby">
      <div className="lobby-brand">
        <div className="lobby-eyebrow">Star Wars</div>
        <h1>Kessel Sabacc</h1>
        <div className="lobby-divider" />
      </div>

      <div className="profile-card">
        <div className="profile-header">
          <div className="profile-avatar">
            {user.displayName.charAt(0).toUpperCase()}
          </div>
          <div className="profile-identity">
            <h2 className="profile-name">{user.displayName}</h2>
            {user.email && (
              <span className="profile-email">{user.email}</span>
            )}
            <span className="profile-member-since">
              Member since {formatDate(user.memberSince)}
            </span>
          </div>
        </div>

        <div className="profile-stats-grid">
          <div className="profile-stat">
            <span className="profile-stat-value">{stats.gamesPlayed}</span>
            <span className="profile-stat-label">Games</span>
          </div>
          <div className="profile-stat">
            <span className="profile-stat-value chip-gain">{stats.wins}</span>
            <span className="profile-stat-label">Wins</span>
          </div>
          <div className="profile-stat">
            <span className="profile-stat-value chip-loss">{stats.losses}</span>
            <span className="profile-stat-label">Losses</span>
          </div>
          <div className="profile-stat">
            <span className="profile-stat-value">
              {stats.gamesPlayed > 0 ? formatWinRate(stats.winRate) : "--"}
            </span>
            <span className="profile-stat-label">Win Rate</span>
          </div>
        </div>

        {stats.bestHand && (
          <div className="profile-best-hand">
            <span className="profile-best-hand-label">Best Hand</span>
            <span
              className={`hand-rank-badge ${handRankClass(stats.bestHand)}`}
            >
              {formatHandRank(stats.bestHand)}
            </span>
          </div>
        )}

        <div className="profile-actions">
          <button className="btn-ghost btn-sm" onClick={refresh}>
            Refresh
          </button>
          <button className="btn-ghost btn-sm" onClick={onBack}>
            Back to Lobby
          </button>
        </div>
      </div>

      {profile.games.length > 0 && (
        <GameHistory
          playerId={userId}
          token={token}
          onClose={onBack}
        />
      )}

      {profile.games.length === 0 && (
        <p
          style={{
            textAlign: "center",
            color: "var(--text-muted)",
            marginTop: "1.5rem",
            fontSize: "0.9rem",
            fontStyle: "italic",
          }}
        >
          No games played yet
        </p>
      )}
    </div>
  );
}
