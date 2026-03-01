import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";

const campaigns = [
  { id: "wc_001", name: "Summer Winback 2025", status: "active", desc: "30% off · Expires 2025-09-01" },
  { id: "wc_002", name: "Q2 Re-Engage", status: "draft", desc: "$5 fixed · Draft" },
  { id: "wc_003", name: "Holiday 2024", status: "ended", desc: "50% off · Ended 2025-01-01" },
];

const dotMap: Record<string, string> = { active: "🟢", draft: "🟡", ended: "⚫" };

export default async function WinbackPage() {
  const t = await getTranslations("winback");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <Button size="sm">{t("newCampaign")}</Button>
      </div>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* Campaigns list */}
        <Card className="lg:col-span-1">
          <CardHeader><CardTitle className="text-sm">{t("campaigns.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {campaigns.map((c) => (
              <div key={c.id} className="cursor-pointer rounded-md border p-2 hover:bg-muted">
                <p className="text-sm font-medium">{dotMap[c.status]} {c.name}</p>
                <p className="text-xs text-muted-foreground">{c.desc}</p>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Edit form */}
        <Card className="lg:col-span-2">
          <CardHeader><CardTitle className="text-sm">{t("editCampaign.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <Input placeholder={t("editCampaign.namePlaceholder")} defaultValue="Summer Winback 2025" />
            <div className="flex gap-2">
              <Select><SelectTrigger className="w-48"><SelectValue placeholder={t("editCampaign.discountTypePlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="pct">{t("editCampaign.percentage")}</SelectItem><SelectItem value="fixed">{t("editCampaign.fixed")}</SelectItem></SelectContent></Select>
              <Input placeholder={t("editCampaign.valuePlaceholder")} defaultValue="30" className="w-32" />
            </div>
            <div>
              <p className="text-sm font-medium mb-2">{t("targeting.title")}</p>
              <div className="flex gap-2">
                <Input placeholder={t("targeting.churnedMin")} className="w-40" />
                <Input placeholder={t("targeting.churnedMax")} className="w-40" />
                <Select><SelectTrigger className="w-40"><SelectValue placeholder={t("targeting.platformPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("targeting.all")}</SelectItem><SelectItem value="ios">{t("targeting.ios")}</SelectItem><SelectItem value="android">{t("targeting.android")}</SelectItem></SelectContent></Select>
              </div>
            </div>
            <Input type="date" placeholder={t("targeting.expiresAt")} />
            <div className="flex items-center gap-2"><Switch id="ab-test" /><label htmlFor="ab-test" className="text-sm">{t("targeting.enableAbTest")}</label></div>

            <Separator />
            {/* Preview */}
            <Card className="bg-muted">
              <CardHeader className="pb-2"><CardTitle className="text-xs">{t("preview.title")}</CardTitle></CardHeader>
              <CardContent>
                <p className="text-sm">{t("preview.message")}</p>
                <Button size="sm" className="mt-2">Subscribe for $6.99/mo (was $9.99)</Button>
                <p className="text-xs text-muted-foreground mt-1">{t("preview.eligibleUsers")}</p>
              </CardContent>
            </Card>

            <div className="flex gap-2">
              <Button size="sm">{t("saveCampaign")}</Button>
              <Button size="sm" variant="default">{t("launchNow")}</Button>
              <Button size="sm" variant="outline">{t("cancel")}</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
