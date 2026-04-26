#!/usr/bin/env python3
"""Cross-platform Terraform helpers for pre-commit."""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

TF_FILE_SUFFIXES = (".tf", ".tfvars", ".tfvars.json", ".tfvars.example")


def write_line(message: str, *, stream: object = sys.stdout) -> None:
    stream.write(f"{message}\n")
    stream.flush()


def is_terraform_file(path: Path) -> bool:
    return path.name.endswith(TF_FILE_SUFFIXES)


def module_dirs(paths: list[str]) -> list[Path]:
    dirs = {
        path.parent if str(path.parent) else Path(".")
        for raw_path in paths
        if is_terraform_file(path := Path(raw_path))
    }
    return sorted(dirs, key=lambda path: path.as_posix())


def tool_path(name: str, *, required: bool) -> str | None:
    path = shutil.which(name)
    if path:
        return path

    if required:
        write_line(f"{name} is required but was not found on PATH", stream=sys.stderr)
        return None

    write_line(f"{name} not installed locally; skipping (CI installs it).")
    return None


def run(command: list[str], *, cwd: Path | None = None) -> int:
    where = f" (cwd={cwd})" if cwd else ""
    write_line(f"+ {' '.join(command)}{where}")
    return subprocess.run(command, cwd=cwd, check=False).returncode  # noqa: S603


def remove_terraform_cache(directory: Path) -> None:
    terraform_cache = directory / ".terraform"
    if terraform_cache.exists():
        write_line(f"Removing stale Terraform cache: {terraform_cache}")
        shutil.rmtree(terraform_cache)


def terraform_fmt(dirs: list[Path]) -> int:
    terraform = tool_path("terraform", required=True)
    if terraform is None:
        return 1

    status = 0
    for directory in dirs:
        status |= run([terraform, "fmt", "-recursive", str(directory)])
    return status


def terraform_validate(dirs: list[Path]) -> int:
    terraform = tool_path("terraform", required=True)
    if terraform is None:
        return 1

    for directory in dirs:
        status = 1
        for attempt in range(2):
            init_status = run(
                [
                    terraform,
                    f"-chdir={directory}",
                    "init",
                    "-backend=false",
                    "-input=false",
                ]
            )
            if init_status == 0:
                status = run([terraform, f"-chdir={directory}", "validate"])
                if status == 0:
                    break
            else:
                status = init_status

            if attempt == 0:
                remove_terraform_cache(directory)
                write_line("Retrying Terraform validate after cache cleanup.")

        if status:
            return status

    return 0


def terraform_tflint(dirs: list[Path]) -> int:
    tflint = tool_path("tflint", required=False)
    if tflint is None:
        return 0

    status = 0
    for directory in dirs:
        status |= run([tflint, "--minimum-failure-severity=warning"], cwd=directory)
    return status


def terraform_docs(dirs: list[Path]) -> int:
    terraform_docs = tool_path("terraform-docs", required=False)
    if terraform_docs is None:
        return 0

    config = Path(".terraform-docs.yml").resolve()
    status = 0
    for directory in dirs:
        status |= run([terraform_docs, "--config", str(config), str(directory)])
    return status


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=["fmt", "validate", "tflint", "docs"])
    parser.add_argument("files", nargs="*")
    args = parser.parse_args()

    dirs = module_dirs(args.files)
    if not dirs:
        return 0

    if args.command == "fmt":
        return terraform_fmt(dirs)
    if args.command == "validate":
        return terraform_validate(dirs)
    if args.command == "tflint":
        return terraform_tflint(dirs)
    if args.command == "docs":
        return terraform_docs(dirs)

    return 1


if __name__ == "__main__":
    raise SystemExit(main())
