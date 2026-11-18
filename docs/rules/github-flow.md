# GitHub Flow

Use GitHub Flow for every task: create a separate branch from the main branch before making task changes, and do not commit task changes directly to the main branch.

Name task branches using `<type>/<feature-name>`, where `<type>` is `feat` for new functionality or `fix` for a correction, and `<feature-name>` is a concise transliterated Russian name written in lowercase kebab-case. For example: `feat/dobavit-proverku-sayta` or `fix/ispravit-vhod`.

Use Conventional Commits (`feat`, `fix`, `refactor`, `test`, `docs`, `chore`) with concise imperative subjects and a 72-character header limit. Do not use emoji or vague promotional wording. Keep each commit focused on one logical step.