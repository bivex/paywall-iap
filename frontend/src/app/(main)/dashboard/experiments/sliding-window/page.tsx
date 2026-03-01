import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";

export default function SlidingWindowPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Sliding Window Test Config</h1>

      <Card>
        <CardHeader><CardTitle className="text-sm">Window Parameters</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <div><p className="text-xs font-medium mb-1">Window Size (days)</p><Input defaultValue="14" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">Step Size (days)</p><Input defaultValue="7" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">Min Observations</p><Input defaultValue="500" className="font-mono" /></div>
            <div><p className="text-xs font-medium mb-1">Significance α</p><Input defaultValue="0.05" className="font-mono" /></div>
          </div>
          <div className="flex items-center gap-2"><Switch id="auto-stop" defaultChecked /><label htmlFor="auto-stop" className="text-sm">Auto-stop when significance reached</label></div>
          <div className="flex items-center gap-2"><Switch id="drift-detect" /><label htmlFor="drift-detect" className="text-sm">Enable drift detection</label></div>

          <Separator />
          <p className="text-sm font-medium">Current Window Progress</p>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between"><span>Window 3 of ~8</span><span className="font-mono text-muted-foreground">Day 8 of 14</span></div>
            <Progress value={57} className="h-2" />
            <div className="grid grid-cols-3 gap-2 text-xs text-muted-foreground pt-1">
              <div><p className="font-medium text-foreground">Observations</p><p>1,204</p></div>
              <div><p className="font-medium text-foreground">p-value</p><p className="font-mono">0.031</p></div>
              <div><p className="font-medium text-foreground">Winner so far</p><p>Variant B</p></div>
            </div>
          </div>

          <div className="flex gap-2">
            <Button size="sm">Save Config</Button>
            <Button size="sm" variant="outline">Reset Window</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
