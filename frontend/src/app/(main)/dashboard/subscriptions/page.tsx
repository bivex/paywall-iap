import { getTranslations } from "next-intl/server";
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

const statusClassMap: Record<string, string> = {
  active: "bg-green-100 text-green-800",
  grace: "bg-yellow-100 text-yellow-800",
  dunning: "bg-orange-100 text-orange-800",
  expired: "bg-red-100 text-red-800",
};

export default async function SubscriptionsPage() {
  const t = await getTranslations("subscriptions");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Select><SelectTrigger className="w-40"><SelectValue placeholder={t("filter.statusPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.statusAll")}</SelectItem><SelectItem value="active">{t("filter.statusActive")}</SelectItem><SelectItem value="grace">{t("filter.statusGrace")}</SelectItem><SelectItem value="dunning">{t("filter.statusDunning")}</SelectItem><SelectItem value="expired">{t("filter.statusExpired")}</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.sourcePlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.sourceAll")}</SelectItem><SelectItem value="apple">Apple IAP</SelectItem><SelectItem value="google">Google Play</SelectItem><SelectItem value="stripe">Stripe</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.planPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.planAll")}</SelectItem><SelectItem value="basic">{t("filter.planBasic")}</SelectItem><SelectItem value="pro">{t("filter.planPro")}</SelectItem><SelectItem value="enterprise">{t("filter.planEnterprise")}</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          <div className="flex gap-2">
            <Button variant="destructive" size="sm">{t("bulkCancel")}</Button>
            <Button variant="outline" size="sm">{t("bulkRenew")}</Button>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>{t("table.user")}</TableHead>
                <TableHead>{t("table.status")}</TableHead>
                <TableHead>{t("table.source")}</TableHead>
                <TableHead>{t("table.plan")}</TableHead>
                <TableHead>{t("table.expires")}</TableHead>
                <TableHead>{t("table.ltv")}</TableHead>
                <TableHead>{t("table.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {subs.map((s) => (
                <TableRow key={s.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell className="font-medium">{s.user}</TableCell>
                  <TableCell><Badge className={statusClassMap[s.status]}>{t(`status.${s.status}`)}</Badge></TableCell>
                  <TableCell>{s.source}</TableCell>
                  <TableCell>{s.plan}</TableCell>
                  <TableCell>{s.expires}</TableCell>
                  <TableCell>{s.ltv}</TableCell>
                  <TableCell className="text-primary text-sm cursor-pointer">⋯</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 89 →  &nbsp; {t("pagination")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
