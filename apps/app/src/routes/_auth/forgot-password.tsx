import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ArrowLeft, Mail } from "lucide-react";

import { LoadingButton } from "@/components/shared/loading-button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { forgotPassword } from "@/features/auth/queries";
import { forgotPasswordSchema } from "@/features/auth/schemas";
import { resolveErrors } from "@/lib/form-errors";

export const Route = createFileRoute("/_auth/forgot-password")({
  component: ForgotPasswordPage,
});

function ForgotPasswordPage() {
  const mutation = useMutation({
    mutationFn: forgotPassword,
  });

  const form = useForm({
    defaultValues: { email: "" },
    validators: {
      onSubmit: forgotPasswordSchema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync(value);
    },
  });

  if (mutation.isSuccess) {
    return (
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="bg-muted flex size-12 items-center justify-center rounded-full">
          <Mail className="text-muted-foreground size-5" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Check your email
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            If an account exists with that email, we&apos;ve sent a password
            reset link.
          </p>
        </div>
        <Link
          to="/login"
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-1.5 text-sm transition-colors"
        >
          <ArrowLeft className="size-3.5" />
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Reset your password
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Enter your email and we&apos;ll send you a reset link.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="flex flex-col gap-4"
      >
        <form.Field name="email">
          {(field) => (
            <Field>
              <FieldLabel htmlFor={field.name}>Email</FieldLabel>
              <Input
                id={field.name}
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                autoFocus
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
              />
              <FieldError errors={resolveErrors(field.state.meta.errors)} />
            </Field>
          )}
        </form.Field>

        <LoadingButton
          type="submit"
          size="lg"
          className="mt-1 w-full"
          loading={mutation.isPending}
        >
          Send reset link
        </LoadingButton>
      </form>

      <p className="text-center">
        <Link
          to="/login"
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-1.5 text-sm transition-colors"
        >
          <ArrowLeft className="size-3.5" />
          Back to login
        </Link>
      </p>
    </div>
  );
}
