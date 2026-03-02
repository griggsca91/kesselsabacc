import { useEffect, useState } from "react";
import type { ToastItem } from "../hooks/useToast";

const TOAST_VISIBLE_MS = 3500;

function Toast({ toast }: { toast: ToastItem }) {
  const [exiting, setExiting] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setExiting(true), TOAST_VISIBLE_MS);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div
      className={`toast toast-${toast.type}${exiting ? " exiting" : ""}`}
      role="status"
      aria-live="polite"
    >
      {toast.message}
    </div>
  );
}

export function ToastContainer({ toasts }: { toasts: ToastItem[] }) {
  if (toasts.length === 0) return null;

  return (
    <div className="toast-container">
      {toasts.map((t) => (
        <Toast key={t.id} toast={t} />
      ))}
    </div>
  );
}
