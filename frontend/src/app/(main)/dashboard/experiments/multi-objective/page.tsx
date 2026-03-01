import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const objectives = [
  { name: "Conversion Rate", weight: 40, direction: "maximize", current: 7.2, baseline: 6.8 },
  { name: "Revenue per User", weight: 35, direction: "maximize", current: 9.14, baseline: 8.91 },
  { name: "Churn Rate", weight: 25, direction: "minimize", current: 1.9, baseline: 2.1 },
];

export default async function MultiObjectivePage() {
  const t = await getTranslations("multiObjective");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>

      <Card>
        <CardHeader><CardTitle className="text-sm">{t("card.title")}</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          {objectives.map((o) => (
            <div key={o.name} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium">{o.name}</span>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">{o.direction}</span>
                  <Input defaultValue={String(o.weight)} className="w-16 font-mono h-7 text-xs" />
                  <span className="text-xs text-muted-foreground">%</span>
                </div>
              </div>
              <Progress value={o.weight} className="h-1.5" />
              <div className="flex gap-4 text-xs text-muted-foreground">
                <span>{t("current")} <span className="font-mono text-foreground">{o.current}</span></span>
                <span>{t("baseline")} <span className="font-mono text-foreground">{o.baseline}</span></span>
                <span className={o.direction === "maximize" ? (o.current > o.baseline ? "text-green-600" : "text-red-600") : (o.current < o.baseline ? "text-green-600" : "text-red-600")}>
                  {o.direction === "maximize" ? (o.current > o.baseline ? "▲" : "▼") : (o.current < o.baseline ? "▲" : "▼")} {Math.abs(o.current - o.baseline).toFixed(2)}
                </span>
              </div>
            </div>
          ))}

          <Separator />
          <p className="text-sm font-medium">{t("paretoFront.title")}</p>
          <Table>
            <TableHeader><TableRow><TableHead>{t("table.variant")}</TableHead><TableHead>{t("table.conv")}</TableHead><TableHead>{t("table.revUser")}</TableHead><TableHead>{t("table.churn")}</TableHead><TableHead>{t("table.paretoScore")}</TableHead></TableRow></TableHeader>
            <TableBody>
              <TableRow><TableCell>Pro Annual $79.99</TableCell><TableCell className="font-mono">7.2%</TableCell><TableCell className="font-mono">$9.14</TableCell><TableCell className="font-mono">1.9%</TableCell><TableCell className="font-bold text-green-600">0.82</TableCell></TableRow>
              <TableRow><TableCell>Pro Annual $69.99</TableCell><TableCell className="font-mono">8.4%</TableCell><TableCell className="font-mono">$8.22</TableCell><TableCell className="font-mono">2.0%</TableCell><TableCell className="font-bold">0.74</TableCell></TableRow>
              <TableRow><TableCell>Pro Annual $89.99</TableCell><TableCell className="font-mono">5.1%</TableCell><TableCell className="font-mono">$10.18</TableCell><TableCell className="font-mono">1.7%</TableCell><TableCell className="font-bold">0.69</TableCell></TableRow>
            </TableBody>
          </Table>

          <div className="flex items-center gap-2"><Switch id="auto-select" /><label htmlFor="auto-select" className="text-sm">{t("autoSelect")}</label></div>
          <div className="flex gap-2">
            <Button size="sm">{t("saveConfig")}</Button>
            <Button size="sm" variant="outline">{t("runAnalysis")}</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
