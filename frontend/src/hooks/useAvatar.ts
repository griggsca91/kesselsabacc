import { useState } from "react";

const STORAGE_KEY = "sabacc_avatar";

export function useAvatar() {
  const [avatarId, setAvatarIdState] = useState<string>(() => {
    return localStorage.getItem(STORAGE_KEY) ?? "";
  });

  function setAvatarId(id: string) {
    localStorage.setItem(STORAGE_KEY, id);
    setAvatarIdState(id);
  }

  return { avatarId, setAvatarId };
}
