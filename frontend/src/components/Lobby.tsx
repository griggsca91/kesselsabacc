import { useState } from "react";

interface LobbyProps {
  onCreateRoom: (name: string) => Promise<void>;
  onJoinRoom: (code: string, name: string) => Promise<void>;
}

export function Lobby({ onCreateRoom, onJoinRoom }: LobbyProps) {
  const [name, setName] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [mode, setMode] = useState<"home" | "create" | "join">("home");
  const [loading, setLoading] = useState(false);
  const [localError, setLocalError] = useState("");

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
            <div className="lobby-home-buttons">
              <button className="btn-primary" onClick={() => setMode("create")}>
                Create Room
              </button>
              <button className="btn-ghost" onClick={() => setMode("join")}>
                Join Room
              </button>
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
            {err && <p className="error">{err}</p>}
            <button className="btn-primary" onClick={handleCreate} disabled={loading}>
              {loading ? "Creating…" : "Create Room"}
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
            <input
              placeholder="Room code — e.g. AB3K"
              value={joinCode}
              onChange={(e) => { setJoinCode(e.target.value.toUpperCase()); setLocalError(""); }}
              onKeyDown={(e) => e.key === "Enter" && handleJoin()}
              maxLength={4}
              style={{ letterSpacing: "0.25em", textTransform: "uppercase" }}
            />
            {err && <p className="error">{err}</p>}
            <button className="btn-primary" onClick={handleJoin} disabled={loading}>
              {loading ? "Joining…" : "Join Room"}
            </button>
            <button className="btn-ghost" onClick={back}>Back</button>
          </>
        )}
      </div>
    </div>
  );
}
