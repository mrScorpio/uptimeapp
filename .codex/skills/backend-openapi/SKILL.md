---
name: backend-openapi
description: Use after adding, changing, or removing a backend HTTP API route to keep its swaggo annotations and generated Swagger and Redoc documentation in sync.
---

# Backend API contract

Keep the backend API contract complete whenever an API route changes. Swaggo annotations on HTTP handlers are the source of the contract, and generated files in `backend/docs/` must be refreshed in the same task.

## Workflow

1. Inspect the route registration and handler changes. Find API routes in backend `RegisterRoutes` methods and router setup, then compare the full API route inventory with the handlers' `@Router` annotations and the current Swagger paths. Include newly added, changed, and removed routes. Ignore middleware and infrastructure endpoints such as CORS preflight and static file serving unless they are intentionally part of the public API.
2. Add or update swaggo annotations on each affected handler. Keep the annotation path relative to `/api/v1` (for example, `/users/me`), and document the HTTP method, purpose, parameters, request body, success and error responses, and bearer authentication where applicable. Use named request and response types where needed to describe schemas. Remove obsolete route annotations when a route is deleted.
3. From the repository root, synchronize both generated references:

   ```sh
   python3 .codex/skills/backend-openapi/scripts/sync_openapi.py
   ```

   The script regenerates Swagger from `backend/`, then builds `backend/docs/redoc.html` from the updated Swagger JSON. Redoc generation runs only if Swagger generation succeeds.
4. Review `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, and `backend/docs/redoc.html`. Confirm every registered API route has exactly the intended method and path in the generated contract, removed routes are absent, and request, response, and authentication details match handler behavior. Include these generated files with the route change.

The helper automates the repeatable generation commands; checking route coverage and whether annotations match handler behavior still requires reviewing the code. The API base path is `/api/v1`, while swaggo `@Router` paths omit that prefix. Follow existing annotation conventions in `backend/internal/` and the generation notes in `backend/docs/README.md`.
