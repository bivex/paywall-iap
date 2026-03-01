import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const tests = [
  { id: "test_001", name: "Pricing A/B: Pro Annual vs Monthly", status: "running", type: "pricing", confidence: 87, threshold: 95, arms: "Control vs Variant A" },
  { id: "test_002", name: "Winback Discount: 30% vs 50%", status: "running", type: "winback", confidence: 71, threshold: 95, arms: "30% vs 50%" },
  { id: "test_003", name: "Paywall CTA: Modal vs Inline", status: "running", type: "paywall", confidence: 54, threshold: 95, arms: "Modal vs Inline" },
  { id: "test_004", name: "Onboarding CTA: Variant B", status: "draft", type: "onboarding", confidence: 0, threshold: 95, arms: "Control vs Variant B" },
  { id: "test_005", name: "Price Point: $9.99 vs $7.99", status: "draft", type: "pricing", confidence: 0, threshold: 95, arms: "Control vs Variant" },
];

const statusMap: Record<string, { label: string; className: string }> = {
  running: { label: "🟢 Running", className: "bg-green-100 text-green-800" },
  draft: { label: "🟡 Draft", className: "bg-yellow-100 text-yellow-800" },
  completed: { label: "⚫ Completed", className: "bg-gray-100 text-gray-700" },
};

function TestCard({ test }: { test: typeof tests[0] }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-sm font-medium">{test.name}</CardTitle>
          <Badge className={statusMap[test.status].className}>{statusMap[test.status].label}</Badge>
        </div>
        <p className="text-xs text-muted-foreground">Type: {test.type} · Arms: {test.arms}</p>
      </CardHeader>
      <CardContent className="space-y-2">
        {test.status === "running" && (
          <>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Confidence: {test.confidence}%</span>
              <span>Threshold: {test.threshold}%</span>
            </div>
            <Progress value={test.confidence} className="h-2" />
          </>
        )}
        {test.status === "draft" && <p className="text-xs text-muted-foreground">Not started</p>}
        <div className="flex gap-2 pt-1">
          {test.status === "running" && <><Button variant="outline" size="sm">View Details</Button><Button variant="destructive" size="sm">Stop Test</Button></>}
          {test.status === "draft" && <><Button variant="outline" size="sm">Edit Draft</Button><Button size="sm">Launch</Button></>}
        </div>
      </CardContent>
    </Card>
  );
}

export default function ExperimentsPage() {
  const running = tests.filter((t) => t.status === "running");
  const drafts = tests.filter((t) => t.status === "draft");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">A/B Test Discovery</h1>
        <Button size="sm">+ New Test</Button>
      </div>
      <Tabs defaultValue="all">
        <TabsList>
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="running">🟢 Running ({running.length})</TabsTrigger>
          <TabsTrigger value="draft">🟡 Draft ({drafts.length})</TabsTrigger>
          <TabsTrigger value="completed">⚫ Completed (7)</TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="mt-4 space-y-3">{tests.map((t) => <TestCard key={t.id} test={t} />)}</TabsContent>
        <TabsContent value="running" className="mt-4 space-y-3">{running.map((t) => <TestCard key={t.id} test={t} />)}</TabsContent>
        <TabsContent value="draft" className="mt-4 space-y-3">{drafts.map((t) => <TestCard key={t.id} test={t} />)}</TabsContent>
        <TabsContent value="completed" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">7 completed tests — contact engineering for archive access.</CardContent></Card></TabsContent>
      </Tabs>
      <p className="text-xs text-muted-foreground">← 1  2  3 →  &nbsp; Showing 1–5 of 12</p>
    </div>
  );
}
