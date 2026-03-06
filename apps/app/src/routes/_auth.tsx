import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
} from "@tanstack/react-router";
import { z } from "zod";
import { Logo } from "@/components/shared/logo";
import { authQueries } from "@/features/auth/queries";

const searchSchema = z.object({
  redirect: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_auth")({
  ssr: false,
  validateSearch: searchSchema,
  beforeLoad: async ({ context }) => {
    const user = await context.queryClient
      .ensureQueryData(authQueries.session())
      .catch(() => null);
    if (user) throw redirect({ to: "/" });
  },
  component: AuthShell,
});

function AuthShell() {
  return (
    <div className="flex min-h-svh flex-col pt-16 pb-12 relative overflow-hidden">
      <main className="flex w-full flex-1 flex-col items-center justify-center px-4">
        <div className="w-full max-w-[340px]">
          <div className="mb-10 flex justify-center">
            <Logo className="size-12" />
          </div>
          <Outlet />
        </div>
      </main>

      <footer className="mt-8 px-4 text-center">
        <p className="text-muted-foreground text-sm">
          By continuing, you agree to Betteroute's{" "}
          <Link
            to="/"
            className="hover:text-foreground underline underline-offset-4 transition-colors"
          >
            Terms of Service
          </Link>{" "}
          and{" "}
          <Link
            to="/"
            className="hover:text-foreground underline underline-offset-4 transition-colors"
          >
            Privacy Policy
          </Link>
          .
        </p>
      </footer>
    </div>
  );
}
