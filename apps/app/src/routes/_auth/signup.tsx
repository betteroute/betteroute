import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { LoadingButton } from "@/components/shared/loading-button";
import {
  Field,
  FieldError,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { OAuthButtons } from "@/features/auth/components/oauth-buttons";
import { PasswordInput } from "@/features/auth/components/password-input";
import { authQueries, signup } from "@/features/auth/queries";
import { signupSchema } from "@/features/auth/schemas";
import { getFieldErrors } from "@/lib/errors";
import { resolveErrors } from "@/lib/form-errors";

export const Route = createFileRoute("/_auth/signup")({
  component: SignupPage,
});

function SignupPage() {
  const queryClient = useQueryClient();
  const navigate = Route.useNavigate();
  const { redirect: redirectTo } = Route.useSearch();

  const mutation = useMutation({
    mutationFn: signup,
    onSuccess: (user) => {
      queryClient.setQueryData(authQueries.session().queryKey, user);
      navigate({ href: redirectTo ?? "/" });
    },
  });

  const form = useForm({
    defaultValues: { name: "", email: "", password: "" },
    validators: {
      onSubmit: signupSchema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync(value);
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  return (
    <div className="flex flex-col gap-8">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Create your account
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Join thousands of developers managing links better.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="flex flex-col gap-4"
      >
        <form.Field name="name">
          {(field) => (
            <Field>
              <FieldLabel htmlFor={field.name}>Name</FieldLabel>
              <Input
                id={field.name}
                type="text"
                placeholder="Your name"
                autoComplete="name"
                autoFocus
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.name
                }
              />
              <FieldError
                errors={resolveErrors(
                  field.state.meta.errors,
                  serverErrors?.name,
                )}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="email">
          {(field) => (
            <Field>
              <FieldLabel htmlFor={field.name}>Email</FieldLabel>
              <Input
                id={field.name}
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.email
                }
              />
              <FieldError
                errors={resolveErrors(
                  field.state.meta.errors,
                  serverErrors?.email,
                )}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="password">
          {(field) => (
            <Field>
              <FieldLabel htmlFor={field.name}>Password</FieldLabel>
              <PasswordInput
                id={field.name}
                placeholder="••••••••"
                autoComplete="new-password"
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
          Create account
        </LoadingButton>
      </form>

      <FieldSeparator>or</FieldSeparator>

      <OAuthButtons disabled={mutation.isPending} />

      <p className="text-muted-foreground text-center text-sm">
        Already have an account?{" "}
        <Link
          to="/login"
          className="text-foreground font-medium hover:underline"
        >
          Log in
        </Link>
      </p>
    </div>
  );
}
