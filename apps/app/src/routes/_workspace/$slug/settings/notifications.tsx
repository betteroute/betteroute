import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_workspace/$slug/settings/notifications",
)({
  staticData: { title: "Notifications" },
  component: NotificationsSettingsPage,
});

function NotificationsSettingsPage() {
  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-lg font-semibold">Notifications</h1>
        <p className="text-sm text-muted-foreground">
          Manage your notification preferences.
        </p>
      </div>
    </div>
  );
}
