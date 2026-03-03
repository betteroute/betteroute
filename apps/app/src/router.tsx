import { createRouter as createTanStackRouter } from "@tanstack/react-router";
import { getContext } from "./integrations/tanstack-query/root-provider";
import { routeTree } from "./routeTree.gen";

export function getRouter() {
  const router = createTanStackRouter({
    routeTree,

    context: getContext(),

    scrollRestoration: true,
    defaultPreload: "intent",
    defaultPreloadStaleTime: 0,
    defaultNotFoundComponent: () => (
      <div className="flex min-h-svh items-center justify-center">
        <p className="text-muted-foreground text-sm">Page not found</p>
      </div>
    ),
  });

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }
}
