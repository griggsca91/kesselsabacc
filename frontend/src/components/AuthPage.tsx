import { useState } from "react";
import { useAuth } from "../hooks/useAuth";

interface AuthPageProps {
  onGuestPlay: () => void;
}

export function AuthPage({ onGuestPlay }: AuthPageProps) {
  const { signup, login } = useAuth();
  const [mode, setMode] = useState<"login" | "signup">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit() {
    setError("");
    setLoading(true);

    let err: string | null;
    if (mode === "signup") {
      if (!displayName.trim()) {
        setError("Enter a display name");
        setLoading(false);
        return;
      }
      if (!email.trim()) {
        setError("Enter your email");
        setLoading(false);
        return;
      }
      if (password.length < 8) {
        setError("Password must be at least 8 characters");
        setLoading(false);
        return;
      }
      err = await signup(email.trim(), password, displayName.trim());
    } else {
      if (!email.trim()) {
        setError("Enter your email");
        setLoading(false);
        return;
      }
      if (!password) {
        setError("Enter your password");
        setLoading(false);
        return;
      }
      err = await login(email.trim(), password);
    }

    if (err) {
      setError(err.trim());
    }
    setLoading(false);
  }

  function switchMode() {
    setMode(mode === "login" ? "signup" : "login");
    setError("");
  }

  return (
    <div className="lobby">
      <div className="lobby-brand">
        <div className="lobby-eyebrow">Star Wars</div>
        <h1>Kessel Sabacc</h1>
        <div className="lobby-divider" />
        <p className="lobby-subtitle">The card game of the galaxy's underworld</p>
      </div>

      <div className="lobby-card">
        <div className="lobby-card-title">
          {mode === "login" ? "Log In" : "Sign Up"}
        </div>

        {mode === "signup" && (
          <input
            placeholder="Display name"
            value={displayName}
            onChange={(e) => {
              setDisplayName(e.target.value);
              setError("");
            }}
            maxLength={20}
            autoFocus
          />
        )}

        <input
          placeholder="Email"
          type="email"
          value={email}
          onChange={(e) => {
            setEmail(e.target.value);
            setError("");
          }}
          autoFocus={mode === "login"}
        />

        <input
          placeholder="Password"
          type="password"
          value={password}
          onChange={(e) => {
            setPassword(e.target.value);
            setError("");
          }}
          onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
        />

        {error && <p className="error">{error}</p>}

        <button
          className="btn-primary"
          onClick={handleSubmit}
          disabled={loading}
        >
          {loading
            ? mode === "login"
              ? "Logging in..."
              : "Creating account..."
            : mode === "login"
              ? "Log In"
              : "Sign Up"}
        </button>

        <button className="btn-ghost" onClick={switchMode}>
          {mode === "login"
            ? "Need an account? Sign up"
            : "Already have an account? Log in"}
        </button>

        <div className="lobby-divider" />

        <button className="btn-ghost" onClick={onGuestPlay}>
          Continue as Guest
        </button>
      </div>
    </div>
  );
}
