# Repository Guidelines

## Repository Layout

This repository contains two independent projects: `backend/` and `frontend/`. Keep each project’s source, tests, configuration, and documentation inside its own directory. Place tests beside the code they cover and keep generated output and dependency directories out of version control.

## General Development Practices

Read the applicable nested guide before changing a project:

- `backend/AGENTS.md` contains Go structure, commands, style, and testing rules.
- `frontend/AGENTS.md` contains Next.js structure, commands, style, and testing rules.

Keep changes focused, use clear names, and update documentation when commands, configuration, or public behavior changes. Run the relevant formatting, linting, build, and test commands before opening a pull request.

## Backend API Documentation

Add or update swaggo annotations for every backend HTTP route whenever its path, parameters, request body, response, or behavior changes. Regenerate the checked-in Swagger contract from `backend/` after changing a route or its annotations:

```sh
go tool swag init -g main.go -d cmd/api,internal/auth,internal/users,internal/monitors,internal/store -o docs --parseInternal --outputTypes json,yaml
```

Commit the regenerated `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, and `backend/docs/redoc.html` with the route change. Refresh the Redoc HTML after regenerating the Swagger contract:

```sh
python3 backend/scripts/generate-redoc.py
```

## When starting a new task use GitHub Flow

details in file: docs/rules/github-flow.md

## When creating Pull Request

details in file: docs/rules/pr.md

## Configuration & Security

Do not commit secrets, tokens, credentials, or local environment files. Use safe example values in tracked configuration documentation and provide real credentials through the deployment environment. Review changes for accidental exposure of sensitive data before committing.

## Комментарии в коде
- комментируй не что, а почему
- очевидное не комментируй, лучше используй правильные наименования
- для публичных функций делай doc comments
- сложную арифметику поясняй рядом
- если меняешь код, то актуализируй комментарии
