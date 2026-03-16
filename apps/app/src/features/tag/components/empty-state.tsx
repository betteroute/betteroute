import { Tag as TagIcon } from "lucide-react";

import { CreateTagDialog } from "./create-dialog";

export function TagsEmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <TagIcon data-slot="icon" className="size-6 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h3 className="text-sm font-medium">No tags yet</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create tags to organize and categorize your links.
        </p>
      </div>
      <CreateTagDialog />
    </div>
  );
}
