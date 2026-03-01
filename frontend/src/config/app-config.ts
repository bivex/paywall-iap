import packageJson from "../../package.json";

const currentYear = new Date().getFullYear();

export const APP_CONFIG = {
  name: "Paywall Admin",
  version: packageJson.version,
  copyright: `© ${currentYear}, Paywall Admin.`,
  meta: {
    title: "Paywall Admin — Subscription & IAP Management",
    description:
      "Paywall Admin is the internal dashboard for managing subscriptions, in-app purchases, A/B experiments, dunning, and revenue operations.",
  },
};
