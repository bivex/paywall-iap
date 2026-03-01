"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { sidebarItems } from "@/navigation/sidebar/sidebar-items";

// Flat map: url → { title, groupLabel }
const routeMap: Record<string, { title: string; groupLabel?: string }> = {};
for (const group of sidebarItems) {
  for (const item of group.items) {
    routeMap[item.url] = { title: item.title, groupLabel: group.label };
    if (item.subItems) {
      for (const sub of item.subItems) {
        routeMap[sub.url] = { title: sub.title, groupLabel: item.title };
      }
    }
  }
}

export function DashboardBreadcrumb() {
  const pathname = usePathname();

  const matched = routeMap[pathname];

  // Always show: Admin > [Group] > Page  (or just Admin > Page for top-level)
  const crumbs: { label: string; href?: string }[] = [
    { label: "Admin", href: "/dashboard/default" },
  ];

  if (matched) {
    if (matched.groupLabel && matched.groupLabel !== "Monitoring") {
      crumbs.push({ label: matched.groupLabel });
    }
    crumbs.push({ label: matched.title });
  } else {
    // Fallback: humanise last path segment
    const segment = pathname.split("/").filter(Boolean).pop() ?? "";
    crumbs.push({ label: segment.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()) });
  }

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {crumbs.map((crumb, i) => {
          const isLast = i === crumbs.length - 1;
          return (
            <span key={i} className="flex items-center gap-1.5">
              {i > 0 && <BreadcrumbSeparator />}
              <BreadcrumbItem>
                {isLast || !crumb.href ? (
                  <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                ) : (
                  <BreadcrumbLink asChild>
                    <Link href={crumb.href}>{crumb.label}</Link>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            </span>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
