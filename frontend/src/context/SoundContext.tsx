import { createContext, useContext } from "react";
import type { SoundEngine } from "../hooks/useSoundEngine";

export const SoundContext = createContext<SoundEngine | null>(null);

export function useSound(): SoundEngine | null {
  return useContext(SoundContext);
}
