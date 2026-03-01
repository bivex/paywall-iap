import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const logs = [
  { id: "al_001", admin: "admin@paywall.local", action: "grant_refund", target: "user:usr_001", ip: "192.168.1.5", ts: "2026-03-01 14:02" },
  { id: "al_002", admin: "admin@paywall.local", action: "cancel_subscription", target: "sub:sub_019", ip: "192.168.1.5", ts: "2026-03-01 13:55" },
  { id: "al_003", admin: "admin@paywall.local", action: "update_plan_price", target: "plan:pro_monthly", ip: "192.168.1.5", ts: "2026-03-01 13:40" },
  { id: "al_004", admin: "ops@paywall.local", action: "trigger_winback", target: "campaign:wc_001", ip: "10.0.0.12", ts: "2026-03-01 13:12" },
  { id: "al_005", admin: "ops@paywall.local", action: "replay_webhook", target: "webhook:wh_003", ip: "10.0.0.12", ts: "2026-03-01 12:58" },
  { id: "al_006", admin: "admin@paywall.local", action: "manual_renewal", target: "sub:sub_022", ip: "192.168.1.5", ts: "2026-03-01 12:30" },
];

const actionColors: Record<string, string> = {
  grant_refund: "bg-blue-100 text-blue-800",
  cancel_subscription: "bg-red-100 text-red-800",
  update_plan_price: "bg-orange-100 text-orange-800",
  trigger_winback: "bg-purple-100 text-purple-800",
  replay_webhook: "bg-yellow-100 text-yellow-800",
  manual_renewal: "bg-green-100 text-green-800",
};

export default function AuditLogPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-semibold">Audit Log Viewer</h1>
        <Button variant="outline" size="sm">Export CSV</Button>
      </div>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Input placeholder="Search admin or target..." className="w-52" />
            <Select><SelectTrigger className="w-48"><SelectValue placeholder="Action: All" /></SelectTrigger><SelectContent><SelectItem value="all">All Actions</SelectItem><SelectItem value="grant_refund">grant_refund</SelectItem><SelectItem value="cancel_subscription">cancel_subscription</SelectItem><SelectItem value="update_plan_price">update_plan_price</SelectItem><SelectItem value="trigger_winback">trigger_winback</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Admin</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Target</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Detail</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{l.ts}</TableCell>
                  <TableCell className="text-sm">{l.admin}</TableCell>
                  <TableCell><Badge className={actionColors[l.action] ?? ""}>{l.action}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{l.target}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{l.ip}</TableCell>
                  <TableCell><Button variant="ghost" size="sm">→</Button></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... →  Showing 1–6 of 842 entries</p>
        </CardContent>
      </Card>
    </div>
  );
}
