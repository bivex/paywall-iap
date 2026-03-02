import { Suspense } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { getUsers } from "@/actions/users";
import { UsersFilters } from "./_components/users-filters";

const subStatusMeta: Record<string, { label: string; className: string }> = {
  active:       { label: "Active",       className: "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400" },
  grace_period: { label: "Grace",        className: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400" },
  dunning:      { label: "Dunning",      className: "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400" },
  expired:      { label: "Expired",      className: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400" },
  cancelled:    { label: "Cancelled",    className: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400" },
  none:         { label: "No Sub",       className: "bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400" },
};

const platformMeta: Record<string, { label: string; className: string }> = {
  ios:     { label: "iOS",     className: "bg-blue-100 text-blue-700" },
  android: { label: "Android", className: "bg-green-100 text-green-700" },
  web:     { label: "Web",     className: "bg-violet-100 text-violet-700" },
};

interface SearchParams { page?: string; search?: string; platform?: string; role?: string; }

export default async function UsersPage({ searchParams }: { searchParams: Promise<SearchParams> }) {
  const sp = await searchParams;
  const page = Math.max(1, parseInt(sp.page ?? "1", 10));

  const data = await getUsers({
    page,
    limit: 20,
    search:   sp.search,
    platform: sp.platform,
    role:     sp.role,
  });

  const buildPageUrl = (p: number) => {
    const params = new URLSearchParams();
    if (sp.search)   params.set("search",   sp.search);
    if (sp.platform) params.set("platform", sp.platform);
    if (sp.role)     params.set("role",     sp.role);
    params.set("page", String(p));
    return `/dashboard/users?${params.toString()}`;
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Users</h1>
        <div className="flex gap-2 text-sm text-muted-foreground items-center">
          <span>{data.total} total</span>
        </div>
      </div>

      <Card>
        <CardContent className="pt-4 space-y-4">
          <Suspense fallback={null}>
            <UsersFilters />
          </Suspense>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead className="w-24">Platform</TableHead>
                <TableHead className="w-28">Sub Status</TableHead>
                <TableHead className="w-24 text-right">LTV</TableHead>
                <TableHead className="w-24">Role</TableHead>
                <TableHead className="w-36">Joined</TableHead>
                <TableHead className="w-20"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.users.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground py-10">
                    No users found.
                  </TableCell>
                </TableRow>
              ) : (
                data.users.map((u) => {
                  const sub = subStatusMeta[u.sub_status] ?? subStatusMeta.none;
                  const plat = platformMeta[u.platform] ?? { label: u.platform, className: "" };
                  return (
                    <TableRow key={u.id}>
                      <TableCell className="font-medium">{u.email}</TableCell>
                      <TableCell>
                        <Badge className={plat.className}>{plat.label}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge className={sub.className}>{sub.label}</Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        ${u.ltv.toFixed(2)}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-xs">{u.role}</Badge>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {new Date(u.created_at).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
                      </TableCell>
                      <TableCell>
                        <Button variant="ghost" size="sm" asChild>
                          <Link href={`/dashboard/users/${u.id}`}>View →</Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          <div className="flex items-center justify-between pt-1">
            <span className="text-xs text-muted-foreground">
              {data.total} users{data.total_pages > 1 && ` · Page ${data.page} of ${data.total_pages}`}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page <= 1} asChild={page > 1}>
                {page > 1 ? <Link href={buildPageUrl(page - 1)}>← Prev</Link> : <span>← Prev</span>}
              </Button>
              <Button variant="outline" size="sm" disabled={page >= data.total_pages} asChild={page < data.total_pages}>
                {page < data.total_pages ? <Link href={buildPageUrl(page + 1)}>Next →</Link> : <span>Next →</span>}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

