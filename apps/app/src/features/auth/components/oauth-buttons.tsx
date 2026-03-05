import { useState } from "react";
import { GitHubIcon, GoogleIcon } from "@/components/shared/icons";
import { LoadingButton } from "@/components/shared/loading-button";
import { getOAuthURL } from "../queries";

const providers = [
  { id: "google" as const, label: "Continue with Google", icon: GoogleIcon },
  { id: "github" as const, label: "Continue with GitHub", icon: GitHubIcon },
];

export function OAuthButtons({ disabled }: { disabled?: boolean }) {
  const [loadingProvider, setLoadingProvider] = useState<
    "google" | "github" | null
  >(null);

  const handleOAuth = (id: "google" | "github") => {
    setLoadingProvider(id);
    // Yield to the React render cycle before navigating so the spinner shows
    requestAnimationFrame(() => {
      window.location.href = getOAuthURL(id);
    });
  };

  return (
    <div className="flex flex-col gap-2">
      {providers.map(({ id, label, icon: Icon }) => (
        <LoadingButton
          key={id}
          variant="outline"
          size="lg"
          className="w-full justify-center gap-2.5 text-sm font-normal"
          disabled={
            disabled || (loadingProvider !== null && loadingProvider !== id)
          }
          loading={loadingProvider === id}
          onClick={() => handleOAuth(id)}
        >
          {loadingProvider !== id && <Icon />}
          {label}
        </LoadingButton>
      ))}
    </div>
  );
}
