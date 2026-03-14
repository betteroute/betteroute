import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { LoadingButton } from "@/components/shared/loading-button";
import {
  Field,
  FieldError,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { OAuthButtons } from "@/features/auth/components/oauth-buttons";
import { sendMagicLink } from "@/features/auth/queries";
import { magicLinkSchema } from "@/features/auth/schemas";
import { getFieldErrors, resolveFieldErrors } from "@/lib/errors";

export const Route = createFileRoute("/_auth/login")({
  component: LoginPage,
});

function LoginPage() {
  const [success, setSuccess] = useState(false);
  const [submittedEmail, setSubmittedEmail] = useState("");

  const mutation = useMutation({
    mutationFn: sendMagicLink,
    onSuccess: (_, variables) => {
      setSubmittedEmail(variables.email);
      setSuccess(true);
    },
  });

  const form = useForm({
    defaultValues: { email: "" },
    validators: {
      onSubmit: magicLinkSchema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync(value);
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  if (success) {
    return (
      <div className="flex flex-col gap-6 text-center">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">
            Please check your email
          </h1>
          <p className="text-muted-foreground text-sm">
            We sent a magic link to <br />
            <span className="text-foreground font-medium">
              {submittedEmail}
            </span>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome to Betteroute
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Please log in to continue
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
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.email
                }
                className="h-9"
              />
              <FieldError
                errors={resolveFieldErrors(
                  field.state.meta.errors,
                  serverErrors?.email,
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
          Continue with email
        </LoadingButton>
      </form>

      <FieldSeparator>or continue with</FieldSeparator>

      <OAuthButtons disabled={mutation.isPending} />
    </div>
  );
}
