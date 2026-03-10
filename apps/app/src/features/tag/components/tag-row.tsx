import { useMutation, useQueryClient } from "@tanstack/react-query";
import { MoreVertical, Pencil, Tag as TagIcon, Trash2 } from "lucide-react";
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
        <TagIcon className="size-3" style={{ color: tag.color }} />
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
            <MoreVertical />
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
