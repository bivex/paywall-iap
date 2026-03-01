import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";

export default function ExperimentStudioPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">A/B Test Studio</h1>
        <Button variant="outline" size="sm">← Back to Experiments</Button>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-sm">Test Configuration</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div><p className="text-xs font-medium mb-1">Test Name</p><Input placeholder="e.g. Pricing A/B: Pro Annual vs Monthly" /></div>
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
            <div><p className="text-xs font-medium mb-1">Test Type</p>
              <Select><SelectTrigger><SelectValue placeholder="Select type" /></SelectTrigger><SelectContent><SelectItem value="pricing">Pricing</SelectItem><SelectItem value="winback">Winback</SelectItem><SelectItem value="paywall">Paywall</SelectItem><SelectItem value="onboarding">Onboarding</SelectItem></SelectContent></Select>
            </div>
            <div><p className="text-xs font-medium mb-1">Traffic Split (%)</p><Input placeholder="50" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">Confidence Threshold (%)</p><Input defaultValue="95" className="font-mono" /></div>
          </div>
          <Separator />
          <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
            <Card className="border-dashed">
              <CardHeader className="pb-2"><CardTitle className="text-xs">Control (A)</CardTitle></CardHeader>
              <CardContent className="space-y-2">
                <Input placeholder="Variant name: Control" defaultValue="Control" />
                <Input placeholder="Plan / config key" />
                <Input placeholder="Price override (optional)" />
              </CardContent>
            </Card>
            <Card className="border-dashed">
              <CardHeader className="pb-2"><CardTitle className="text-xs">Variant (B)</CardTitle></CardHeader>
              <CardContent className="space-y-2">
                <Input placeholder="Variant name: Variant A" defaultValue="Variant A" />
                <Input placeholder="Plan / config key" />
                <Input placeholder="Price override (optional)" />
              </CardContent>
            </Card>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="text-sm font-medium">Targeting</p>
            <div className="flex flex-wrap gap-2">
              <Select><SelectTrigger className="w-36"><SelectValue placeholder="Platform: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="ios">iOS</SelectItem><SelectItem value="android">Android</SelectItem><SelectItem value="web">Web</SelectItem></SelectContent></Select>
              <Input placeholder="Country code (optional)" className="w-40" />
              <Input placeholder="New users only: days since reg ≤" className="w-60" />
            </div>
            <div className="flex items-center gap-2"><Switch id="holdout" /><label htmlFor="holdout" className="text-sm">Enable holdout group (5%)</label></div>
          </div>
          <div className="flex gap-2 pt-2">
            <Button size="sm">Save Draft</Button>
            <Button size="sm" variant="default">Launch Test</Button>
            <Button size="sm" variant="outline">Cancel</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
