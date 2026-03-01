import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function MatomoPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-semibold">Matomo Web Analytics</h1>
        <div className="flex gap-2">
          <Select>
            <SelectTrigger className="w-44"><SelectValue placeholder="Date Range: Last 30 days" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="7">Last 7 days</SelectItem>
              <SelectItem value="30">Last 30 days</SelectItem>
              <SelectItem value="90">Last 90 days</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm">Open Matomo</Button>
        </div>
      </div>

      {/* Summary KPIs */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {[
          { label: "Unique Visitors", value: "18,432" },
          { label: "Page Views", value: "74,112" },
          { label: "Avg Session", value: "3m 24s" },
          { label: "Bounce Rate", value: "34.2%" },
        ].map((k) => (
          <Card key={k.label}>
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase">{k.label}</CardTitle></CardHeader>
            <CardContent><div className="text-2xl font-bold">{k.value}</div></CardContent>
          </Card>
        ))}
      </div>

      {/* Iframe embed placeholder */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Matomo Dashboard Embed</CardTitle><p className="text-xs text-muted-foreground">Set MATOMO_URL + MATOMO_SITE_ID in env to enable live iframe embed</p></CardHeader>
        <CardContent>
          <div className="w-full rounded-md border-2 border-dashed border-muted-foreground/30 flex items-center justify-center bg-muted/30" style={{ height: 480 }}>
            <div className="text-center space-y-2">
              <p className="text-sm text-muted-foreground">📊 Matomo Iframe</p>
              <p className="text-xs text-muted-foreground">MATOMO_URL not configured</p>
              <Button variant="outline" size="sm">Configure Matomo</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Top pages */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-sm">Top Pages</CardTitle></CardHeader>
          <CardContent className="space-y-1 text-sm">
            <div className="flex justify-between"><span>/pricing</span><span className="font-mono text-muted-foreground">12,340</span></div>
            <div className="flex justify-between"><span>/checkout</span><span className="font-mono text-muted-foreground">8,721</span></div>
            <div className="flex justify-between"><span>/dashboard</span><span className="font-mono text-muted-foreground">7,209</span></div>
            <div className="flex justify-between"><span>/onboarding</span><span className="font-mono text-muted-foreground">5,112</span></div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm">Traffic Sources</CardTitle></CardHeader>
          <CardContent className="space-y-1 text-sm">
            <div className="flex justify-between"><span>Direct</span><span className="font-mono text-muted-foreground">38%</span></div>
            <div className="flex justify-between"><span>Organic Search</span><span className="font-mono text-muted-foreground">27%</span></div>
            <div className="flex justify-between"><span>Referral</span><span className="font-mono text-muted-foreground">19%</span></div>
            <div className="flex justify-between"><span>Social</span><span className="font-mono text-muted-foreground">16%</span></div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
