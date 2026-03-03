interface NotificationToggleProps {
  enabled: boolean;
  onToggle: () => void;
}

export function NotificationToggle({
  enabled,
  onToggle,
}: NotificationToggleProps) {
  return (
    <div className="notification-toggle-wrapper">
      <button
        className="btn-notification-toggle"
        onClick={onToggle}
        title={enabled ? "Mute turn notifications" : "Enable turn notifications"}
        aria-label={
          enabled ? "Mute turn notifications" : "Enable turn notifications"
        }
      >
        {enabled ? "\u{1F514}" : "\u{1F515}"}
      </button>
    </div>
  );
}
