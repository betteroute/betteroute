import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, useRef } from "react";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { verifyMagicLink } from "@/features/auth/queries";
import { isApiError } from "@/lib/errors";

const searchSchema = z.object({
  token: z.string().optional().catch(undefined),
  redirect: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_auth/verify")({
  validateSearch: searchSchema,
  component: VerifyPage,
});

function VerifyPage() {
  const { token, redirect: redirectTo } = Route.useSearch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const attempted = useRef(false);

  const { mutate, isError, error } = useMutation({
    mutationFn: verifyMagicLink,
    onSuccess: async () => {
      queryClient.clear();
      await navigate({ to: redirectTo ?? "/" });
    },
  });

  useEffect(() => {
    if (token && !attempted.current) {
      attempted.current = true;
      mutate({ token });
    }
  }, [token, mutate]);

  // Missing token — broken link
  if (!token) {
    return (
      <div className="flex flex-col gap-6 text-center">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">
            Invalid link
          </h1>
          <p className="text-muted-foreground text-sm">
            This verification link is broken or malformed.
          </p>
        </div>
        <Button asChild variant="outline" size="lg" className="w-full">
          <Link to="/login">Return to login</Link>
        </Button>
      </div>
    );
  }

  // Error state — token was present but verification failed
  if (isError) {
    const detail =
      error && isApiError(error) ? error.apiError.detail : undefined;

    return (
      <div className="flex flex-col gap-6 text-center">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">
            Verification failed
          </h1>
          <p className="text-muted-foreground text-sm">
            {detail || (
              <>
                This link has expired or is no longer valid.
                <br />
                Let's get you a fresh one.
              </>
            )}
          </p>
        </div>
        <Button asChild variant="outline" size="lg" className="w-full">
          <Link to="/login">Request new link</Link>
        </Button>
      </div>
    );
  }

  // Default: loading — shown immediately on mount
  return (
    <div className="flex flex-col gap-6 text-center">
      <div className="flex justify-center">
        <Loader2 className="text-muted-foreground size-8 animate-spin" />
      </div>
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Welcome back</h1>
        <p className="text-muted-foreground text-sm">
          We're verifying your account. One moment…
        </p>
      </div>
    </div>
  );
}
