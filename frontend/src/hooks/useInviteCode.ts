export function useInviteCode(): string | null {
  const match = window.location.pathname.match(/^\/join\/([A-Z0-9]{4})$/i);
  return match ? match[1].toUpperCase() : null;
}
