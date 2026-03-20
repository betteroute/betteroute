import { Link, useMatchRoute } from "@tanstack/react-router";
import {
  ArrowLeft,
  BarChart3,
  Bell,
  CircleHelp,
  FolderIcon,
  Globe,
  Key,
  LinkIcon,
  Settings2,
  TagIcon,
  Users,
  Wallet,
  Webhook,
} from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

import { useWorkspace } from "../hooks";
import { NavGroup, type NavItem } from "./nav-group";
import { NavUser } from "./nav-user";
import { WorkspaceSwitcher } from "./switcher";

const manageNav: NavItem[] = [
  { label: "Links", icon: LinkIcon, to: "/$slug/links" },
  { label: "Analytics", icon: BarChart3, to: "/$slug/analytics" },
  { label: "Domains", icon: Globe, to: "/$slug/domains" },
];

const organizeNav: NavItem[] = [
  { label: "Folders", icon: FolderIcon, to: "/$slug/folders" },
  { label: "Tags", icon: TagIcon, to: "/$slug/tags" },
];

const footerNav: NavItem[] = [
  { label: "Settings", icon: Settings2, to: "/$slug/settings" },
  { label: "Help", icon: CircleHelp, to: "/$slug/help" },
];

const settingsWorkspaceNav: NavItem[] = [
  { label: "General", icon: Settings2, to: "/$slug/settings" },
  { label: "Members", icon: Users, to: "/$slug/settings/members" },
  { label: "Billing", icon: Wallet, to: "/$slug/settings/billing" },
  { label: "Domains", icon: Globe, to: "/$slug/settings/domains" },
];

const settingsDeveloperNav: NavItem[] = [
  { label: "API Keys", icon: Key, to: "/$slug/settings/api-keys" },
  { label: "Webhooks", icon: Webhook, to: "/$slug/settings/webhooks" },
];

const settingsPersonalNav: NavItem[] = [
  { label: "Notifications", icon: Bell, to: "/$slug/settings/notifications" },
];

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  const { user, workspaces, workspace } = useWorkspace();
  const matchRoute = useMatchRoute();
  const isSettings = !!matchRoute({ to: "/$slug/settings", fuzzy: true });

  return (
    <Sidebar collapsible="icon" variant="inset" {...props}>
      <SidebarHeader>
        {isSettings ? (
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild tooltip="Back to workspace">
                <Link to="/$slug" params={{ slug: workspace.slug }}>
                  <Avatar className="size-7 rounded-lg after:rounded-lg">
                    <AvatarFallback className="rounded-lg bg-primary text-primary-foreground text-xs font-semibold">
                      <ArrowLeft data-slot="icon" className="size-4" />
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid min-w-0 flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">Settings</span>
                    <span className="truncate text-xs text-muted-foreground">
                      {workspace.name}
                    </span>
                  </div>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        ) : (
          <WorkspaceSwitcher workspaces={workspaces} current={workspace} />
        )}
      </SidebarHeader>

      <SidebarContent>
        {isSettings ? (
          <>
            <NavGroup items={settingsWorkspaceNav} label="Workspace" />
            <NavGroup items={settingsDeveloperNav} label="Developer" />
            <NavGroup items={settingsPersonalNav} label="Personal" />
          </>
        ) : (
          <>
            <NavGroup items={manageNav} label="Manage" />
            <NavGroup items={organizeNav} label="Organize" />
          </>
        )}
      </SidebarContent>

      <SidebarFooter>
        {!isSettings && (
          <SidebarMenu>
            {footerNav.map((item) => (
              <SidebarMenuItem key={item.label}>
                <SidebarMenuButton asChild tooltip={item.label}>
                  <Link to={item.to} params={{ slug: workspace.slug }}>
                    <item.icon data-slot="icon" />
                    <span>{item.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        )}
        <NavUser user={user} />
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}
