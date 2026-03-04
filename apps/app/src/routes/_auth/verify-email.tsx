import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { CheckCircle2, Mail, XCircle } from "lucide-react";
import { z } from "zod";

import { LoadingButton } from "@/components/shared/loading-button";

import {
  authQueries,
  resendVerification,
  verifyEmail,
} from "@/features/auth/queries";
import { isApiError } from "@/lib/api";

const searchSchema = z.object({
  token: z.string().catch(""),
  email: z.string().catch(""),
});

export const Route = createFileRoute("/_auth/verify-email")({
  validateSearch: searchSchema,
  ssr: false,
  loaderDeps: ({ search }) => ({ token: search.token }),
  async loader({ context, deps }) {
    if (!deps.token) return { status: "no-token" as const };
    try {
      await verifyEmail(deps.token);
      context.queryClient.invalidateQueries(authQueries.session());
      return { status: "success" as const };
    } catch (error) {
      return {
        status: "error" as const,
        detail: isApiError(error)
          ? error.apiError.detail
          : "Something went wrong",
      };
    }
  },
  component: VerifyEmailPage,
});

function VerifyEmailPage() {
  const { email } = Route.useSearch();
  const { status, ...rest } = Route.useLoaderData();
  const detail = "detail" in rest ? rest.detail : undefined;

  const resend = useMutation({
    mutationFn: () => resendVerification(email),
  });

  if (status === "no-token") {
    return (
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="bg-muted flex size-12 items-center justify-center rounded-full">
          <Mail className="text-muted-foreground size-5" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Verify your email
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            Check your inbox for a verification link.
          </p>
        </div>
        {email && (
          <LoadingButton
            variant="outline"
            size="lg"
            className="h-10"
            loading={resend.isPending}
            disabled={resend.isSuccess}
            onClick={() => resend.mutate()}
          >
            {resend.isSuccess ? "Email sent!" : "Resend verification email"}
          </LoadingButton>
        )}
        <Link
          to="/login"
          className="text-muted-foreground hover:text-foreground text-sm transition-colors"
        >
          Back to login
        </Link>
      </div>
    );
  }

  if (status === "error") {
    return (
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="bg-destructive/10 flex size-12 items-center justify-center rounded-full">
          <XCircle className="text-destructive size-5" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Verification failed
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm">{detail}</p>
        </div>
        <Link
          to="/login"
          className="text-foreground text-sm font-medium hover:underline"
        >
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-6 text-center">
      <div className="bg-emerald-500/10 flex size-12 items-center justify-center rounded-full">
        <CheckCircle2 className="size-5 text-emerald-600 dark:text-emerald-400" />
      </div>
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Email verified
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Your email has been verified successfully.
        </p>
      </div>
      <Link
        to="/login"
        className="text-foreground text-sm font-medium hover:underline"
      >
        Continue to login
      </Link>
    </div>
  );
}
