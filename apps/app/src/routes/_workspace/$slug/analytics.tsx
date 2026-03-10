import { createFileRoute } from "@tanstack/react-router";
import { Calendar } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/_workspace/$slug/analytics")({
  component: AnalyticsPage,
});

function AnalyticsPage() {
  return (
    <>
      <PageHeader
        title="Analytics"
        actions={
          <Button variant="outline" className="hidden sm:flex">
            <Calendar className="mr-2 size-4" />
            Last 30 days
          </Button>
        }
      />
      <div className="flex flex-1 flex-col gap-4 p-4">
        <div className="text-muted-foreground text-center text-sm py-20">
          Click analytics will appear here
        </div>
      </div>
    </>
  );
}
