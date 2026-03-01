import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const transactions = [
  { id: "txn_001", date: "2025-07-01", user: "alice@example.com", amount: "$49.99", currency: "USD", status: "success", source: "Apple IAP", receipt: "rcp_abc...123" },
  { id: "txn_002", date: "2025-06-28", user: "bob@example.com", amount: "$9.99", currency: "USD", status: "failed", source: "Google Play", receipt: "rcp_def...456" },
  { id: "txn_003", date: "2025-06-25", user: "carol@example.com", amount: "$29.99", currency: "USD", status: "success", source: "Stripe", receipt: "rcp_ghi...789" },
  { id: "txn_004", date: "2025-06-20", user: "dave@example.com", amount: "€9.24", currency: "EUR", status: "refunded", source: "Stripe", receipt: "rcp_jkl...012" },
  { id: "txn_005", date: "2025-06-18", user: "eve@example.com", amount: "$99.99", currency: "USD", status: "success", source: "Apple IAP", receipt: "rcp_mno...345" },
];

const statusClassMap: Record<string, string> = {
  success: "bg-green-100 text-green-800",
  failed: "bg-red-100 text-red-800",
  refunded: "bg-gray-100 text-gray-800",
};

export default async function TransactionsPage() {
  const t = await getTranslations("transactions");
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Input placeholder={t("filter.searchPlaceholder")} className="max-w-sm" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.statusPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.statusAll")}</SelectItem><SelectItem value="success">{t("filter.statusSuccess")}</SelectItem><SelectItem value="failed">{t("filter.statusFailed")}</SelectItem><SelectItem value="refunded">{t("filter.statusRefunded")}</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.currencyPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.currencyAll")}</SelectItem><SelectItem value="usd">USD</SelectItem><SelectItem value="eur">EUR</SelectItem><SelectItem value="gbp">GBP</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          {/* Summary row */}
          <div className="flex gap-4 rounded-md bg-muted p-3 text-sm">
            <span className="font-medium">TOTAL: $48,234</span>
            <span className="text-green-700">✅ 412 {t("summary.success")}</span>
            <span className="text-red-700">❌ 8 {t("summary.failed")}</span>
            <span className="text-muted-foreground">↩️ 3 {t("summary.refunded")}</span>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>{t("table.date")}</TableHead>
                <TableHead>{t("table.user")}</TableHead>
                <TableHead className="text-right">{t("table.amount")}</TableHead>
                <TableHead>{t("table.currency")}</TableHead>
                <TableHead>{t("table.status")}</TableHead>
                <TableHead>{t("table.source")}</TableHead>
                <TableHead>{t("table.receipt")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactions.map((tx) => (
                <TableRow key={tx.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell>{tx.date}</TableCell>
                  <TableCell className="font-medium">{tx.user}</TableCell>
                  <TableCell className="text-right font-mono">{tx.amount}</TableCell>
                  <TableCell>{tx.currency}</TableCell>
                  <TableCell><Badge className={statusClassMap[tx.status]}>{t(`status.${tx.status}`)}</Badge></TableCell>
                  <TableCell>{tx.source}</TableCell>
                  <TableCell className="text-xs text-muted-foreground font-mono">{tx.receipt}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 217 →  &nbsp; {t("pagination")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
