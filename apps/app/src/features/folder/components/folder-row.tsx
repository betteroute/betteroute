import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Folder as FolderIcon,
  MoreVertical,
  Pencil,
  Trash2,
} from "lucide-react";
import { useState } from "react";

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
        <FolderIcon className="size-3" style={{ color: folder.color }} />
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
            <MoreVertical />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onClick={() => {
              setOpen(false);
              onEditClick(folder);
            }}
          >
            <Pencil className="mr-2 size-4" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => {
              setOpen(false);
              deleteMutation.mutate();
            }}
            className="text-destructive focus:text-destructive"
            disabled={deleteMutation.isPending}
          >
            <Trash2 className="mr-2 size-4" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
