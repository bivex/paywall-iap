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

const statusMap: Record<string, { label: string; className: string }> = {
  success: { label: "✅ Success", className: "bg-green-100 text-green-800" },
  failed: { label: "❌ Failed", className: "bg-red-100 text-red-800" },
  refunded: { label: "↩️ Refunded", className: "bg-gray-100 text-gray-800" },
};

export default function TransactionsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">Transaction Reconciliation</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <div className="flex flex-wrap gap-2">
            <Input placeholder="receipt_hash / user email..." className="max-w-sm" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Status: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="success">Success</SelectItem><SelectItem value="failed">Failed</SelectItem><SelectItem value="refunded">Refunded</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Currency: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="usd">USD</SelectItem><SelectItem value="eur">EUR</SelectItem><SelectItem value="gbp">GBP</SelectItem></SelectContent></Select>
            <Input type="date" className="w-40" />
            <Input type="date" className="w-40" />
          </div>
          {/* Summary row */}
          <div className="flex gap-4 rounded-md bg-muted p-3 text-sm">
            <span className="font-medium">TOTAL: $48,234</span>
            <span className="text-green-700">✅ 412 success</span>
            <span className="text-red-700">❌ 8 failed</span>
            <span className="text-muted-foreground">↩️ 3 refunded</span>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>Date</TableHead>
                <TableHead>User</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead>Currency</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Receipt</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactions.map((t) => (
                <TableRow key={t.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell>{t.date}</TableCell>
                  <TableCell className="font-medium">{t.user}</TableCell>
                  <TableCell className="text-right font-mono">{t.amount}</TableCell>
                  <TableCell>{t.currency}</TableCell>
                  <TableCell><Badge className={statusMap[t.status].className}>{statusMap[t.status].label}</Badge></TableCell>
                  <TableCell>{t.source}</TableCell>
                  <TableCell className="text-xs text-muted-foreground font-mono">{t.receipt}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 217 →  &nbsp; Showing 1–5 of 1,085 transactions</p>
        </CardContent>
      </Card>
    </div>
  );
}
