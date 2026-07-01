"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import {
  Apple,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Globe,
  KeyRound,
  Loader2,
  Save,
  Settings2,
  Smartphone,
  Webhook,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { useAppStore } from "@/stores/app-store";

// ── Types ─────────────────────────────────────────────────────────────────────

interface AppSettings {
  grace_period_days: number;
  trial_enabled: boolean;
  trial_days: number;
  default_currency: string;
  webhook_url: string;
  webhook_secret: string;
  store_environment: "production" | "sandbox";
  entitlements: Record<string, string[]>;
  subscription_required_for: string[];
}

interface CredentialStatus {
  provider: string;
  apple_team_id?: string;
  apple_key_id?: string;
  apple_bundle_id?: string;
  apple_environment?: string;
  apple_shared_secret_set: boolean;
  apple_private_key_set: boolean;
  google_package_name?: string;
  google_service_account_set: boolean;
  stripe_publishable_key?: string;
  stripe_secret_key_set: boolean;
  stripe_webhook_secret_set: boolean;
  paddle_vendor_id?: string;
  paddle_api_key_set: boolean;
  paddle_webhook_secret_set: boolean;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const ConfiguredBadge = ({ set }: { set: boolean }) =>
  set ? (
    <Badge variant="default" className="gap-1 text-xs bg-green-600">
      <CheckCircle2 className="h-3 w-3" /> Configured
    </Badge>
  ) : (
    <Badge variant="secondary" className="text-xs">Not set</Badge>
  );

// ── Settings tab ──────────────────────────────────────────────────────────────

function SettingsTab({ appId }: { appId: string }) {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [s, setS] = useState<AppSettings>({
    grace_period_days: 3,
    trial_enabled: false,
    trial_days: 0,
    default_currency: "USD",
    webhook_url: "",
    webhook_secret: "",
    store_environment: "production",
    entitlements: {},
    subscription_required_for: [],
  });
  const [entitlementsRaw, setEntitlementsRaw] = useState("{}");
  const [entitlementsError, setEntitlementsError] = useState("");

  useEffect(() => {
    fetch(`/api/admin/apps/${appId}/settings`)
      .then((r) => r.json())
      .then((data) => {
        if (data.settings) {
          setS(data.settings);
          setEntitlementsRaw(JSON.stringify(data.settings.entitlements ?? {}, null, 2));
        }
      })
      .catch(() => toast.error("Failed to load settings"))
      .finally(() => setLoading(false));
  }, [appId]);

  const save = async () => {
    // validate entitlements JSON
    let entitlements: Record<string, string[]> = {};
    try {
      entitlements = JSON.parse(entitlementsRaw);
    } catch {
      setEntitlementsError("Invalid JSON");
      return;
    }
    setEntitlementsError("");
    setSaving(true);
    try {
      const res = await fetch(`/api/admin/apps/${appId}/settings`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...s, entitlements }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        toast.error(err.error ?? "Failed to save settings");
        return;
      }
      toast.success("Settings saved");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* General */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Settings2 className="h-4 w-4" /> General
          </CardTitle>
          <CardDescription>Paywall behaviour settings for this app.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-5 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>Default Currency (ISO-4217)</Label>
            <Input
              value={s.default_currency}
              maxLength={3}
              className="uppercase"
              onChange={(e) => setS({ ...s, default_currency: e.target.value.toUpperCase() })}
            />
          </div>
          <div className="space-y-2">
            <Label>Store Environment</Label>
            <Select
              value={s.store_environment}
              onValueChange={(v) => setS({ ...s, store_environment: v as "production" | "sandbox" })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="production">Production</SelectItem>
                <SelectItem value="sandbox">Sandbox</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Grace Period (days after expiry)</Label>
            <Input
              type="number"
              min={0}
              max={90}
              value={s.grace_period_days}
              onChange={(e) => setS({ ...s, grace_period_days: Number(e.target.value) })}
            />
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label>Trial Enabled</Label>
                <p className="text-xs text-muted-foreground mt-0.5">Allow free trial for new users</p>
              </div>
              <Switch
                checked={s.trial_enabled}
                onCheckedChange={(v) => setS({ ...s, trial_enabled: v })}
              />
            </div>
            {s.trial_enabled && (
              <div className="space-y-2">
                <Label>Trial Duration (days)</Label>
                <Input
                  type="number"
                  min={1}
                  max={365}
                  value={s.trial_days}
                  onChange={(e) => setS({ ...s, trial_days: Number(e.target.value) })}
                />
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Webhook */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Webhook className="h-4 w-4" /> Webhook
          </CardTitle>
          <CardDescription>Endpoint that receives subscription events from this app.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-5 sm:grid-cols-2">
          <div className="space-y-2 sm:col-span-2">
            <Label>Webhook URL</Label>
            <Input
              type="url"
              placeholder="https://your-server.com/webhook"
              value={s.webhook_url}
              onChange={(e) => setS({ ...s, webhook_url: e.target.value })}
            />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label>Webhook Secret</Label>
            <Input
              type="password"
              placeholder="••••••••"
              value={s.webhook_secret}
              onChange={(e) => setS({ ...s, webhook_secret: e.target.value })}
            />
          </div>
        </CardContent>
      </Card>

      {/* Entitlements */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <KeyRound className="h-4 w-4" /> Entitlements
          </CardTitle>
          <CardDescription>
            Map product IDs to feature keys. Format: {`{"com.app.monthly": ["premium","ads_free"]}`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Textarea
            rows={8}
            className="font-mono text-sm"
            value={entitlementsRaw}
            onChange={(e) => setEntitlementsRaw(e.target.value)}
          />
          {entitlementsError && (
            <p className="text-sm text-destructive mt-1">{entitlementsError}</p>
          )}
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={save} disabled={saving} className="gap-2">
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          Save Settings
        </Button>
      </div>
    </div>
  );
}

// ── Credentials tab ───────────────────────────────────────────────────────────

interface CredFormState {
  apple_shared_secret: string;
  apple_team_id: string;
  apple_key_id: string;
  apple_private_key: string;
  apple_bundle_id: string;
  apple_environment: string;
  google_package_name: string;
  google_service_account: string;
  stripe_publishable_key: string;
  stripe_secret_key: string;
  stripe_webhook_secret: string;
  paddle_vendor_id: string;
  paddle_api_key: string;
  paddle_webhook_secret: string;
}

const emptyCredForm = (): CredFormState => ({
  apple_shared_secret: "",
  apple_team_id: "",
  apple_key_id: "",
  apple_private_key: "",
  apple_bundle_id: "",
  apple_environment: "production",
  google_package_name: "",
  google_service_account: "",
  stripe_publishable_key: "",
  stripe_secret_key: "",
  stripe_webhook_secret: "",
  paddle_vendor_id: "",
  paddle_api_key: "",
  paddle_webhook_secret: "",
});

function CredentialSection({
  title,
  icon: Icon,
  expanded,
  onToggle,
  status,
  children,
}: {
  title: string;
  icon: React.ElementType;
  expanded: boolean;
  onToggle: () => void;
  status?: CredentialStatus;
  children: React.ReactNode;
}) {
  const isConfigured = status !== undefined;
  return (
    <Card>
      <CardHeader
        className="cursor-pointer select-none"
        onClick={onToggle}
      >
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <Icon className="h-4 w-4" />
            {title}
            {isConfigured && <ConfiguredBadge set={true} />}
          </CardTitle>
          {expanded ? <ChevronUp className="h-4 w-4 text-muted-foreground" /> : <ChevronDown className="h-4 w-4 text-muted-foreground" />}
        </div>
      </CardHeader>
      {expanded && (
        <>
          <Separator />
          <CardContent className="pt-5">{children}</CardContent>
        </>
      )}
    </Card>
  );
}

function CredentialsTab({ appId }: { appId: string }) {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [statuses, setStatuses] = useState<CredentialStatus[]>([]);
  const [form, setForm] = useState<CredFormState>(emptyCredForm());
  const [expanded, setExpanded] = useState<Record<string, boolean>>({
    apple: false,
    google: false,
    stripe: false,
    paddle: false,
  });

  const toggle = (p: string) => setExpanded((e) => ({ ...e, [p]: !e[p] }));
  const statusFor = (p: string) => statuses.find((s) => s.provider === p);

  useEffect(() => {
    fetch(`/api/admin/apps/${appId}/credentials`)
      .then((r) => r.json())
      .then((data) => {
        if (data.credentials) {
          setStatuses(data.credentials);
          // pre-fill non-sensitive fields from first status
          const apple = data.credentials.find((c: CredentialStatus) => c.provider === "apple");
          const google = data.credentials.find((c: CredentialStatus) => c.provider === "google");
          const stripe = data.credentials.find((c: CredentialStatus) => c.provider === "stripe");
          const paddle = data.credentials.find((c: CredentialStatus) => c.provider === "paddle");
          setForm((f) => ({
            ...f,
            apple_team_id: apple?.apple_team_id ?? "",
            apple_key_id: apple?.apple_key_id ?? "",
            apple_bundle_id: apple?.apple_bundle_id ?? "",
            apple_environment: apple?.apple_environment ?? "production",
            google_package_name: google?.google_package_name ?? "",
            stripe_publishable_key: stripe?.stripe_publishable_key ?? "",
            paddle_vendor_id: paddle?.paddle_vendor_id ?? "",
          }));
        }
      })
      .catch(() => toast.error("Failed to load credentials"))
      .finally(() => setLoading(false));
  }, [appId]);

  const saveProvider = async (provider: string) => {
    setSaving(true);
    try {
      const payload: Record<string, string> = { provider };
      if (provider === "apple") {
        Object.assign(payload, {
          apple_shared_secret: form.apple_shared_secret,
          apple_team_id: form.apple_team_id,
          apple_key_id: form.apple_key_id,
          apple_private_key: form.apple_private_key,
          apple_bundle_id: form.apple_bundle_id,
          apple_environment: form.apple_environment,
        });
      } else if (provider === "google") {
        Object.assign(payload, {
          google_package_name: form.google_package_name,
          google_service_account: form.google_service_account,
        });
      } else if (provider === "stripe") {
        Object.assign(payload, {
          stripe_publishable_key: form.stripe_publishable_key,
          stripe_secret_key: form.stripe_secret_key,
          stripe_webhook_secret: form.stripe_webhook_secret,
        });
      } else if (provider === "paddle") {
        Object.assign(payload, {
          paddle_vendor_id: form.paddle_vendor_id,
          paddle_api_key: form.paddle_api_key,
          paddle_webhook_secret: form.paddle_webhook_secret,
        });
      }

      const res = await fetch(`/api/admin/apps/${appId}/credentials`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        toast.error(err.error ?? "Failed to save credentials");
        return;
      }
      const data = await res.json();
      // update status
      setStatuses((prev) => {
        const next = prev.filter((s) => s.provider !== provider);
        if (data.credentials) next.push(data.credentials);
        return next;
      });
      // clear sensitive fields after save
      setForm((f) => ({
        ...f,
        apple_shared_secret: "",
        apple_private_key: "",
        google_service_account: "",
        stripe_secret_key: "",
        stripe_webhook_secret: "",
        paddle_api_key: "",
        paddle_webhook_secret: "",
      }));
      toast.success(`${provider} credentials saved`);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Apple */}
      <CredentialSection
        title="Apple App Store"
        icon={Apple}
        expanded={expanded.apple}
        onToggle={() => toggle("apple")}
        status={statusFor("apple")}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          {statusFor("apple") && (
            <div className="sm:col-span-2 flex gap-3 text-sm text-muted-foreground">
              <span>Shared secret: <ConfiguredBadge set={statusFor("apple")!.apple_shared_secret_set} /></span>
              <span>Private key: <ConfiguredBadge set={statusFor("apple")!.apple_private_key_set} /></span>
            </div>
          )}
          <div className="space-y-2">
            <Label>Team ID</Label>
            <Input value={form.apple_team_id} onChange={(e) => setForm({ ...form, apple_team_id: e.target.value })} placeholder="ABCD1234EF" />
          </div>
          <div className="space-y-2">
            <Label>Key ID</Label>
            <Input value={form.apple_key_id} onChange={(e) => setForm({ ...form, apple_key_id: e.target.value })} placeholder="ABCD1234EF" />
          </div>
          <div className="space-y-2">
            <Label>Bundle ID</Label>
            <Input value={form.apple_bundle_id} onChange={(e) => setForm({ ...form, apple_bundle_id: e.target.value })} placeholder="com.company.app" />
          </div>
          <div className="space-y-2">
            <Label>Environment</Label>
            <Select value={form.apple_environment} onValueChange={(v) => setForm({ ...form, apple_environment: v })}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="production">Production</SelectItem>
                <SelectItem value="sandbox">Sandbox</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label>Shared Secret {statusFor("apple")?.apple_shared_secret_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Input type="password" value={form.apple_shared_secret} onChange={(e) => setForm({ ...form, apple_shared_secret: e.target.value })} placeholder="••••••••" />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label>Private Key (.p8) {statusFor("apple")?.apple_private_key_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Textarea rows={5} className="font-mono text-xs" value={form.apple_private_key} onChange={(e) => setForm({ ...form, apple_private_key: e.target.value })} placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----" />
          </div>
        </div>
        <div className="flex justify-end mt-4">
          <Button onClick={() => saveProvider("apple")} disabled={saving} size="sm" className="gap-2">
            {saving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
            Save Apple Credentials
          </Button>
        </div>
      </CredentialSection>

      {/* Google */}
      <CredentialSection
        title="Google Play"
        icon={Smartphone}
        expanded={expanded.google}
        onToggle={() => toggle("google")}
        status={statusFor("google")}
      >
        <div className="grid gap-4">
          {statusFor("google") && (
            <div className="text-sm text-muted-foreground">
              Service account: <ConfiguredBadge set={statusFor("google")!.google_service_account_set} />
            </div>
          )}
          <div className="space-y-2">
            <Label>Package Name</Label>
            <Input value={form.google_package_name} onChange={(e) => setForm({ ...form, google_package_name: e.target.value })} placeholder="com.company.app" />
          </div>
          <div className="space-y-2">
            <Label>Service Account JSON {statusFor("google")?.google_service_account_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Textarea rows={8} className="font-mono text-xs" value={form.google_service_account} onChange={(e) => setForm({ ...form, google_service_account: e.target.value })} placeholder='{"type":"service_account",...}' />
          </div>
        </div>
        <div className="flex justify-end mt-4">
          <Button onClick={() => saveProvider("google")} disabled={saving} size="sm" className="gap-2">
            {saving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
            Save Google Credentials
          </Button>
        </div>
      </CredentialSection>

      {/* Stripe */}
      <CredentialSection
        title="Stripe"
        icon={Globe}
        expanded={expanded.stripe}
        onToggle={() => toggle("stripe")}
        status={statusFor("stripe")}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          {statusFor("stripe") && (
            <div className="sm:col-span-2 flex gap-3 text-sm text-muted-foreground">
              <span>Secret key: <ConfiguredBadge set={statusFor("stripe")!.stripe_secret_key_set} /></span>
              <span>Webhook secret: <ConfiguredBadge set={statusFor("stripe")!.stripe_webhook_secret_set} /></span>
            </div>
          )}
          <div className="space-y-2 sm:col-span-2">
            <Label>Publishable Key</Label>
            <Input value={form.stripe_publishable_key} onChange={(e) => setForm({ ...form, stripe_publishable_key: e.target.value })} placeholder="pk_live_..." />
          </div>
          <div className="space-y-2">
            <Label>Secret Key {statusFor("stripe")?.stripe_secret_key_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Input type="password" value={form.stripe_secret_key} onChange={(e) => setForm({ ...form, stripe_secret_key: e.target.value })} placeholder="sk_live_..." />
          </div>
          <div className="space-y-2">
            <Label>Webhook Secret {statusFor("stripe")?.stripe_webhook_secret_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Input type="password" value={form.stripe_webhook_secret} onChange={(e) => setForm({ ...form, stripe_webhook_secret: e.target.value })} placeholder="whsec_..." />
          </div>
        </div>
        <div className="flex justify-end mt-4">
          <Button onClick={() => saveProvider("stripe")} disabled={saving} size="sm" className="gap-2">
            {saving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
            Save Stripe Credentials
          </Button>
        </div>
      </CredentialSection>

      {/* Paddle */}
      <CredentialSection
        title="Paddle"
        icon={Globe}
        expanded={expanded.paddle}
        onToggle={() => toggle("paddle")}
        status={statusFor("paddle")}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          {statusFor("paddle") && (
            <div className="sm:col-span-2 flex gap-3 text-sm text-muted-foreground">
              <span>API key: <ConfiguredBadge set={statusFor("paddle")!.paddle_api_key_set} /></span>
              <span>Webhook secret: <ConfiguredBadge set={statusFor("paddle")!.paddle_webhook_secret_set} /></span>
            </div>
          )}
          <div className="space-y-2 sm:col-span-2">
            <Label>Vendor ID</Label>
            <Input value={form.paddle_vendor_id} onChange={(e) => setForm({ ...form, paddle_vendor_id: e.target.value })} placeholder="12345" />
          </div>
          <div className="space-y-2">
            <Label>API Key {statusFor("paddle")?.paddle_api_key_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Input type="password" value={form.paddle_api_key} onChange={(e) => setForm({ ...form, paddle_api_key: e.target.value })} placeholder="••••••••" />
          </div>
          <div className="space-y-2">
            <Label>Webhook Secret {statusFor("paddle")?.paddle_webhook_secret_set && <span className="text-xs text-muted-foreground ml-1">(leave blank to keep existing)</span>}</Label>
            <Input type="password" value={form.paddle_webhook_secret} onChange={(e) => setForm({ ...form, paddle_webhook_secret: e.target.value })} placeholder="••••••••" />
          </div>
        </div>
        <div className="flex justify-end mt-4">
          <Button onClick={() => saveProvider("paddle")} disabled={saving} size="sm" className="gap-2">
            {saving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
            Save Paddle Credentials
          </Button>
        </div>
      </CredentialSection>
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────────────────────────

export function AppConfigPageClient() {
  const params = useParams<{ id: string }>();
  const appId = params.id;
  const app = useAppStore((s) => s.apps.find((a) => a.id === appId));

  return (
    <div className="container mx-auto max-w-3xl py-8 px-4 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
          <Settings2 className="h-5 w-5" />
          App Configuration
        </h1>
        {app && (
          <p className="text-muted-foreground mt-1 text-sm">
            {app.display_name || app.name} · <span className="font-mono">{app.bundle_id}</span>
          </p>
        )}
      </div>

      <Tabs defaultValue="settings">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="credentials">Store Credentials</TabsTrigger>
        </TabsList>
        <TabsContent value="settings" className="mt-6">
          <SettingsTab appId={appId} />
        </TabsContent>
        <TabsContent value="credentials" className="mt-6">
          <CredentialsTab appId={appId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
