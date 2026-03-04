import { queryOptions } from "@tanstack/react-query";
import { env } from "@/env";
import { api } from "@/lib/api";
import type {
  ForgotPasswordInput,
  LoginInput,
  ResetPasswordInput,
  SignupInput,
} from "./schemas";
import type { User } from "./types";

export const authKeys = {
  all: ["auth"] as const,
  session: () => [...authKeys.all, "session"] as const,
};

export const authQueries = {
  session: () =>
    queryOptions({
      queryKey: authKeys.session(),
      queryFn: () => api.get("auth/me").json<User>(),
      staleTime: 5 * 60 * 1000,
      retry: false,
      refetchOnWindowFocus: true,
    }),
};

export async function login(input: LoginInput) {
  return api.post("auth/login", { json: input }).json<User>();
}

export async function signup(input: SignupInput) {
  return api.post("auth/signup", { json: input }).json<User>();
}

export async function logout() {
  await api.post("auth/logout");
}

export async function forgotPassword(input: ForgotPasswordInput) {
  await api.post("auth/forgot-password", { json: input });
}

export async function resetPassword(input: ResetPasswordInput) {
  await api.post("auth/reset-password", { json: input });
}

export async function verifyEmail(token: string) {
  await api.post("auth/verify-email", { json: { token } });
}

export async function resendVerification(email: string) {
  await api.post("auth/resend-verification", { json: { email } });
}

export function getOAuthURL(provider: "google" | "github") {
  const base = env.VITE_API_URL?.replace(/\/+$/, "") ?? "";
  return `${base}/auth/oauth/${provider}`;
}
