import { useState } from "react";

export function CopyInviteButton({ roomCode }: { roomCode: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    const url = `${window.location.origin}/join/${roomCode}`;
    navigator.clipboard.writeText(url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <button
      className={`btn-invite${copied ? " copied" : ""}`}
      onClick={handleCopy}
    >
      {copied ? "Copied!" : "Copy Invite Link"}
    </button>
  );
}
