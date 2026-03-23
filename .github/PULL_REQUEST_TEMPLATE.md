## What does this PR do?

<!-- A brief description of the change and why it's needed. Link related issues: Closes #123 -->

## How was it tested?

<!-- How did you verify this works? A sentence or two is enough. -->

## Checklist

- [ ] I have self-reviewed my own code
- [ ] `moon run api:lint` and `moon run app:check` pass
- [ ] `moon run api:test` passes
- [ ] Production build succeeds (`moon run api:build && moon run app:build`)
- [ ] SQL bindings regenerated if queries changed (`moon run api:sqlc`)
- [ ] Migration generated if schema changed (`moon run api:db-migrate`)
- [ ] Branch is up to date with `main`
- [ ] "Allow edits from maintainers" is checked
