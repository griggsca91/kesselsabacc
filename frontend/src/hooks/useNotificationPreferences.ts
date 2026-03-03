import { useState, useCallback, useEffect } from "react";

const STORAGE_KEY = "sabacc-notifications-enabled";

export function useNotificationPreferences() {
  const [notificationsEnabled, setNotificationsEnabled] = useState<boolean>(
    () => {
      try {
        const stored = localStorage.getItem(STORAGE_KEY);
        return stored === null ? true : stored === "true";
      } catch {
        return true;
      }
    },
  );

  // Persist to localStorage whenever it changes
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, String(notificationsEnabled));
    } catch {
      // Ignore storage errors
    }
  }, [notificationsEnabled]);

  // Request browser notification permission when enabled
  useEffect(() => {
    if (
      notificationsEnabled &&
      typeof Notification !== "undefined" &&
      Notification.permission === "default"
    ) {
      Notification.requestPermission();
    }
  }, [notificationsEnabled]);

  const toggleNotifications = useCallback(() => {
    setNotificationsEnabled((prev) => !prev);
  }, []);

  return { notificationsEnabled, toggleNotifications };
}
