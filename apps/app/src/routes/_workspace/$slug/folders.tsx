import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { z } from "zod";

import { DebouncedSearchInput } from "@/components/shared/debounced-search-input";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { CreateFolderDialog } from "@/features/folder/components/create-dialog";

import { FoldersEmptyState } from "@/features/folder/components/empty-state";

import { FolderRow } from "@/features/folder/components/folder-row";

import { folderQueries } from "@/features/folder/queries";
import type { Folder } from "@/features/folder/types";

import { useWorkspace } from "@/features/workspace/hooks";

const folderSearchSchema = z.object({
  search: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_workspace/$slug/folders")({
  validateSearch: folderSearchSchema,
  loader: async ({ context, params }) => {
    await context.queryClient.ensureQueryData(folderQueries.list(params.slug));
  },
  pendingComponent: PageLoader,
  component: FoldersPage,
});

function FoldersPage() {
  const { workspace } = useWorkspace();
  const navigate = Route.useNavigate();
  const searchParams = Route.useSearch();

  const [editFolder, setEditFolder] = useState<Folder | null>(null);

  const query = useSuspenseQuery(folderQueries.list(workspace.slug));

  const filteredFolders = useMemo(() => {
    const searchLower = (searchParams.search || "").toLowerCase();
    return query.data.filter((folder) => {
      return (
        !searchParams.search || folder.name.toLowerCase().includes(searchLower)
      );
    });
  }, [query.data, searchParams.search]);

  return (
    <>
      <PageHeader
        title="Folders"
        actions={
          <CreateFolderDialog
            editFolder={editFolder ?? undefined}
            onSuccess={() => {
              setEditFolder(null);
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
            placeholder="Search folders…"
          />
        </div>
        {filteredFolders.length > 0 ? (
          <div className="rounded-lg border">
            {filteredFolders.map((folder) => (
              <FolderRow
                key={folder.id}
                folder={folder}
                onEditClick={(f) => {
                  setEditFolder(f);
                }}
              />
            ))}
          </div>
        ) : (
          <FoldersEmptyState />
        )}
      </div>
    </>
  );
}
