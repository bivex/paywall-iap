import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold">Admin Dashboard</h1>
        <p className="text-sm text-muted-foreground">Last Updated: 2026-03-01 13:40 UTC</p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Active Users</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">14,205</div><p className="text-xs text-green-600 mt-1">▲ +2.1% vs last month</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">MRR (USD)</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">$45,230</div><p className="text-xs text-green-600 mt-1">▲ +8.3% vs last month</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Active Subs</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">12,100</div><p className="text-xs text-green-600 mt-1">▲ +5.2% vs last month</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Churn Risk</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">345</div><p className="text-xs text-orange-600 mt-1">Dunning in progress</p></CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Recent Admin Actions */}
        <Card>
          <CardHeader><CardTitle className="text-sm">Recent Admin Actions</CardTitle><p className="text-xs text-muted-foreground">admin_audit_log</p></CardHeader>
          <CardContent className="space-y-2">
            <div className="text-sm">[13:40] <span className="font-medium">Admin_01</span> updated pricing tier <span className="text-muted-foreground">Pro Annual → $39.99</span></div>
            <Separator />
            <div className="text-sm">[13:35] <span className="font-medium">Admin_02</span> refunded transaction <span className="text-muted-foreground">txn_8821 · $49.99</span></div>
            <Separator />
            <div className="text-sm">[13:30] <span className="font-medium">System</span> auto-retry dunning <span className="text-muted-foreground">sub_3341 · attempt 2/4</span></div>
            <a href="/dashboard/audit-log" className="text-xs text-primary mt-2 block">View Full Audit Log →</a>
          </CardContent>
        </Card>

        {/* Subscription Status Distribution */}
        <Card>
          <CardHeader><CardTitle className="text-sm">Subscription Status Distribution</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {[
              { label: "Active", value: 85, color: "bg-green-500" },
              { label: "Grace Period", value: 5, color: "bg-yellow-500" },
              { label: "Cancelled", value: 7, color: "bg-gray-400" },
              { label: "Expired", value: 3, color: "bg-red-500" },
            ].map((item) => (
              <div key={item.label} className="space-y-1">
                <div className="flex justify-between text-xs"><span>{item.label}</span><span className="text-muted-foreground">{item.value}%</span></div>
                <Progress value={item.value} className="h-2" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Webhook Health */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Webhook Health</CardTitle><p className="text-xs text-muted-foreground">webhook_events</p></CardHeader>
        <CardContent>
          <div className="flex gap-4 text-sm mb-2">
            <span>✅ Stripe</span><span>✅ Apple</span><span>⚠️ Google (2 failed, retrying)</span>
          </div>
          <p className="text-xs text-muted-foreground">0 Unprocessed Events</p>
        </CardContent>
      </Card>
    </div>
  );
}
