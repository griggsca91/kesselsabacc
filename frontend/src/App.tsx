import { useGame } from "./hooks/useGame";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import { ConnectionBanner } from "./components/ConnectionBanner";
import { ErrorBanner } from "./components/ErrorBanner";
import "./App.css";

function App() {
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
  } = useGame();

  if (!gameState) {
    return (
      <>
        <ErrorBanner message={error} onDismiss={clearError} />
        <Lobby
          onCreateRoom={createRoom}
          onJoinRoom={joinRoom}
          playerId={playerId}
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

export default App;
