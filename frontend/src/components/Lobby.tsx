import { useState } from "react";
import { GameHistory } from "./GameHistory";
import { AvatarPicker, AVATARS } from "./AvatarPicker";
import { useAvatar } from "../hooks/useAvatar";
import { QuickMatch } from "./QuickMatch";

interface LobbyProps {
  onCreateRoom: (name: string) => Promise<void>;
  onJoinRoom: (code: string, name: string) => Promise<void>;
  playerId: string;
  displayName?: string;
  isAuthenticated?: boolean;
  token?: string | null;
  onLogout?: () => void;
  /** Callback to view a player's profile. */
  onViewProfile?: (userId: string) => void;
  /** Pre-filled room code from an invite link. */
  inviteCode?: string;
}

export function Lobby({
  onCreateRoom,
  onJoinRoom,
  playerId,
  displayName,
  isAuthenticated,
  token,
  onLogout,
  onViewProfile,
  inviteCode,
}: LobbyProps) {
  const [name, setName] = useState(displayName ?? "");
  const [joinCode, setJoinCode] = useState(inviteCode ?? "");
  const [mode, setMode] = useState<"home" | "create" | "join">(inviteCode ? "join" : "home");
  const [loading, setLoading] = useState(false);
  const [localError, setLocalError] = useState("");
  const [showHistory, setShowHistory] = useState(false);
  const { avatarId, setAvatarId } = useAvatar();

  if (!avatarId) {
    const random = AVATARS[Math.floor(Math.random() * AVATARS.length)];
    setAvatarId(random.id);
  }

  async function handleCreate() {
    if (!name.trim()) return setLocalError("Enter your name");
    setLoading(true);
    setLocalError("");
    await onCreateRoom(name.trim());
    setLoading(false);
  }

  async function handleJoin() {
    if (!name.trim()) return setLocalError("Enter your name");
    if (joinCode.trim().length !== 4) return setLocalError("Enter a 4-character room code");
    setLoading(true);
    setLocalError("");
    await onJoinRoom(joinCode.trim(), name.trim());
    setLoading(false);
  }

  function back() {
    setMode("home");
    setLocalError("");
  }

  const err = localError;

  return (
    <div className="lobby">
      <div className="lobby-brand">
        <div className="lobby-eyebrow">Star Wars</div>
        <h1>Kessel Sabacc</h1>
        <div className="lobby-divider" />
        <p className="lobby-subtitle">The card game of the galaxy's underworld</p>
      </div>

      <div className="lobby-card">
        {mode === "home" && (
          <>
            <div className="lobby-card-title">Choose your table</div>
            {isAuthenticated && (
              <p style={{ textAlign: "center", fontSize: "0.85rem", color: "var(--text-muted)" }}>
                Signed in as <strong style={{ color: "var(--text)" }}>{displayName}</strong>
              </p>
            )}
            <div className="lobby-home-buttons">
              <button className="btn-primary" onClick={() => setMode("create")}>
                Create Room
              </button>
              <button className="btn-ghost" onClick={() => setMode("join")}>
                Join Room
              </button>
              {isAuthenticated && onViewProfile && (
                <button className="btn-ghost" onClick={() => onViewProfile(playerId)}>
                  My Profile
                </button>
              )}
              {isAuthenticated && onLogout && (
                <button className="btn-ghost" onClick={onLogout}>
                  Log Out
                </button>
              )}
            </div>
          </>
        )}

        {mode === "create" && (
          <>
            <div className="lobby-card-title">New Room</div>
            <input
              placeholder="Your name"
              value={name}
              onChange={(e) => { setName(e.target.value); setLocalError(""); }}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              maxLength={20}
              autoFocus
            />
            <AvatarPicker selected={avatarId} onSelect={setAvatarId} />
            {err && <p className="error">{err}</p>}
            <button className="btn-primary" onClick={handleCreate} disabled={loading}>
              {loading ? "Creating..." : "Create Room"}
            </button>
            <button className="btn-ghost" onClick={back}>Back</button>
          </>
        )}

        {mode === "join" && (
          <>
            <div className="lobby-card-title">Join Room</div>
            <input
              placeholder="Your name"
              value={name}
              onChange={(e) => { setName(e.target.value); setLocalError(""); }}
              maxLength={20}
              autoFocus
            />
            <AvatarPicker selected={avatarId} onSelect={setAvatarId} />
            <input
              placeholder="Room code -- e.g. AB3K"
              value={joinCode}
              onChange={(e) => { setJoinCode(e.target.value.toUpperCase()); setLocalError(""); }}
              onKeyDown={(e) => e.key === "Enter" && handleJoin()}
              maxLength={4}
              style={{ letterSpacing: "0.25em", textTransform: "uppercase" }}
            />
            {err && <p className="error">{err}</p>}
            <button className="btn-primary" onClick={handleJoin} disabled={loading}>
              {loading ? "Joining..." : "Join Room"}
            </button>
            <button className="btn-ghost" onClick={back}>Back</button>
          </>
        )}
      </div>

      <div className="lobby-history-toggle">
        <button
          className="btn-ghost"
          onClick={() => setShowHistory(!showHistory)}
        >
          {showHistory ? "Hide Game History" : "Game History"}
        </button>
      </div>

      {showHistory && (
        <GameHistory
          playerId={playerId}
          token={token}
          onClose={() => setShowHistory(false)}
        />
      )}
    </div>
  );
}
