import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/features/workspace/components/sidebar";
import { resolveDefaultWorkspace, setLastWorkspace } from "@/lib/session";

export const Route = createFileRoute("/_workspace/$slug")({
  ssr: false,
  beforeLoad: ({ context, params }) => {
    const workspace = context.workspaces.find((w) => w.slug === params.slug);

    if (!workspace) {
      const slug = resolveDefaultWorkspace(context.workspaces);
      if (slug) throw redirect({ to: "/$slug", params: { slug } });
      throw redirect({ to: "/onboarding" });
    }

    setLastWorkspace(workspace.slug);
    return { workspace };
  },
  component: WorkspaceShell,
});

function WorkspaceShell() {
  const defaultOpen =
    document.cookie.includes("sidebar_state=true") ||
    !document.cookie.includes("sidebar_state=");

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      <AppSidebar />
      <SidebarInset>
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  );
}
