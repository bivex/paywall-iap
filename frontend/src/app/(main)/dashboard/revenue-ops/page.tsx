import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const dunningQueue = [
  { id: "usr_019", email: "alice@example.com", plan: "Pro Monthly", attempt: 2, nextRetry: "2026-03-03", amount: 9.99 },
  { id: "usr_031", email: "bob@example.com", plan: "Basic Monthly", attempt: 1, nextRetry: "2026-03-02", amount: 4.99 },
  { id: "usr_047", email: "carol@example.com", plan: "Pro Annual", attempt: 3, nextRetry: "2026-03-08", amount: 79.99 },
];

const webhookQueue = [
  { id: "wh_011", provider: "Stripe", type: "invoice.payment_failed", ts: "2026-03-01 13:45", status: "pending" },
  { id: "wh_012", provider: "Apple", type: "DID_FAIL_TO_RENEW", ts: "2026-03-01 13:20", status: "pending" },
  { id: "wh_013", provider: "Google", type: "subscription_on_hold", ts: "2026-03-01 12:55", status: "failed" },
];

const provColors: Record<string, string> = {
  Stripe: "bg-purple-100 text-purple-800",
  Apple: "bg-blue-100 text-blue-800",
  Google: "bg-emerald-100 text-emerald-800",
};

const whStatus: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  failed: "bg-red-100 text-red-800",
};

export default function RevenueOpsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Revenue Ops Center</h1>
      <Tabs defaultValue="dunning">
        <TabsList>
          <TabsTrigger value="dunning">Dunning Queue (43)</TabsTrigger>
          <TabsTrigger value="webhooks">Webhook Inbox (12)</TabsTrigger>
          <TabsTrigger value="matomo">Matomo Overview</TabsTrigger>
        </TabsList>

        {/* DUNNING QUEUE */}
        <TabsContent value="dunning" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Active Dunning Cases</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex gap-2">
                <Input placeholder="Search user..." className="w-52" />
                <Select>
                  <SelectTrigger className="w-36">
                    <SelectValue placeholder="Attempt: All" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All</SelectItem>
                    <SelectItem value="1">Attempt 1</SelectItem>
                    <SelectItem value="2">Attempt 2</SelectItem>
                    <SelectItem value="3">Attempt 3</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Plan</TableHead>
                    <TableHead>Attempt</TableHead>
                    <TableHead>Next Retry</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {dunningQueue.map((d) => (
                    <TableRow key={d.id}>
                      <TableCell className="text-sm">{d.email}</TableCell>
                      <TableCell>{d.plan}</TableCell>
                      <TableCell>
                        <Badge variant="outline">#{d.attempt}</Badge>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">{d.nextRetry}</TableCell>
                      <TableCell className="font-mono">${d.amount.toFixed(2)}</TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button variant="outline" size="sm">Retry Now</Button>
                          <Button variant="outline" size="sm">Skip</Button>
                          <Button variant="ghost" size="sm">View User</Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <p className="text-xs text-muted-foreground">
                Showing 3 of 43 —{" "}
                <Button variant="link" size="sm" className="h-auto p-0 text-xs">
                  View All in Dunning Config →
                </Button>
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* WEBHOOK INBOX */}
        <TabsContent value="webhooks" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Unprocessed Webhook Events</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Provider</TableHead>
                    <TableHead>Event</TableHead>
                    <TableHead>Received</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {webhookQueue.map((w) => (
                    <TableRow key={w.id}>
                      <TableCell>
                        <Badge className={provColors[w.provider]}>{w.provider}</Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs">{w.type}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{w.ts}</TableCell>
                      <TableCell>
                        <Badge className={whStatus[w.status]}>{w.status}</Badge>
                      </TableCell>
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
              <p className="text-xs text-muted-foreground">
                Showing 3 of 12 unprocessed —{" "}
                <Button variant="link" size="sm" className="h-auto p-0 text-xs">
                  View All in Webhook Inspector →
                </Button>
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* MATOMO OVERVIEW */}
        <TabsContent value="matomo" className="mt-4">
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
              {[
                { label: "Unique Visitors", value: "18,432" },
                { label: "Page Views", value: "74,112" },
                { label: "Avg Session", value: "3m 24s" },
                { label: "Bounce Rate", value: "34.2%" },
              ].map((k) => (
                <Card key={k.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-xs font-medium text-muted-foreground uppercase">
                      {k.label}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-xl font-bold">{k.value}</div>
                  </CardContent>
                </Card>
              ))}
            </div>
            <Card>
              <CardContent className="pt-4 space-y-2 text-sm text-muted-foreground">
                <p>
                  Top page:{" "}
                  <span className="font-mono text-foreground">/pricing</span> — 12,340 views
                </p>
                <p>Top source: Direct (38%)</p>
                <Button variant="outline" size="sm" className="mt-2">
                  Open Full Matomo Dashboard →
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
