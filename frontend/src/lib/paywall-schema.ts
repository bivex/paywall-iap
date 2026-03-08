import { z } from "zod";

const hexColor = z
  .string()
  .trim()
  .regex(/^#(?:[0-9A-Fa-f]{3}){1,2}$/, "Use a valid hex color");

const paywallPlanSchema = z.object({
  id: z.string().trim().min(1),
  title: z.string().trim().min(1),
  price: z.string().trim().min(1),
  period: z.string().trim().min(1),
  caption: z.string().trim().optional().default(""),
  badge: z.string().trim().optional().default(""),
  highlight: z.boolean().optional().default(false),
});

export const paywallSchema = z
  .object({
    id: z.string().trim().min(1),
    name: z.string().trim().min(1),
    platform: z.enum(["universal", "web", "mobile"]).default("universal"),
    layout: z.enum(["centered", "split", "stacked"]).default("split"),
    theme: z
      .object({
        mode: z.enum(["light", "dark"]).default("dark"),
        accentColor: hexColor.default("#7C3AED"),
        backgroundColor: hexColor.default("#0F172A"),
        surfaceColor: hexColor.default("#111827"),
        textColor: hexColor.default("#F8FAFC"),
      })
      .default({
        mode: "dark",
        accentColor: "#7C3AED",
        backgroundColor: "#0F172A",
        surfaceColor: "#111827",
        textColor: "#F8FAFC",
      }),
    hero: z.object({
      badge: z.string().trim().optional().default(""),
      title: z.string().trim().min(1),
      subtitle: z.string().trim().min(1),
      socialProof: z.string().trim().optional().default(""),
    }),
    features: z.array(z.string().trim().min(1)).min(2).max(8),
    plans: z.array(paywallPlanSchema).min(1).max(3),
    cta: z.object({
      primaryLabel: z.string().trim().min(1),
      secondaryLabel: z.string().trim().optional().default(""),
    }),
    footer: z.object({
      restoreLabel: z.string().trim().min(1),
      legalText: z.string().trim().min(1),
    }),
  })
  .superRefine((value, ctx) => {
    const highlightedPlans = value.plans.filter((plan) => plan.highlight).length;
    if (highlightedPlans > 1) {
      ctx.addIssue({
        code: "custom",
        message: "Use only one highlighted plan",
        path: ["plans"],
      });
    }
  });

export type PaywallDefinition = z.infer<typeof paywallSchema>;

export const PAYWALL_TEMPLATES: Record<string, PaywallDefinition> = {
  mobileStreaming: {
    id: "mobile-premium",
    name: "Mobile Premium Upsell",
    platform: "mobile",
    layout: "stacked",
    theme: {
      mode: "dark",
      accentColor: "#F59E0B",
      backgroundColor: "#020617",
      surfaceColor: "#111827",
      textColor: "#F8FAFC",
    },
    hero: {
      badge: "7-day free trial",
      title: "Unlock Premium listening",
      subtitle: "Offline mode, no ads, exclusive episodes and better audio quality in one tap.",
      socialProof: "Trusted by 120k+ paying listeners",
    },
    features: ["Ad-free experience", "Offline downloads", "Exclusive episodes", "Priority support"],
    plans: [
      {
        id: "monthly",
        title: "Monthly",
        price: "$9.99",
        period: "/month",
        caption: "Flexible billing",
        badge: "",
        highlight: false,
      },
      {
        id: "annual",
        title: "Annual",
        price: "$79.99",
        period: "/year",
        caption: "33% savings compared to monthly",
        badge: "Best value",
        highlight: true,
      },
    ],
    cta: { primaryLabel: "Start free trial", secondaryLabel: "Maybe later" },
    footer: {
      restoreLabel: "Restore purchases",
      legalText: "Subscription auto-renews until canceled. Cancel anytime in Settings.",
    },
  },
  webSaas: {
    id: "web-growth",
    name: "Growth SaaS Paywall",
    platform: "web",
    layout: "split",
    theme: {
      mode: "light",
      accentColor: "#2563EB",
      backgroundColor: "#F8FAFC",
      surfaceColor: "#FFFFFF",
      textColor: "#0F172A",
    },
    hero: {
      badge: "For teams shipping faster",
      title: "Upgrade to Growth",
      subtitle: "Collaborate with your whole team, unlock premium analytics and remove trial limits.",
      socialProof: "4.9/5 average rating from 2,400+ teams",
    },
    features: ["Unlimited projects", "Advanced dashboards", "Role-based access", "Priority onboarding"],
    plans: [
      {
        id: "starter",
        title: "Starter",
        price: "$19",
        period: "/mo",
        caption: "For solo makers",
        badge: "",
        highlight: false,
      },
      {
        id: "growth",
        title: "Growth",
        price: "$49",
        period: "/mo",
        caption: "For scaling teams",
        badge: "Most popular",
        highlight: true,
      },
      {
        id: "scale",
        title: "Scale",
        price: "$99",
        period: "/mo",
        caption: "For larger orgs",
        badge: "",
        highlight: false,
      },
    ],
    cta: { primaryLabel: "Upgrade workspace", secondaryLabel: "Talk to sales" },
    footer: {
      restoreLabel: "Already subscribed? Restore access",
      legalText: "Prices exclude tax where applicable. You can change or cancel the plan anytime.",
    },
  },
};

export const DEFAULT_PAYWALL_TEMPLATE = PAYWALL_TEMPLATES.mobileStreaming;

export function stringifyPaywallDefinition(value: PaywallDefinition) {
  return JSON.stringify(value, null, 2);
}

export function parsePaywallDefinition(
  raw: string,
): { success: true; data: PaywallDefinition } | { success: false; errors: string[] } {
  try {
    const json = JSON.parse(raw);
    const parsed = paywallSchema.safeParse(json);
    if (!parsed.success) {
      return {
        success: false,
        errors: parsed.error.issues.map((issue) => {
          const path = issue.path.length > 0 ? `${issue.path.join(".")}: ` : "";
          return `${path}${issue.message}`;
        }),
      };
    }
    return { success: true, data: parsed.data };
  } catch (error) {
    return {
      success: false,
      errors: [error instanceof Error ? error.message : "Invalid JSON input"],
    };
  }
}
