import { useNavigate } from "@tanstack/react-router";
import { ChevronsUpDown, Plus } from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { CreateWorkspaceDialog } from "@/features/workspace/components/create-dialog";
import type { WorkspaceWithRole } from "@/features/workspace/types";
import { getInitials } from "@/lib/initials";

export function WorkspaceSwitcher({
  workspaces,
  current,
}: {
  workspaces: WorkspaceWithRole[];
  current: WorkspaceWithRole;
}) {
  const { isMobile, setOpenMobile } = useSidebar();
  const navigate = useNavigate();
  const [createOpen, setCreateOpen] = useState(false);

  function switchTo(slug: string) {
    setOpenMobile(false);
    navigate({ to: "/$slug", params: { slug } });
  }

  return (
    <>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground focus-visible:ring-0"
              >
                <Avatar className="size-7 rounded-lg after:rounded-lg">
                  <AvatarFallback className="rounded-lg bg-primary text-primary-foreground text-xs font-semibold">
                    {getInitials(current.name)}
                  </AvatarFallback>
                </Avatar>
                <div className="grid min-w-0 flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{current.name}</span>
                  <span className="truncate text-xs text-muted-foreground capitalize">
                    {current.plan_id}
                  </span>
                </div>
                <ChevronsUpDown data-slot="icon" className="ml-auto" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-(--radix-dropdown-menu-trigger-width) min-w-64 rounded-lg"
              align="start"
              side={isMobile ? "bottom" : "right"}
              sideOffset={4}
            >
              <DropdownMenuLabel className="text-xs text-muted-foreground">
                Workspaces
              </DropdownMenuLabel>
              {workspaces.map((ws) => (
                <DropdownMenuItem
                  key={ws.id}
                  onClick={() => switchTo(ws.slug)}
                  className={`gap-2 ${ws.id === current.id ? "bg-accent" : ""}`}
                >
                  <Avatar className="size-6 rounded-lg after:rounded-lg">
                    <AvatarFallback className="rounded-lg border bg-background text-[10px] font-semibold">
                      {getInitials(ws.name)}
                    </AvatarFallback>
                  </Avatar>
                  <span className="min-w-0 flex-1 truncate">{ws.name}</span>
                  <span className="text-[10px] capitalize text-muted-foreground">
                    {ws.role}
                  </span>
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="gap-2"
                onClick={() => setCreateOpen(true)}
              >
                <div className="flex size-6 items-center justify-center rounded-lg border bg-background">
                  <Plus data-slot="icon" className="size-3.5" />
                </div>
                Create workspace
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>

      <CreateWorkspaceDialog open={createOpen} onOpenChange={setCreateOpen} />
    </>
  );
}
