#!/usr/bin/env python3
"""Regenerates the generated parts of docs/OPENTEACHER_INVENTORY.md from
scripts/openteacher_inventory.py: the status summary table, the "Covered
elsewhere" and "Dropped, and why" tables, and the module table (section 5).

    python3 scripts/update_inventory_doc.py
"""
import collections
import csv
import importlib.util
import io
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DOC = ROOT / "docs/OPENTEACHER_INVENTORY.md"
ORDER = ["working", "untested", "partial", "covered", "scaffold", "missing", "dropped", "test suite"]
AREAS = ["data", "interfaces", "logic", "misc", "profileRunners"]


def main():
    spec = importlib.util.spec_from_file_location("inv", ROOT / "scripts/openteacher_inventory.py")
    inv = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(inv)
    script = ["python3", "scripts/openteacher_inventory.py"]
    rows = list(csv.DictReader(io.StringIO(subprocess.check_output(script + ["--csv"], cwd=ROOT).decode())))
    table = subprocess.check_output(script, cwd=ROOT).decode()

    count = collections.Counter((r["area"], r["status"]) for r in rows)
    total = collections.Counter(r["status"] for r in rows)
    summary = "| Area | " + " | ".join(ORDER) + " | Total |\n|---|" + "---:|" * (len(ORDER) + 1) + "\n"
    for a in AREAS:
        summary += f"| {a} | " + " | ".join(str(count[(a, s)] or "") for s in ORDER) + \
            f" | {sum(count[(a, s)] for s in ORDER)} |\n"
    summary += "| **all** | " + " | ".join(f"**{total[s]}**" for s in ORDER) + f" | **{len(rows)}** |\n"

    reasons = "## Covered elsewhere\n\nOpenTeacher modules whose job Recuerdo does in other code (besides the\n" \
        "file formats, see **covered** above):\n\n| Module | Where | Why |\n|---|---|---|\n"
    for k, (where, why) in sorted(inv.COVERED.items()):
        reasons += f"| `{k}` | `{where}` | {why} |\n"
    reasons += "\n## Dropped, and why\n\nBesides the areas dropped as a whole (the web version, packaging and\n" \
        "generator tooling, test mode, web services; see **dropped** above):\n\n| Module | Why |\n|---|---|\n"
    for k, why in sorted(inv.DROPPED_REASONS.items()):
        reasons += f"| `{k}` | {why} |\n"

    doc = DOC.read_text()
    doc = re.sub(r"\| Area \| working.*?\n\n", summary + "\n", doc, count=1, flags=re.S)
    doc = re.sub(r"## Covered elsewhere\n.*?(?=\n## 5\. Module table)", reasons, doc, count=1, flags=re.S)
    i = doc.index("## 5. Module table")
    DOC.write_text(doc[:i] + "## 5. Module table\n\n" + table)
    print(dict(total))


if __name__ == "__main__":
    main()
