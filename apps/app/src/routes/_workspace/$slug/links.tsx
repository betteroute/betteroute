import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
export const Route = createFileRoute("/_workspace/$slug/settings/domains")({
  staticData: { title: "Links" },
  component: LinksPage,
});

function LinksPage() {
  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Links</h1>
          <p className="text-sm text-muted-foreground">
            Manage your short links.
          </p>
        </div>
        <Button>Add Link</Button>
      </div>
    </div>
  );
}
