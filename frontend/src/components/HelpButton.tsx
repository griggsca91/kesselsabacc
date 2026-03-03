interface HelpButtonProps {
  onClick: () => void;
}

export function HelpButton({ onClick }: HelpButtonProps) {
  return (
    <button
      className="help-button"
      onClick={onClick}
      aria-label="Game rules and help"
      title="Rules & Help"
    >
      ?
    </button>
  );
}
