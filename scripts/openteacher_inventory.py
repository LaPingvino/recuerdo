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


RANK = ["missing", "scaffold", "partial", "untested", "working"]  # plus "covered", "dropped", "test suite"  # "covered": see central_format and COVERED


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


EXT = re.compile(r"""["'](\w+)["']\s*:""")


def declared_extensions(d, attr):
    """File extensions a legacy loader/saver declares in self.loads/self.saves."""
    exts = set()
    for py in d.glob("*.py"):
        src = py.read_text(errors="replace")
        # a multi-line dict, or one on a single line ({"stp": ["words"]})
        m = re.search(r"self\." + attr + r"\s*=\s*\{(.*?)\n\t\t\}", src, re.S) or \
            re.search(r"self\." + attr + r"\s*=\s*\{([^\n]*)\}", src)
        if m:
            exts.update(e for e in EXT.findall(m.group(1)) if e not in ("words", "topo", "media"))
    return exts


def central_format(rel):
    """Loaders/savers are implemented in internal/lesson rather than per module.
    Returns the internal/lesson file handling one of the module's extensions."""
    parts = rel.parts
    # OpenTeacher's generic loader (choosing the load module for a file) is
    # FileLoader.LoadFile itself
    if parts == ("logic", "loader"):
        return "internal/lesson"
    # and the generic saver is FileSaver.SaveFile with export.Save
    if parts == ("logic", "saver"):
        return "internal/lesson"
    if len(parts) != 3 or parts[:2] not in (("logic", "loaders"), ("logic", "savers")):
        return ""
    loader = parts[1] == "loaders"
    exts = declared_extensions(LEGACY / rel, "loads" if loader else "saves")
    exts.add(parts[2].rstrip("_").lower())
    target = LESSON / ("loader.go" if loader else "saver.go")
    src = target.read_text(errors="replace") if target.exists() else ""
    # only the dispatching function: LoadFile / SaveFile
    fn = "LoadFile" if loader else "SaveFile"
    m = re.search(r"\nfunc \([^)]*\) " + fn + r"\(.*?\n}\n", src, re.S)
    src = m.group(0) if m else ""
    for e in sorted(exts):
        # only formats the load/save switch dispatches on, not comments or names
        if re.search(r'^\s*case [^\n]*"\.' + re.escape(e) + r'"', src, re.I | re.M):
            return "internal/lesson"
    if not loader:
        # PDF, ODT and LibreOffice formats: the Qt export package
        export = (ROOT / "internal/modules/interfaces/qt/export/export.go")
        esrc = export.read_text(errors="replace") if export.exists() else ""
        for e in sorted(exts):
            if re.search(r'"\.' + re.escape(e) + r'"', esrc, re.I):
                return "internal/modules/interfaces/qt/export"
    return ""


# Areas removed from the Go port on purpose (web version, packaging and
# generator tooling, classroom test mode, web services): see
# docs/OPENTEACHER_INVENTORY.md.
DROPPED = (
    "javaScript", "interfaces/qt/testMode", "interfaces/qt/webServices",
    "interfaces/webServicesServer", "logic/webDatabase", "logic/moduleGraphBuilder",
    "profileRunners/packagers", "profileRunners/backgroundImageGenerator",
    "profileRunners/businessCardGenerator", "profileRunners/codeComplexity",
    "profileRunners/getTranslationAuthors", "profileRunners/languageCodeGuesserTableGenerator",
    "profileRunners/moduleGraph", "profileRunners/rosettaUpdater", "profileRunners/translationUpdater",
    "profileRunners/ircBot", "profileRunners/webServicesServerRunner", "profileRunners/gtkGui",
    "data/profileDescriptions/codeComplexity", "data/profileDescriptions/codeDocumentation",
    "data/profileDescriptions/generate", "data/profileDescriptions/getTranslationAuthors",
    "data/profileDescriptions/ircBot", "data/profileDescriptions/moduleGraph",
    "data/profileDescriptions/package", "data/profileDescriptions/update",
    "data/profileDescriptions/webServicesServer",
)


# Modules whose job Recuerdo does elsewhere (not in a module of their own),
# with where and why. Listed in docs/OPENTEACHER_INVENTORY.md.
COVERED = {
    "data/chars/greek": ("interfaces/qt/lessons/words/unicode_picker.go",
                         "OpenTeacher's Greek table is a built-in set of the special characters picker"),
    "data/chars/cyrillic": ("interfaces/qt/lessons/words/unicode_picker.go",
                            "OpenTeacher's Cyrillic table is a built-in set of the special characters picker"),
    "data/chars/symbols": ("interfaces/qt/lessons/words/unicode_picker.go",
                           "the accented letters of OpenTeacher's symbols table are the picker's Latin Accents "
                           "set; data/character_sets.json adds per-language sets"),
    "data/maps/africa": ("internal/maps", "the map picture and places in data/maps/africa, loaded by MapManager"),
    "data/maps/asia": ("internal/maps", "the map picture and places in data/maps/asia, loaded by MapManager"),
    "data/maps/europe": ("internal/maps", "the map picture and places in data/maps/europe, loaded by MapManager"),
    "data/maps/latinamerica": ("internal/maps", "the map picture and places in data/maps/latinamerica, "
                               "loaded by MapManager"),
    "data/maps/usa": ("internal/maps", "the map picture and places in data/maps/usa, loaded by MapManager"),
    "data/maps/world": ("internal/maps", "the map picture and places in data/maps/world, loaded by MapManager"),
    "data/metadata": ("internal/modules/metadata.go", "Recuerdo's own metadata module (name, version, "
                      "application ID)"),
}


def dropped(rel):
    r = str(rel)
    return any(r.startswith(d) or "/" + d in "/" + r for d in DROPPED)


def rows():
    for d in legacy_modules():
        rel = d.relative_to(LEGACY)
        types, deps, pylines = module_info(d)
        st = best_status(rel)
        central = central_format(rel)
        if (central or str(rel) in COVERED) and st["status"] in ("missing", "scaffold"):
            st["status"] = "covered"
        if dropped(rel):
            st["status"] = "dropped"
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
