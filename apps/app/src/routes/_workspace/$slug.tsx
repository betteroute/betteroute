import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
} from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/features/workspace/components/sidebar";
import { useWorkspace } from "@/features/workspace/hooks";
import type { SessionContext } from "@/lib/session";
import { resolveDefaultWorkspace, setLastWorkspace } from "@/lib/session";

export const Route = createFileRoute("/_workspace/$slug")({
  ssr: false,
  beforeLoad: async ({ context, params }) => {
    // workspaces already fetched by parent _workspace.tsx via ensureSession()
    const { workspaces } = context as SessionContext;
    const workspace = workspaces.find((w) => w.slug === params.slug);

    if (!workspace) {
      const slug = resolveDefaultWorkspace(workspaces);
      if (slug) throw redirect({ to: "/$slug", params: { slug } });
      throw redirect({ to: "/onboarding" });
    }

    setLastWorkspace(workspace.slug);
    return { workspace };
  },
  component: WorkspaceShell,
});

function WorkspaceShell() {
  const { workspace } = useWorkspace();
  const defaultOpen =
    document.cookie.includes("sidebar_state=true") ||
    !document.cookie.includes("sidebar_state=");

  if (workspace.status === "suspended") {
    return (
      <div className="bg-background flex min-h-svh flex-col items-center justify-center p-4 text-center">
        <div className="bg-destructive/10 text-destructive mb-6 flex size-12 items-center justify-center rounded-full">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={2}
            stroke="currentColor"
            className="size-6"
            role="img"
          >
            <title>Suspended icon</title>
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 12.752zM12 15.75h.007v.008H12v-.008z"
            />
          </svg>
        </div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Workspace Suspended
        </h1>
        <p className="text-muted-foreground mt-2 max-w-sm text-sm">
          Your workspace has been suspended due to a billing issue. Please
          resolve your payment method to restore access.
        </p>
        <div className="mt-8 flex gap-3">
          <Button variant="outline" onClick={() => window.location.reload()}>
            Refresh
          </Button>
          <Button asChild>
            <Link
              to="/$slug/settings/billing"
              params={{ slug: workspace.slug }}
            >
              Resolve Billing
            </Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      <AppSidebar />
      <SidebarInset>
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  );
}
