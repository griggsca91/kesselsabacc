import { useCallback, useEffect, useRef, useState } from "react";

const MUTE_KEY = "sabacc_sound_muted";

export interface SoundEngine {
  playCardShuffle: () => void;
  playCardFlip: () => void;
  playCardDraw: () => void;
  playChipClink: () => void;
  playWinFanfare: () => void;
  playLoseStinger: () => void;
  playButtonClick: () => void;
  isMuted: boolean;
  toggleMute: () => void;
}

export function useSoundEngine(): SoundEngine {
  const ctxRef = useRef<AudioContext | null>(null);
  const [isMuted, setIsMuted] = useState(
    () => localStorage.getItem(MUTE_KEY) === "true",
  );

  useEffect(() => {
    localStorage.setItem(MUTE_KEY, String(isMuted));
  }, [isMuted]);

  const toggleMute = useCallback(() => setIsMuted((m) => !m), []);

  const ensureCtx = useCallback((): AudioContext | null => {
    if (isMuted) return null;
    if (!ctxRef.current) {
      ctxRef.current = new AudioContext();
    }
    const ctx = ctxRef.current;
    if (ctx.state === "suspended") {
      ctx.resume();
    }
    return ctx;
  }, [isMuted]);

  const noiseBurst = useCallback(
    (ctx: AudioContext, startTime: number, duration: number, freq: number, volume = 0.15) => {
      const bufferSize = Math.floor(ctx.sampleRate * duration);
      const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
      const data = buffer.getChannelData(0);
      for (let i = 0; i < bufferSize; i++) {
        data[i] = Math.random() * 2 - 1;
      }
      const src = ctx.createBufferSource();
      src.buffer = buffer;
      const filter = ctx.createBiquadFilter();
      filter.type = "bandpass";
      filter.frequency.value = freq;
      filter.Q.value = 1.5;
      const gain = ctx.createGain();
      gain.gain.setValueAtTime(volume, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration);
      src.connect(filter).connect(gain).connect(ctx.destination);
      src.start(startTime);
      src.stop(startTime + duration);
    },
    [],
  );

  const sinePing = useCallback(
    (
      ctx: AudioContext,
      startTime: number,
      freq: number,
      duration: number,
      volume = 0.15,
      harmonics: number[] = [],
    ) => {
      const playTone = (f: number, v: number) => {
        const osc = ctx.createOscillator();
        osc.type = "sine";
        osc.frequency.value = f;
        const gain = ctx.createGain();
        gain.gain.setValueAtTime(v, startTime);
        gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration);
        osc.connect(gain).connect(ctx.destination);
        osc.start(startTime);
        osc.stop(startTime + duration);
      };
      playTone(freq, volume);
      for (const h of harmonics) {
        playTone(freq * h, volume * 0.3);
      }
    },
    [],
  );

  const playCardShuffle = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    for (let i = 0; i < 6; i++) {
      noiseBurst(ctx, now + i * 0.045, 0.035, 3000 + i * 400, 0.08);
    }
  }, [ensureCtx, noiseBurst]);

  const playCardFlip = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    const bufSize = Math.floor(ctx.sampleRate * 0.12);
    const buf = ctx.createBuffer(1, bufSize, ctx.sampleRate);
    const d = buf.getChannelData(0);
    for (let i = 0; i < bufSize; i++) d[i] = Math.random() * 2 - 1;
    const src = ctx.createBufferSource();
    src.buffer = buf;
    const filter = ctx.createBiquadFilter();
    filter.type = "highpass";
    filter.frequency.setValueAtTime(2000, now);
    filter.frequency.exponentialRampToValueAtTime(6000, now + 0.12);
    const gain = ctx.createGain();
    gain.gain.setValueAtTime(0.12, now);
    gain.gain.exponentialRampToValueAtTime(0.001, now + 0.12);
    src.connect(filter).connect(gain).connect(ctx.destination);
    src.start(now);
    src.stop(now + 0.12);
  }, [ensureCtx]);

  const playCardDraw = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    const bufSize = Math.floor(ctx.sampleRate * 0.15);
    const buf = ctx.createBuffer(1, bufSize, ctx.sampleRate);
    const d = buf.getChannelData(0);
    for (let i = 0; i < bufSize; i++) d[i] = Math.random() * 2 - 1;
    const src = ctx.createBufferSource();
    src.buffer = buf;
    const filter = ctx.createBiquadFilter();
    filter.type = "bandpass";
    filter.frequency.setValueAtTime(1200, now);
    filter.frequency.exponentialRampToValueAtTime(3500, now + 0.15);
    filter.Q.value = 2;
    const gain = ctx.createGain();
    gain.gain.setValueAtTime(0.1, now);
    gain.gain.exponentialRampToValueAtTime(0.001, now + 0.15);
    src.connect(filter).connect(gain).connect(ctx.destination);
    src.start(now);
    src.stop(now + 0.15);
  }, [ensureCtx]);

  const playChipClink = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    sinePing(ctx, now, 2800, 0.15, 0.1, [2.4, 3.1]);
    sinePing(ctx, now + 0.06, 3400, 0.1, 0.06, [1.8]);
  }, [ensureCtx, sinePing]);

  const playWinFanfare = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    const notes = [523.25, 659.25, 783.99, 1046.5];
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      osc.type = "triangle";
      osc.frequency.value = freq;
      const gain = ctx.createGain();
      const t = now + i * 0.12;
      gain.gain.setValueAtTime(0, t);
      gain.gain.linearRampToValueAtTime(0.12, t + 0.03);
      gain.gain.exponentialRampToValueAtTime(0.001, t + 0.3);
      osc.connect(gain).connect(ctx.destination);
      osc.start(t);
      osc.stop(t + 0.3);
    });
  }, [ensureCtx]);

  const playLoseStinger = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    const notes = [311.13, 233.08];
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      osc.type = "triangle";
      osc.frequency.value = freq;
      const gain = ctx.createGain();
      const t = now + i * 0.2;
      gain.gain.setValueAtTime(0, t);
      gain.gain.linearRampToValueAtTime(0.1, t + 0.03);
      gain.gain.exponentialRampToValueAtTime(0.001, t + 0.4);
      osc.connect(gain).connect(ctx.destination);
      osc.start(t);
      osc.stop(t + 0.4);
    });
  }, [ensureCtx]);

  const playButtonClick = useCallback(() => {
    const ctx = ensureCtx();
    if (!ctx) return;
    const now = ctx.currentTime;
    noiseBurst(ctx, now, 0.025, 4000, 0.05);
  }, [ensureCtx, noiseBurst]);

  return {
    playCardShuffle,
    playCardFlip,
    playCardDraw,
    playChipClink,
    playWinFanfare,
    playLoseStinger,
    playButtonClick,
    isMuted,
    toggleMute,
  };
}
