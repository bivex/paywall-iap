import { useAppStore } from "@/stores/app-store";

/**
 * Appends ?app_id=<selectedAppId> to a path if an app is selected.
 * Uses Zustand getState() so it works outside React components too.
 */
export function buildUrl(path: string): string {
  const { selectedAppId } = useAppStore.getState();
  if (!selectedAppId) return path;
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}app_id=${encodeURIComponent(selectedAppId)}`;
}

/**
 * Drop-in replacement for fetch() that automatically injects the selected app_id.
 * Use this in all client-side components instead of raw fetch().
 */
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  return fetch(buildUrl(path), init);
}
