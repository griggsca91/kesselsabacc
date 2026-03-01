import { useGame } from "./hooks/useGame";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import { ConnectionBanner } from "./components/ConnectionBanner";
import "./App.css";

function App() {
  const {
    gameState,
    error,
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
      <Lobby
        onCreateRoom={createRoom}
        onJoinRoom={joinRoom}
        error={error}
      />
    );
  }

  return (
    <>
      <ConnectionBanner status={connectionStatus} onRetry={reconnect} />
      <GameBoard
        state={gameState}
        playerId={playerId}
        roomCode={roomCode}
        onStartGame={startGame}
        onDraw={draw}
        onStand={stand}
        onNextRound={nextRound}
        error={error}
      />
    </>
  );
}

export default App;
