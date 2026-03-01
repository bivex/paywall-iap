import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const retryRules = [
  { attempt: 1, interval: 1, action: "Charge card", notification: "Email: payment_failed" },
  { attempt: 2, interval: 3, action: "Charge card", notification: "Email: retry_notice" },
  { attempt: 3, interval: 7, action: "Charge card + SMS", notification: "Email+SMS: final_warning" },
  { attempt: 4, interval: 14, action: "Escalate to grace", notification: "Email: grace_start" },
];

export default function DunningPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Dunning Campaign Config</h1>

      {/* Retry Rules */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Retry Rules</CardTitle><p className="text-xs text-muted-foreground">dunning table</p></CardHeader>
        <CardContent>
          <Table>
            <TableHeader><TableRow><TableHead>Attempt</TableHead><TableHead>Interval (days)</TableHead><TableHead>Action</TableHead><TableHead>Notification</TableHead></TableRow></TableHeader>
            <TableBody>
              {retryRules.map((r) => (
                <TableRow key={r.attempt}>
                  <TableCell>{r.attempt}</TableCell>
                  <TableCell>{r.interval}</TableCell>
                  <TableCell>{r.action}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{r.notification}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Grace Period Config */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Grace Period Config</CardTitle><p className="text-xs text-muted-foreground">grace_periods table</p></CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-3 items-center">
            <Input placeholder="Grace Duration (days): 7" defaultValue="7" className="w-52" />
            <Input placeholder="Max grace periods: 2" defaultValue="2" className="w-52" />
            <div className="flex items-center gap-2"><Switch id="notify-grace" defaultChecked /><label htmlFor="notify-grace" className="text-sm">Notify on grace start</label></div>
          </div>
        </CardContent>
      </Card>

      {/* Escalation Rules */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Escalation Rules</CardTitle></CardHeader>
        <CardContent className="space-y-1 text-sm">
          <p>• After 4 failed attempts → winback campaign trigger</p>
          <p>• After grace end → admin_alert + subscription cancel</p>
        </CardContent>
      </Card>

      {/* Flow Preview */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Flow Preview</CardTitle></CardHeader>
        <CardContent className="font-mono text-xs space-y-1 text-muted-foreground">
          <p>❌ Payment fails  →  Day +1: Retry #1  →  Day +3: Retry #2</p>
          <p>→ Day +7: Retry #3  →  Day +14: Grace Period Starts (7 days)</p>
          <p>→ Day +21: Grace Ends  →  Sub Cancelled + Winback triggered</p>
        </CardContent>
      </Card>

      <div className="flex items-center gap-4">
        <Card className="flex-1"><CardContent className="pt-4 text-sm font-medium">Currently in Dunning: <span className="text-orange-600 font-bold">43 users</span></CardContent></Card>
        <Button size="sm">Save Config</Button>
        <Button size="sm" variant="outline">Preview Email Template</Button>
      </div>
    </div>
  );
}
