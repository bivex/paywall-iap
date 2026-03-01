import { getTranslations } from "next-intl/server";
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

const statusClassMap: Record<string, string> = {
  active: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  grace: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
  dunning: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
  expired: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
};

export default async function UsersPage() {
  const t = await getTranslations("users");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <Button variant="outline" size="sm">{t("exportCsv")}</Button>
      </div>

      <Card>
        <CardContent className="pt-4 space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap gap-2">
            <Input placeholder={t("filter.searchPlaceholder")} className="max-w-sm" />
            <Select><SelectTrigger className="w-36"><SelectValue placeholder={t("filter.platformPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.platformAll")}</SelectItem><SelectItem value="ios">{t("filter.platformIos")}</SelectItem><SelectItem value="android">{t("filter.platformAndroid")}</SelectItem><SelectItem value="web">{t("filter.platformWeb")}</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-32"><SelectValue placeholder={t("filter.rolePlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.roleAll")}</SelectItem><SelectItem value="user">{t("filter.roleUser")}</SelectItem><SelectItem value="admin">{t("filter.roleAdmin")}</SelectItem></SelectContent></Select>
            <Select><SelectTrigger className="w-40"><SelectValue placeholder={t("filter.subStatusPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.subStatusAll")}</SelectItem><SelectItem value="active">{t("filter.subStatusActive")}</SelectItem><SelectItem value="grace">{t("filter.subStatusGrace")}</SelectItem><SelectItem value="dunning">{t("filter.subStatusDunning")}</SelectItem><SelectItem value="expired">{t("filter.subStatusExpired")}</SelectItem></SelectContent></Select>
            <Input placeholder={t("filter.ltvMin")} className="w-28" />
            <Input placeholder={t("filter.ltvMax")} className="w-28" />
          </div>
          <div className="flex gap-2">
            <Button variant="destructive" size="sm">{t("bulkCancel")}</Button>
            <Button variant="outline" size="sm">{t("bulkGrace")}</Button>
          </div>

          {/* Table */}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"><input type="checkbox" /></TableHead>
                <TableHead>{t("table.email")}</TableHead>
                <TableHead>{t("table.platform")}</TableHead>
                <TableHead>{t("table.ltv")}</TableHead>
                <TableHead>{t("table.status")}</TableHead>
                <TableHead>{t("table.role")}</TableHead>
                <TableHead>{t("table.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((u) => (
                <TableRow key={u.id}>
                  <TableCell><input type="checkbox" /></TableCell>
                  <TableCell className="font-medium">{u.email}</TableCell>
                  <TableCell>{u.platform}</TableCell>
                  <TableCell>{u.ltv}</TableCell>
                  <TableCell><Badge className={statusClassMap[u.status]}>{t(`status.${u.status}`)}</Badge></TableCell>
                  <TableCell><Badge variant="outline">{u.role}</Badge></TableCell>
                  <TableCell><Link href={`/dashboard/users/${u.id}`} className="text-primary text-sm hover:underline">{t("viewProfile")}</Link></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground">← 1  2  3 ... 576 →  &nbsp; {t("pagination")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
