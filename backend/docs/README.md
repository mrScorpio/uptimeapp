# OpenAPI

The API contract is generated from the annotations in `cmd/api/main.go` and the
HTTP controller handlers. After changing an endpoint or its request/response
shape, regenerate the checked-in Swagger 2.0 files from `backend/`:

```sh
go tool swag init -g main.go -d cmd/api,internal/auth,internal/users,internal/monitors,internal/store -o docs --parseInternal --outputTypes json,yaml
```

The generated contract is available in [`swagger.json`](swagger.json) and
[`swagger.yaml`](swagger.yaml).

Build an offline HTML API reference from the generated contract, from the
repository root:

```sh
python3 backend/scripts/generate-redoc.py
```

The script embeds the endpoints and schemas from `swagger.json` in
[`redoc.html`](redoc.html), so the API structure remains visible when opened
without network access.
