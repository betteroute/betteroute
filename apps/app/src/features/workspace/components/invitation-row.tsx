import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Mail, MoreVertical, RotateCcw, X } from "lucide-react";

import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { UserAvatar } from "@/components/shared/user-avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { TableCell, TableRow } from "@/components/ui/table";
import { useWorkspace } from "@/features/workspace/hooks";
import { cancelInvitation, workspaceKeys } from "@/features/workspace/queries";
import type { Invitation } from "@/features/workspace/types";
import { expiresIn } from "@/lib/relative-time";
import { cn } from "@/lib/utils";

export function InvitationRow({ invitation }: { invitation: Invitation }) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();

  const cancelMutation = useMutation({
    mutationFn: () => cancelInvitation(workspace.slug, invitation.id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: workspaceKeys.invitations(workspace.slug),
      });
    },
  });

  return (
    <TableRow className="group">
      <TableCell className="py-3 sm:py-4">
        <div className="flex items-center gap-3">
          <UserAvatar
            name={invitation.email}
            fallbackIcon={
              <Mail data-slot="icon" className="size-3.5 opacity-70" />
            }
          />
          <div className="flex min-w-0 flex-col">
            <div className="flex items-center gap-2">
              <span className="truncate text-sm">{invitation.email}</span>
              <Badge
                variant="secondary"
                className="h-5 shrink-0 px-1.5 py-0 text-[10px] uppercase tracking-wider font-medium bg-amber-500/15 text-amber-700 hover:bg-amber-500/25 dark:text-amber-400"
              >
                Invited
              </Badge>
            </div>
            <span
              className={cn(
                "text-xs",
                new Date(invitation.expires_at) < new Date()
                  ? "text-destructive"
                  : "text-muted-foreground",
              )}
            >
              {expiresIn(invitation.expires_at)}
            </span>
          </div>
        </div>
      </TableCell>

      <TableCell className="text-muted-foreground capitalize">
        {invitation.role}
      </TableCell>

      <TableCell className="text-right">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon-sm">
              <MoreVertical data-slot="icon" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="min-w-48">
            <DropdownMenuItem>
              <RotateCcw data-slot="icon" />
              Resend invitation
            </DropdownMenuItem>
            <ConfirmDialog
              trigger={
                <DropdownMenuItem
                  onSelect={(e) => e.preventDefault()}
                  className="text-destructive focus:text-destructive"
                >
                  <X data-slot="icon" />
                  Cancel invitation
                </DropdownMenuItem>
              }
              title="Cancel invitation"
              description={`Revoke the invitation for ${invitation.email}? They will no longer be able to use the link.`}
              confirmLabel="Cancel invitation"
              pendingLabel="Cancelling…"
              onConfirm={() => cancelMutation.mutateAsync()}
              pending={cancelMutation.isPending}
            />
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
}
