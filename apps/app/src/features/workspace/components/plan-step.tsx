import { useNavigate } from "@tanstack/react-router";
import { Check } from "lucide-react";

import { Button } from "@/components/ui/button";

const FREE_FEATURES = [
  "25 Links / month",
  "1,000 Clicks / month",
  "3 Custom domains",
  "10 Folders & 25 Tags",
  "30-day analytics",
  "API Access included",
  "3 Workspace members",
];

export function PlanStep({
  slug,
  onContinue,
}: {
  slug: string;
  onContinue?: () => void;
}) {
  const navigate = useNavigate();

  function handleContinue() {
    onContinue?.();
    navigate({ to: "/$slug", params: { slug } });
  }

  return (
    <div className="mx-auto flex w-full max-w-sm flex-col items-center gap-6">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Choose your plan
        </h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Start free — upgrade anytime from workspace settings.
        </p>
      </div>

      <div className="bg-card text-card-foreground w-full rounded-xl border p-6 shadow-sm">
        <div className="mb-4">
          <h3 className="text-lg font-semibold">Free</h3>
          <p className="text-muted-foreground text-sm">
            Everything you need to get started in beta.
          </p>
        </div>

        <div className="mb-6 flex items-baseline gap-1">
          <span className="text-4xl font-bold">$0</span>
          <span className="text-muted-foreground text-sm font-medium">/mo</span>
        </div>

        <ul className="mb-6 flex flex-col gap-2.5">
          {FREE_FEATURES.map((feature) => (
            <li key={feature} className="flex items-center gap-2.5 text-sm">
              <Check className="size-4 shrink-0 text-emerald-600 dark:text-emerald-400" />
              <span className="text-muted-foreground">{feature}</span>
            </li>
          ))}
        </ul>

        <Button size="lg" className="w-full" onClick={handleContinue}>
          Get started
        </Button>
      </div>

      <p className="text-muted-foreground text-center text-[10px] font-medium opacity-70">
        Pro, Team, and Enterprise plans coming soon.
      </p>
    </div>
  );
}
