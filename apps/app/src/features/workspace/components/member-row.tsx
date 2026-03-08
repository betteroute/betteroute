import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, LogOut, MoreVertical, SquarePen, Trash2 } from "lucide-react";

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
import { useSession, useWorkspace } from "@/features/workspace/hooks";
import {
  removeMember,
  updateMemberRole,
  workspaceKeys,
} from "@/features/workspace/queries";
import {
  ASSIGNABLE_ROLES,
  hasRole,
  type Member,
  ROLE_INFO,
} from "@/features/workspace/types";

export function MemberRow({ member }: { member: Member }) {
  const { workspace } = useWorkspace();
  const { user } = useSession();
  const queryClient = useQueryClient();

  const isCurrentUser = member.user_id === user.id;
  const isOwner = member.role === "owner";
  const canEditRole = !isOwner && !isCurrentUser;
  const canManageMembers = hasRole(workspace.role, "admin");
  const canLeaveWorkspace = isCurrentUser && !isOwner; // Current user can leave unless they're the owner
  const canRemoveMember = canManageMembers && !isCurrentUser && !isOwner; // Admins can remove non-owners who aren't themselves

  const roleMutation = useMutation({
    mutationFn: (role: string) =>
      updateMemberRole(workspace.slug, member.user_id, role),
    onSuccess: () => {
      queryClient.refetchQueries({
        queryKey: workspaceKeys.members(workspace.slug),
      });
    },
  });

  const removeMutation = useMutation({
    mutationFn: () => removeMember(workspace.slug, member.user_id),
    onSuccess: () => {
      queryClient.refetchQueries({
        queryKey: workspaceKeys.members(workspace.slug),
      });
    },
  });

  return (
    <TableRow className="group">
      <TableCell className="py-3 sm:py-4">
        <div className="flex items-center gap-3">
          <UserAvatar name={member.name} src={member.avatar_url} />
          <div className="flex min-w-0 items-center gap-2">
            <span className="truncate text-sm">{member.email}</span>
            {isCurrentUser && (
              <Badge
                variant="outline"
                className="shrink-0 text-[10px] uppercase h-5 px-1.5 py-0 tracking-wider text-muted-foreground border-border/60 bg-muted/30"
              >
                You
              </Badge>
            )}
          </div>
        </div>
      </TableCell>

      <TableCell>
        {canEditRole ? (
          <DropdownMenu>
            <DropdownMenuTrigger className="group/role flex items-center gap-1.5 rounded-md px-1.5 py-1 -ml-1.5 text-sm capitalize text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground data-[state=open]:bg-muted data-[state=open]:text-foreground">
              <span>{ROLE_INFO[member.role]?.label ?? member.role}</span>
              <SquarePen className="size-3.5 text-muted-foreground/50 transition-colors group-hover/role:text-foreground" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-[140px]">
              {ASSIGNABLE_ROLES.map((role) => (
                <DropdownMenuItem
                  key={role}
                  onClick={() => roleMutation.mutate(role)}
                  className="flex items-center justify-between text-sm capitalize"
                >
                  <span>{ROLE_INFO[role]?.label ?? role}</span>
                  {role === member.role && <Check className="size-4" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <span className="flex items-center px-1.5 py-1 -ml-1.5 text-sm capitalize text-muted-foreground">
            {ROLE_INFO[member.role]?.label ?? member.role}
          </span>
        )}
      </TableCell>

      <TableCell className="text-right">
        {(canLeaveWorkspace || canRemoveMember) && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground hover:text-foreground"
              >
                <MoreVertical />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              {canLeaveWorkspace && (
                <ConfirmDialog
                  trigger={
                    <DropdownMenuItem
                      onSelect={(e) => e.preventDefault()}
                      className="text-destructive focus:text-destructive"
                    >
                      <LogOut className="size-4 mr-2" />
                      Leave workspace
                    </DropdownMenuItem>
                  }
                  title="Leave workspace"
                  description={`Leave ${workspace.name}? You will lose access to all links and settings.`}
                  confirmLabel="Leave"
                  pendingLabel="Leaving…"
                  onConfirm={() => removeMutation.mutateAsync()}
                  pending={removeMutation.isPending}
                />
              )}
              {canRemoveMember && (
                <ConfirmDialog
                  trigger={
                    <DropdownMenuItem
                      onSelect={(e) => e.preventDefault()}
                      className="text-destructive focus:text-destructive cursor-pointer"
                    >
                      <Trash2 className="size-4 mr-2" />
                      Remove member
                    </DropdownMenuItem>
                  }
                  title="Remove member"
                  description={`Remove ${member.name} from this workspace? They will lose access immediately.`}
                  confirmLabel="Remove"
                  pendingLabel="Removing…"
                  onConfirm={() => removeMutation.mutateAsync()}
                  pending={removeMutation.isPending}
                />
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </TableCell>
    </TableRow>
  );
}
