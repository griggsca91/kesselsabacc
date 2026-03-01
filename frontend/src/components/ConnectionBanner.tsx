import type { ConnectionStatus } from "../types";

interface ConnectionBannerProps {
  status: ConnectionStatus;
  onRetry: () => void;
}

export function ConnectionBanner({ status, onRetry }: ConnectionBannerProps) {
  if (status === "connected") return null;

  return (
    <div className={`connection-banner connection-${status}`}>
      {status === "reconnecting" && (
        <>
          <span className="connection-spinner" />
          <span>Reconnecting...</span>
        </>
      )}
      {status === "disconnected" && (
        <>
          <span>Disconnected from server</span>
          <button className="btn-retry" onClick={onRetry}>
            Retry
          </button>
        </>
      )}
    </div>
  );
}
