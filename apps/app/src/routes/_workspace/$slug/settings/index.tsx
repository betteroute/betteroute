import { createFileRoute } from "@tanstack/react-router";

import { DeleteWorkspaceDialog } from "@/features/workspace/components/delete-dialog";
import { GeneralForm } from "@/features/workspace/components/general-form";

export const Route = createFileRoute("/_workspace/$slug/settings/")({
  staticData: { title: "General" },

  component: GeneralSettingsPage,
});

function GeneralSettingsPage() {
  return (
    <div className="space-y-10 p-6">
      <div>
        <h1 className="text-lg font-semibold">General</h1>

        <p className="text-muted-foreground text-sm">
          Manage your workspace name, slug, and other details.
        </p>
      </div>

      <GeneralForm />

      <DeleteWorkspaceDialog />
    </div>
  );
}
