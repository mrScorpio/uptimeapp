# Repository Guidelines

## Repository Layout

This repository contains two independent projects: `backend/` and `frontend/`. Keep each project’s source, tests, configuration, and documentation inside its own directory. Place tests beside the code they cover and keep generated output and dependency directories out of version control.

## General Development Practices

Read the applicable nested guide before changing a project:

- `backend/AGENTS.md` contains Go structure, commands, style, and testing rules.
- `frontend/AGENTS.md` contains Next.js structure, commands, style, and testing rules.

Keep changes focused, use clear names, and update documentation when commands, configuration, or public behavior changes. Run the relevant formatting, linting, build, and test commands before opening a pull request.

## Commit & Pull Request Guidelines

This checkout has no Git history to establish a project-specific convention. Use concise imperative commit subjects, for example `Add uptime check configuration`, and keep each commit focused. Pull requests should explain the change, identify the affected project, link the relevant issue, and list validation commands. Include screenshots or a short recording for visible frontend changes.
Format: Convetional Commits (feat, fix, refactor, test, docs, chore).
Header limit is 72 symbols.
No emoji, no "significantly improve" phrases and other garbage.
One logical step - one commit.

## Configuration & Security

Do not commit secrets, tokens, credentials, or local environment files. Use safe example values in tracked configuration documentation and provide real credentials through the deployment environment. Review changes for accidental exposure of sensitive data before committing.
