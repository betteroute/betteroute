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
import { authQueries, login } from "@/features/auth/queries";
import { loginSchema } from "@/features/auth/schemas";
import { getFieldErrors } from "@/lib/api";
import { resolveErrors } from "@/lib/form-errors";

export const Route = createFileRoute("/_auth/login")({
  component: LoginPage,
});

function LoginPage() {
  const queryClient = useQueryClient();
  const navigate = Route.useNavigate();
  const { redirect: redirectTo } = Route.useSearch();

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (user) => {
      queryClient.setQueryData(authQueries.session().queryKey, user);
      navigate({ href: redirectTo ?? "/" });
    },
  });

  const form = useForm({
    defaultValues: { email: "", password: "" },
    validators: {
      onSubmit: loginSchema,
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
          Log in to Betteroute
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Welcome back. Please log in to continue.
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
              <div className="flex items-center justify-between">
                <FieldLabel htmlFor={field.name}>Password</FieldLabel>
                <Link
                  to="/forgot-password"
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  Forgot password?
                </Link>
              </div>
              <PasswordInput
                id={field.name}
                placeholder="••••••••"
                autoComplete="current-password"
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
          Log in
        </LoadingButton>
      </form>

      <FieldSeparator>or</FieldSeparator>

      <OAuthButtons disabled={mutation.isPending} />

      <p className="text-muted-foreground text-center text-sm">
        Don&apos;t have an account?{" "}
        <Link
          to="/signup"
          className="text-foreground font-medium hover:underline"
        >
          Sign up
        </Link>
      </p>
    </div>
  );
}
