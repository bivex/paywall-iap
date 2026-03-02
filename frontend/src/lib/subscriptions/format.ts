/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 06:28
 * Last Updated: 2026-03-02 06:28
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

/** Map source+platform to a human-readable label. */
export function formatSource(source: string, platform: string): string {
  if (source === "iap") {
    if (platform === "ios") return "Apple IAP";
    if (platform === "android") return "Google Play";
    return "IAP";
  }
  if (source === "stripe") return "Stripe";
  if (source === "paddle") return "Paddle";
  return source;
}

/** Map plan_type to a human-readable label. */
export function formatPlanType(planType: string): string {
  if (planType === "monthly") return "Monthly";
  if (planType === "annual") return "Annual";
  if (planType === "lifetime") return "Lifetime";
  return planType;
}
