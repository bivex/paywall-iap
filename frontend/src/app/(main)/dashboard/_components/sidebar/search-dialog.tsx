"use client";
import * as React from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

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
  LayoutDashboard,
  Layers,
  ScrollText,
  Search,
  Settings,
  Settings2,
  Tag,
  Target,
  TrendingUp,
  Users,
  Webhook,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";

const searchItems = [
  // Monitoring
  { group: "Monitoring", icon: LayoutDashboard,       label: "Dashboard",           url: "/dashboard/default" },
  { group: "Monitoring", icon: TrendingUp,             label: "Analytics Reports",   url: "/dashboard/analytics" },
  { group: "Monitoring", icon: ChartNoAxesCombined,   label: "Matomo Analytics",    url: "/dashboard/matomo" },
  // User Management
  { group: "User Management", icon: Users,            label: "User List",           url: "/dashboard/users" },
  { group: "User Management", icon: ScrollText,       label: "Audit Log",           url: "/dashboard/audit-log" },
  // Revenue Ops
  { group: "Revenue Ops", icon: Layers,               label: "Revenue Ops Center",  url: "/dashboard/revenue-ops" },
  { group: "Revenue Ops", icon: CreditCard,           label: "Subscriptions",       url: "/dashboard/subscriptions" },
  { group: "Revenue Ops", icon: DollarSign,           label: "Transactions",        url: "/dashboard/transactions" },
  { group: "Revenue Ops", icon: AlertTriangle,        label: "Dunning",             url: "/dashboard/dunning" },
  { group: "Revenue Ops", icon: Gift,                 label: "Winback",             url: "/dashboard/winback" },
  { group: "Revenue Ops", icon: Tag,                  label: "Pricing Tiers",       url: "/dashboard/pricing" },
  { group: "Revenue Ops", icon: Webhook,              label: "Webhooks",            url: "/dashboard/webhooks" },
  // Experiments
  { group: "Experiments", icon: FlaskConical,         label: "A/B Tests",           url: "/dashboard/experiments" },
  { group: "Experiments", icon: Settings2,            label: "Experiment Studio",   url: "/dashboard/experiments/studio" },
  { group: "Experiments", icon: Brain,                label: "Bandit Model",        url: "/dashboard/experiments/bandit" },
  { group: "Experiments", icon: Activity,             label: "Delayed Feedback",    url: "/dashboard/experiments/feedback" },
  { group: "Experiments", icon: BarChart2,            label: "Sliding Window",      url: "/dashboard/experiments/sliding-window" },
  { group: "Experiments", icon: Target,               label: "Multi-Objective",     url: "/dashboard/experiments/multi-objective" },
  // Config
  { group: "Config", icon: Settings,                  label: "Platform Settings",   url: "/dashboard/settings" },
];

export function SearchDialog() {
  const [open, setOpen] = React.useState(false);
  const router = useRouter();
  const t = useTranslations("search");

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "j" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const handleSelect = (url: string) => {
    setOpen(false);
    router.push(url);
  };

  return (
    <>
      <Button
        variant="link"
        className="!px-0 font-normal text-muted-foreground hover:no-underline"
        onClick={() => setOpen(true)}
      >
        <Search className="size-4" />
        Search
        <kbd className="inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-medium text-[10px]">
          <span className="text-xs">⌘</span>J
        </kbd>
      </Button>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder={t("placeholder")} />
        <CommandList>
          <CommandEmpty>{t("noResults")}</CommandEmpty>
          {[...new Set(searchItems.map((item) => item.group))].map((group, i) => (
            <React.Fragment key={group}>
              {i !== 0 && <CommandSeparator />}
              <CommandGroup heading={group}>
                {searchItems
                  .filter((item) => item.group === group)
                  .map((item) => (
                    <CommandItem
                      className="!py-1.5"
                      key={item.label}
                      onSelect={() => handleSelect(item.url)}
                    >
                      {item.icon && <item.icon className="size-4" />}
                      <span>{item.label}</span>
                    </CommandItem>
                  ))}
              </CommandGroup>
            </React.Fragment>
          ))}
        </CommandList>
      </CommandDialog>
    </>
  );
}
