import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const retryRules = [
  { attempt: 1, interval: 1, action: "Charge card", notification: "Email: payment_failed" },
  { attempt: 2, interval: 3, action: "Charge card", notification: "Email: retry_notice" },
  { attempt: 3, interval: 7, action: "Charge card + SMS", notification: "Email+SMS: final_warning" },
  { attempt: 4, interval: 14, action: "Escalate to grace", notification: "Email: grace_start" },
];

export default async function DunningPage() {
  const t = await getTranslations("dunning");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>

      {/* Retry Rules */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("retryRules.title")}</CardTitle><p className="text-xs text-muted-foreground">dunning table</p></CardHeader>
        <CardContent>
          <Table>
            <TableHeader><TableRow><TableHead>{t("retryRules.attempt")}</TableHead><TableHead>{t("retryRules.interval")}</TableHead><TableHead>{t("retryRules.action")}</TableHead><TableHead>{t("retryRules.notification")}</TableHead></TableRow></TableHeader>
            <TableBody>
              {retryRules.map((r) => (
                <TableRow key={r.attempt}>
                  <TableCell>{r.attempt}</TableCell>
                  <TableCell>{r.interval}</TableCell>
                  <TableCell>{r.action}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{r.notification}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Grace Period Config */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("gracePeriod.title")}</CardTitle><p className="text-xs text-muted-foreground">grace_periods table</p></CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-3 items-center">
            <Input placeholder={t("gracePeriod.durationPlaceholder")} defaultValue="7" className="w-52" />
            <Input placeholder={t("gracePeriod.maxPlaceholder")} defaultValue="2" className="w-52" />
            <div className="flex items-center gap-2"><Switch id="notify-grace" defaultChecked /><label htmlFor="notify-grace" className="text-sm">{t("gracePeriod.notifyOnStart")}</label></div>
          </div>
        </CardContent>
      </Card>

      {/* Escalation Rules */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("escalation.title")}</CardTitle></CardHeader>
        <CardContent className="space-y-1 text-sm">
          <p>{t("escalation.rule1")}</p>
          <p>{t("escalation.rule2")}</p>
        </CardContent>
      </Card>

      {/* Flow Preview */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("flowPreview.title")}</CardTitle></CardHeader>
        <CardContent className="font-mono text-xs space-y-1 text-muted-foreground">
          <p>❌ Payment fails  →  Day +1: Retry #1  →  Day +3: Retry #2</p>
          <p>→ Day +7: Retry #3  →  Day +14: Grace Period Starts (7 days)</p>
          <p>→ Day +21: Grace Ends  →  Sub Cancelled + Winback triggered</p>
        </CardContent>
      </Card>

      <div className="flex items-center gap-4">
        <Card className="flex-1"><CardContent className="pt-4 text-sm font-medium">{t("currentlyInDunning")} <span className="text-orange-600 font-bold">{t("currentlyInDunningUsers")}</span></CardContent></Card>
        <Button size="sm">{t("saveConfig")}</Button>
        <Button size="sm" variant="outline">{t("previewEmailTemplate")}</Button>
      </div>
    </div>
  );
}
