"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface PlatformSettings {
  general: {
    platform_name: string;
    support_email: string;
    default_currency: string;
    dark_mode_default: boolean;
  };
  integrations: {
    stripe_api_key: string;
    stripe_webhook_secret: string;
    stripe_test_mode: boolean;
    apple_issuer_id: string;
    apple_bundle_id: string;
    google_service_account: string;
    google_package_name: string;
    matomo_url: string;
    matomo_site_id: string;
    matomo_auth_token: string;
  };
  notifications: {
    new_subscription: boolean;
    payment_failed: boolean;
    subscription_cancelled: boolean;
    refund_issued: boolean;
    webhook_failed: boolean;
    dunning_started: boolean;
  };
  security: {
    jwt_expiry_hours: number;
    require_mfa: boolean;
    enable_ip_allowlist: boolean;
  };
}

export const DEFAULT_PLATFORM_SETTINGS: PlatformSettings = {
  general: {
    platform_name: "Paywall SaaS",
    support_email: "support@paywall.local",
    default_currency: "USD",
    dark_mode_default: false,
  },
  integrations: {
    stripe_api_key: "",
    stripe_webhook_secret: "",
    stripe_test_mode: false,
    apple_issuer_id: "",
    apple_bundle_id: "",
    google_service_account: "",
    google_package_name: "",
    matomo_url: "",
    matomo_site_id: "",
    matomo_auth_token: "",
  },
  notifications: {
    new_subscription: true,
    payment_failed: true,
    subscription_cancelled: true,
    refund_issued: true,
    webhook_failed: true,
    dunning_started: true,
  },
  security: {
    jwt_expiry_hours: 24,
    require_mfa: false,
    enable_ip_allowlist: false,
  },
};

function cloneDefaults(): PlatformSettings {
  return JSON.parse(JSON.stringify(DEFAULT_PLATFORM_SETTINGS)) as PlatformSettings;
}

async function getAdminToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get("admin_access_token")?.value;
}

async function parseResponse<T>(res: Response): Promise<{ ok: boolean; data?: T; error?: string }> {
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    return {
      ok: false,
      error:
        (body as { message?: string; error?: string }).message ??
        (body as { error?: string }).error ??
        `HTTP ${res.status}`,
    };
  }
  return { ok: true, data: ((body as { data?: T }).data ?? body) as T };
}

export async function getPlatformSettings(): Promise<PlatformSettings> {
  const token = await getAdminToken();
  if (!token) return cloneDefaults();

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<PlatformSettings>(res);
    return parsed.ok && parsed.data ? parsed.data : cloneDefaults();
  } catch {
    return cloneDefaults();
  }
}

export async function updatePlatformSettings(payload: PlatformSettings) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings`, {
      method: "PUT",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const parsed = await parseResponse<PlatformSettings>(res);
    if (parsed.ok) revalidatePath("/dashboard/settings");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) };
  }
}

export async function changeAdminPasswordAction(input: {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings/password`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        current_password: input.currentPassword,
        new_password: input.newPassword,
        confirm_password: input.confirmPassword,
      }),
    });
    return await parseResponse<{ ok: boolean }>(res);
  } catch (error) {
    return { ok: false, error: String(error) };
  }
}
