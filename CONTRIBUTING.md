# Contributing to Betteroute

Thanks for contributing. Here's how we work.

## Where to Start

**Bug fixes, typos, docs, performance** — open a PR directly.

**New features or architecture changes** — open an issue first. The core is moving fast and a quick discussion keeps things aligned.

[`good first issue`](https://github.com/betteroute/betteroute/labels/good%20first%20issue) labels are scoped and well-described.

Check [Issues](https://github.com/betteroute/betteroute/issues) and [PRs](https://github.com/betteroute/betteroute/pulls) before starting to avoid duplicate work.

## Setup

You need [Docker](https://docs.docker.com/get-docker/) (must be running — Atlas needs it), [proto](https://moonrepo.dev/docs/proto/install), and [moon](https://moonrepo.dev/docs/install). Go, Node, and pnpm are provisioned via `.prototools`.

```sh
git clone https://github.com/betteroute/betteroute.git
cd betteroute
pnpm install

cp apps/api/.env.example apps/api/.env
# set DATABASE_URL at minimum

moon run api:db-push
moon run api:db-seed

moon run api:dev        # API → localhost:8080
moon run app:dev        # dashboard → localhost:3000
```

Email and OAuth are optional locally. If setup takes more than five minutes, that's a bug.

## Project Structure

```
apps/
├── api/              Go API (Fiber v3, sqlc, PostgreSQL)
│   ├── cmd/          entrypoint
│   ├── internal/     feature packages (handler + service + store)
│   └── sql/          schema, migrations, queries
└── app/              dashboard (TanStack Start, shadcn)
    └── src/
        ├── routes/   file-based routing
        └── features/ domain logic
```

See **[STYLE.md](STYLE.md)** for conventions and patterns.

## Commands

```sh
# Backend
moon run api:dev          # live reload
moon run api:test         # tests with race detection
moon run api:lint         # golangci-lint
moon run api:sqlc         # regenerate SQL bindings
moon run api:db-migrate   # generate migration

# Frontend
moon run app:dev          # dev server
moon run app:build        # production build
moon run app:check        # lint + format
```

## Pull Requests

One PR, one purpose. Aim for under 500 lines and under 10 files — those get reviewed same-day.

For bigger work, split naturally: schema, then backend, then frontend. Each PR should be mergeable on its own.

Read your own diff before submitting. Your PR description should cover what changed, why, and how you tested it. Use `Closes #123` or `Fixes #456` to auto-link and close issues on merge.

### Checklist

- [ ] `moon run api:lint` and `moon run app:check` pass
- [ ] `moon run api:test` passes
- [ ] Builds clean
- [ ] SQL bindings regenerated if queries changed
- [ ] Migration generated if schema changed
- [ ] Branch is up to date with `main`

### Commits

[Conventional Commits](https://www.conventionalcommits.org/) with scopes:

```
feat(api): add custom domain dns verification
fix(app): resolve stale cache on workspace switch
chore: update toolchain deps
```

All lowercase, imperative mood, no period.

## Reporting Bugs

Open an issue with steps to reproduce, expected vs actual behavior, and your environment. Not sure if it's a bug? Report it anyway.

## License

Contributions are licensed under the [AGPLv3](LICENSE). Enterprise features in `/ee` are under a separate commercial license — don't add code there without discussing first.
