import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_workspace/$slug/settings/billing")({
  staticData: { title: "Billing" },
  component: BillingSettingsPage,
});

function BillingSettingsPage() {
  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-lg font-semibold">Billing</h1>
        <p className="text-sm text-muted-foreground">
          Manage your subscription, plan, and usage.
        </p>
      </div>
    </div>
  );
}
