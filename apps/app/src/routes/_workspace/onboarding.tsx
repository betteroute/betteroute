import { createFileRoute, redirect } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { useState } from "react";
import { Logo } from "@/components/shared/logo";
import { Button } from "@/components/ui/button";
import { CreateWorkspaceForm } from "@/features/workspace/components/create-form";
import { useSession } from "@/features/workspace/hooks";
import { resolveDefaultWorkspace } from "@/lib/session";

export const Route = createFileRoute("/_workspace/onboarding")({
  ssr: false,
  beforeLoad: ({ context }) => {
    const slug = resolveDefaultWorkspace(context.workspaces);
    if (slug) throw redirect({ to: "/$slug", params: { slug } });
    return { user: context.user };
  },
  component: OnboardingPage,
});

function OnboardingPage() {
  const { user } = useSession();
  const [step, setStep] = useState<"welcome" | "workspace">("welcome");

  return (
    <div className="flex min-h-svh flex-col">
      <main className="flex w-full flex-1 flex-col items-center justify-center px-4 py-12">
        {step === "welcome" ? (
          <WelcomeStep
            userName={user.name}
            onContinue={() => setStep("workspace")}
          />
        ) : (
          <WorkspaceStep />
        )}
      </main>

      <footer className="px-6 py-4">
        <p className="text-muted-foreground text-sm">
          Signed in as{" "}
          <span className="text-foreground font-medium">{user.email}</span>
        </p>
      </footer>
    </div>
  );
}

function WelcomeStep({
  userName,
  onContinue,
}: {
  userName?: string;
  onContinue: () => void;
}) {
  const firstName = userName?.split(" ")[0];

  return (
    <div className="flex flex-col items-center gap-8 text-center">
      <Logo className="size-16" />

      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          {firstName ? `Welcome, ${firstName}!` : "Welcome to Betteroute"}
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Betteroute helps you create, manage, and track
          <br />
          your short links in one place.
        </p>
      </div>

      <Button size="lg" className="w-56" onClick={onContinue}>
        Get started
        <ArrowRight className="ml-2" />
      </Button>
    </div>
  );
}

function WorkspaceStep() {
  return (
    <div className="w-full max-w-sm">
      <div className="mb-10 flex justify-center">
        <Logo className="size-12" />
      </div>

      <div className="flex flex-col gap-8">
        <div className="text-center">
          <h1 className="text-2xl font-semibold tracking-tight">
            Create your workspace
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            Your shared space for links, analytics, and your team.
          </p>
        </div>

        <CreateWorkspaceForm />
      </div>
    </div>
  );
}
