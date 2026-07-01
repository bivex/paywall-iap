"use client";

import { useEffect, useState } from "react";
import { Pencil, Plus, Smartphone, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { type App, useAppStore } from "@/stores/app-store";

type FormState = {
  name: string;
  display_name: string;
  bundle_id: string;
  platform: string;
  is_active: boolean;
};

type FormErrors = Partial<Record<keyof FormState, string>>;

const emptyForm = (): FormState => ({
  name: "",
  display_name: "",
  bundle_id: "",
  platform: "ios",
  is_active: true,
});

const BUNDLE_RE = /^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z][a-zA-Z0-9]*)+$/;

function validateForm(form: FormState, isEdit: boolean): FormErrors {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = "Name is required.";
  } else if (!BUNDLE_RE.test(form.name.trim())) {
    errors.name = "Use reverse-DNS format, e.g. com.mothsalt.game1";
  } else if (form.name.length > 128) {
    errors.name = "Max 128 characters.";
  }

  if (!form.bundle_id.trim()) {
    errors.bundle_id = "Bundle ID is required.";
  } else if (!BUNDLE_RE.test(form.bundle_id.trim())) {
    errors.bundle_id = "Use reverse-DNS format, e.g. com.mothsalt.game1";
  } else if (form.bundle_id.length > 256) {
    errors.bundle_id = "Max 256 characters.";
  }

  if (form.display_name.length > 128) {
    errors.display_name = "Max 128 characters.";
  }

  if (!isEdit && !["ios", "android", "both"].includes(form.platform)) {
    errors.platform = "Select a platform.";
  }

  return errors;
}

export function AppsPageClient() {
  const { apps, setApps } = useAppStore();
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editApp, setEditApp] = useState<App | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm());
  const [saving, setSaving] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [errors, setErrors] = useState<FormErrors>({});

  useEffect(() => {
    fetch("/api/admin/apps")
      .then((r) => r.json())
      .then((body: { apps?: App[] }) => setApps(body?.apps ?? []))
      .finally(() => setLoading(false));
  }, [setApps]);

  function openCreate() {
    setEditApp(null);
    setForm(emptyForm());
    setErrors({});
    setDialogOpen(true);
  }

  function openEdit(app: App) {
    setEditApp(app);
    setForm({
      name: app.name,
      display_name: app.display_name,
      bundle_id: app.bundle_id,
      platform: app.platform,
      is_active: app.is_active,
    });
    setErrors({});
    setDialogOpen(true);
  }

  async function handleSave() {
    const errs = validateForm(form, !!editApp);
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    setErrors({});
    setSaving(true);
    try {
      if (editApp) {
        const res = await fetch(`/api/admin/apps/${editApp.id}`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(form),
        });
        if (res.ok) {
          const updated: App = await res.json();
          setApps(apps.map((a) => (a.id === updated.id ? updated : a)));
        }
      } else {
        const res = await fetch("/api/admin/apps", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(form),
        });
        if (res.ok) {
          const created: App = await res.json();
          setApps([...apps, created]);
        }
      }
      setDialogOpen(false);
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    await fetch(`/api/admin/apps/${id}`, { method: "DELETE" });
    setApps(apps.filter((a) => a.id !== id));
    setDeleteId(null);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Smartphone className="size-5 text-muted-foreground" />
          <h1 className="text-xl font-semibold">Apps</h1>
          <Badge variant="secondary">{apps.length}</Badge>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus className="size-4 mr-1" />
          Add app
        </Button>
      </div>

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : apps.length === 0 ? (
        <p className="text-sm text-muted-foreground">No apps yet. Add your first app.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Bundle ID</TableHead>
              <TableHead>Platform</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-20" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {apps.map((app) => (
              <TableRow key={app.id}>
                <TableCell>
                  <p className="font-medium">{app.display_name || app.name}</p>
                  <p className="text-xs text-muted-foreground">{app.name}</p>
                </TableCell>
                <TableCell className="font-mono text-sm">{app.bundle_id}</TableCell>
                <TableCell>
                  <Badge variant="outline" className="capitalize">
                    {app.platform}
                  </Badge>
                </TableCell>
                <TableCell>
                  {app.is_active ? (
                    <Badge variant="default">Active</Badge>
                  ) : (
                    <Badge variant="secondary">Inactive</Badge>
                  )}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button variant="ghost" size="icon" className="size-8" onClick={() => openEdit(app)}>
                      <Pencil className="size-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-8 text-destructive hover:text-destructive"
                      onClick={() => setDeleteId(app.id)}
                    >
                      <Trash2 className="size-3.5" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      {/* Create / Edit dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editApp ? "Edit app" : "Add app"}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-2">
            <div className="grid gap-1.5">
              <Label htmlFor="name">Name (internal)</Label>
              <Input
                id="name"
                placeholder="com.mothsalt.game1"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                aria-invalid={!!errors.name}
                className={errors.name ? "border-destructive focus-visible:ring-destructive" : ""}
              />
              {errors.name && <p className="text-xs text-destructive">{errors.name}</p>}
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="display_name">Display name</Label>
              <Input
                id="display_name"
                placeholder="Mothsalt Game 1"
                value={form.display_name}
                onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                aria-invalid={!!errors.display_name}
                className={errors.display_name ? "border-destructive focus-visible:ring-destructive" : ""}
              />
              {errors.display_name && <p className="text-xs text-destructive">{errors.display_name}</p>}
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="bundle_id">Bundle ID / Package name</Label>
              <Input
                id="bundle_id"
                placeholder="com.mothsalt.game1"
                value={form.bundle_id}
                onChange={(e) => setForm({ ...form, bundle_id: e.target.value })}
                aria-invalid={!!errors.bundle_id}
                className={errors.bundle_id ? "border-destructive focus-visible:ring-destructive" : ""}
              />
              {errors.bundle_id && <p className="text-xs text-destructive">{errors.bundle_id}</p>}
            </div>
            <div className="grid gap-1.5">
              <Label>Platform</Label>
              <Select
                value={form.platform}
                onValueChange={(v) => setForm({ ...form, platform: v })}
              >
                <SelectTrigger className={errors.platform ? "border-destructive" : ""}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ios">iOS</SelectItem>
                  <SelectItem value="android">Android</SelectItem>
                  <SelectItem value="both">Both</SelectItem>
                </SelectContent>
              </Select>
              {errors.platform && <p className="text-xs text-destructive">{errors.platform}</p>}
            </div>
            {editApp && (
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={form.is_active}
                  onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                  className="size-4 rounded"
                />
                <Label htmlFor="is_active">Active</Label>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? "Saving…" : editApp ? "Save changes" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <Dialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete app?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            This will deactivate the app. Existing data is preserved.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteId && handleDelete(deleteId)}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
