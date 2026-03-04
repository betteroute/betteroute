import { createFileRoute, redirect } from "@tanstack/react-router";
import { resolveDefaultWorkspace } from "@/lib/session";

export const Route = createFileRoute("/_workspace/")({
  beforeLoad: ({ context }) => {
    const slug = resolveDefaultWorkspace(context.workspaces);
    if (!slug) throw redirect({ to: "/onboarding" });
    throw redirect({ to: "/$slug", params: { slug } });
  },
});
