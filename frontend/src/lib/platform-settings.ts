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
