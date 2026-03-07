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
  type LucideIcon,
  ScrollText,
  Settings,
  Settings2,
  Tag,
  Target,
  TrendingUp,
  Users,
  Webhook,
} from "lucide-react";

export interface NavSubItem {
  title: string;
  url: string;
  icon?: LucideIcon;
  comingSoon?: boolean;
  newTab?: boolean;
  isNew?: boolean;
}

export interface NavMainItem {
  title: string;
  url: string;
  icon?: LucideIcon;
  subItems?: NavSubItem[];
  comingSoon?: boolean;
  newTab?: boolean;
  isNew?: boolean;
}

export interface NavGroup {
  id: number;
  label?: string;
  items: NavMainItem[];
}

export const comingSoonUrls = new Set([
  "/dashboard/matomo",
  "/dashboard/dunning",
  "/dashboard/winback",
  "/dashboard/pricing",
  "/dashboard/experiments",
  "/dashboard/experiments/studio",
  "/dashboard/experiments/bandit",
  "/dashboard/experiments/feedback",
  "/dashboard/experiments/sliding-window",
  "/dashboard/experiments/multi-objective",
]);

export function isComingSoonUrl(url: string): boolean {
  return comingSoonUrls.has(url);
}

export const sidebarItems: NavGroup[] = [
  {
    id: 1,
    label: "Admin Dashboard",
    items: [
      { title: "Dashboard", url: "/dashboard/default", icon: LayoutDashboard },
      { title: "Analytics Reports", url: "/dashboard/analytics", icon: TrendingUp },
      { title: "Matomo Analytics", url: "/dashboard/matomo", icon: ChartNoAxesCombined, comingSoon: true },
      { title: "Platform Settings", url: "/dashboard/settings", icon: Settings },
    ],
  },
  {
    id: 2,
    label: "User 360° Profile",
    items: [
      { title: "User List", url: "/dashboard/users", icon: Users },
      { title: "Audit Log", url: "/dashboard/audit-log", icon: ScrollText },
    ],
  },
  {
    id: 3,
    label: "Revenue Ops Center",
    items: [
      { title: "Overview", url: "/dashboard/revenue-ops", icon: Layers },
      { title: "Subscriptions", url: "/dashboard/subscriptions", icon: CreditCard },
      { title: "Transactions", url: "/dashboard/transactions", icon: DollarSign },
      { title: "Dunning", url: "/dashboard/dunning", icon: AlertTriangle, comingSoon: true },
      { title: "Winback", url: "/dashboard/winback", icon: Gift, comingSoon: true },
      { title: "Pricing Tiers", url: "/dashboard/pricing", icon: Tag, comingSoon: true },
      { title: "Webhooks", url: "/dashboard/webhooks", icon: Webhook },
    ],
  },
  {
    id: 4,
    label: "Experiment Studio",
    items: [
      { title: "A/B Tests", url: "/dashboard/experiments", icon: FlaskConical, comingSoon: true },
      { title: "Studio", url: "/dashboard/experiments/studio", icon: Settings2, comingSoon: true },
      { title: "Bandit Model", url: "/dashboard/experiments/bandit", icon: Brain, comingSoon: true },
      { title: "Delayed Feedback", url: "/dashboard/experiments/feedback", icon: Activity, comingSoon: true },
      { title: "Sliding Window", url: "/dashboard/experiments/sliding-window", icon: BarChart2, comingSoon: true },
      { title: "Multi-Objective", url: "/dashboard/experiments/multi-objective", icon: Target, comingSoon: true },
    ],
  },
];
