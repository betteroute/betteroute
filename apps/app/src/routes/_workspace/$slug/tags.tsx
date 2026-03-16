import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { z } from "zod";

import { DebouncedSearchInput } from "@/components/shared/debounced-search-input";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { CreateTagDialog } from "@/features/tag/components/create-dialog";

import { TagsEmptyState } from "@/features/tag/components/empty-state";

import { TagRow } from "@/features/tag/components/tag-row";

import { tagQueries } from "@/features/tag/queries";
import type { Tag } from "@/features/tag/types";

import { useWorkspace } from "@/features/workspace/hooks";

const tagSearchSchema = z.object({
  search: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_workspace/$slug/tags")({
  validateSearch: tagSearchSchema,
  loader: async ({ context, params }) => {
    await context.queryClient.ensureQueryData(tagQueries.list(params.slug));
  },
  pendingComponent: PageLoader,
  component: TagsPage,
});

function TagsPage() {
  const { workspace } = useWorkspace();
  const navigate = Route.useNavigate();
  const searchParams = Route.useSearch();

  const [editTag, setEditTag] = useState<Tag | null>(null);

  const query = useSuspenseQuery(tagQueries.list(workspace.slug));

  const filteredTags = useMemo(() => {
    const searchLower = (searchParams.search || "").toLowerCase();
    return query.data.filter((tag) => {
      return (
        !searchParams.search || tag.name.toLowerCase().includes(searchLower)
      );
    });
  }, [query.data, searchParams.search]);

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

      <div className="flex flex-1 flex-col gap-4 p-4">
        <div className="flex items-center gap-2">
          <DebouncedSearchInput
            value={searchParams.search ?? ""}
            onChange={(value) =>
              navigate({
                search: (prev) => ({
                  ...prev,
                  search: value || undefined,
                }),
                replace: true,
              })
            }
            placeholder="Search tags…"
          />
        </div>
        {filteredTags.length > 0 ? (
          <div className="rounded-lg border">
            {filteredTags.map((tag) => (
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
          <TagsEmptyState />
        )}
      </div>
    </>
  );
}
