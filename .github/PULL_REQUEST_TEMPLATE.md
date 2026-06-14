## Summary

<!--
  One paragraph. What problem does this PR solve, and how?
  If it fixes an issue: "Fixes #<number>"
-->

## Type of change

<!-- Check all that apply -->

- [ ] `feat` — new feature
- [ ] `fix` — bug fix
- [ ] `refactor` — no behaviour change
- [ ] `docs` — documentation only
- [ ] `test` — tests only
- [ ] `chore` — build, tooling, or dependency update
- [ ] Breaking change (existing behaviour changes in a backward-incompatible way)

## Affected layers

- [ ] Backend (Go / Gin / PostgreSQL)
- [ ] Frontend (Next.js / React / TypeScript)
- [ ] Mobile (Android / Kotlin / Compose)
- [ ] CI / infrastructure

## Checklist

### Code quality

- [ ] `go vet ./...` passes (backend)
- [ ] `pnpm lint && pnpm build` passes (frontend)
- [ ] `./gradlew lint` passes (mobile)

### Tests

- [ ] New behaviour is covered by tests
- [ ] Integration tests use Testcontainers — no database mocks
- [ ] `make test` (backend) / `pnpm test` (frontend) / `./gradlew test` (mobile) passes locally

### Documentation

- [ ] Relevant doc in `backend/docs/`, `frontend/docs/`, or `mobile/docs/` is updated
- [ ] `last_verified` date updated in any touched doc files
- [ ] `CLAUDE.md` / `AGENTS.md` updated if project setup or structure changed

## Screenshots / recordings

<!--
  Required for any UI change (frontend or mobile).
  Before / After screenshots help reviewers verify the intent without running the app.
  Delete this section if not applicable.
-->

| Before | After |
|--------|-------|
|        |       |

## Additional context

<!--
  Anything else reviewers should know: migration steps, performance impact,
  follow-up issues, known limitations, etc.
-->
