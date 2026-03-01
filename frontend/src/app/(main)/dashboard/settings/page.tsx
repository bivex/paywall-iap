import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default async function SettingsPage() {
  const t = await getTranslations("settings");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">{t("tabs.general")}</TabsTrigger>
          <TabsTrigger value="integrations">{t("tabs.integrations")}</TabsTrigger>
          <TabsTrigger value="notifications">{t("tabs.notifications")}</TabsTrigger>
          <TabsTrigger value="security">{t("tabs.security")}</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("general.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-3">
              <div><p className="text-xs font-medium mb-1">{t("general.platformName")}</p><Input defaultValue="Paywall SaaS" /></div>
              <div><p className="text-xs font-medium mb-1">{t("general.supportEmail")}</p><Input defaultValue="support@paywall.local" type="email" /></div>
              <div><p className="text-xs font-medium mb-1">{t("general.defaultCurrency")}</p><Input defaultValue="USD" className="w-24" /></div>
              <div className="flex items-center gap-2"><Switch id="dark-mode" /><label htmlFor="dark-mode" className="text-sm">{t("general.darkMode")}</label></div>
              <Button size="sm">{t("general.save")}</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="integrations" className="mt-4 space-y-3">
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("stripe.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder={t("stripe.apiKey")} type="password" />
              <Input placeholder={t("stripe.webhookSecret")} type="password" />
              <div className="flex items-center gap-2"><Switch id="stripe-test" /><label htmlFor="stripe-test" className="text-sm">{t("stripe.testMode")}</label></div>
              <Button size="sm">{t("stripe.save")}</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("apple.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder={t("apple.issuerId")} />
              <Input placeholder={t("apple.bundleId")} />
              <Button size="sm">{t("apple.save")}</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("google.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder={t("google.serviceAccount")} />
              <Input placeholder={t("google.packageName")} />
              <Button size="sm">{t("google.save")}</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("matomo.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder={t("matomo.url")} />
              <Input placeholder={t("matomo.siteId")} className="w-32" />
              <Input placeholder={t("matomo.authToken")} type="password" />
              <Button size="sm">{t("matomo.save")}</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("notifications.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              {[
                { key: "newSubscription", label: t("notifications.newSubscription") },
                { key: "paymentFailed", label: t("notifications.paymentFailed") },
                { key: "subscriptionCancelled", label: t("notifications.subscriptionCancelled") },
                { key: "refundIssued", label: t("notifications.refundIssued") },
                { key: "webhookFailed", label: t("notifications.webhookFailed") },
                { key: "dunningStarted", label: t("notifications.dunningStarted") },
              ].map((n) => (
                <div key={n.key} className="flex items-center gap-2"><Switch id={n.key} defaultChecked /><label htmlFor={n.key} className="text-sm">{n.label}</label></div>
              ))}
              <Button size="sm" className="mt-2">{t("notifications.save")}</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">{t("security.title")}</CardTitle></CardHeader>
            <CardContent className="space-y-3">
              <div><p className="text-xs font-medium mb-1">{t("security.jwtExpiry")}</p><Input defaultValue="24" className="w-28" /></div>
              <div className="flex items-center gap-2"><Switch id="mfa" /><label htmlFor="mfa" className="text-sm">{t("security.mfa")}</label></div>
              <div className="flex items-center gap-2"><Switch id="ip-whitelist" /><label htmlFor="ip-whitelist" className="text-sm">{t("security.ipAllowlist")}</label></div>
              <Separator />
              <p className="text-xs font-medium">{t("security.changePassword")}</p>
              <Input placeholder={t("security.currentPassword")} type="password" />
              <Input placeholder={t("security.newPassword")} type="password" />
              <Input placeholder={t("security.confirmPassword")} type="password" />
              <Button size="sm" variant="destructive">{t("security.updatePassword")}</Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
