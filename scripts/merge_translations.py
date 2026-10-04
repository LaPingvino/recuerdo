#!/usr/bin/env python3
"""Merges OpenTeacher's translations (a translations/<lang>.po in each of
its modules under legacy/) into one file per language,
data/translations/<lang>.po, which internal/i18n reads.

Fuzzy and empty translations are left out; when modules translate the
same text differently, the most common translation wins.

    python3 scripts/merge_translations.py
"""
import collections
import glob
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
LEGACY = ROOT / "legacy/modules/org/openteacher"
OUT = ROOT / "data/translations"
MIN = 10  # languages with fewer translated texts are left out (mus: 1)


def unquote(s):
    s = s.strip()
    assert s.startswith('"') and s.endswith('"'), s
    out, i = [], 1
    while i < len(s) - 1:
        c = s[i]
        if c == "\\":
            i += 1
            out.append({"n": "\n", "t": "\t", '"': '"', "\\": "\\"}.get(s[i], s[i]))
        else:
            out.append(c)
        i += 1
    return "".join(out)


def parse(path):
    entries, msgid, msgstr, field, fuzzy = {}, None, None, None, False

    def flush():
        if msgid and msgstr and not fuzzy:
            entries[msgid] = msgstr

    for line in open(path, encoding="utf-8", errors="replace"):
        line = line.rstrip("\n")
        if line.startswith("#,") and "fuzzy" in line:
            fuzzy = True
        elif line.startswith("msgid "):
            flush()
            msgid, msgstr, field = unquote(line[6:]), None, "id"
        elif line.startswith("msgstr "):
            msgstr, field = unquote(line[7:]), "str"
        elif line.startswith('"'):
            if field == "id":
                msgid += unquote(line)
            elif field == "str":
                msgstr += unquote(line)
        elif not line.strip():
            flush()
            msgid, msgstr, field, fuzzy = None, None, None, False
    flush()
    return entries


def quote(s):
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n").replace("\t", "\\t") + '"'


def main():
    votes = collections.defaultdict(lambda: collections.defaultdict(collections.Counter))
    for path in glob.glob(str(LEGACY / "**/translations/*.po"), recursive=True):
        lang = os.path.basename(path)[:-3]
        for k, v in parse(path).items():
            votes[lang][k][v] += 1
    OUT.mkdir(parents=True, exist_ok=True)
    for lang, entries in sorted(votes.items()):
        if len(entries) < MIN:
            continue
        with open(OUT / f"{lang}.po", "w", encoding="utf-8") as f:
            f.write("# OpenTeacher's translations, merged by scripts/merge_translations.py.\n")
            f.write('msgid ""\nmsgstr "Content-Type: text/plain; charset=UTF-8\\n"\n\n')
            for k in sorted(entries):
                f.write(f"msgid {quote(k)}\nmsgstr {quote(entries[k].most_common(1)[0][0])}\n\n")
        print(lang, len(entries))


if __name__ == "__main__":
    main()
