import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { CreateFolderDialog } from "@/features/folder/components/create-dialog";

import { FoldersEmptyState } from "@/features/folder/components/empty-state";

import { FolderRow } from "@/features/folder/components/folder-row";

import { folderQueries } from "@/features/folder/queries";
import type { Folder } from "@/features/folder/types";

import { useWorkspace } from "@/features/workspace/hooks";

export const Route = createFileRoute("/_workspace/$slug/folders")({
  component: FoldersPage,
});

function FoldersPage() {
  const { workspace } = useWorkspace();

  const [editFolder, setEditFolder] = useState<Folder | null>(null);

  const query = useQuery(folderQueries.list(workspace.slug));

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

      <div className="flex flex-1 flex-col p-4">
        {query.isLoading ? (
          <PageLoader />
        ) : query.data && query.data.length > 0 ? (
          <div className="rounded-lg border">
            {query.data.map((folder) => (
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
