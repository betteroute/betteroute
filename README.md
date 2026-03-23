<p align="center">
  <h1 align="center">Betteroute</h1>
</p>

<p align="center">
  The open-source deep-link-first link management platform.
  <br />
  <br />
  <a href="https://github.com/betteroute/betteroute/issues">Issues</a>
  ·
  <a href="https://github.com/betteroute/betteroute/blob/main/CONTRIBUTING.md">Contributing</a>
  ·
  <a href="https://github.com/betteroute/betteroute/blob/main/STYLE.md">Style Guide</a>
</p>

> **Work in progress.** Core architecture is settling fast — expect breaking changes until the first public beta.

## About

Betteroute lets you create deep-link-first short links with custom domains, device-aware redirects, and smart routing — all in one platform you can self-host or use as a managed service.

**Self-hosted or cloud.** Run it on your own infrastructure with full control, or use our managed cloud when you'd rather not. Both are first-class.

## Tech Stack

| Layer | Technology |
| --- | --- |
| **API** | [Go](https://go.dev) + [Fiber v3](https://gofiber.io) |
| **Database** | [PostgreSQL](https://postgresql.org) + [sqlc](https://sqlc.dev) (no ORM) |
| **Migrations** | [Atlas](https://atlasgo.io) |
| **Frontend** | [TanStack Start](https://tanstack.com/start) / [Router](https://tanstack.com/router) / [Query](https://tanstack.com/query) / [Form](https://tanstack.com/form) |
| **UI** | [shadcn/ui](https://ui.shadcn.com) (Nova) + [Tailwind v4](https://tailwindcss.com) |
| **Monorepo** | [Moon](https://moonrepo.dev) task runner |

## Getting Started

### Prerequisites

| Tool | Why |
| --- | --- |
| [Docker](https://docs.docker.com/get-docker/) | Atlas uses it for schema diffs |
| [proto](https://moonrepo.dev/docs/proto/install) + [moon](https://moonrepo.dev/docs/install) | Monorepo toolchain |

> Go, Node, and pnpm are automatically provisioned via `.prototools` — no manual version management required.

### Quick Start

```sh
git clone https://github.com/betteroute/betteroute.git
cd betteroute
pnpm install

cp apps/api/.env.example apps/api/.env
# edit apps/api/.env — at minimum, set your DATABASE_URL

moon run api:db-push    # apply schema (Docker must be running)
moon run api:db-seed    # seed development data

moon run api:dev        # Go API → localhost:8080
moon run app:dev        # dashboard → localhost:3000
```

Email and OAuth are optional for local development. Leave those env vars blank — the API falls back to a no-op notifier and OAuth buttons won't appear.

## Project Structure

```
apps/
├── api/              Go API (Fiber v3, sqlc, PostgreSQL)
│   ├── cmd/          entrypoint
│   ├── internal/     feature packages (handler + service + store)
│   └── sql/          schema, migrations, queries
└── app/              dashboard (TanStack Start, shadcn Nova)
    └── src/
        ├── routes/   file-based routing
        └── features/ domain logic (queries, components, types)
```

The codebase follows a consistent pattern — every backend feature uses the same four-file structure, and every frontend feature uses the same flat layout. Once you've read one feature, you've read them all.

## Development Commands

### Backend

```sh
moon run api:dev          # live reload via air
moon run api:build        # compile production binary
moon run api:test         # tests with race detection
moon run api:lint         # golangci-lint
moon run api:sqlc         # regenerate type-safe SQL bindings
moon run api:docs         # regenerate OpenAPI spec
moon run api:db-migrate   # generate migration from schema changes
moon run api:db-push      # apply migrations
moon run api:db-seed      # seed development data
```

### Frontend

```sh
moon run app:dev          # TanStack Start dev server
moon run app:build        # production build
moon run app:check        # biome lint + format
```

## Self-Hosting

Betteroute is designed to be self-hosted from day one. You need PostgreSQL, a reverse proxy for custom domains, and optionally ClickHouse for analytics.

Detailed deployment guides are coming. If you're self-hosting today and run into issues, [open an issue](https://github.com/betteroute/betteroute/issues) — self-hosting bugs are high priority.

## Contributing

We welcome contributions of all sizes. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, PR guidelines, and how we work together.

For code style and architecture patterns, see [STYLE.md](STYLE.md).

## License

[AGPLv3](LICENSE) with an open-core model. The core platform is fully open source. Enterprise features will live in `/ee` under a separate commercial license.
