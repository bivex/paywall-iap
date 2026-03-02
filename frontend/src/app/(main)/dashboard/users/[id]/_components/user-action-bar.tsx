"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { forceCancelAction, forceRenewAction, grantGraceAction } from "@/actions/user-actions";

interface Props {
  userId: string;
  hasActiveSub: boolean;
}

function useAction() {
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  const run = (fn: () => Promise<{ ok: boolean; error?: string }>) => {
    setError(null);
    startTransition(async () => {
      const result = await fn();
      if (!result.ok) {
        setError(result.error ?? "Unknown error");
      } else {
        router.refresh();
      }
    });
  };

  return { isPending, error, run, setError };
}

function ForceCancelDialog({ userId }: { userId: string }) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("admin_force_cancel");
  const { isPending, error, run } = useAction();

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="destructive" size="sm">Force Cancel</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Force Cancel Subscription</DialogTitle>
          <DialogDescription>
            Immediately cancels the active subscription. This action is logged.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Label htmlFor="cancel-reason">Reason</Label>
          <Input id="cancel-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="destructive" disabled={isPending}
            onClick={() => run(async () => {
              const r = await forceCancelAction(userId, reason);
              if (r.ok) setOpen(false);
              return r;
            })}>
            {isPending ? "Cancelling…" : "Confirm Cancel"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ForceRenewDialog({ userId }: { userId: string }) {
  const [open, setOpen] = useState(false);
  const [days, setDays] = useState(30);
  const [reason, setReason] = useState("admin_force_renew");
  const { isPending, error, run } = useAction();

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">Force Renew</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Force Renew Subscription</DialogTitle>
          <DialogDescription>
            Extends expiry by the specified days. Reactivates if subscription is expired.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label htmlFor="renew-days">Days to extend</Label>
              <Input id="renew-days" type="number" min={1} max={365} value={days}
                onChange={(e) => setDays(parseInt(e.target.value) || 30)} />
            </div>
            <div>
              <Label htmlFor="renew-reason">Reason</Label>
              <Input id="renew-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
            </div>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button disabled={isPending}
            onClick={() => run(async () => {
              const r = await forceRenewAction(userId, days, reason);
              if (r.ok) setOpen(false);
              return r;
            })}>
            {isPending ? "Renewing…" : `Extend +${days} days`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function GrantGraceDialog({ userId }: { userId: string }) {
  const [open, setOpen] = useState(false);
  const [days, setDays] = useState(7);
  const [reason, setReason] = useState("admin_grant_grace");
  const { isPending, error, run } = useAction();

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">Grant Grace Period</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Grant Grace Period</DialogTitle>
          <DialogDescription>
            Creates an active grace period. Subscription status is set to &quot;grace&quot; and access is preserved.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label htmlFor="grace-days">Grace days</Label>
              <Input id="grace-days" type="number" min={1} max={90} value={days}
                onChange={(e) => setDays(parseInt(e.target.value) || 7)} />
            </div>
            <div>
              <Label htmlFor="grace-reason">Reason</Label>
              <Input id="grace-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
            </div>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button disabled={isPending}
            onClick={() => run(async () => {
              const r = await grantGraceAction(userId, days, reason);
              if (r.ok) setOpen(false);
              return r;
            })}>
            {isPending ? "Granting…" : `Grant ${days}-day grace`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function UserActionBar({ userId, hasActiveSub }: Props) {
  return (
    <div className="flex flex-wrap gap-2 border-t pt-4">
      <ForceCancelDialog userId={userId} />
      <ForceRenewDialog userId={userId} />
      <GrantGraceDialog userId={userId} />
    </div>
  );
}
