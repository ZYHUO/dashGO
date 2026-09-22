# Contributing to dashGO

Thanks for taking the time to contribute. This project is a Go backend with a
Vue 3 frontend, so most changes fall into one of those two halves.

## Reporting bugs

Please use the **Bug report** template and include:

- dashGO version or commit SHA
- deploy method (`install.sh` / Docker Compose / manual build)
- database (SQLite or MySQL + version)
- browser and version, if it's a frontend rendering issue

Bugs without a reproduction path often can't be acted on, so the more specific
the steps, the faster it gets looked at.

## Suggesting features

Open a **Feature request** first. Describe the problem you're trying to solve,
not just the solution — several dashGO features exist because someone explained
a workflow that was painful.

## Pull requests

1. Fork and branch from `main`. Use a descriptive branch name
   (`fix/status-page-crash`, not `patch-1`).
2. Keep the diff focused. One logical change per PR.
3. **Never commit credentials, `.env` files, tokens or real site URLs** into
   the repo. This includes demo accounts and test instances.
4. If you touch the frontend, make sure the build passes.
5. If you touch the backend, make sure `go build ./...` and the tests pass.

## Security issues

Do **not** open a public issue for anything security-sensitive. See
[SECURITY.md](SECURITY.md).

## Code of conduct

Be decent to each other. Disagreement about a technical approach is fine;
disagreement about a person is not.
