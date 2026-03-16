import { createFileRoute, isRedirect, redirect } from "@tanstack/react-router";
import { toast } from "sonner";
import { z } from "zod";
import { authQueries } from "@/features/auth/queries";

const searchSchema = z.object({
  error: z.string().catch(""),
  error_description: z.string().catch(""),
});

export const Route = createFileRoute("/auth/callback")({
  validateSearch: searchSchema,
  ssr: false,
  beforeLoad: async ({ context, search }) => {
    // Handle OAuth provider errors (e.g. user denied consent)
    if (search.error) {
      toast.error(
        search.error_description || "Authentication was cancelled or failed.",
      );
      throw redirect({ to: "/login" });
    }

    // Backend already set the session cookie — verify by fetching /me
    try {
      context.queryClient.clear();
      await context.queryClient.fetchQuery(authQueries.session());
      // Successful auth → go to root (which redirects to first workspace)
      throw redirect({ to: "/" });
    } catch (e) {
      if (isRedirect(e)) throw e;
      throw redirect({ to: "/login" });
    }
  },
  component: () => null,
});
