import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function UserProfilePage({ params }: { params: { id: string } }) {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <Link href="/dashboard/users" className="text-sm text-muted-foreground hover:text-foreground">← Back to Users</Link>
        <span className="text-muted-foreground">/</span>
        <span className="text-sm font-medium">User 360° Profile: {params.id}</span>
      </div>

      {/* Identity + LTV row */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-sm">Identity Card</CardTitle><p className="text-xs text-muted-foreground">users table</p></CardHeader>
          <CardContent className="grid grid-cols-2 gap-y-2 text-sm">
            <span className="text-muted-foreground">Platform ID</span><span>usr_apple_123abc</span>
            <span className="text-muted-foreground">Device ID</span><span>dev_iphone14_xyz</span>
            <span className="text-muted-foreground">Platform</span><span><Badge variant="outline">iOS</Badge></span>
            <span className="text-muted-foreground">App Version</span><span>3.2.1</span>
            <span className="text-muted-foreground">Email</span><span>alice@example.com</span>
            <span className="text-muted-foreground">Role</span><span><Badge variant="outline">user</Badge></span>
            <span className="text-muted-foreground">Created</span><span>2024-01-15</span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm">LTV & Metrics</CardTitle><p className="text-xs text-muted-foreground">bandit_user_contexts</p></CardHeader>
          <CardContent className="grid grid-cols-2 gap-y-2 text-sm">
            <span className="text-muted-foreground">Lifetime Value</span><span className="font-bold text-lg">$184.20</span>
            <span className="text-muted-foreground">Total Payments</span><span>7</span>
            <span className="text-muted-foreground">Last Payment</span><span>2026-03-01</span>
            <span className="text-muted-foreground">Churn Risk</span><span><Badge className="bg-green-100 text-green-800">Low</Badge></span>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="subscription">
        <TabsList>
          <TabsTrigger value="subscription">Subscription</TabsTrigger>
          <TabsTrigger value="billing">Billing History</TabsTrigger>
          <TabsTrigger value="experiments">Experiments</TabsTrigger>
          <TabsTrigger value="grace">Grace & Winback</TabsTrigger>
          <TabsTrigger value="audit">Audit</TabsTrigger>
        </TabsList>
        <TabsContent value="subscription" className="space-y-4 mt-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Current Subscription</CardTitle></CardHeader>
            <CardContent className="grid grid-cols-2 gap-y-2 text-sm lg:grid-cols-4">
              <span className="text-muted-foreground">Plan</span><span className="font-medium">Pro Annual</span>
              <span className="text-muted-foreground">Status</span><span><Badge className="bg-green-100 text-green-800">✅ Active</Badge></span>
              <span className="text-muted-foreground">Expires</span><span>2027-03-01</span>
              <span className="text-muted-foreground">Source</span><span>Apple IAP</span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-sm">Recent Transactions</CardTitle></CardHeader>
            <CardContent>
              <Table>
                <TableHeader><TableRow><TableHead>Date</TableHead><TableHead>Amount</TableHead><TableHead>Currency</TableHead><TableHead>Status</TableHead><TableHead>Source</TableHead></TableRow></TableHeader>
                <TableBody>
                  <TableRow><TableCell>2026-03-01</TableCell><TableCell>$49.99</TableCell><TableCell>USD</TableCell><TableCell><Badge className="bg-green-100 text-green-800">✅ Success</Badge></TableCell><TableCell>Apple IAP</TableCell></TableRow>
                  <TableRow><TableCell>2025-03-01</TableCell><TableCell>$49.99</TableCell><TableCell>USD</TableCell><TableCell><Badge className="bg-green-100 text-green-800">✅ Success</Badge></TableCell><TableCell>Apple IAP</TableCell></TableRow>
                  <TableRow><TableCell>2025-02-28</TableCell><TableCell>$49.99</TableCell><TableCell>USD</TableCell><TableCell><Badge className="bg-red-100 text-red-800">❌ Failed</Badge></TableCell><TableCell>Apple IAP</TableCell></TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="billing" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">Full billing history — 7 transactions total.</CardContent></Card></TabsContent>
        <TabsContent value="experiments" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">No active experiments for this user.</CardContent></Card></TabsContent>
        <TabsContent value="grace" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">No active grace period or winback offers.</CardContent></Card></TabsContent>
        <TabsContent value="audit" className="mt-4"><Card><CardContent className="pt-4 text-sm text-muted-foreground">No admin actions on this user yet.</CardContent></Card></TabsContent>
      </Tabs>

      {/* Action buttons */}
      <div className="flex flex-wrap gap-2 border-t pt-4">
        <Button variant="destructive" size="sm">Force Cancel</Button>
        <Button variant="outline" size="sm">Force Renew</Button>
        <Button variant="outline" size="sm">Grant Grace Period</Button>
        <Button variant="outline" size="sm">Impersonate</Button>
      </div>
    </div>
  );
}
