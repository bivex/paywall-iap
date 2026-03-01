import { getTranslations } from "next-intl/server";
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

const whStatusClass: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  failed: "bg-red-100 text-red-800",
};

export default async function RevenueOpsPage() {
  const t = await getTranslations("revenueOps");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Tabs defaultValue="dunning">
        <TabsList>
          <TabsTrigger value="dunning">{t("tabs.dunning")}</TabsTrigger>
          <TabsTrigger value="webhooks">{t("tabs.webhooks")}</TabsTrigger>
          <TabsTrigger value="matomo">{t("tabs.matomo")}</TabsTrigger>
        </TabsList>

        {/* DUNNING QUEUE */}
        <TabsContent value="dunning" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("dunning.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex gap-2">
                <Input placeholder={t("dunning.searchPlaceholder")} className="w-52" />
                <Select>
                  <SelectTrigger className="w-36">
                    <SelectValue placeholder={t("dunning.attemptPlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("dunning.all")}</SelectItem>
                    <SelectItem value="1">{t("dunning.attempt1")}</SelectItem>
                    <SelectItem value="2">{t("dunning.attempt2")}</SelectItem>
                    <SelectItem value="3">{t("dunning.attempt3")}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("dunning.table.user")}</TableHead>
                    <TableHead>{t("dunning.table.plan")}</TableHead>
                    <TableHead>{t("dunning.table.attempt")}</TableHead>
                    <TableHead>{t("dunning.table.nextRetry")}</TableHead>
                    <TableHead>{t("dunning.table.amount")}</TableHead>
                    <TableHead>{t("dunning.table.actions")}</TableHead>
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
                          <Button variant="outline" size="sm">{t("dunning.retryNow")}</Button>
                          <Button variant="outline" size="sm">{t("dunning.skip")}</Button>
                          <Button variant="ghost" size="sm">{t("dunning.viewUser")}</Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <p className="text-xs text-muted-foreground">
                {t("dunning.showing")}{" "}
                <Button variant="link" size="sm" className="h-auto p-0 text-xs">
                  {t("dunning.viewAllLink")}
                </Button>
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* WEBHOOK INBOX */}
        <TabsContent value="webhooks" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("webhooks.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("webhooks.table.provider")}</TableHead>
                    <TableHead>{t("webhooks.table.event")}</TableHead>
                    <TableHead>{t("webhooks.table.received")}</TableHead>
                    <TableHead>{t("webhooks.table.status")}</TableHead>
                    <TableHead>{t("webhooks.table.actions")}</TableHead>
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
                        <Badge className={whStatusClass[w.status]}>{w.status}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button variant="outline" size="sm">{t("webhooks.view")}</Button>
                          <Button variant="outline" size="sm">{t("webhooks.replay")}</Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <p className="text-xs text-muted-foreground">
                {t("webhooks.showing")}{" "}
                <Button variant="link" size="sm" className="h-auto p-0 text-xs">
                  {t("webhooks.viewAllLink")}
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
                { label: t("matomo.kpi.uniqueVisitors"), value: "18,432" },
                { label: t("matomo.kpi.pageViews"), value: "74,112" },
                { label: t("matomo.kpi.avgSession"), value: "3m 24s" },
                { label: t("matomo.kpi.bounceRate"), value: "34.2%" },
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
                  {t("matomo.topPage")}{" "}
                  <span className="font-mono text-foreground">/pricing</span> — 12,340 {t("matomo.views")}
                </p>
                <p>{t("matomo.topSource")}</p>
                <Button variant="outline" size="sm" className="mt-2">
                  {t("matomo.openFull")}
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
