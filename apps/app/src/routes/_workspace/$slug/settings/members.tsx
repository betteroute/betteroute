import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useMemo } from "react";
import { z } from "zod";

import { DebouncedSearchInput } from "@/components/shared/debounced-search-input";
import { FilterSheet } from "@/components/shared/filter-sheet";

import { PageLoader } from "@/components/shared/page-loader";

import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { InvitationRow } from "@/features/workspace/components/invitation-row";

import { InviteDialog } from "@/features/workspace/components/invite-dialog";

import { MemberRow } from "@/features/workspace/components/member-row";
import { ROLE_FILTERS } from "@/features/workspace/constants";
import { useWorkspace } from "@/features/workspace/hooks";
import { workspaceQueries } from "@/features/workspace/queries";
import type { Invitation, Member } from "@/features/workspace/types";

const membersSearchSchema = z.object({
  search: z.string().optional().catch(undefined),
  role: z.array(z.string()).optional().catch(undefined),
});

export const Route = createFileRoute("/_workspace/$slug/settings/members")({
  validateSearch: membersSearchSchema,
  staticData: { title: "Members" },
  component: MembersSettingsPage,
});

function MembersSettingsPage() {
  const { workspace } = useWorkspace();
  const navigate = Route.useNavigate();
  const { search, role: selectedRoles } = Route.useSearch();
  const searchLower = (search || "").toLowerCase();

  const matchesRole = useCallback(
    (role: string) => !selectedRoles?.length || selectedRoles.includes(role),
    [selectedRoles],
  );

  const membersQuery = useQuery(workspaceQueries.members(workspace.slug));

  const invitationsQuery = useQuery(
    workspaceQueries.invitations(workspace.slug),
  );

  const filteredMembers = useMemo(() => {
    if (!membersQuery.data) return [];
    return membersQuery.data.filter(
      (m) =>
        (!search ||
          m.name?.toLowerCase().includes(searchLower) ||
          m.email?.toLowerCase().includes(searchLower)) &&
        matchesRole(m.role),
    );
  }, [membersQuery.data, search, searchLower, matchesRole]);

  const filteredInvitations = useMemo(() => {
    if (!invitationsQuery.data) return [];
    return invitationsQuery.data.filter(
      (inv) =>
        (!search || inv.email.toLowerCase().includes(searchLower)) &&
        matchesRole(inv.role),
    );
  }, [invitationsQuery.data, search, searchLower, matchesRole]);

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-lg font-semibold">Members</h1>

          <p className="text-muted-foreground text-sm">
            Invite team members and manage roles.
          </p>
        </div>

        <InviteDialog />
      </div>

      <div className="flex items-center gap-2">
        <FilterSheet
          filters={ROLE_FILTERS}
          values={{ role: selectedRoles }}
          onChange={(v) => {
            navigate({
              search: (prev) => ({ ...prev, role: v.role }),
              replace: true,
            });
          }}
        />

        <DebouncedSearchInput
          value={search ?? ""}
          onChange={(value) =>
            navigate({
              search: (prev) => ({
                ...prev,
                search: value || undefined,
              }),
              replace: true,
            })
          }
          placeholder="Search members…"
        />
      </div>

      <section>
        {membersQuery.isLoading ? (
          <PageLoader />
        ) : filteredMembers.length ? (
          <div className="space-y-4">
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-full">Member</TableHead>

                    <TableHead className="w-[150px]">Role</TableHead>

                    <TableHead className="w-[50px]" />
                  </TableRow>
                </TableHeader>

                <TableBody>
                  {filteredMembers.map((member: Member) => (
                    <MemberRow key={member.user_id} member={member} />
                  ))}
                </TableBody>
              </Table>
            </div>

            {filteredMembers.length > 10 && (
              <div className="flex items-center justify-between px-2">
                <p className="text-sm text-muted-foreground">
                  Showing {filteredMembers.length > 0 ? 1 : 0}–
                  {filteredMembers.length} of {filteredMembers.length}
                </p>
              </div>
            )}
          </div>
        ) : (
          <div className="text-muted-foreground text-sm py-8 text-center">
            {search || selectedRoles?.length
              ? "No members match your filters."
              : "No members found."}
          </div>
        )}
      </section>

      {filteredInvitations.length > 0 && (
        <section className="space-y-3">
          <h2 className="text-sm font-medium">
            Pending invitations{" "}
            <span className="text-muted-foreground font-normal">
              ({filteredInvitations.length})
            </span>
          </h2>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-full">Email</TableHead>

                  <TableHead className="w-[150px]">Role</TableHead>

                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>

              <TableBody>
                {filteredInvitations.map((invitation: Invitation) => (
                  <InvitationRow key={invitation.id} invitation={invitation} />
                ))}
              </TableBody>
            </Table>
          </div>
        </section>
      )}
    </div>
  );
}
