import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Copy,
  ExternalLink,
  MoreVertical,
  Pencil,
  Power,
  QrCode,
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
import { deleteLink, linkKeys } from "@/features/link/queries";
import type { Link } from "@/features/link/types";
import { useWorkspace } from "@/features/workspace/hooks";
import { copyToClipboard } from "@/lib/clipboard";

export function LinkActions({ link }: { link: Link }) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [menuOpen, setMenuOpen] = useState(false);

  const deleteMutation = useMutation({
    mutationFn: () => deleteLink(workspace.slug, link.id),
    onSuccess: async () => {
      await queryClient.refetchQueries({ queryKey: linkKeys.all });
    },
  });

  return (
    <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon-sm">
          <MoreVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="min-w-48"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <DropdownMenuItem
          onClick={() => copyToClipboard(link.short_url || link.short_code)}
        >
          <Copy />
          Copy short URL
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => window.open(link.dest_url, "_blank")}>
          <ExternalLink />
          Open destination
        </DropdownMenuItem>
        <DropdownMenuItem>
          <QrCode />
          QR Code
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Pencil />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Power />
          {link.is_active ? "Disable" : "Enable"}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <ConfirmDialog
          trigger={
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onSelect={(e) => e.preventDefault()}
            >
              <Trash2 />
              Delete
            </DropdownMenuItem>
          }
          title="Delete link"
          description={`This will permanently delete the short link "${link.short_code}". Existing redirects will stop working immediately.`}
          onConfirm={() => deleteMutation.mutateAsync()}
          pending={deleteMutation.isPending}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
