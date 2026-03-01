import { getTranslations } from "next-intl/server";
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

const statusClassMap: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  processed: "bg-green-100 text-green-800",
  failed: "bg-red-100 text-red-800",
};

const providerMap: Record<string, string> = {
  Stripe: "bg-purple-100 text-purple-800",
  Apple: "bg-blue-100 text-blue-800",
  Google: "bg-emerald-100 text-emerald-800",
};

export default async function WebhooksPage() {
  const t = await getTranslations("webhooks");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.providerPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.providerAll")}</SelectItem><SelectItem value="stripe">Stripe</SelectItem><SelectItem value="apple">Apple</SelectItem><SelectItem value="google">Google</SelectItem></SelectContent></Select>
            <Input placeholder={t("filter.eventTypePlaceholder")} className="w-52" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.statusPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.statusAll")}</SelectItem><SelectItem value="pending">{t("filter.statusPending")}</SelectItem><SelectItem value="processed">{t("filter.statusProcessed")}</SelectItem><SelectItem value="failed">{t("filter.statusFailed")}</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("table.provider")}</TableHead>
                <TableHead>{t("table.eventType")}</TableHead>
                <TableHead>{t("table.received")}</TableHead>
                <TableHead>{t("table.status")}</TableHead>
                <TableHead>{t("table.payloadPreview")}</TableHead>
                <TableHead>{t("table.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((e) => (
                <TableRow key={e.id}>
                  <TableCell><Badge className={providerMap[e.provider] ?? ""}>{e.provider}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{e.type}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{e.received}</TableCell>
                  <TableCell><Badge className={statusClassMap[e.status]}>{t(`status.${e.status}`)}</Badge></TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">{e.preview}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="outline" size="sm">{t("actions.view")}</Button>
                      <Button variant="outline" size="sm">{t("actions.replay")}</Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 28 →  &nbsp; {t("pagination")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
