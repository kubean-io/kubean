#!/usr/bin/env python3
"""Backport ansible/ansible#87842 to Ansible distribution.py.

Classifies Anolis, Anolis OS, and Alibaba Cloud Linux under the RedHat OS
family in ansible-core releases prior to 2.23, while remaining a no-op when
upstream already includes those entries.
"""
from __future__ import annotations

import ast
from pathlib import Path
import sys

REDHAT_DISTRIBUTIONS = ("Anolis", "Anolis OS", "Alibaba Cloud Linux")


def _find_redhat_family_list(source: str) -> ast.List:
    tree = ast.parse(source)
    for node in tree.body:
        if not isinstance(node, ast.ClassDef) or node.name != "Distribution":
            continue
        for stmt in node.body:
            if not isinstance(stmt, ast.Assign):
                continue
            if not any(isinstance(t, ast.Name) and t.id == "OS_FAMILY_MAP" for t in stmt.targets):
                continue
            if not isinstance(stmt.value, ast.Dict):
                raise RuntimeError("Distribution.OS_FAMILY_MAP is not a dict literal")
            for key, value in zip(stmt.value.keys, stmt.value.values):
                if isinstance(key, ast.Constant) and key.value == "RedHat":
                    if not isinstance(value, ast.List) or not value.elts:
                        raise RuntimeError("Distribution.OS_FAMILY_MAP['RedHat'] is not a non-empty list")
                    return value
    raise RuntimeError("Could not locate Distribution.OS_FAMILY_MAP['RedHat'] in distribution.py")


def _char_offset(source: str, lineno: int, col_byte_offset: int) -> int:
    lines = source.splitlines(keepends=True)
    line_start = sum(len(line) for line in lines[: lineno - 1])
    line_prefix = lines[lineno - 1].encode("utf-8")[:col_byte_offset].decode("utf-8")
    return line_start + len(line_prefix)


def patch_source(source: str) -> tuple[str, bool]:
    """Return (updated_source, changed) after ensuring RedHat OS_FAMILY_MAP entries."""
    redhat_list = _find_redhat_family_list(source)
    existing = {
        elt.value
        for elt in redhat_list.elts
        if isinstance(elt, ast.Constant) and isinstance(elt.value, str)
    }
    missing = [name for name in REDHAT_DISTRIBUTIONS if name not in existing]
    if not missing:
        return source, False

    last_elt = redhat_list.elts[-1]
    if last_elt.end_lineno is None or last_elt.end_col_offset is None:
        raise RuntimeError("AST node is missing end position information")
    insert_at = _char_offset(source, last_elt.end_lineno, last_elt.end_col_offset)
    addition = ", " + ", ".join(repr(name) for name in missing)
    updated = source[:insert_at] + addition + source[insert_at:]

    # Validate syntax and verify all target distributions are present.
    updated_list = _find_redhat_family_list(updated)
    updated_names = {
        elt.value
        for elt in updated_list.elts
        if isinstance(elt, ast.Constant) and isinstance(elt.value, str)
    }
    still_missing = [name for name in REDHAT_DISTRIBUTIONS if name not in updated_names]
    if still_missing:
        raise RuntimeError(f"Failed to insert RedHat family distributions: {still_missing}")
    return updated, True


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <path/to/distribution.py>", file=sys.stderr)
        return 1
    path = Path(argv[1])
    source = path.read_text(encoding="utf-8")
    updated, changed = patch_source(source)
    if changed:
        path.write_text(updated, encoding="utf-8")
        print(f"Patched {path} with {REDHAT_DISTRIBUTIONS}")
    else:
        print(f"{path} already classifies {REDHAT_DISTRIBUTIONS} as RedHat")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
