import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const users = [
  { id: "usr_001", email: "alice@example.com", platform: "iOS", ltv: "$184.20", status: "active", role: "user" },
  { id: "usr_002", email: "bob@example.com", platform: "Android", ltv: "$92.40", status: "grace", role: "user" },
  { id: "usr_003", email: "carol@example.com", platform: "iOS", ltv: "$312.60", status: "dunning", role: "user" },
  { id: "usr_004", email: "dave@example.com", platform: "Android", ltv: "$47.00", status: "expired", role: "user" },
  { id: "usr_005", email: "eve@example.com", platform: "Web", ltv: "$220.00", status: "active", role: "admin" },
];

const statusMap: Record<string, { label: string; className: string }> = {
  active: { label: "✅ Active", className: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200" },
  grace: { label: "🔶 Grace", className: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200" },
  dunning: { label: "⚠️ Dunning", className: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200" },
  expired: { label: "❌ Expired", className: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200" },
};

export default function UsersPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">User List</h1>
        <Button variant="outline" size="sm">Export CSV</Button>
      </div>

      <Card>
        <CardContent className="pt-4 space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap gap-2">
            <Input placeholder="Search by email / platform_user_id / device_id..." className="max-w-sm" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder="Platform: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="ios">iOS</SelectItem><SelectItem value="android">Android</SelectItem><SelectItem value="web">Web</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-32"><SelectValue placeholder="Role: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="user">User</SelectItem><SelectItem value="admin">Admin</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-40"><SelectValue placeholder="Sub Status: All" /></SelectTrigger><SelectContent><SelectItem value="all">All</SelectItem><SelectItem value="active">Active</SelectItem><SelectItem value="grace">Grace</SelectItem><SelectItem value="dunning">Dunning</SelectItem><SelectItem value="expired">Expired</SelectItem></SelectContent></Select>
            <Input placeholder="LTV Min $" className="w-28" />
            <Input placeholder="LTV Max $" className="w-28" />
          </div>
          <div className="flex gap-2">
            <Button variant="destructive" size="sm">Bulk: Cancel Subscriptions</Button>
            <Button variant="outline" size="sm">Bulk: Grant Grace</Button>
          </div>

          {/* Table */}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Platform</TableHead>
                <TableHead>LTV</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((u) => (
                <TableRow key={u.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell className="font-medium">{u.email}</TableCell>
                  <TableCell>{u.platform}</TableCell>
                  <TableCell>{u.ltv}</TableCell>
                  <TableCell><Badge className={statusMap[u.status].className}>{statusMap[u.status].label}</Badge></TableCell>
                  <TableCell><Badge variant="outline">{u.role}</Badge></TableCell>
                  <TableCell><Link href={`/dashboard/users/${u.id}`} className="text-primary text-sm hover:underline">View Profile →</Link></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 576 →  &nbsp; Showing 1–5 of 2,876 users</p>
        </CardContent>
      </Card>
    </div>
  );
}
