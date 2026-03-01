import { useState } from "react";
import { AuthProvider } from "./context/AuthContext";
import { useAuth } from "./hooks/useAuth";
import { AuthPage } from "./components/AuthPage";
import { useGame } from "./hooks/useGame";
import { useAvatar } from "./hooks/useAvatar";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import { ConnectionBanner } from "./components/ConnectionBanner";
import { ErrorBanner } from "./components/ErrorBanner";
import "./App.css";

function AppInner() {
  const { user, token, isLoading, logout } = useAuth();
  const [guestMode, setGuestMode] = useState(false);

  const {
    gameState,
    error,
    clearError,
    connectionStatus,
    playerId,
    roomCode,
    createRoom,
    joinRoom,
    startGame,
    draw,
    stand,
    nextRound,
    reconnect,
  } = useGame({
    playerId: user?.id,
    token,
  });

  const { avatarId } = useAvatar();

  if (isLoading) {
    return (
      <div className="lobby">
        <div className="lobby-brand">
          <div className="lobby-eyebrow">Star Wars</div>
          <h1>Kessel Sabacc</h1>
          <div className="lobby-divider" />
          <p className="lobby-subtitle">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user && !guestMode) {
    return <AuthPage onGuestPlay={() => setGuestMode(true)} />;
  }

  if (!gameState) {
    return (
      <>
        <ErrorBanner message={error} onDismiss={clearError} />
        <Lobby
          onCreateRoom={createRoom}
          onJoinRoom={joinRoom}
          playerId={playerId}
          displayName={user?.displayName}
          isAuthenticated={!!user}
          token={token}
          onLogout={() => {
            logout();
            setGuestMode(false);
          }}
        />
      </>
    );
  }

  return (
    <>
      <ConnectionBanner status={connectionStatus} onRetry={reconnect} />
      <ErrorBanner message={error} onDismiss={clearError} />
      <GameBoard
        state={gameState}
        playerId={playerId}
        roomCode={roomCode}
        avatarId={avatarId}
        onStartGame={startGame}
        onDraw={draw}
        onStand={stand}
        onNextRound={nextRound}
      />
    </>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppInner />
    </AuthProvider>
  );
}

export default App;
