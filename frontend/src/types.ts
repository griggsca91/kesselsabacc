export type ConnectionStatus = "connected" | "reconnecting" | "disconnected";

export type CardSuit = "sand" | "blood";
export type CardKind = "value" | "impostor" | "sylop";
export type ShiftToken =
  | "free_draw"
  | "refund"
  | "general_tariff"
  | "markdown"
  | "immunity"
  | "major_fraud"
  | "cook_the_books"
  | "direct_transaction"
  | "prime_sabacc";

export type Phase =
  | "lobby"
  | "dealing"
  | "turn"
  | "reveal"
  | "round_end"
  | "game_over";

export type HandRank = 0 | 1 | 2; // pure_sabacc, sabacc, no_sabacc

export interface Card {
  suit: CardSuit;
  kind: CardKind;
  value: number;
  id: number;
}

export interface HandResult {
  rank: HandRank;
  value: number;
  sandCard: Card;
  bloodCard: Card;
}

export interface PlayerView {
  id: string;
  name: string;
  chips: number;
  invested: number;
  isHost: boolean;
  eliminated: boolean;
  tokensLeft: number;
  stood: boolean;
  sandCard?: Card;
  bloodCard?: Card;
}

export interface HandView {
  sandCard: Card;
  bloodCard: Card;
  tokens: ShiftToken[];
}

export interface RoundResult {
  winnerIds: string[];
  playerHands: Record<string, HandResult>;
  chipChanges: Record<string, number>;
}

export interface GameState {
  phase: Phase;
  round: number;
  turnInRound: number;
  currentTurnPlayerId: string;
  players: PlayerView[];
  yourHand: HandView | null;
  lastResult: RoundResult | null;
  winnerId: string;
  sandRemaining: number;
  bloodRemaining: number;
}

export interface ServerEnvelope {
  type: "game_state" | "error";
  payload: GameState | { message: string };
}

export interface GameHistoryEntry {
  gameId: string;
  roomCode: string;
  finishedAt: string;
  rounds: number;
  players: GamePlayerSummary[];
  winnerName: string;
}

export interface GamePlayerSummary {
  userId: string;
  displayName: string;
  finalChips: number;
  isWinner: boolean;
}

export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
}

export interface ProfileUser {
  id: string;
  displayName: string;
  email?: string;
  memberSince: string;
}

export interface PlayerStats {
  gamesPlayed: number;
  wins: number;
  losses: number;
  winRate: number;
  bestHand: string | null;
}

export interface PlayerProfile {
  user: ProfileUser;
  stats: PlayerStats;
  games: GameHistoryEntry[];
}
