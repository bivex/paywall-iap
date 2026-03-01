import { getTranslations } from "next-intl/server";
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

const statusClassMap: Record<string, string> = {
  running: "bg-green-100 text-green-800",
  draft: "bg-yellow-100 text-yellow-800",
  completed: "bg-gray-100 text-gray-700",
};

type TranslationFn = (key: string) => string;

function TestCard({ test, t }: { test: (typeof tests)[0]; t: TranslationFn }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-sm font-medium">{test.name}</CardTitle>
          <Badge className={statusClassMap[test.status]}>{t(`status.${test.status}`)}</Badge>
        </div>
        <p className="text-xs text-muted-foreground">{t("card.type")} {test.type} · {t("card.arms")} {test.arms}</p>
      </CardHeader>
      <CardContent className="space-y-2">
        {test.status === "running" && (
          <>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{t("card.confidence")} {test.confidence}%</span>
              <span>{t("card.threshold")} {test.threshold}%</span>
            </div>
            <Progress value={test.confidence} className="h-2" />
          </>
        )}
        {test.status === "draft" && <p className="text-xs text-muted-foreground">{t("card.notStarted")}</p>}
        <div className="flex gap-2 pt-1">
          {test.status === "running" && <><Button variant="outline" size="sm">{t("card.viewDetails")}</Button><Button variant="destructive" size="sm">{t("card.stopTest")}</Button></>}
          {test.status === "draft" && <><Button variant="outline" size="sm">{t("card.editDraft")}</Button><Button size="sm">{t("card.launch")}</Button></>}
        </div>
      </CardContent>
    </Card>
  );
}

export default async function ExperimentsPage() {
  const t = await getTranslations("experiments");
  const running = tests.filter((t) => t.status === "running");
  const drafts = tests.filter((t) => t.status === "draft");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <Button size="sm">{t("newTest")}</Button>
      </div>
      <Tabs defaultValue="all">
        <TabsList>
          <TabsTrigger value="all">{t("tabs.all")}</TabsTrigger>
          <TabsTrigger value="running">{t("tabs.running")} ({running.length})</TabsTrigger>
          <TabsTrigger value="draft">{t("tabs.draft")} ({drafts.length})</TabsTrigger>
          <TabsTrigger value="completed">{t("tabs.completed")} (7)</TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="mt-4 space-y-3">{tests.map((test) => <TestCard key={test.id} test={test} t={t as unknown as TranslationFn} />)}</TabsContent>
        <TabsContent value="running" className="mt-4 space-y-3">{running.map((test) => <TestCard key={test.id} test={test} t={t as unknown as TranslationFn} />)}</TabsContent>
        <TabsContent value="draft" className="mt-4 space-y-3">{drafts.map((test) => <TestCard key={test.id} test={test} t={t as unknown as TranslationFn} />)}</TabsContent>
        <TabsContent value="completed" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">{t("completedArchive")}</CardContent></Card></TabsContent>
      </Tabs>
      <p className="text-xs text-muted-foreground">← 1  2  3 →  &nbsp; {t("pagination")}</p>
    </div>
  );
}
