import { useState } from "react";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { AuthPage } from "./components/AuthPage";
import { useGame } from "./hooks/useGame";
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

  // Show loading spinner while checking stored token
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

  // If not authenticated and not in guest mode, show auth page
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
