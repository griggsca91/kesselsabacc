import { useCallback, useEffect, useState } from "react";

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

interface PublicRoom {
  code: string;
  playerCount: number;
  maxPlayers: number;
}

interface RoomBrowserProps {
  onJoinRoom: (code: string) => void;
}

export function RoomBrowser({ onJoinRoom }: RoomBrowserProps) {
  const [rooms, setRooms] = useState<PublicRoom[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchRooms = useCallback(async () => {
    try {
      const res = await fetch(`${API}/api/rooms`);
      if (res.ok) {
        const data = await res.json();
        setRooms(data ?? []);
      }
    } catch {
      // ignore transient errors
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRooms();
    const id = setInterval(fetchRooms, 5000);
    return () => clearInterval(id);
  }, [fetchRooms]);

  return (
    <div className="room-browser">
      <div className="room-browser-header">
        <span className="room-browser-title">Open Rooms</span>
        <button className="btn-ghost room-browser-refresh" onClick={fetchRooms}>
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="room-browser-empty">Loading...</div>
      ) : rooms.length === 0 ? (
        <div className="room-browser-empty">No public rooms available</div>
      ) : (
        <div className="room-browser-list">
          {rooms.map((r) => (
            <div key={r.code} className="room-card">
              <span className="room-card-code">{r.code}</span>
              <span className="room-card-players">
                {r.playerCount} / {r.maxPlayers} players
              </span>
              <button
                className="btn-primary room-card-join"
                onClick={() => onJoinRoom(r.code)}
              >
                Join
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
