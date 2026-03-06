import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
export const Route = createFileRoute("/_workspace/$slug/folders")({
  staticData: { title: "Folders" },
  component: FoldersPage,
});

function FoldersPage() {
  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Folders</h1>
          <p className="text-sm text-muted-foreground">
            Organize your short links into folders.
          </p>
        </div>
        <Button>Add Folder</Button>
      </div>
    </div>
  );
}
