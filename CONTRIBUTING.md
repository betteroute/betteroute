# Contributing to Betteroute

First off — thank you. Whether you're fixing a typo, reporting a bug, or building an entire feature, your contribution matters and we're glad you're here.

If anything in this guide is unclear, please open an issue.

## Where to Start

**Bug fixes, typos, docs, performance improvements** — open a PR directly. No prior approval needed.

**New features or architectural changes** — open an issue first to discuss the approach. The core is evolving quickly, and a short conversation upfront saves time and ensures your work gets merged.

Not sure where to begin? Look for [`good first issue`](https://github.com/betteroute/betteroute/labels/good%20first%20issue) — these are scoped, well-described, and ideal for getting familiar with the codebase.

Before starting any work, check existing [Issues](https://github.com/betteroute/betteroute/issues) and [Pull Requests](https://github.com/betteroute/betteroute/pulls) to avoid duplicate work. If an issue exists but isn't assigned, comment on it — we'll assign it to you.

## Getting Set Up

You'll need:

- [Docker](https://docs.docker.com/get-docker/) running (Atlas needs it for schema diffs)
- [proto](https://moonrepo.dev/docs/proto/install) + [moon](https://moonrepo.dev/docs/install) installed

That's it. Go, Node, and pnpm are automatically provisioned via `.prototools` — no manual version management.

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

Email and OAuth are optional for local development. Leave those env vars blank — the API falls back to a no-op notifier and OAuth buttons won't render. No third-party credentials required.

If setup takes longer than five minutes or something breaks, open an issue.

## Project Overview

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

The codebase is intentionally uniform — every backend feature uses the same four-file structure, every frontend feature uses the same flat layout. Once you've read one, you've read them all.

See **[STYLE.md](STYLE.md)** for naming, architecture, error handling, and patterns across Go and TypeScript.

## Development Commands

### Backend

```sh
moon run api:dev          # live reload via air
moon run api:test         # tests with race detection
moon run api:lint         # golangci-lint
moon run api:sqlc         # regenerate SQL bindings after query changes
moon run api:db-migrate   # generate migration from schema changes
moon run api:db-push      # apply migrations
```

### Frontend

```sh
moon run app:dev          # TanStack Start dev server
moon run app:build        # production build
moon run app:check        # biome lint + format
```

## Pull Requests

### Keep them focused

One PR, one purpose — a single feature, bug fix, or refactor. Cross-cutting changes across backend and frontend are fine if they're one logical unit.

**Target under 500 lines changed and under 10 files.** Not a hard rule, but PRs of this size get reviewed same-day. Larger ones tend to sit.

For larger efforts, split into a natural sequence: schema first, then backend, then frontend. Each PR should be independently reviewable and mergeable.

### Think like a reviewer

Before you open the PR, read your own diff as if you're seeing it for the first time. Ask yourself:

- Can someone understand this without asking me questions?
- Are there trade-offs or edge cases I should document?
- Did I leave any debugging code, TODOs, or half-finished thoughts?

Your PR description is the first thing a reviewer reads:

- **What** does this do and **why**?
- Explain non-obvious decisions or trade-offs
- Link related issues: `Closes #123` or `Fixes #456`
- How did you test it? A sentence or two is enough

Every PR should be self-contained. If context lives in a Slack thread or your head, bring it into the description.

### Before you submit

- [ ] `moon run api:lint` and `moon run app:check` pass
- [ ] `moon run api:test` passes
- [ ] Production build succeeds (`moon run api:build && moon run app:build`)
- [ ] SQL bindings regenerated if you changed queries (`moon run api:sqlc`)
- [ ] Migration generated if you changed schema (`moon run api:db-migrate`)
- [ ] Branch is up to date with `main`
- [ ] "Allow edits from maintainers" is checked (this lets us make small fixes without back-and-forth)

Moon's pre-commit hook handles staged-file linting automatically — it'll catch most issues before they leave your machine.

### Keep your branch updated

Rebase on `main` before requesting review. If `main` moves while your PR is open, rebase again before merge. This keeps history clean and avoids merge conflicts piling up.

```sh
git fetch origin
git rebase origin/main
```

### Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/) with scopes:

```
feat(api): add custom domain dns verification
fix(app): resolve stale cache on workspace switch
refactor(api): extract mapError to package function
chore: update toolchain deps
```

All lowercase, imperative mood, no period. Details in [STYLE.md](STYLE.md).

## Reporting Bugs

Found something broken? Open an issue with:

- Steps to reproduce
- What you expected vs what happened
- Your environment (OS, browser, self-hosted or cloud)

Even if you're not sure it's a bug — report it. We'd rather triage a false alarm than miss a real issue.

## Priorities

| Type | Priority |
|---|---|
| Core bugs — auth, redirects, data integrity | Urgent |
| Core features — redirect engine, deep linking, workspaces | High |
| UX issues, missing edge cases | Medium |
| Minor improvements, non-core requests | Low |

This isn't about gatekeeping — it's about making sure the most impactful work lands first. Every contribution matters, and we'll review everything that comes in.

## Cloud and Self-Hosting

Betteroute is available as a managed cloud service and as a self-hosted deployment. Both are first-class — we don't treat self-hosters as second-class citizens or gate critical features behind the cloud tier.

If you encounter issues specific to your deployment — whether cloud or self-hosted — please report them. They're high priority for us.

## License

By contributing, you agree that your contributions will be licensed under the [AGPLv3](LICENSE). Enterprise features in `/ee` are under a separate commercial license — do not add code there without prior discussion.

---

Thank you for contributing. Every issue filed, every PR opened, every bug reported makes Betteroute better for everyone.
