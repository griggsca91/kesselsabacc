import { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "../types";

interface ChatPanelProps {
  messages: ChatMessage[];
  onSend: (text: string) => void;
  playerId: string;
}

function formatTime(ts: number): string {
  const d = new Date(ts);
  const h = d.getHours().toString().padStart(2, "0");
  const m = d.getMinutes().toString().padStart(2, "0");
  return `${h}:${m}`;
}

export function ChatPanel({ messages, onSend, playerId }: ChatPanelProps) {
  const [collapsed, setCollapsed] = useState(false);
  const [text, setText] = useState("");
  const [unread, setUnread] = useState(0);
  const prevLenRef = useRef(messages.length);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Track unread when collapsed
  useEffect(() => {
    if (messages.length > prevLenRef.current) {
      if (collapsed) {
        setUnread((n) => n + (messages.length - prevLenRef.current));
      }
    }
    prevLenRef.current = messages.length;
  }, [messages.length, collapsed]);

  // Clear unread on expand
  useEffect(() => {
    if (!collapsed) setUnread(0);
  }, [collapsed]);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    if (!collapsed && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, collapsed]);

  function handleSend() {
    const trimmed = text.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setText("");
  }

  return (
    <div className={`chat-panel${collapsed ? " collapsed" : ""}`}>
      <div className="chat-header" onClick={() => setCollapsed((c) => !c)}>
        <span className="chat-header-title">Chat</span>
        {collapsed && unread > 0 && (
          <span className="chat-unread-badge">{unread}</span>
        )}
        <span className="chat-collapse-icon">{collapsed ? "▲" : "▼"}</span>
      </div>

      {!collapsed && (
        <>
          <div className="chat-messages" ref={scrollRef}>
            {messages.length === 0 && (
              <div className="chat-empty">No messages yet</div>
            )}
            {messages.map((msg, i) => (
              <div
                key={i}
                className={`chat-message${msg.playerId === playerId ? " own" : ""}`}
                title={formatTime(msg.timestamp)}
              >
                <span className="chat-message-name">{msg.playerName}</span>
                <span className="chat-message-text">{msg.text}</span>
              </div>
            ))}
          </div>
          <div className="chat-input-row">
            <input
              className="chat-input"
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSend()}
              placeholder="Message..."
              maxLength={200}
            />
            <button className="btn-chat-send" onClick={handleSend}>
              Send
            </button>
          </div>
        </>
      )}
    </div>
  );
}
