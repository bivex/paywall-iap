import {
  Activity,
  AlertTriangle,
  BarChart2,
  Brain,
  CreditCard,
  DollarSign,
  FlaskConical,
  Flag,
  Gift,
  Globe,
  LayoutDashboard,
  Layers,
  type LucideIcon,
  Settings2,
  Tag,
  Target,
  TrendingUp,
  Users,
  ClipboardList,
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

export const sidebarItems: NavGroup[] = [
  {
    id: 1,
    label: "Monitoring",
    items: [
      { title: "Dashboard", url: "/dashboard/default", icon: LayoutDashboard },
      { title: "Analytics Reports", url: "/dashboard/analytics", icon: TrendingUp },
    ],
  },
  {
    id: 2,
    label: "User Management",
    items: [
      { title: "User List", url: "/dashboard/users", icon: Users },
      { title: "Audit Log", url: "/dashboard/audit-log", icon: ClipboardList },
    ],
  },
  {
    id: 3,
    label: "Revenue Ops",
    items: [
      { title: "Subscriptions", url: "/dashboard/subscriptions", icon: CreditCard },
      { title: "Transactions", url: "/dashboard/transactions", icon: DollarSign },
      { title: "Revenue Ops Center", url: "/dashboard/revenue-ops", icon: Layers },
      { title: "Dunning", url: "/dashboard/dunning", icon: AlertTriangle },
      { title: "Winback", url: "/dashboard/winback", icon: Gift },
    ],
  },
  {
    id: 4,
    label: "Experiments",
    items: [
      { title: "A/B Tests", url: "/dashboard/experiments", icon: FlaskConical },
      { title: "Experiment Studio", url: "/dashboard/experiments/studio", icon: Settings2 },
      { title: "Bandit Model", url: "/dashboard/experiments/bandit", icon: Brain },
      { title: "Delayed Feedback", url: "/dashboard/experiments/feedback", icon: Activity },
      { title: "Sliding Window", url: "/dashboard/experiments/sliding-window", icon: BarChart2 },
      { title: "Multi-Objective", url: "/dashboard/experiments/multi-objective", icon: Target },
    ],
  },
  {
    id: 5,
    label: "Config",
    items: [
      { title: "Pricing Tiers", url: "/dashboard/pricing", icon: Tag },
      { title: "Platform Settings", url: "/dashboard/settings", icon: Globe },
      { title: "Webhooks", url: "/dashboard/webhooks", icon: Webhook },
      { title: "Feature Flags", url: "/dashboard/settings", icon: Flag },
    ],
  },
];
