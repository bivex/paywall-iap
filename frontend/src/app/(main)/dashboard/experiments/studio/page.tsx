import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";

export default async function ExperimentStudioPage() {
  const t = await getTranslations("experimentStudio");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <Button variant="outline" size="sm">{t("backToExperiments")}</Button>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-sm">{t("config.title")}</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div><p className="text-xs font-medium mb-1">{t("config.testName")}</p><Input placeholder={t("config.testNamePlaceholder")} /></div>
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
            <div><p className="text-xs font-medium mb-1">{t("config.testType")}</p>
              <Select><SelectTrigger><SelectValue placeholder={t("config.selectType")} /></SelectTrigger><SelectContent><SelectItem value="pricing">{t("config.typePricing")}</SelectItem><SelectItem value="winback">{t("config.typeWinback")}</SelectItem><SelectItem value="paywall">{t("config.typePaywall")}</SelectItem><SelectItem value="onboarding">{t("config.typeOnboarding")}</SelectItem></SelectContent></Select>
            </div>
            <div><p className="text-xs font-medium mb-1">{t("config.trafficSplit")}</p><Input placeholder="50" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">{t("config.confidenceThreshold")}</p><Input defaultValue="95" className="font-mono" /></div>
          </div>
          <Separator />
          <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
            <Card className="border-dashed">
              <CardHeader className="pb-2"><CardTitle className="text-xs">{t("control.title")}</CardTitle></CardHeader>
              <CardContent className="space-y-2">
                <Input placeholder={t("control.variantName")} defaultValue="Control" />
                <Input placeholder={t("control.planKey")} />
                <Input placeholder={t("control.priceOverride")} />
              </CardContent>
            </Card>
            <Card className="border-dashed">
              <CardHeader className="pb-2"><CardTitle className="text-xs">{t("variant.title")}</CardTitle></CardHeader>
              <CardContent className="space-y-2">
                <Input placeholder={t("variant.variantName")} defaultValue="Variant A" />
                <Input placeholder={t("variant.planKey")} />
                <Input placeholder={t("variant.priceOverride")} />
              </CardContent>
            </Card>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="text-sm font-medium">{t("targeting.title")}</p>
            <div className="flex flex-wrap gap-2">
              <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("targeting.platformAll")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("targeting.all")}</SelectItem><SelectItem value="ios">{t("targeting.ios")}</SelectItem><SelectItem value="android">{t("targeting.android")}</SelectItem><SelectItem value="web">{t("targeting.web")}</SelectItem></SelectContent></Select>
              <Input placeholder={t("targeting.countryCode")} className="w-40" />
              <Input placeholder={t("targeting.newUsersOnly")} className="w-60" />
            </div>
            <div className="flex items-center gap-2"><Switch id="holdout" /><label htmlFor="holdout" className="text-sm">{t("targeting.holdout")}</label></div>
          </div>
          <div className="flex gap-2 pt-2">
            <Button size="sm">{t("saveDraft")}</Button>
            <Button size="sm" variant="default">{t("launchTest")}</Button>
            <Button size="sm" variant="outline">{t("cancel")}</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
