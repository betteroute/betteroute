import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, stripSearchParams } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { z } from "zod";

import { DebouncedSearchInput } from "@/components/shared/debounced-search-input";
import { FilterSheet } from "@/components/shared/filter-sheet";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { Button } from "@/components/ui/button";

import { CreateLinkDialog } from "@/features/link/components/create-dialog";

import {
  LinksEmptyState,
  LinksFilteredEmptyState,
} from "@/features/link/components/empty-state";

import { LinkCard } from "@/features/link/components/link-card";

import { LINK_FILTERS } from "@/features/link/constants";
import { linkQueries } from "@/features/link/queries";

import { PAGINATION } from "@/lib/constants";
import { useWorkspace } from "@/features/workspace/hooks";

const searchSchema = z.object({
  offset: z.number().default(0).catch(0),
  search: z.string().optional().catch(undefined),
  status: z.array(z.string()).optional().catch(undefined),
});

export const Route = createFileRoute("/_workspace/$slug/links")({
  validateSearch: searchSchema,
  search: {
    middlewares: [stripSearchParams({ offset: 0 })],
  },
  loaderDeps: ({ search }) => ({
    offset: search.offset,
    search: search.search,
    status: search.status,
  }),
  loader: async ({ context, params, deps }) => {
    await context.queryClient.ensureQueryData(
      linkQueries.list(params.slug, deps),
    );
  },
  pendingComponent: PageLoader,
  component: LinksPage,
});

function LinksPage() {
  const { workspace } = useWorkspace();
  const navigate = Route.useNavigate();
  const searchParams = Route.useSearch();

  const hasFilters = !!(searchParams.search || searchParams.status?.length);
  const perPage = PAGINATION.DEFAULT_PER_PAGE;

  const query = useSuspenseQuery(
    linkQueries.list(workspace.slug, {
      offset: searchParams.offset,
      search: searchParams.search,
      status: searchParams.status,
    }),
  );

  const links = query.data.data;
  const pagination = query.data.pagination;
  const isEmpty = links.length === 0;
  const isFirstPage = searchParams.offset === 0;

  function clearAll() {
    navigate({
      search: { offset: 0 },
      replace: true,
    });
  }

  return (
    <>
      <PageHeader title="Links" actions={<CreateLinkDialog />} />

      <div className="flex flex-1 flex-col gap-4 p-4">
        {/* Toolbar — search + filters */}

        <div className="flex items-center gap-2">
          <FilterSheet
            filters={LINK_FILTERS}
            values={{ status: searchParams.status }}
            onChange={(v) => {
              navigate({
                search: (prev) => ({
                  ...prev,
                  status: v.status,
                  offset: 0,
                }),
                replace: true,
              });
            }}
          />

          <DebouncedSearchInput
            value={searchParams.search ?? ""}
            onChange={(value) =>
              navigate({
                search: (prev) => ({
                  ...prev,
                  search: value || undefined,
                  offset: 0,
                }),
                replace: true,
              })
            }
            placeholder="Search links…"
          />
        </div>

        {/* Link cards or empty state */}

        {isEmpty ? (
          <div className="rounded-lg border">
            {hasFilters ? (
              <LinksFilteredEmptyState onClear={clearAll} />
            ) : (
              <LinksEmptyState />
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {links.map((link) => (
              <LinkCard key={link.id} link={link} />
            ))}
          </div>
        )}

        {/* Pagination */}

        {!isEmpty && (
          <div className="flex items-center justify-end px-2">
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  navigate({
                    search: (p) => ({
                      ...p,
                      offset: Math.max(0, (p.offset ?? 0) - perPage),
                    }),
                  })
                }
                disabled={isFirstPage}
              >
                <ChevronLeft data-slot="icon" />
                Previous
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  navigate({
                    search: (p) => ({
                      ...p,
                      offset: (p.offset ?? 0) + perPage,
                    }),
                  })
                }
                disabled={!pagination.has_more}
              >
                Next
                <ChevronRight data-slot="icon" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
