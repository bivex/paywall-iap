"use client";

import {
  Activity,
  AlertTriangle,
  BarChart2,
  Brain,
  ChartNoAxesCombined,
  CreditCard,
  DollarSign,
  FlaskConical,
  Gift,
  Layers,
  LayoutDashboard,
  ScrollText,
  Settings,
  Settings2,
  Tag,
  Target,
  TrendingUp,
  Users,
  Webhook,
} from "lucide-react";
import { useTranslations } from "next-intl";

import { isComingSoonUrl, type NavGroup } from "./sidebar-items";

export function useSidebarItems(): NavGroup[] {
  const g = useTranslations("nav.groups");
  const i = useTranslations("nav.items");

  return [
    {
      id: 1,
      label: g("adminDashboard"),
      items: [
        { title: i("dashboard"), url: "/dashboard/default", icon: LayoutDashboard },
        { title: i("analyticsReports"), url: "/dashboard/analytics", icon: TrendingUp },
        {
          title: i("matomoAnalytics"),
          url: "/dashboard/matomo",
          icon: ChartNoAxesCombined,
          comingSoon: isComingSoonUrl("/dashboard/matomo"),
        },
        { title: i("platformSettings"), url: "/dashboard/settings", icon: Settings },
      ],
    },
    {
      id: 2,
      label: g("user360"),
      items: [
        { title: i("userList"), url: "/dashboard/users", icon: Users },
        { title: i("auditLog"), url: "/dashboard/audit-log", icon: ScrollText },
      ],
    },
    {
      id: 3,
      label: g("revenueOps"),
      items: [
        { title: i("overview"), url: "/dashboard/revenue-ops", icon: Layers },
        { title: i("subscriptions"), url: "/dashboard/subscriptions", icon: CreditCard },
        { title: i("transactions"), url: "/dashboard/transactions", icon: DollarSign },
        { title: i("dunning"), url: "/dashboard/dunning", icon: AlertTriangle },
        {
          title: i("winback"),
          url: "/dashboard/winback",
          icon: Gift,
          comingSoon: isComingSoonUrl("/dashboard/winback"),
        },
        { title: i("pricingTiers"), url: "/dashboard/pricing", icon: Tag },
        { title: i("webhooks"), url: "/dashboard/webhooks", icon: Webhook },
      ],
    },
    {
      id: 4,
      label: g("experimentStudio"),
      items: [
        {
          title: i("abTests"),
          url: "/dashboard/experiments",
          icon: FlaskConical,
          comingSoon: isComingSoonUrl("/dashboard/experiments"),
        },
        {
          title: i("studio"),
          url: "/dashboard/experiments/studio",
          icon: Settings2,
          comingSoon: isComingSoonUrl("/dashboard/experiments/studio"),
        },
        {
          title: i("banditModel"),
          url: "/dashboard/experiments/bandit",
          icon: Brain,
          comingSoon: isComingSoonUrl("/dashboard/experiments/bandit"),
        },
        {
          title: i("delayedFeedback"),
          url: "/dashboard/experiments/feedback",
          icon: Activity,
          comingSoon: isComingSoonUrl("/dashboard/experiments/feedback"),
        },
        {
          title: i("slidingWindow"),
          url: "/dashboard/experiments/sliding-window",
          icon: BarChart2,
          comingSoon: isComingSoonUrl("/dashboard/experiments/sliding-window"),
        },
        {
          title: i("multiObjective"),
          url: "/dashboard/experiments/multi-objective",
          icon: Target,
          comingSoon: isComingSoonUrl("/dashboard/experiments/multi-objective"),
        },
      ],
    },
  ];
}
