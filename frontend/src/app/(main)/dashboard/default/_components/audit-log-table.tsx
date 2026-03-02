import {
  ShieldCheck,
  RotateCcw,
  XCircle,
  DollarSign,
  RefreshCw,
  Settings,
  Activity,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { AuditLogEntry } from "@/actions/dashboard";

type ActionMeta = {
  label: string;
  icon: React.ReactNode;
  badge: string; // tailwind class string for Badge override
};

function getActionMeta(action: string): ActionMeta {
  const a = action.toLowerCase();
  if (a.includes("grant"))
    return {
      label: "Grant",
      icon: <ShieldCheck className="h-3.5 w-3.5" />,
      badge: "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300",
    };
  if (a.includes("revoke"))
    return {
      label: "Revoke",
      icon: <XCircle className="h-3.5 w-3.5" />,
      badge: "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300",
    };
  if (a.includes("refund"))
    return {
      label: "Refund",
      icon: <DollarSign className="h-3.5 w-3.5" />,
      badge: "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300",
    };
  if (a.includes("retry") || a.includes("dunning"))
    return {
      label: "Dunning",
      icon: <RefreshCw className="h-3.5 w-3.5" />,
      badge: "bg-orange-100 text-orange-800 dark:bg-orange-900/40 dark:text-orange-300",
    };
  if (a.includes("pricing") || a.includes("updated"))
    return {
      label: "Pricing",
      icon: <Settings className="h-3.5 w-3.5" />,
      badge: "bg-violet-100 text-violet-800 dark:bg-violet-900/40 dark:text-violet-300",
    };
  return {
    label: action.replace(/_/g, " "),
    icon: <Activity className="h-3.5 w-3.5" />,
    badge: "bg-muted text-muted-foreground",
  };
}

function formatTime(iso: string) {
  const d = new Date(iso);
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export function AuditLogTable({ entries }: { entries: AuditLogEntry[] }) {
  if (entries.length === 0) {
    return (
      <p className="py-6 text-center text-sm text-muted-foreground">
        No recent actions.
      </p>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[130px] text-xs">Time</TableHead>
          <TableHead className="w-[110px] text-xs">Action</TableHead>
          <TableHead className="text-xs">Details</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {entries.map((entry, i) => {
          const meta = getActionMeta(entry.Action);
          return (
            <TableRow key={i} className="hover:bg-muted/40">
              <TableCell className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                {formatTime(entry.Time)}
              </TableCell>
              <TableCell>
                <Badge
                  variant="outline"
                  className={`gap-1 border-0 text-[11px] font-medium px-2 py-0.5 ${meta.badge}`}
                >
                  {meta.icon}
                  {meta.label}
                </Badge>
              </TableCell>
              <TableCell className="text-xs text-muted-foreground max-w-[280px] truncate">
                {entry.Detail || "—"}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
