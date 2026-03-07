import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
export const Route = createFileRoute("/_workspace/$slug/settings/webhooks")({
  staticData: { title: "Webhooks" },
  component: WebhooksSettingsPage,
});

function WebhooksSettingsPage() {
  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Webhooks</h1>
          <p className="text-sm text-muted-foreground">
            Receive real-time notifications when events happen.
          </p>
        </div>
        <Button>Add Webhook</Button>
      </div>
    </div>
  );
}
