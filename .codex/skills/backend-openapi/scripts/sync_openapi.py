"""Regenerate the backend Swagger contract, then its offline HTML reference."""

from pathlib import Path
import subprocess
import sys


def find_project_root() -> Path:
    for candidate in Path(__file__).resolve().parents:
        if (candidate / "backend/go.mod").is_file() and (
            candidate / "backend/scripts/generate-redoc.py"
        ).is_file():
            return candidate
    raise RuntimeError("Could not find the project root containing backend/go.mod")


def main() -> None:
    root = find_project_root()
    backend = root / "backend"

    print("Generating Swagger contract...", flush=True)
    subprocess.run(
        [
            "go",
            "tool",
            "swag",
            "init",
            "-g",
            "main.go",
            "-d",
            "cmd/api,internal/auth,internal/users,internal/monitors,internal/store",
            "-o",
            "docs",
            "--parseInternal",
            "--outputTypes",
            "json,yaml",
        ],
        cwd=backend,
        check=True,
    )

    print("Generating Redoc HTML from the updated Swagger contract...", flush=True)
    subprocess.run(
        [sys.executable, str(backend / "scripts/generate-redoc.py")],
        cwd=root,
        check=True,
    )


if __name__ == "__main__":
    main()
