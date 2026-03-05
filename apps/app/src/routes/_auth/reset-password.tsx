import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { XCircle } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";

import { LoadingButton } from "@/components/shared/loading-button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { PasswordInput } from "@/features/auth/components/password-input";
import { resetPassword } from "@/features/auth/queries";
import { resetPasswordSchema } from "@/features/auth/schemas";
import { getFieldErrors } from "@/lib/errors";
import { resolveErrors } from "@/lib/form-errors";

const searchSchema = z.object({
  token: z.string().catch(""),
});

export const Route = createFileRoute("/_auth/reset-password")({
  validateSearch: searchSchema,
  component: ResetPasswordPage,
});

function ResetPasswordPage() {
  const { token } = Route.useSearch();
  const navigate = Route.useNavigate();

  const mutation = useMutation({
    mutationFn: resetPassword,
    onSuccess: () => {
      toast.success("Password reset successfully");
      navigate({ to: "/login" });
    },
  });

  const form = useForm({
    defaultValues: { password: "" },
    validators: {
      onSubmit: resetPasswordSchema.pick({ password: true }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ ...value, token });
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  if (!token) {
    return (
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="bg-destructive/10 flex size-12 items-center justify-center rounded-full">
          <XCircle className="text-destructive size-5" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Invalid reset link
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            This link is invalid or has expired. Please request a new one.
          </p>
        </div>
        <Link
          to="/forgot-password"
          className="text-foreground text-sm font-medium hover:underline"
        >
          Request new link
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Set new password
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Enter a new password for your account.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="flex flex-col gap-4"
      >
        <form.Field name="password">
          {(field) => (
            <Field>
              <FieldLabel htmlFor={field.name}>New password</FieldLabel>
              <PasswordInput
                id={field.name}
                placeholder="••••••••"
                autoComplete="new-password"
                autoFocus
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.password
                }
              />
              <FieldError
                errors={resolveErrors(
                  field.state.meta.errors,
                  serverErrors?.password,
                )}
              />
            </Field>
          )}
        </form.Field>

        <LoadingButton
          type="submit"
          size="lg"
          className="mt-1 w-full"
          loading={mutation.isPending}
        >
          Reset password
        </LoadingButton>
      </form>
    </div>
  );
}
