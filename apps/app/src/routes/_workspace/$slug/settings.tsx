import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
  useMatches,
} from "@tanstack/react-router";
import { PageHeader } from "@/components/shared/page-header";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { hasRole } from "@/features/workspace/types";

export const Route = createFileRoute("/_workspace/$slug/settings")({
  beforeLoad: ({ context }) => {
    const { workspace } = context;
    if (!hasRole(workspace.role, "admin")) {
      throw redirect({ to: "/$slug", params: { slug: workspace.slug } });
    }
  },
  component: SettingsLayout,
});

function SettingsLayout() {
  const { slug } = Route.useParams();
  const matches = useMatches();
  const deepest = matches[matches.length - 1];
  const pageTitle =
    (deepest?.staticData as { title?: string })?.title ?? "Settings";

  return (
    <div className="flex h-full flex-col">
      <PageHeader>
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink asChild>
                <Link to="/$slug/settings" params={{ slug }}>
                  Settings
                </Link>
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{pageTitle}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </PageHeader>
      <div className="flex-1 overflow-auto">
        <div className="mx-auto">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
