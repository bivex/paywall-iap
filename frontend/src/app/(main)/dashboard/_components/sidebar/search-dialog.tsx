"use client";
import * as React from "react";

import { useRouter } from "next/navigation";

import { Search, Settings2, Smartphone } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { useSidebarItems } from "@/navigation/sidebar/use-sidebar-items";
import { useAppStore } from "@/stores/app-store";

export function SearchDialog() {
  const [open, setOpen] = React.useState(false);
  const router = useRouter();
  const t = useTranslations("search");
  const sidebarGroups = useSidebarItems();
  const apps = useAppStore((s) => s.apps);

  const searchItems = React.useMemo(
    () =>
      sidebarGroups.flatMap((group) =>
        group.items.flatMap((item) => {
          if (item.comingSoon) return [];

          if (item.subItems?.length) {
            return item.subItems
              .filter((subItem) => !subItem.comingSoon)
              .map((subItem) => ({
                group: group.label ?? t("navigation"),
                icon: subItem.icon ?? item.icon,
                label: subItem.title,
                url: subItem.url,
                sublabel: undefined as string | undefined,
              }));
          }

          return [
            {
              group: group.label ?? t("navigation"),
              icon: item.icon,
              label: item.title,
              url: item.url,
              sublabel: undefined as string | undefined,
            },
          ];
        }),
      ),
    [sidebarGroups, t],
  );

  const appItems = React.useMemo(
    () =>
      apps.flatMap((app) => [
        {
          group: "Apps",
          icon: Smartphone,
          label: app.display_name || app.name,
          sublabel: app.bundle_id,
          url: `/dashboard/apps`,
        },
        {
          group: "Apps",
          icon: Settings2,
          label: `${app.display_name || app.name} — Configure`,
          sublabel: app.bundle_id,
          url: `/dashboard/apps/${app.id}/settings`,
        },
      ]),
    [apps],
  );

  const allItems = React.useMemo(() => [...searchItems, ...appItems], [searchItems, appItems]);

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "j" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const handleSelect = (url: string) => {
    setOpen(false);
    router.push(url);
  };

  return (
    <>
      <Button
        variant="link"
        className="!px-0 font-normal text-muted-foreground hover:no-underline"
        onClick={() => setOpen(true)}
      >
        <Search className="size-4" />
        Search
        <kbd className="inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-medium text-[10px]">
          <span className="text-xs">⌘</span>J
        </kbd>
      </Button>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder={t("placeholder")} />
        <CommandList>
          <CommandEmpty>{t("noResults")}</CommandEmpty>
          {[...new Set(allItems.map((item) => item.group))].map((group, i) => (
            <React.Fragment key={group}>
              {i !== 0 && <CommandSeparator />}
              <CommandGroup heading={group}>
                {allItems
                  .filter((item) => item.group === group)
                  .map((item) => (
                    <CommandItem
                      className="!py-1.5"
                      key={item.label + item.url}
                      onSelect={() => handleSelect(item.url)}
                    >
                      {item.icon && <item.icon className="size-4 shrink-0" />}
                      <span>{item.label}</span>
                      {item.sublabel && (
                        <span className="ml-auto text-xs text-muted-foreground font-mono truncate max-w-[160px]">
                          {item.sublabel}
                        </span>
                      )}
                    </CommandItem>
                  ))}
              </CommandGroup>
            </React.Fragment>
          ))}
        </CommandList>
      </CommandDialog>
    </>
  );
}
