#!/usr/bin/env python3
"""Inventory of the OpenTeacher modules in legacy/ and their Go ports.

For every OpenTeacher module (a directory under
legacy/modules/org/openteacher holding Python code) this prints its
OpenTeacher type, the module types it uses/requires, and how far the Go
counterpart in internal/modules has got, measured from the code itself:
lines of real code, stub markers (TODO, "not implemented", ...) and tests.

Unlike verify_coverage.py, which counts a module as converted as soon as a
Go file exists, this looks at what the Go file contains.

Usage: scripts/openteacher_inventory.py [--csv]
"""

import csv
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
LEGACY = ROOT / "legacy/modules/org/openteacher"
GO = ROOT / "internal/modules"
LESSON = ROOT / "internal/lesson"

STUB = re.compile(r"TODO|FIXME|not (yet )?implemented|\bstub\b|placeholder", re.I)
TYPE = re.compile(r"""self\.type\s*=\s*["'](\w+)["']""")
DEP = re.compile(r"""mods\([^)]*type\s*=\s*["'](\w+)["']""")


def legacy_modules():
    seen = set()
    for py in LEGACY.rglob("*.py"):
        d = py.parent
        if "translations" in d.parts or d in seen:
            continue
        seen.add(d)
    return sorted(seen)


def module_info(d):
    """OpenTeacher type and dependency types from a module's Python code."""
    types, deps, lines = set(), set(), 0
    for py in d.glob("*.py"):
        src = py.read_text(errors="replace")
        lines += sum(1 for l in src.splitlines() if l.strip() and not l.strip().startswith("#"))
        types.update(TYPE.findall(src))
        deps.update(DEP.findall(src))
    return sorted(types), sorted(deps - types), lines


def go_dir(rel):
    exact = GO / rel
    if exact.is_dir():
        return exact
    # case-insensitive match (profileRunners vs profilerunners)
    want = str(rel).lower()
    for d in GO.rglob("*"):
        if d.is_dir() and str(d.relative_to(GO)).lower() == want:
            return d
    return None


def code_lines(src):
    n, in_block = 0, False
    for l in src.splitlines():
        s = l.strip()
        if in_block:
            in_block = "*/" not in s
            continue
        if not s or s.startswith("//"):
            continue
        if s.startswith("/*"):
            in_block = "*/" not in s
            continue
        n += 1
    return n


RANK = ["missing", "scaffold", "partial", "untested", "working"]


def go_candidates(rel):
    """Where a module's Go code may live: its own directory, or a file
    named after it at the top of internal/modules (the hand-written core
    modules such as settings.go and event.go live there)."""
    cands = []
    d = go_dir(rel)
    if d is not None:
        cands.append(sorted(d.glob("*.go")))
    if len(rel.parts) == 2 and rel.parts[0] in ("logic", "misc"):
        base = rel.parts[-1].lower()
        top = [f for f in GO.glob("*.go") if f.stem.lower() in (base, base + "_test")]
        if top:
            cands.append(sorted(top))
    return cands


def best_status(rel):
    best = go_status([])
    for files in go_candidates(rel):
        st = go_status(files)
        if (RANK.index(st["status"]), st["code"]) > (RANK.index(best["status"]), best["code"]):
            best = st
    return best


def go_status(all_files):
    if not all_files:
        return dict(go="", files=0, code=0, stubs=0, tests=0, status="missing")
    files = [f for f in all_files if not f.name.endswith("_test.go")]
    tests = [f for f in all_files if f.name.endswith("_test.go")]
    code = stubs = 0
    for f in files:
        src = f.read_text(errors="replace")
        code += code_lines(src)
        stubs += len(STUB.findall(src))
    if not files:
        status = "missing"
    elif stubs == 0 and tests:
        status = "working"
    elif stubs == 0:
        status = "untested"
    elif code >= 150 and code >= 40 * stubs:
        status = "partial"
    else:
        status = "scaffold"
    where = sorted({str(f.parent.relative_to(ROOT)) if f.parent != GO else str(f.relative_to(ROOT)) for f in files})
    return dict(go=" ".join(where), files=len(files), code=code,
                stubs=stubs, tests=len(tests), status=status)


def central_format(rel):
    """Loaders/savers are implemented in internal/lesson rather than per module."""
    parts = rel.parts
    if len(parts) == 3 and parts[:2] in (("logic", "loaders"), ("logic", "savers")):
        name = parts[2].rstrip("_").lower()
        for f in LESSON.glob("*.go"):
            if f.name.endswith("_test.go"):
                continue
            if re.search(r'"\.' + re.escape(name) + r'"', f.read_text(errors="replace"), re.I):
                return f.name
    return ""


def rows():
    for d in legacy_modules():
        rel = d.relative_to(LEGACY)
        types, deps, pylines = module_info(d)
        st = best_status(rel)
        central = central_format(rel)
        if central and st["status"] in ("missing", "scaffold"):
            st["status"] = "in " + central
        name = rel.parts[-1]
        if name.endswith("Test") or name in ("test", "testRunner", "testserver", "testServer", "testSuite") \
                or "testserver" in rel.parts:
            # OpenTeacher's own test suites; their Go equivalent is _test.go files
            st["status"] = "test suite"
        yield dict(module=str(rel), area=rel.parts[0], type=",".join(types),
                   uses=",".join(deps), py_lines=pylines, **st)


def main():
    data = list(rows())
    if "--csv" in sys.argv:
        w = csv.DictWriter(sys.stdout, fieldnames=list(data[0]))
        w.writeheader()
        w.writerows(data)
        return
    print("| Module | Type | Python lines | Go code lines | Stub markers | Tests | Status |")
    print("|---|---|---:|---:|---:|---:|---|")
    for r in data:
        print(f"| `{r['module']}` | {r['type']} | {r['py_lines']} | {r['code']} | {r['stubs']} | {r['tests']} | {r['status']} |")


if __name__ == "__main__":
    main()
