import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function SettingsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Admin Settings & Config</h1>
      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="integrations">Integrations</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Platform Config</CardTitle></CardHeader>
            <CardContent className="space-y-3">
              <div><p className="text-xs font-medium mb-1">Platform Name</p><Input defaultValue="Paywall SaaS" /></div>
              <div><p className="text-xs font-medium mb-1">Support Email</p><Input defaultValue="support@paywall.local" type="email" /></div>
              <div><p className="text-xs font-medium mb-1">Default Currency</p><Input defaultValue="USD" className="w-24" /></div>
              <div className="flex items-center gap-2"><Switch id="dark-mode" /><label htmlFor="dark-mode" className="text-sm">Dark Mode Default</label></div>
              <Button size="sm">Save General</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="integrations" className="mt-4 space-y-3">
          <Card>
            <CardHeader><CardTitle className="text-sm">Stripe</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder="Stripe API Key (sk_...)" type="password" />
              <Input placeholder="Stripe Webhook Secret (whsec_...)" type="password" />
              <div className="flex items-center gap-2"><Switch id="stripe-test" /><label htmlFor="stripe-test" className="text-sm">Test Mode</label></div>
              <Button size="sm">Save Stripe</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">Apple In-App</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder="App Store Issuer ID" />
              <Input placeholder="Bundle ID (com.yourapp)" />
              <Button size="sm">Save Apple</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">Google Play</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder="Service Account JSON" />
              <Input placeholder="Package Name" />
              <Button size="sm">Save Google</Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">Matomo</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              <Input placeholder="Matomo URL (https://...)" />
              <Input placeholder="Site ID" className="w-32" />
              <Input placeholder="Auth Token" type="password" />
              <Button size="sm">Save Matomo</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Email Notifications</CardTitle></CardHeader>
            <CardContent className="space-y-2">
              {[
                "New subscription",
                "Payment failed",
                "Subscription cancelled",
                "Refund issued",
                "Webhook failed",
                "Dunning started",
              ].map((n) => (
                <div key={n} className="flex items-center gap-2"><Switch id={n} defaultChecked /><label htmlFor={n} className="text-sm">{n}</label></div>
              ))}
              <Button size="sm" className="mt-2">Save Notifications</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Access & Auth</CardTitle></CardHeader>
            <CardContent className="space-y-3">
              <div><p className="text-xs font-medium mb-1">JWT Expiry (hours)</p><Input defaultValue="24" className="w-28" /></div>
              <div className="flex items-center gap-2"><Switch id="mfa" /><label htmlFor="mfa" className="text-sm">Require MFA for admin</label></div>
              <div className="flex items-center gap-2"><Switch id="ip-whitelist" /><label htmlFor="ip-whitelist" className="text-sm">Enable IP Allowlist</label></div>
              <Separator />
              <p className="text-xs font-medium">Change Admin Password</p>
              <Input placeholder="Current password" type="password" />
              <Input placeholder="New password" type="password" />
              <Input placeholder="Confirm new password" type="password" />
              <Button size="sm" variant="destructive">Update Password</Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
