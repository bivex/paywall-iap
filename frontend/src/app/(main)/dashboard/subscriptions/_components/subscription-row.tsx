"use client";

import { SubscriptionDetailSheet } from "./subscription-detail-sheet";
import type { SubscriptionRow } from "@/actions/subscriptions";
import { Badge } from "@/components/ui/badge";
import { TableCell, TableRow } from "@/components/ui/table";
import { formatSource, formatPlanType } from "@/lib/subscriptions/format";

const statusClassMap: Record<string, string> = {
  active:    "bg-green-100 text-green-800",
  grace:     "bg-yellow-100 text-yellow-800",
  cancelled: "bg-orange-100 text-orange-800",
  expired:   "bg-red-100 text-red-800",
};

export function SubscriptionRow({ s, showCreatedAt }: { s: SubscriptionRow; showCreatedAt?: boolean }) {
  return (
    <SubscriptionDetailSheet
      subscriptionId={s.id}
      trigger={
        <TableRow className="cursor-pointer hover:bg-muted/50">
          <TableCell className="font-medium">{s.email || s.user_id}</TableCell>
          <TableCell>
            <Badge className={statusClassMap[s.status] ?? "bg-muted text-muted-foreground"}>
              {s.status}
            </Badge>
          </TableCell>
          <TableCell>{formatSource(s.source, s.platform)}</TableCell>
          <TableCell>{formatPlanType(s.plan_type)}</TableCell>
          <TableCell>
            {new Date(s.expires_at).toLocaleDateString("en-US", {
              year: "numeric", month: "short", day: "numeric",
            })}
          </TableCell>
          <TableCell>${s.ltv.toFixed(2)}</TableCell>
          {showCreatedAt && (
            <TableCell className="text-xs text-muted-foreground whitespace-nowrap" suppressHydrationWarning>
              {new Date(s.created_at).toLocaleString("en-US", {
                month: "short", day: "numeric", year: "numeric",
                hour: "2-digit", minute: "2-digit",
              })}
            </TableCell>
          )}
        </TableRow>
      }
    />
  );
}
