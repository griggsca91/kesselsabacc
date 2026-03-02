import { useCallback, useRef, useState } from "react";

export type ToastType = "info" | "success" | "warning";

export interface ToastItem {
  id: string;
  message: string;
  type: ToastType;
}

const TOAST_DURATION_MS = 3500;
const EXIT_ANIMATION_MS = 300;

let nextId = 0;

export function useToast() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(
    new Map(),
  );

  const removeToast = useCallback((id: string) => {
    const timer = timersRef.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timersRef.current.delete(id);
    }
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback(
    (message: string, type: ToastType = "info") => {
      const id = `toast-${++nextId}`;
      const item: ToastItem = { id, message, type };
      setToasts((prev) => [...prev, item]);

      // Schedule removal after duration + exit animation time
      const timer = setTimeout(() => {
        timersRef.current.delete(id);
        removeToast(id);
      }, TOAST_DURATION_MS + EXIT_ANIMATION_MS);

      timersRef.current.set(id, timer);
      return id;
    },
    [removeToast],
  );

  return { toasts, addToast, removeToast };
}
