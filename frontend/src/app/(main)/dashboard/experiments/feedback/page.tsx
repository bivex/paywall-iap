import { getTranslations } from "next-intl/server";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

const questions = [
  { id: 1, type: "nps", text: "How likely are you to recommend us? (0–10)" },
  { id: 2, type: "multiple_choice", text: "Why did you cancel?" },
  { id: 3, type: "open_ended", text: "Any additional feedback?" },
];

export default async function FeedbackPage() {
  const t = await getTranslations("feedback");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">{t("previewForm")}</Button>
          <Button size="sm">{t("addQuestion")}</Button>
        </div>
      </div>

      {/* Form meta */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("formSettings.title")}</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <div><p className="text-xs font-medium mb-1">{t("formSettings.formName")}</p><Input defaultValue="Post-Cancellation Survey" /></div>
          <div className="flex flex-wrap gap-3">
            <div><p className="text-xs font-medium mb-1">{t("formSettings.trigger")}</p>
              <Select><SelectTrigger className="w-48"><SelectValue placeholder="on_cancellation" /></SelectTrigger><SelectContent><SelectItem value="cancel">on_cancellation</SelectItem><SelectItem value="downgrade">on_downgrade</SelectItem><SelectItem value="trial_end">on_trial_end</SelectItem></SelectContent></Select>
            </div>
            <div><p className="text-xs font-medium mb-1">{t("formSettings.displayDelay")}</p><Input defaultValue="2" className="w-28 font-mono" /></div>
          </div>
        </CardContent>
      </Card>

      {/* Questions */}
      <div className="space-y-3">
        {questions.map((q, i) => (
          <Card key={q.id} className="border-l-4 border-l-primary">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-xs text-muted-foreground">{t("question")} {i + 1} · {q.type}</CardTitle>
                <div className="flex gap-1">
                  <Button variant="ghost" size="sm">↑</Button>
                  <Button variant="ghost" size="sm">↓</Button>
                  <Button variant="ghost" size="sm" className="text-destructive">✕</Button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              <Input defaultValue={q.text} />
              {q.type === "multiple_choice" && (
                <Textarea placeholder="Option 1&#10;Option 2&#10;Option 3" rows={3} defaultValue={"Price too high\nNot using enough\nMissing features\nOther"} />
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <Separator />
      <div className="flex gap-2">
        <Button size="sm">{t("saveForm")}</Button>
        <Button size="sm" variant="outline">{t("cancel")}</Button>
      </div>
    </div>
  );
}
