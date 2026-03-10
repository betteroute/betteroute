import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { CreateTagDialog } from "@/features/tag/components/create-dialog";

import { EmptyState } from "@/features/tag/components/empty-state";

import { TagRow } from "@/features/tag/components/tag-row";

import { tagQueries } from "@/features/tag/queries";
import type { Tag } from "@/features/tag/types";

import { useWorkspace } from "@/features/workspace/hooks";

export const Route = createFileRoute("/_workspace/$slug/tags")({
  component: TagsPage,
});

function TagsPage() {
  const { workspace } = useWorkspace();

  const [editTag, setEditTag] = useState<Tag | null>(null);

  const query = useQuery(tagQueries.list(workspace.slug));

  return (
    <>
      <PageHeader
        title="Tags"
        actions={
          <CreateTagDialog
            editTag={editTag ?? undefined}
            onSuccess={() => {
              setEditTag(null);
            }}
          />
        }
      />

      <div className="flex flex-1 flex-col p-4">
        {query.isLoading ? (
          <PageLoader />
        ) : query.data && query.data.length > 0 ? (
          <div className="rounded-lg border">
            {query.data.map((tag) => (
              <TagRow
                key={tag.id}
                tag={tag}
                onEditClick={(t) => {
                  setEditTag(t);
                }}
              />
            ))}
          </div>
        ) : (
          <EmptyState />
        )}
      </div>
    </>
  );
}
