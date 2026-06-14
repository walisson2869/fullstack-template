---
topic: testing
last_verified: 2026-06-15
sources:
  - vitest.config.ts
  - vitest.setup.ts
  - __tests__/page.test.tsx
---

# Testing

## Framework

Vitest v4 + @testing-library/react v16 + jsdom v29.

| Package | Version | Role |
|---|---|---|
| `vitest` | ^4.1.8 | Test runner and assertion library |
| `@testing-library/react` | ^16.3.2 | Component rendering utilities |
| `@testing-library/jest-dom` | ^6.9.1 | DOM matchers (`toBeInTheDocument`, etc.) |
| `jsdom` | ^29.1.1 | Browser environment for Node |
| `@vitejs/plugin-react` | ^6.0.2 | JSX transform for Vitest |

## Commands

```bash
pnpm test          # run all tests once (CI)
pnpm test:watch    # watch mode (TDD inner loop)
pnpm test:ui       # browser UI
```

These map to the `scripts` in `package.json`:
- `test` → `vitest run`
- `test:watch` → `vitest`
- `test:ui` → `vitest --ui`

## Directory structure

```
__tests__/          # route-level and page-level tests
  page.test.tsx
__mocks__/next/     # Next.js built-in mocks (image.tsx, link.tsx) — aliased in vitest.config.ts
vitest.config.ts    # test config: jsdom env, @vitejs/plugin-react, next/* aliases
vitest.setup.ts     # imports @testing-library/jest-dom matchers
```

Component and utility tests co-locate with their source file:
```
components/
  MyButton.tsx
  MyButton.test.tsx   # co-located component test
lib/
  format.ts
  format.test.ts      # co-located utility test
```

## Configuration

`vitest.config.ts`:
```ts
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '.'),
      'next/image': path.resolve(__dirname, './__mocks__/next/image.tsx'),
      'next/link': path.resolve(__dirname, './__mocks__/next/link.tsx'),
    },
  },
})
```

`vitest.setup.ts`:
```ts
import '@testing-library/jest-dom'
```

## Next.js mock aliases

`vitest.config.ts` maps `next/image` and `next/link` to files under `__mocks__/next/`. These mocks render real `<img>` and `<a>` elements so jsdom can match against them. Create these files when a component imports either module.

## What to test and where

| Subject | Location | Pattern |
|---|---|---|
| `lib/` functions | Co-located `lib/foo.test.ts` | Pure unit — no DOM, no `render` |
| Client Components (`"use client"`) | Co-located `ComponentName.test.tsx` | `render` + `screen` queries |
| Route-level pages | `__tests__/<route>.test.tsx` | Renders the component in jsdom |
| Server Components | Extract logic to `lib/`; test that | Full rendering via Playwright (not yet set up) |

## Patterns

### Page / route test

```tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import Home from '../app/page'

describe('Home page', () => {
  it('renders the getting-started heading', () => {
    render(<Home />)
    expect(screen.getByText(/get started/i)).toBeInTheDocument()
  })

  it('renders the Deploy Now link', () => {
    render(<Home />)
    expect(screen.getByText('Deploy Now')).toBeInTheDocument()
  })
})
```

### Client Component test

```tsx
import { render, screen } from '@testing-library/react'
import MyButton from '@/components/MyButton'

it('renders label', () => {
  render(<MyButton label="Save" />)
  expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument()
})
```

### Utility function test

```ts
import { formatDate } from '@/lib/format'

it('formats ISO date', () => {
  expect(formatDate('2026-01-15')).toBe('Jan 15, 2026')
})
```

## TDD mandate

Write a failing test before writing the implementation. `pnpm test:watch` is the inner loop — keep it running while developing.

## Quality gate

All three must pass before committing:

```bash
pnpm test    # no failing tests
pnpm lint    # no ESLint errors
pnpm build   # no TypeScript errors, no build failures
```
