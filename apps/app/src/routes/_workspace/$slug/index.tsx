import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_workspace/$slug/")({
  beforeLoad: ({ params }) => {
    throw redirect({
      to: "/$slug/links",
      params,
    });
  },
});
