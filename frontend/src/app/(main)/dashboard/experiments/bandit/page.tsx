import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";

const arms = [
  { name: "Pro Annual $79.99", pulls: 1240, conversions: 87, rate: 7.02, weight: 38 },
  { name: "Pro Annual $69.99", pulls: 980, conversions: 82, rate: 8.37, weight: 42 },
  { name: "Pro Annual $89.99", pulls: 640, conversions: 31, rate: 4.84, weight: 20 },
];

export default function BanditPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Bandit Algorithm Config</h1>
        <Badge className="bg-green-100 text-green-800">🟢 Running</Badge>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-sm">Bandit: Pricing Optimization (Thompson Sampling)</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <Select><SelectTrigger className="w-52"><SelectValue placeholder="Algorithm: Thompson Sampling" /></SelectTrigger><SelectContent><SelectItem value="thompson">Thompson Sampling</SelectItem><SelectItem value="ucb1">UCB1</SelectItem><SelectItem value="epsilon">ε-Greedy</SelectItem></SelectContent></Select>
            <Input placeholder="Exploration rate ε" defaultValue="0.1" className="w-36 font-mono" />
            <Input placeholder="Min pulls per arm" defaultValue="100" className="w-36 font-mono" />
          </div>
          <Separator />
          <p className="text-sm font-medium">Arm Performance</p>
          {arms.map((a) => (
            <div key={a.name} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium">{a.name}</span>
                <span className="font-mono text-muted-foreground">{a.conversions}/{a.pulls} ({a.rate}%)</span>
              </div>
              <div className="flex items-center gap-2">
                <Progress value={a.weight} className="h-2 flex-1" />
                <span className="text-xs font-mono text-muted-foreground w-10">{a.weight}%</span>
              </div>
            </div>
          ))}
          <Separator />
          <div className="flex gap-2">
            <Button size="sm" variant="outline">Pause</Button>
            <Button size="sm" variant="destructive">Stop Bandit</Button>
            <Button size="sm">Save Config</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
