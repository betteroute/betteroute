import { createFileRoute, redirect } from "@tanstack/react-router";
import { requireAuth } from "@/lib/auth-guard";

export const Route = createFileRoute("/")({
  ssr: false,
  beforeLoad: async ({ context }) => {
    const { workspaces } = await requireAuth(context.queryClient);

    if (!workspaces.length) {
      throw redirect({ to: "/onboarding" });
    }

    throw redirect({
      to: "/$slug",
      params: { slug: workspaces[0]!.slug },
    });
  },
});
