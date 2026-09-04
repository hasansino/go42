#!/usr/bin/env python3
"""Strict v-prefixed SemVer validation and ordering for release workflows."""

from __future__ import annotations

import functools
import re
import sys
from pathlib import Path
from tempfile import TemporaryDirectory


PATTERN = re.compile(
    r"^v(0|[1-9][0-9]*)\."
    r"(0|[1-9][0-9]*)\."
    r"(0|[1-9][0-9]*)"
    r"(?:-((?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)"
    r"(?:\.(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?$"
)


def parse(version: str) -> tuple[tuple[int, int, int], list[str] | None]:
    match = PATTERN.fullmatch(version)
    if match is None or len(version) > 128:
        raise ValueError(f"invalid strict v-prefixed SemVer: {version}")
    core = tuple(int(part) for part in match.group(1, 2, 3))
    prerelease = match.group(4)
    return core, None if prerelease is None else prerelease.split(".")


def compare(left: str, right: str) -> int:
    left_core, left_pre = parse(left)
    right_core, right_pre = parse(right)
    if left_core != right_core:
        return (left_core > right_core) - (left_core < right_core)
    if left_pre is None or right_pre is None:
        if left_pre is None and right_pre is None:
            return 0
        return 1 if left_pre is None else -1
    for left_id, right_id in zip(left_pre, right_pre):
        if left_id == right_id:
            continue
        left_numeric = left_id.isdigit()
        right_numeric = right_id.isdigit()
        if left_numeric and right_numeric:
            return (int(left_id) > int(right_id)) - (int(left_id) < int(right_id))
        if left_numeric != right_numeric:
            return -1 if left_numeric else 1
        return (left_id > right_id) - (left_id < right_id)
    return (len(left_pre) > len(right_pre)) - (len(left_pre) < len(right_pre))


def published_versions(path: Path) -> list[str]:
    versions: list[str] = []
    for line in path.read_text(encoding="utf-8").splitlines():
        try:
            parse(line)
        except ValueError:
            continue
        versions.append(line)
    return versions


def latest_version(path: Path) -> str:
    versions = published_versions(path)
    if not versions:
        raise ValueError("no published strict SemVer release establishes a baseline")
    return max(versions, key=functools.cmp_to_key(compare))


def prepare_release(
    requested: str,
    current: str,
    published_releases: Path,
) -> str:
    """Return the latest published baseline for a new release."""
    parse(requested)
    parse(current)
    previous = latest_version(published_releases)
    if requested == current:
        raise ValueError(f"{requested} already matches the Chart appVersion")
    if compare(current, previous) < 0:
        raise ValueError(
            f"Chart appVersion {current} is older than published release {previous}"
        )
    if compare(requested, previous) <= 0:
        raise ValueError(f"{requested} must be newer than published release {previous}")
    return previous


def self_test() -> None:
    ordered = [
        "v0.9.71",
        "v1.0.0-0",
        "v1.0.0-alpha",
        "v1.0.0-alpha.1",
        "v1.0.0-alpha.beta",
        "v1.0.0-beta",
        "v1.0.0-beta.2",
        "v1.0.0-beta.11",
        "v1.0.0-rc.1",
        "v1.0.0",
        "v1.0.1-build.7",
        "v1.0.1",
    ]
    for lower, higher in zip(ordered, ordered[1:]):
        assert compare(higher, lower) > 0, (lower, higher)
        assert compare(lower, higher) < 0, (lower, higher)
    for invalid in ("1.2.3", "v1.2", "v01.2.3", "v1.2.3+build", "v1.2.3-01"):
        try:
            parse(invalid)
        except ValueError:
            continue
        raise AssertionError(f"accepted invalid version: {invalid}")

    with TemporaryDirectory() as directory:
        published_file = Path(directory) / "published-releases"
        published_file.write_text("v0.9.71\nv0.9.72\n", encoding="utf-8")
        assert prepare_release("v0.9.73", "v0.9.72", published_file) == "v0.9.72"
        assert prepare_release("v0.9.74", "v0.9.73", published_file) == "v0.9.72"
        # A failed version is not an ordering baseline: a lower unused version
        # remains valid when it is newer than the latest published version.
        assert prepare_release("v0.9.73", "v0.9.74", published_file) == "v0.9.72"

        rejected = [
            ("v0.9.72", "v0.9.73", published_file),
            ("v0.9.73", "v0.9.73", published_file),
            ("v0.9.74", "v0.9.71", published_file),
        ]
        for arguments in rejected:
            try:
                prepare_release(*arguments)
            except ValueError:
                continue
            raise AssertionError(f"accepted invalid release transition: {arguments}")


def main(argv: list[str]) -> int:
    try:
        command = argv[1]
        if command == "validate" and len(argv) == 3:
            parse(argv[2])
        elif command == "newer" and len(argv) == 4:
            if compare(argv[2], argv[3]) <= 0:
                raise ValueError(f"{argv[2]} must be newer than {argv[3]}")
        elif command == "prepare" and len(argv) == 5:
            requested, current, releases_path = argv[2], argv[3], Path(argv[4])
            print(prepare_release(requested, current, releases_path))
        elif command == "self-test" and len(argv) == 2:
            self_test()
        else:
            raise ValueError(
                "usage: release-semver.py "
                "{validate VERSION|newer NEW OLD|prepare NEW CURRENT RELEASES_FILE|self-test}"
            )
    except (IndexError, OSError, ValueError) as error:
        print(error, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
