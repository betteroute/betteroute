<p align="center">
  <h1 align="center">Betteroute</h1>
</p>

<p align="center">
  Open-source link infrastructure platform.
  <br />
  <br />
  <a href="https://github.com/betteroute/betteroute/issues">Issues</a>
  ·
  <a href="https://github.com/betteroute/betteroute/blob/main/CONTRIBUTING.md">Contributing</a>
  ·
  <a href="https://github.com/betteroute/betteroute/blob/main/STYLE.md">Style Guide</a>
</p>

> **Work in progress.** Core architecture is settling — expect breaking changes until the first public beta.

## About

Short links, deep linking, routing, and attribution in one platform. Self-host or let us host it.

## Tech Stack

| Layer | Technology |
| --- | --- |
| **API** | [Go](https://go.dev) + [Fiber v3](https://gofiber.io) |
| **Database** | [PostgreSQL](https://postgresql.org) + [sqlc](https://sqlc.dev) |
| **Analytics** | [ClickHouse](https://clickhouse.com) |
| **Migrations** | [Atlas](https://atlasgo.io) |
| **Frontend** | [TanStack Start](https://tanstack.com/start) / [Router](https://tanstack.com/router) / [Query](https://tanstack.com/query) / [Form](https://tanstack.com/form) |
| **UI** | [shadcn/ui](https://ui.shadcn.com) + [Tailwind v4](https://tailwindcss.com) |
| **Monorepo** | [Moon](https://moonrepo.dev) |

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) — must be running (Atlas uses it for migrations and schema diffs)
- [proto](https://moonrepo.dev/docs/proto/install) + [moon](https://moonrepo.dev/docs/install)

Go, Node, and pnpm are provisioned via `.prototools`.

### Quick Start

```sh
git clone https://github.com/betteroute/betteroute.git
cd betteroute
pnpm install

cp apps/api/.env.example apps/api/.env
# edit .env — at minimum, set DATABASE_URL

moon run api:db-push    # apply schema
moon run api:db-seed    # seed development data

moon run api:dev        # API → localhost:8080
moon run app:dev        # dashboard → localhost:3000
```

Email and OAuth are optional locally. Leave those env vars blank and the API falls back to no-op defaults.

## Project Structure

```
apps/
├── api/              Go API
│   ├── cmd/          entrypoint
│   ├── internal/     feature packages
│   └── sql/          schema, migrations, queries
└── app/              dashboard
    └── src/
        ├── routes/   file-based routing
        └── features/ domain logic
```

## Development Commands

```sh
# Backend
moon run api:dev          # live reload
moon run api:build        # production binary
moon run api:test         # tests with race detection
moon run api:lint         # golangci-lint
moon run api:sqlc         # regenerate SQL bindings
moon run api:db-migrate   # generate migration
moon run api:db-push      # apply migrations

# Frontend
moon run app:dev          # dev server
moon run app:build        # production build
moon run app:check        # lint + format
```

## Self-Hosting

You need PostgreSQL, a reverse proxy for custom domains, and optionally ClickHouse for analytics. Deployment guides are coming.

If you're self-hosting and hit issues, [open an issue](https://github.com/betteroute/betteroute/issues).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup and guidelines. For code conventions, see [STYLE.md](STYLE.md).

## License

[AGPLv3](LICENSE). Enterprise features in `/ee` are under a separate commercial license.
