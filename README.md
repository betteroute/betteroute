# Betteroute

Open-source link management platform built for creators, teams, developers, and enterprises.

> **Work in progress.** Core architecture is settling fast — expect breaking changes until the first public beta.

## Modules

### Stable

| Module | Description |
|---|---|
| `workspace` `rbac` `entitlement` | Multi-tenant architecture with tier-based feature gating |
| `auth` `apikey` | Session and API key authentication with scoped permissions |
| `link` `tag` `folder` | Link management and organization |
| `redirect` | High-speed programmatic redirect engine |
| `notify` | Email and team notifications via [Resend](https://resend.com) |

### In development

| Module | Description |
|---|---|
| `deeplink` `wellknown` | Device-aware resolution and app associations |
| `domain` | Custom domains with automated TLS provisioning |
| `billing` `usage` | Monetization and API quota tracking |
| `ch` | ClickHouse integration for high-volume analytics |

### Planned

- Bio pages — link-in-bio with deep linking built in by default
- Advanced conversion tracking and cohort analytics
- Expanded deep link fallback logic for iOS and Android edge cases

## Stack

Monorepo managed by [Moon](https://moonrepo.dev). Toolchain versions are pinned in `.prototools` — Go and Node are provisioned automatically, no manual setup needed.

**Backend** (`apps/api`) — Go, Fiber v3, sqlc, Atlas, air, Caddy, swag

**Frontend** (`apps/app`) — TypeScript, TanStack Start, TanStack Router, TanStack Query, shadcn Nova

## Getting started

Install [moonrepo](https://moonrepo.dev/docs/install) and [proto](https://moonrepo.dev/docs/proto/install), then:

```sh
# Install dependencies — auto-provisions Go + Node via .prototools
pnpm install

# Set up the database
moon run api:db-push    # apply schema
moon run api:db-seed    # seed test data

# Start development servers
moon run api:dev        # backend with live reload
moon run app:dev        # frontend dashboard
```

## Commands

**Backend**

```sh
moon run api:build        # compile production binary
moon run api:test         # run tests with race detection and coverage
moon run api:lint         # golangci-lint
moon run api:sqlc         # regenerate type-safe SQL bindings
moon run api:docs         # regenerate OpenAPI spec
moon run api:db-migrate   # diff and generate migrations
moon run api:db-push      # apply migrations
moon run api:db-status    # check migration status
moon run api:db-seed      # seed development data
```

**Frontend**

```sh
moon run app:build        # production build
moon run app:test         # run test suite
moon run app:lint         # lint TypeScript
moon run app:format       # apply formatting
moon run app:check        # strict lint and format validation
```

## Contributing

The core architecture is still moving fast. If you're planning a larger contribution, open an issue first — it helps avoid duplicate effort.

Bug reports and small fixes are welcome at any stage.

## License

[AGPLv3](LICENSE) with an open-core model. The core platform is fully open source. Enterprise features live in `/ee` under a separate commercial license.