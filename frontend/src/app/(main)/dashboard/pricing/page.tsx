import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const plans = [
  { id: "plan_001", name: "Basic Monthly", price: 4.99, currency: "USD", interval: "monthly", active: true, trial: 7, subs: 512 },
  { id: "plan_002", name: "Pro Monthly", price: 9.99, currency: "USD", interval: "monthly", active: true, trial: 7, subs: 934 },
  { id: "plan_003", name: "Pro Annual", price: 79.99, currency: "USD", interval: "annual", active: true, trial: 14, subs: 1204 },
  { id: "plan_004", name: "Enterprise", price: 299.00, currency: "USD", interval: "annual", active: false, trial: 30, subs: 38 },
];

export default function PricingPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Pricing Plan Editor</h1>
        <Button size="sm">+ New Plan</Button>
      </div>

      {/* Plans table */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Plans</CardTitle><p className="text-xs text-muted-foreground">plans table</p></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Price</TableHead>
                <TableHead>Interval</TableHead>
                <TableHead>Trial</TableHead>
                <TableHead>Active Subs</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {plans.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell className="font-mono">${p.price.toFixed(2)} {p.currency}</TableCell>
                  <TableCell><Badge variant="outline">{p.interval}</Badge></TableCell>
                  <TableCell className="text-muted-foreground">{p.trial} days</TableCell>
                  <TableCell className="font-mono">{p.subs.toLocaleString()}</TableCell>
                  <TableCell>
                    {p.active
                      ? <Badge className="bg-green-100 text-green-800">Active</Badge>
                      : <Badge variant="secondary">Inactive</Badge>}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="outline" size="sm">Edit</Button>
                      {p.active
                        ? <Button variant="destructive" size="sm">Deactivate</Button>
                        : <Button size="sm">Activate</Button>}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Edit Plan inline form */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Edit Plan: Pro Monthly</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <div><p className="text-xs font-medium mb-1">Plan Name</p><Input defaultValue="Pro Monthly" /></div>
            <div><p className="text-xs font-medium mb-1">Price</p><Input defaultValue="9.99" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">Currency</p><Select><SelectTrigger><SelectValue defaultValue="USD" placeholder="USD" /></SelectTrigger><SelectContent><SelectItem value="usd">USD</SelectItem><SelectItem value="eur">EUR</SelectItem><SelectItem value="gbp">GBP</SelectItem></SelectContent></Select></div>
            <div><p className="text-xs font-medium mb-1">Interval</p><Select><SelectTrigger><SelectValue placeholder="monthly" /></SelectTrigger><SelectContent><SelectItem value="monthly">Monthly</SelectItem><SelectItem value="annual">Annual</SelectItem><SelectItem value="weekly">Weekly</SelectItem></SelectContent></Select></div>
          </div>
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
            <div><p className="text-xs font-medium mb-1">Trial Days</p><Input defaultValue="7" className="w-24" /></div>
            <div className="flex items-center gap-2 mt-4"><Switch id="grace-enabled" defaultChecked /><label htmlFor="grace-enabled" className="text-sm">Grace Period Enabled</label></div>
            <div><p className="text-xs font-medium mb-1">Grace Duration (days)</p><Input defaultValue="3" className="w-24" /></div>
          </div>
          <div className="flex gap-2">
            <Button size="sm">Save Plan</Button>
            <Button size="sm" variant="outline">Cancel</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
