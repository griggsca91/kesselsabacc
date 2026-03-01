import { useGame } from "./hooks/useGame";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import "./App.css";

function App() {
  const {
    gameState,
    error,
    playerId,
    roomCode,
    createRoom,
    joinRoom,
    startGame,
    draw,
    stand,
    nextRound,
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
  );
}

export default App;
