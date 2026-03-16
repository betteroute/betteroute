import { useMutation, useQueryClient } from "@tanstack/react-query";
import { MoreVertical, Pencil, Tag as TagIcon, Trash2 } from "lucide-react";
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
import { deleteTag, tagKeys } from "../queries";
import type { Tag } from "../types";

interface TagRowProps {
  tag: Tag;
  onEditClick: (tag: Tag) => void;
}

export function TagRow({ tag, onEditClick }: TagRowProps) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const deleteMutation = useMutation({
    mutationFn: () => deleteTag(workspace.slug, tag.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: tagKeys.list(workspace.slug),
      });
    },
  });

  return (
    <div className="flex items-center gap-4 border-b px-4 py-3 transition-colors last:border-b-0 hover:bg-muted/50">
      {/* Color icon */}
      <div
        className="shrink-0 rounded-md border p-1.5"
        style={{
          backgroundColor: `${tag.color}20`,
          borderColor: `${tag.color}40`,
        }}
      >
        <TagIcon
          data-slot="icon"
          className="size-3"
          style={{ color: tag.color }}
        />
      </div>

      {/* Tag name */}
      <div className="min-w-0 flex-1">
        <span className="text-sm font-medium">{tag.name}</span>
      </div>

      {/* Created date */}
      <span className="text-muted-foreground shrink-0 text-xs">
        {timeAgo(tag.created_at)}
      </span>

      {/* Actions */}
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="size-8">
            <MoreVertical data-slot="icon" />
            <span className="sr-only">Open menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onClick={() => {
              setOpen(false);
              onEditClick(tag);
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
            title="Delete tag"
            description={`This will permanently delete the tag "${tag.name}". It will be removed from all links that use it.`}
            onConfirm={() => deleteMutation.mutateAsync()}
            pending={deleteMutation.isPending}
          />
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
