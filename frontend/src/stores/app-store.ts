import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface App {
  id: string;
  name: string;
  display_name: string;
  bundle_id: string;
  platform: "ios" | "android" | "both";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface AppStoreState {
  apps: App[];
  selectedAppId: string | null;
  setApps: (apps: App[]) => void;
  setSelectedAppId: (id: string | null) => void;
}

export const useAppStore = create<AppStoreState>()(
  persist(
    (set) => ({
      apps: [],
      selectedAppId: null,
      setApps: (apps) => set({ apps }),
      setSelectedAppId: (id) => set({ selectedAppId: id }),
    }),
    {
      name: "paywall-selected-app",
      partialize: (state) => ({ selectedAppId: state.selectedAppId }),
    },
  ),
);

export function getSelectedApp(state: AppStoreState): App | null {
  return state.apps.find((a) => a.id === state.selectedAppId) ?? null;
}
