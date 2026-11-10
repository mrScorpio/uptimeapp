# Repository Guidelines

## Repository Layout

This repository contains two independent projects: `backend/` and `frontend/`. Keep each project’s source, tests, configuration, and documentation inside its own directory. Place tests beside the code they cover and keep generated output and dependency directories out of version control.

## General Development Practices

Read the applicable nested guide before changing a project:

- `backend/AGENTS.md` contains Go structure, commands, style, and testing rules.
- `frontend/AGENTS.md` contains Next.js structure, commands, style, and testing rules.

Keep changes focused, use clear names, and update documentation when commands, configuration, or public behavior changes. Run the relevant formatting, linting, build, and test commands before opening a pull request.

## Commit & Pull Request Guidelines

Use GitHub Flow for every task: create a separate branch from the main branch before making task changes, and do not commit task changes directly to the main branch.

Name task branches using `<type>/<feature-name>`, where `<type>` is `feat` for new functionality or `fix` for a correction, and `<feature-name>` is a concise transliterated Russian name written in lowercase kebab-case. For example: `feat/dobavit-proverku-sayta` or `fix/ispravit-vhod`.

Use Conventional Commits (`feat`, `fix`, `refactor`, `test`, `docs`, `chore`) with concise imperative subjects and a 72-character header limit. Do not use emoji or vague promotional wording. Keep each commit focused on one logical step.

### Pull Request workflow

For each task, open a pull request from the task branch into `master` after implementation and validation are complete. Do not include unrelated working-tree changes in the branch or pull request.

- Use a Conventional Commit style title with a type and imperative summary, such as `feat: add monitor creation flow`. Keep the title concise and within 72 characters.
- Write the description so a reviewer can understand the change without reading the whole diff. Include a summary of user-visible behavior, implementation details for each affected project, any important scope limits, and a checklist of validation commands with their results.
- Link the related issue when one exists; do not invent an issue reference when none was provided.
- Include screenshots or a short recording for visible frontend changes when available.
- Push the task branch and confirm the PR base is `master`. Report the PR URL, title, and validation results after creation.

## Configuration & Security

Do not commit secrets, tokens, credentials, or local environment files. Use safe example values in tracked configuration documentation and provide real credentials through the deployment environment. Review changes for accidental exposure of sensitive data before committing.

## Additional commit rules

Format: Convetional Commits (feat, fix, refactor, test, docs, chore).
Header limit is 72 symbols.
No emoji, no "significantly improve" phrases and other garbage.
One logical step - one commit.
