import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const events = [
  { id: "wh_001", provider: "Stripe", type: "invoice.payment_failed", received: "2026-03-01 13:45", status: "pending", preview: '{"subscription":"sub_...","amount":4999}' },
  { id: "wh_002", provider: "Apple", type: "RENEWAL", received: "2026-03-01 13:40", status: "processed", preview: '{"original_transaction_id":"..."}' },
  { id: "wh_003", provider: "Google", type: "subscription_renewed", received: "2026-03-01 13:30", status: "failed", preview: '{"packageName":"com.app","..."}' },
  { id: "wh_004", provider: "Stripe", type: "customer.subscription.deleted", received: "2026-03-01 13:20", status: "processed", preview: '{"id":"sub_...","status":"canceled"}' },
  { id: "wh_005", provider: "Apple", type: "DID_FAIL_TO_RENEW", received: "2026-03-01 13:10", status: "pending", preview: '{"notification_type":"DID_FAIL..."}' },
];

const statusMap: Record<string, { label: string; className: string }> = {
  pending: { label: "⏳ Pending", className: "bg-yellow-100 text-yellow-800" },
  processed: { label: "✅ Processed", className: "bg-green-100 text-green-800" },
  failed: { label: "❌ Failed", className: "bg-red-100 text-red-800" },
};

const providerMap: Record<string, string> = {
  Stripe: "bg-purple-100 text-purple-800",
  Apple: "bg-blue-100 text-blue-800",
  Google: "bg-emerald-100 text-emerald-800",
};

export default function WebhooksPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Webhook Event Inspector</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Provider: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="stripe">Stripe</SelectItem><SelectItem value="apple">Apple</SelectItem><SelectItem value="google">Google</SelectItem></SelectContent></Select>
            <Input placeholder="Event type..." className="w-52" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Status: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="pending">Pending</SelectItem><SelectItem value="processed">Processed</SelectItem><SelectItem value="failed">Failed</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Provider</TableHead>
                <TableHead>Event Type</TableHead>
                <TableHead>Received</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Payload Preview</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((e) => (
                <TableRow key={e.id}>
                  <TableCell><Badge className={providerMap[e.provider] ?? ""}>{e.provider}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{e.type}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{e.received}</TableCell>
                  <TableCell><Badge className={statusMap[e.status].className}>{statusMap[e.status].label}</Badge></TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">{e.preview}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="outline" size="sm">View</Button>
                      <Button variant="outline" size="sm">Replay</Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 28 →  &nbsp; Showing 1–5 of 140 events</p>
        </CardContent>
      </Card>
    </div>
  );
}
