import { Link2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { CreateLinkDialog } from "./create-dialog";

/** Shown when the workspace has zero links. */
export function LinksEmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Link2 data-slot="icon" className="size-6 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h3 className="text-sm font-medium">No links yet</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create your first short link to get started.
        </p>
      </div>
      <CreateLinkDialog />
    </div>
  );
}

/** Shown when filters return no results. */
export function LinksFilteredEmptyState({ onClear }: { onClear: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Link2 data-slot="icon" className="size-6 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h3 className="text-sm font-medium">No links match your filters</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Try adjusting your search or filters.
        </p>
      </div>
      <Button variant="outline" onClick={onClear}>
        Clear filters
      </Button>
    </div>
  );
}
