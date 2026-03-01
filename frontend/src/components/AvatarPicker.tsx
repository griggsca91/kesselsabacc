export interface AvatarOption {
  id: string;
  label: string;
  emoji: string;
  bg: string;
}

export const AVATARS: AvatarOption[] = [
  { id: "bounty-hunter", label: "Bounty Hunter", emoji: "\u2694", bg: "linear-gradient(135deg, #4a3020, #6a4a30)" },
  { id: "smuggler",      label: "Smuggler",      emoji: "\uD83D\uDE80", bg: "linear-gradient(135deg, #1a2a4a, #2a4060)" },
  { id: "jedi",          label: "Jedi",           emoji: "\u2726",  bg: "linear-gradient(135deg, #1a3a2a, #2a5a3a)" },
  { id: "sith",          label: "Sith",           emoji: "\u2604",  bg: "linear-gradient(135deg, #4a1010, #6a2020)" },
  { id: "droid",         label: "Droid",          emoji: "\uD83E\uDD16", bg: "linear-gradient(135deg, #2a2a3a, #3a3a5a)" },
  { id: "pilot",         label: "Pilot",          emoji: "\u2605",  bg: "linear-gradient(135deg, #3a2a10, #5a4020)" },
  { id: "wookiee",       label: "Wookiee",        emoji: "\uD83D\uDC3B", bg: "linear-gradient(135deg, #3a2a1a, #5a4a2a)" },
  { id: "mandalorian",   label: "Mandalorian",    emoji: "\uD83D\uDEE1",  bg: "linear-gradient(135deg, #1a2030, #2a3a50)" },
  { id: "scoundrel",     label: "Scoundrel",      emoji: "\uD83C\uDFB2", bg: "linear-gradient(135deg, #2a1a3a, #4a2a5a)" },
  { id: "mystic",        label: "Mystic",         emoji: "\uD83D\uDD2E", bg: "linear-gradient(135deg, #1a1a3a, #30306a)" },
];

export function getAvatar(id: string): AvatarOption {
  return AVATARS.find((a) => a.id === id) ?? AVATARS[0];
}

export function avatarForPlayerId(playerId: string): AvatarOption {
  let hash = 0;
  for (let i = 0; i < playerId.length; i++) {
    hash = (hash * 31 + playerId.charCodeAt(i)) | 0;
  }
  return AVATARS[Math.abs(hash) % AVATARS.length];
}

interface AvatarPanelProps {
  avatarId: string;
  size?: number;
  active?: boolean;
}

export function Avatar({ avatarId, size = 30, active = false }: AvatarPanelProps) {
  const avatar = getAvatar(avatarId);
  return (
    <div
      className={`avatar${active ? " avatar-active" : ""}`}
      style={{
        width: size,
        height: size,
        background: avatar.bg,
        fontSize: size * 0.48,
      }}
      title={avatar.label}
    >
      {avatar.emoji}
    </div>
  );
}

interface AvatarPickerProps {
  selected: string;
  onSelect: (id: string) => void;
}

export function AvatarPicker({ selected, onSelect }: AvatarPickerProps) {
  return (
    <div className="avatar-picker">
      <div className="avatar-picker-label">Choose your avatar</div>
      <div className="avatar-picker-grid">
        {AVATARS.map((a) => (
          <button
            key={a.id}
            type="button"
            className={`avatar-picker-option${a.id === selected ? " selected" : ""}`}
            onClick={() => onSelect(a.id)}
            title={a.label}
          >
            <div
              className="avatar"
              style={{
                width: 44,
                height: 44,
                background: a.bg,
                fontSize: 20,
              }}
            >
              {a.emoji}
            </div>
            <span className="avatar-picker-option-label">{a.label}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
