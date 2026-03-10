import { createFileRoute } from "@tanstack/react-router";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/_workspace/$slug/domains")({
  component: DomainsPage,
});

function DomainsPage() {
  return (
    <>
      <PageHeader title="Domains" actions={<Button>Add Domain</Button>} />
      <div className="flex flex-1 flex-col gap-4 p-4">
        <div className="text-muted-foreground text-center text-sm py-20">
          Custom domains will appear here
        </div>
      </div>
    </>
  );
}
