import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Folder as FolderIcon,
  MoreVertical,
  Pencil,
  Trash2,
} from "lucide-react";
import { useState } from "react";

import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useWorkspace } from "@/features/workspace/hooks";
import { timeAgo } from "@/lib/relative-time";
import { deleteFolder, folderKeys } from "../queries";
import type { Folder } from "../types";

interface FolderRowProps {
  folder: Folder;
  onEditClick: (folder: Folder) => void;
}

export function FolderRow({ folder, onEditClick }: FolderRowProps) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const deleteMutation = useMutation({
    mutationFn: () => deleteFolder(workspace.slug, folder.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: folderKeys.list(workspace.slug),
      });
    },
  });

  return (
    <div className="flex items-center gap-4 border-b px-4 py-3 transition-colors last:border-b-0 hover:bg-muted/50">
      {/* Color icon */}
      <div
        className="shrink-0 rounded-md border p-1.5"
        style={{
          backgroundColor: `${folder.color}20`,
          borderColor: `${folder.color}40`,
        }}
      >
        <FolderIcon
          data-slot="icon"
          className="size-3"
          style={{ color: folder.color }}
        />
      </div>

      {/* Folder name */}
      <div className="min-w-0 flex-1">
        <span className="text-sm font-medium">{folder.name}</span>
      </div>

      {/* Created date */}
      <span className="text-muted-foreground shrink-0 text-xs">
        {timeAgo(folder.created_at)}
      </span>

      {/* Actions */}
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="size-8">
            <MoreVertical data-slot="icon" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onClick={() => {
              setOpen(false);
              onEditClick(folder);
            }}
          >
            <Pencil data-slot="icon" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <ConfirmDialog
            trigger={
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onSelect={(e) => e.preventDefault()}
              >
                <Trash2 data-slot="icon" />
                Delete
              </DropdownMenuItem>
            }
            title="Delete folder"
            description={`This will permanently delete the folder "${folder.name}". Links inside will be moved to the root level.`}
            onConfirm={() => deleteMutation.mutateAsync()}
            pending={deleteMutation.isPending}
          />
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
