"use client";

import { useEffect } from "react";
import { installNetworkMonitor } from "@/lib/network-monitor.client";

/** Mounts the dev-only fetch interceptor. Renders nothing. */
export function NetworkMonitor() {
  useEffect(() => {
    installNetworkMonitor();
  }, []);
  return null;
}
