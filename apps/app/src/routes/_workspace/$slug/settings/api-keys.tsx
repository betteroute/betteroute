import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
export const Route = createFileRoute("/_workspace/$slug/settings/api-keys")({
  staticData: { title: "API Keys" },
  component: ApiKeysSettingsPage,
});

function ApiKeysSettingsPage() {
  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">API Keys</h1>
          <p className="text-sm text-muted-foreground">
            Create and manage API keys for programmatic access.
          </p>
        </div>
        <Button>Create API Key</Button>
      </div>
    </div>
  );
}
