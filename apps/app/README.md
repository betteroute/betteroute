# @betteroute/app

Betteroute dashboard — built with TanStack Start, shadcn/ui (Nova), and Tailwind CSS.

## Development

```bash
moon run app:dev
```

## Scripts

| Command | Description |
|---------|-------------|
| `moon run app:dev` | Start dev server on port 3000 |
| `moon run app:build` | Production build |
| `moon run app:test` | Run tests with Vitest |
| `moon run app:lint` | Lint with Biome |
| `moon run app:format` | Format with Biome |
| `moon run app:check` | Lint + format |

## Adding shadcn components

```bash
pnpm dlx shadcn@latest add button
```

## Stack

- **Framework:** TanStack Start (SSR, server functions, file-based routing)
- **UI:** shadcn/ui (Nova preset, Neutral theme)
- **Styling:** Tailwind CSS v4, Geist font
- **Forms:** TanStack Form + Zod
- **Data:** TanStack Query + Table
- **Linting:** Biome
- **Testing:** Vitest
