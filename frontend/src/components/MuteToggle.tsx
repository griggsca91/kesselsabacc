import { useSound } from "../context/SoundContext";

export function MuteToggle() {
  const sound = useSound();
  if (!sound) return null;

  return (
    <button
      className="mute-toggle"
      onClick={sound.toggleMute}
      aria-label={sound.isMuted ? "Unmute sounds" : "Mute sounds"}
      title={sound.isMuted ? "Unmute sounds" : "Mute sounds"}
    >
      {sound.isMuted ? "\u{1F507}" : "\u{1F50A}"}
    </button>
  );
}
