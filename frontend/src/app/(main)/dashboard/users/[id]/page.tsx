import Link from "next/link";
import { notFound } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getUserProfile } from "@/actions/user-profile";
import { UserActionBar } from "./_components/user-action-bar";

const subStatusMeta: Record<string, { label: string; className: string }> = {
  active:    { label: "Active",     className: "bg-emerald-100 text-emerald-800" },
  grace:     { label: "Grace",      className: "bg-yellow-100 text-yellow-800" },
  dunning:   { label: "Dunning",    className: "bg-orange-100 text-orange-800" },
  expired:   { label: "Expired",    className: "bg-red-100 text-red-800" },
  cancelled: { label: "Cancelled",  className: "bg-gray-100 text-gray-600" },
};

const txStatusMeta: Record<string, { label: string; className: string }> = {
  success:  { label: "✅ Success",  className: "bg-emerald-100 text-emerald-800" },
  failed:   { label: "❌ Failed",   className: "bg-red-100 text-red-800" },
  refunded: { label: "↩ Refunded", className: "bg-blue-100 text-blue-800" },
};

const dunStatusMeta: Record<string, { label: string; className: string }> = {
  pending:     { label: "Pending",     className: "bg-yellow-100 text-yellow-800" },
  in_progress: { label: "In Progress", className: "bg-orange-100 text-orange-800" },
  recovered:   { label: "Recovered",   className: "bg-emerald-100 text-emerald-800" },
  failed:      { label: "Failed",      className: "bg-red-100 text-red-800" },
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", { month: "short", day: "numeric", year: "numeric", hour: "2-digit", minute: "2-digit", hour12: false });
}
function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

export default async function UserProfilePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const data = await getUserProfile(id);
  if (!data) notFound();

  const { user, subscriptions, transactions, audit_log, dunning } = data;
  const activeSub = subscriptions.find((s) => s.status === "active") ?? subscriptions[0];
  const totalSpend = transactions.filter((t) => t.status === "success").reduce((sum, t) => sum + t.amount, 0);

  return (
    <div className="flex flex-col gap-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm">
        <Link href="/dashboard/users" className="text-muted-foreground hover:text-foreground">← Users</Link>
        <span className="text-muted-foreground">/</span>
        <span className="font-medium">{user.email}</span>
      </div>

      {/* Identity + Stats */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader><CardTitle className="text-sm">Identity</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-2 gap-y-2 text-sm">
            <span className="text-muted-foreground">Email</span>
            <span className="font-medium">{user.email}</span>
            <span className="text-muted-foreground">Platform ID</span>
            <span className="font-mono text-xs">{user.platform_user_id}</span>
            <span className="text-muted-foreground">Device ID</span>
            <span className="font-mono text-xs">{user.device_id ?? "—"}</span>
            <span className="text-muted-foreground">Platform</span>
            <span><Badge variant="outline">{user.platform}</Badge></span>
            <span className="text-muted-foreground">App Version</span>
            <span>{user.app_version}</span>
            <span className="text-muted-foreground">Role</span>
            <span><Badge variant="outline">{user.role}</Badge></span>
            <span className="text-muted-foreground">Joined</span>
            <span>{fmtDate(user.created_at)}</span>
            <span className="text-muted-foreground">User ID</span>
            <span className="font-mono text-xs text-muted-foreground">{user.id}</span>
          </CardContent>
        </Card>

        <div className="flex flex-col gap-4">
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold">${user.ltv.toFixed(2)}</div>
              <div className="text-xs text-muted-foreground mt-1">Lifetime Value</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold">{transactions.length}</div>
              <div className="text-xs text-muted-foreground mt-1">Transactions · ${totalSpend.toFixed(2)} paid</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              {activeSub ? (
                <>
                  <Badge className={subStatusMeta[activeSub.status]?.className ?? ""}>
                    {subStatusMeta[activeSub.status]?.label ?? activeSub.status}
                  </Badge>
                  <div className="text-xs text-muted-foreground mt-2">{activeSub.product_id}</div>
                  <div className="text-xs text-muted-foreground">Expires {fmtDate(activeSub.expires_at)}</div>
                </>
              ) : (
                <span className="text-sm text-muted-foreground">No subscription</span>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="subscriptions">
        <TabsList>
          <TabsTrigger value="subscriptions">Subscriptions ({subscriptions.length})</TabsTrigger>
          <TabsTrigger value="transactions">Transactions ({transactions.length})</TabsTrigger>
          <TabsTrigger value="dunning">Dunning ({dunning.length})</TabsTrigger>
          <TabsTrigger value="audit">Audit ({audit_log.length})</TabsTrigger>
        </TabsList>

        {/* Subscriptions tab */}
        <TabsContent value="subscriptions" className="mt-4">
          <Card>
            <CardContent className="pt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Product</TableHead>
                    <TableHead>Plan</TableHead>
                    <TableHead>Source</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Auto-Renew</TableHead>
                    <TableHead>Expires</TableHead>
                    <TableHead>Started</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {subscriptions.length === 0 ? (
                    <TableRow><TableCell colSpan={7} className="text-center text-muted-foreground py-8">No subscriptions.</TableCell></TableRow>
                  ) : subscriptions.map((s) => (
                    <TableRow key={s.id}>
                      <TableCell className="font-mono text-xs">{s.product_id}</TableCell>
                      <TableCell><Badge variant="outline">{s.plan_type}</Badge></TableCell>
                      <TableCell className="text-sm">{s.source}</TableCell>
                      <TableCell><Badge className={subStatusMeta[s.status]?.className ?? ""}>{subStatusMeta[s.status]?.label ?? s.status}</Badge></TableCell>
                      <TableCell>{s.auto_renew ? "✅" : "❌"}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{fmtDate(s.expires_at)}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{fmtDate(s.created_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Transactions tab */}
        <TabsContent value="transactions" className="mt-4">
          <Card>
            <CardContent className="pt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Currency</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Provider TX ID</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transactions.length === 0 ? (
                    <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground py-8">No transactions.</TableCell></TableRow>
                  ) : transactions.map((t) => (
                    <TableRow key={t.id}>
                      <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmt(t.date)}</TableCell>
                      <TableCell className="text-right font-mono font-medium">${t.amount.toFixed(2)}</TableCell>
                      <TableCell className="text-sm">{t.currency}</TableCell>
                      <TableCell><Badge className={txStatusMeta[t.status]?.className ?? ""}>{txStatusMeta[t.status]?.label ?? t.status}</Badge></TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{t.provider_tx_id ?? "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Dunning tab */}
        <TabsContent value="dunning" className="mt-4">
          <Card>
            <CardContent className="pt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Attempts</TableHead>
                    <TableHead>Next Attempt</TableHead>
                    <TableHead>Started</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {dunning.length === 0 ? (
                    <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground py-8">No dunning records.</TableCell></TableRow>
                  ) : dunning.map((d, i) => (
                    <TableRow key={i}>
                      <TableCell><Badge className={dunStatusMeta[d.status]?.className ?? ""}>{dunStatusMeta[d.status]?.label ?? d.status}</Badge></TableCell>
                      <TableCell>{d.attempt_count} / {d.max_attempts}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{d.next_attempt_at ? fmt(d.next_attempt_at) : "—"}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{fmtDate(d.created_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Audit tab */}
        <TabsContent value="audit" className="mt-4">
          <Card>
            <CardContent className="pt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Action</TableHead>
                    <TableHead>By Admin</TableHead>
                    <TableHead>Detail</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {audit_log.length === 0 ? (
                    <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground py-8">No admin actions recorded.</TableCell></TableRow>
                  ) : audit_log.map((a, i) => (
                    <TableRow key={i}>
                      <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmt(a.date)}</TableCell>
                      <TableCell><Badge variant="outline">{a.action}</Badge></TableCell>
                      <TableCell className="text-sm">{a.admin_email}</TableCell>
                      <TableCell className="text-xs text-muted-foreground max-w-xs truncate" title={a.detail}>{a.detail || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Action bar */}
      <UserActionBar userId={user.id} hasActiveSub={!!activeSub && activeSub.status === "active"} />
    </div>
  );
}

