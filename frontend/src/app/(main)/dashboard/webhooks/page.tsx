import { AlertCircle } from "lucide-react";
import { getWebhooks } from "@/actions/webhooks";
import type { WebhooksParams } from "@/actions/webhooks";
import { WebhookEventsTable } from "./_components/webhook-events-table";

const PAGE_SIZE = 20;

interface Props {
  searchParams: Promise<Record<string, string | undefined>>;
}

export default async function WebhooksPage({ searchParams }: Props) {
  const sp = await searchParams;
  const page = Math.max(1, parseInt(sp.page ?? "1", 10) || 1);

  const params: WebhooksParams = {
    page, limit: PAGE_SIZE,
    provider: sp.provider,
    status: sp.status,
    search: sp.search,
    date_from: sp.date_from,
    date_to: sp.date_to,
  };

  const data = await getWebhooks(params);

  if (!data) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
        <AlertCircle className="h-8 w-8" />
        <p className="text-sm">Failed to load — make sure you are logged in.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold mb-1">Webhook Events</h1>
        <p className="text-muted-foreground">Monitor and manage incoming webhook events from all providers</p>
      </div>

      <WebhookEventsTable
        webhooks={data.webhooks}
        summary={data.summary}
        total={data.total}
        page={data.page}
        totalPages={data.total_pages}
        initialProvider={sp.provider ?? "all"}
        initialStatus={sp.status ?? "all"}
        initialSearch={sp.search ?? ""}
      />
    </div>
  );
}

