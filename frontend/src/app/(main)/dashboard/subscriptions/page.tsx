import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const subs = [
  { id: "sub_001", user: "alice@example.com", status: "active", source: "Apple IAP", plan: "Pro Annual", expires: "2027-03-01", ltv: "$184.20" },
  { id: "sub_002", user: "bob@example.com", status: "grace", source: "Google Play", plan: "Basic Monthly", expires: "2026-03-08", ltv: "$92.40" },
  { id: "sub_003", user: "carol@example.com", status: "dunning", source: "Stripe", plan: "Pro Monthly", expires: "2026-02-28", ltv: "$312.60" },
  { id: "sub_004", user: "dave@example.com", status: "expired", source: "Apple IAP", plan: "Basic Monthly", expires: "2026-01-15", ltv: "$47.00" },
  { id: "sub_005", user: "eve@example.com", status: "grace", source: "Stripe", plan: "Enterprise", expires: "2026-03-05", ltv: "$220.00" },
];

const statusMap: Record<string, { label: string; className: string }> = {
  active: { label: "✅ Active", className: "bg-green-100 text-green-800" },
  grace: { label: "🔶 Grace", className: "bg-yellow-100 text-yellow-800" },
  dunning: { label: "⚠️ Dunning", className: "bg-orange-100 text-orange-800" },
  expired: { label: "❌ Expired", className: "bg-red-100 text-red-800" },
};

export default function SubscriptionsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Subscription Management</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Select><SelectTrigger className="w-40"><SelectValue placeholder="Status: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="active">Active</SelectItem><SelectItem value="grace">Grace</SelectItem><SelectItem value="dunning">Dunning</SelectItem><SelectItem value="expired">Expired</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Source: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="apple">Apple IAP</SelectItem><SelectItem value="google">Google Play</SelectItem><SelectItem value="stripe">Stripe</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Plan: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="basic">Basic</SelectItem><SelectItem value="pro">Pro</SelectItem><SelectItem value="enterprise">Enterprise</SelectItem></SelectContent></Select>
            <Input type="date" placeholder="Expires from" className="w-40" />
            <Input type="date" placeholder="Expires to" className="w-40" />
          </div>
          <div className="flex gap-2">
            <Button variant="destructive" size="sm">Bulk: Cancel</Button>
            <Button variant="outline" size="sm">Bulk: Renew</Button>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>User</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Plan</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>LTV</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {subs.map((s) => (
                <TableRow key={s.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell className="font-medium">{s.user}</TableCell>
                  <TableCell><Badge className={statusMap[s.status].className}>{statusMap[s.status].label}</Badge></TableCell>
                  <TableCell>{s.source}</TableCell>
                  <TableCell>{s.plan}</TableCell>
                  <TableCell>{s.expires}</TableCell>
                  <TableCell>{s.ltv}</TableCell>
                  <TableCell className="text-primary text-sm cursor-pointer">⋯</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 89 →  &nbsp; Showing 1–5 of 445</p>
        </CardContent>
      </Card>
    </div>
  );
}
