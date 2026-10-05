#!/usr/bin/env python3
"""Lists Recuerdo's translatable texts in data/translations/recuerdo.pot:
every i18n.T("...") in the Go code, and the names, descriptions and
categories of registered settings (translated where the settings dialog
shows them). Then reports per language how many are translated, by
OpenTeacher's translations (with internal/i18n's aliases and "..." rule)
and by Recuerdo's own recuerdo-<lang>.po.

    python3 scripts/extract_strings.py [--missing LANG]
"""
import glob
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from merge_translations import parse, quote  # noqa: E402

ROOT = Path(__file__).resolve().parent.parent
TR = ROOT / "data/translations"
GOSTR = r'"((?:[^"\\]|\\.)*)"'


def gounquote(s):
    return s.encode().decode("unicode_escape").encode("latin-1").decode("utf-8")


def texts():
    found = {}
    for f in sorted(glob.glob(str(ROOT / "internal/**/*.go"), recursive=True)):
        if f.endswith("_test.go"):
            continue
        src = open(f, encoding="utf-8").read()
        rel = str(Path(f).relative_to(ROOT))
        for m in re.finditer(r'i18n\.Tf?\(' + GOSTR + r'((?:\s*\+\s*' + GOSTR + r')*)', src):
            parts = [m.group(1)] + re.findall(GOSTR, m.group(2))
            found.setdefault(gounquote("".join(parts)), rel)
        # values shown translated (valuecombo): marked constant blocks,
        # and the note calculators' display names
        for m in re.finditer(r'// i18n:values[^\n]*\n(?:const|var) \((.*?)\n\)', src, re.S):
            for v in re.findall(r'=\s*' + GOSTR, m.group(1)):
                found.setdefault(gounquote(v), rel)
        for m in re.finditer(r'\) DisplayName\(\) string\s*\{\s*return ' + GOSTR, src):
            found.setdefault(gounquote(m.group(1)), rel)
        for m in re.finditer(r'\bcombo\(' + GOSTR, src):
            found.setdefault(gounquote(m.group(1)), rel)
        for m in re.finditer(r'settingsdefs\.Def\{(.*?)\n\t\}\)', src, re.S):
            for field in ("Category", "Name", "Help"):
                fm = re.search(field + r':\s*' + GOSTR + r'((?:\s*\+\s*' + GOSTR + r')*)', m.group(1))
                if fm:
                    parts = [fm.group(1)] + re.findall(GOSTR, fm.group(2))
                    found.setdefault(gounquote("".join(parts)), rel)
    # the formula builder's group names and button tips
    pal = (ROOT / "internal/richtext/palette.go").read_text(encoding="utf-8")
    for m in re.finditer(r'^\t\{"(\w+)", (?:\[\]PaletteItem\{|greek\(\))', pal, re.M):
        found.setdefault(m.group(1), "internal/richtext/palette.go")
    for m in re.finditer(r'\{"\w+", "[^"]*", (?:`[^`]*`|"(?:[^"\\]|\\.)*"), "([^"]+)"\}', pal):
        found.setdefault(m.group(1), "internal/richtext/palette.go")
    # the web version: texts marked data-i18n / data-i18n-placeholder in
    # web/index.html, and t("...") in web/app.js
    import html as htmlmod
    page = (ROOT / "web/index.html").read_text(encoding="utf-8")
    for m in re.finditer(r'<(\w+)[^>]*\bdata-i18n\b[^>]*>(.*?)</\1>', page, re.S):
        found.setdefault(htmlmod.unescape(m.group(2).strip()), "web/index.html")
    for m in re.finditer(r'data-i18n-placeholder="([^"]*)"', page):
        found.setdefault(htmlmod.unescape(m.group(1)), "web/index.html")
    js = (ROOT / "web/app.js").read_text(encoding="utf-8")
    for m in re.finditer(r'\bt\(' + GOSTR, js):
        found.setdefault(gounquote(m.group(1)), "web/app.js")
    for c in ("Practice", "Results", "Files", "Interface", "System language"):
        found.setdefault(c, "internal/settingsdefs")
    return found


def aliases():
    src = open(ROOT / "internal/i18n/aliases.go", encoding="utf-8").read()
    return {gounquote(a): gounquote(b) for a, b in re.findall(GOSTR + r':\s*' + GOSTR, src)}


def translated(text, ot, own, al):
    if text in own:
        return True
    for t in (text, text[:-3] if text.endswith("...") else None, text[:-1] if text.endswith("…") else None):
        if t and (t in ot or (t in al and al[t] in ot)):
            return True
    return False


def main():
    found = texts()
    with open(TR / "recuerdo.pot", "w", encoding="utf-8") as f:
        f.write("# Recuerdo's translatable texts, made by scripts/extract_strings.py.\n")
        f.write("# Translate them in data/translations/recuerdo-<lang>.po.\n")
        f.write('msgid ""\nmsgstr "Content-Type: text/plain; charset=UTF-8\\n"\n\n')
        for t in sorted(found):
            f.write(f"#: {found[t]}\nmsgid {quote(t)}\nmsgstr \"\"\n\n")
    al = aliases()
    print(f"{len(found)} texts")
    missing_for = sys.argv[sys.argv.index("--missing") + 1] if "--missing" in sys.argv else None
    for po in sorted(TR.glob("*.po")):
        lang = po.stem
        if lang.startswith("recuerdo-"):
            continue
        ot = parse(po)
        own_path = TR / f"recuerdo-{lang}.po"
        own = parse(own_path) if own_path.exists() else {}
        if lang.startswith("en_"):
            # British and Australian English: Recuerdo's English already
            # uses British spelling (practise, cancelled), so the English
            # source is their text where OpenTeacher's file has none
            print(f"{lang:6} {len(found):4}/{len(found)} (100%) English source")
            continue
        n = sum(translated(t, ot, own, al) for t in found)
        print(f"{lang:6} {n:4}/{len(found)} ({100 * n // len(found)}%)" + (" + own" if own else ""))
        if lang == missing_for:
            for t in sorted(found):
                if not translated(t, ot, own, al):
                    print("   missing:", repr(t))


if __name__ == "__main__":
    main()
