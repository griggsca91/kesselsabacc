import { useState } from "react";
import { AuthProvider } from "./context/AuthContext";
import { SoundContext } from "./context/SoundContext";
import { useAuth } from "./hooks/useAuth";
import { AuthPage } from "./components/AuthPage";
import { useGame } from "./hooks/useGame";
import { useAvatar } from "./hooks/useAvatar";
import { useToast } from "./hooks/useToast";
import { useGameEvents } from "./hooks/useGameEvents";
import { useSoundEngine } from "./hooks/useSoundEngine";
import { useGameSounds } from "./hooks/useGameSounds";
import { useNotificationPreferences } from "./hooks/useNotificationPreferences";
import { useTurnNotification } from "./hooks/useTurnNotification";
import { useInviteCode } from "./hooks/useInviteCode";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import { MuteToggle } from "./components/MuteToggle";
import { ProfilePage } from "./components/ProfilePage";
import { ConnectionBanner } from "./components/ConnectionBanner";
import { ErrorBanner } from "./components/ErrorBanner";
import { ToastContainer } from "./components/Toast";
import { NotificationToggle } from "./components/NotificationToggle";
import "./App.css";

function AppInner() {
  const { user, token, isLoading, logout } = useAuth();
  const [guestMode, setGuestMode] = useState(false);
  const [profileUserId, setProfileUserId] = useState<string | null>(null);
  const inviteCode = useInviteCode();

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
  const { toasts, addToast } = useToast();
  const sound = useSoundEngine();
  useGameEvents(gameState, playerId, addToast);
  useGameSounds(gameState, playerId, sound);
  const { notificationsEnabled, toggleNotifications } =
    useNotificationPreferences();
  useTurnNotification(gameState, playerId, notificationsEnabled);

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

  if (!user && !guestMode) {
    return <AuthPage onGuestPlay={() => setGuestMode(true)} />;
  }

  if (!gameState && profileUserId) {
    return (
      <>
        <ErrorBanner message={error} onDismiss={clearError} />
        <ProfilePage
          userId={profileUserId}
          isOwnProfile={profileUserId === user?.id}
          token={token}
          onBack={() => setProfileUserId(null)}
          onViewProfile={(id) => setProfileUserId(id)}
        />
      </>
    );
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
          onViewProfile={(id) => setProfileUserId(id)}
          inviteCode={inviteCode ?? undefined}
        />
      </>
    );
  }

  return (
    <SoundContext.Provider value={sound}>
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
      <MuteToggle />
      <NotificationToggle
        enabled={notificationsEnabled}
        onToggle={toggleNotifications}
      />
      <ToastContainer toasts={toasts} />
    </SoundContext.Provider>
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
