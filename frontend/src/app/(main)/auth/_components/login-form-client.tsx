"use client";

import dynamic from "next/dynamic";

// Render LoginForm only on the client to avoid useId() hydration mismatch
// from react-hook-form's FormItem generating IDs that differ between SSR and CSR.
export const LoginFormClient = dynamic(
  () => import("./login-form").then((m) => ({ default: m.LoginForm })),
  {
    ssr: false,
    loading: () => <div className="h-48" />,
  },
);
