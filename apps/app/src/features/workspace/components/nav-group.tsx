import { Link, useMatchRoute, useParams } from "@tanstack/react-router";
import { ChevronRight, type LucideIcon } from "lucide-react";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";

export type NavItem = {
  label: string;
  icon: LucideIcon;
  to: string;
  items?: { label: string; to: string }[];
};

export function NavGroup({
  items,
  label,
}: {
  items: NavItem[];
  label?: string;
}) {
  const { slug } = useParams({ from: "/_workspace/$slug" });
  const matchRoute = useMatchRoute();

  return (
    <SidebarGroup>
      {label && <SidebarGroupLabel>{label}</SidebarGroupLabel>}
      <SidebarMenu>
        {items.map((item) => {
          const isActive = !!matchRoute({
            to: item.to,
            params: { slug },
          });

          if (!item.items?.length) {
            return (
              <SidebarMenuItem key={item.label}>
                <SidebarMenuButton
                  asChild
                  isActive={isActive}
                  tooltip={item.label}
                >
                  <Link to={item.to} params={{ slug }}>
                    <item.icon data-slot="icon" />
                    <span>{item.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            );
          }

          return (
            <Collapsible
              key={item.label}
              asChild
              defaultOpen={isActive}
              className="group/collapsible"
            >
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton isActive={isActive} tooltip={item.label}>
                    <item.icon data-slot="icon" />
                    <span>{item.label}</span>
                    <ChevronRight
                      data-slot="icon"
                      className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
                    />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {item.items.map((sub) => (
                      <SidebarMenuSubItem key={sub.label}>
                        <SidebarMenuSubButton
                          asChild
                          isActive={
                            !!matchRoute({
                              to: sub.to,
                              params: { slug },
                            })
                          }
                        >
                          <Link to={sub.to} params={{ slug }}>
                            <span>{sub.label}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
}
