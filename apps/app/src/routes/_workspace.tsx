import { createFileRoute } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { ensureSession } from "@/lib/session";

export const Route = createFileRoute("/_workspace")({
  ssr: false,
  beforeLoad: async ({ context, location }) => {
    return ensureSession(context.queryClient, location.href);
  },
  pendingComponent: () => (
    <div className="flex min-h-svh items-center justify-center">
      <Loader2 className="text-muted-foreground size-6 animate-spin" />
    </div>
  ),
});
