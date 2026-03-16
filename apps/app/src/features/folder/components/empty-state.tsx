import { Folder } from "lucide-react";

import { CreateFolderDialog } from "./create-dialog";

/** Shown when the workspace has zero folders. */
export function FoldersEmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Folder data-slot="icon" className="size-6 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h3 className="text-sm font-medium">No folders yet</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create folders to organize your links.
        </p>
      </div>
      <CreateFolderDialog />
    </div>
  );
}
