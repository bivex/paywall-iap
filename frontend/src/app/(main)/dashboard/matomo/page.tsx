import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default async function MatomoPage() {
  const t = await getTranslations("matomo");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <div className="flex gap-2">
          <Select>
            <SelectTrigger className="w-44"><SelectValue placeholder={t("dateRangePlaceholder")} /></SelectTrigger>
            <SelectContent>
              <SelectItem value="7">{t("last7")}</SelectItem>
              <SelectItem value="30">{t("last30")}</SelectItem>
              <SelectItem value="90">{t("last90")}</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm">{t("openMatomo")}</Button>
        </div>
      </div>

      {/* Summary KPIs */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {[
          { label: t("kpi.uniqueVisitors"), value: "18,432" },
          { label: t("kpi.pageViews"), value: "74,112" },
          { label: t("kpi.avgSession"), value: "3m 24s" },
          { label: t("kpi.bounceRate"), value: "34.2%" },
        ].map((k) => (
          <Card key={k.label}>
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase">{k.label}</CardTitle></CardHeader>
            <CardContent><div className="text-2xl font-bold">{k.value}</div></CardContent>
          </Card>
        ))}
      </div>

      {/* Iframe embed placeholder */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("embed.title")}</CardTitle><p className="text-xs text-muted-foreground">{t("embed.envHint")}</p></CardHeader>
        <CardContent>
          <div className="w-full rounded-md border-2 border-dashed border-muted-foreground/30 flex items-center justify-center bg-muted/30" style={{ height: 480 }}>
            <div className="text-center space-y-2">
              <p className="text-sm text-muted-foreground">{t("embed.iframeLabel")}</p>
              <p className="text-xs text-muted-foreground">{t("embed.notConfigured")}</p>
              <Button variant="outline" size="sm">{t("embed.configure")}</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Top pages */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("topPages.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-1 text-sm">
            <div className="flex justify-between"><span>/pricing</span><span className="font-mono text-muted-foreground">12,340</span></div>
            <div className="flex justify-between"><span>/checkout</span><span className="font-mono text-muted-foreground">8,721</span></div>
            <div className="flex justify-between"><span>/dashboard</span><span className="font-mono text-muted-foreground">7,209</span></div>
            <div className="flex justify-between"><span>/onboarding</span><span className="font-mono text-muted-foreground">5,112</span></div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("trafficSources.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-1 text-sm">
            <div className="flex justify-between"><span>{t("trafficSources.direct")}</span><span className="font-mono text-muted-foreground">38%</span></div>
            <div className="flex justify-between"><span>{t("trafficSources.organic")}</span><span className="font-mono text-muted-foreground">27%</span></div>
            <div className="flex justify-between"><span>{t("trafficSources.referral")}</span><span className="font-mono text-muted-foreground">19%</span></div>
            <div className="flex justify-between"><span>{t("trafficSources.social")}</span><span className="font-mono text-muted-foreground">16%</span></div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
