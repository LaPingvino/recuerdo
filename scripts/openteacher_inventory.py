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

# (placeholder as a word: not Qt's SetPlaceholderText)
STUB = re.compile(r"TODO|FIXME|not (yet )?implemented|\bstub\b|\bplaceholder\b", re.I)
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


RANK = ["missing", "scaffold", "partial", "untested", "working"]  # plus "covered", "planned", "dropped", "test suite"  # "covered": see central_format and COVERED


def ancestor_tests(d):
    """Tests in a parent package that import d's package: the lesson
    types, list modifiers and note calculators are tested together that
    way."""
    if any(d.glob("*_test.go")):
        return []
    imp = '"github.com/LaPingvino/recuerdo/' + str(d.relative_to(ROOT)) + '"'
    found, p = [], d.parent
    while p != GO.parent:
        found += [t for t in p.glob("*_test.go") if imp in t.read_text(errors="replace")]
        p = p.parent
    return found


def go_candidates(rel):
    """Where a module's Go code may live: its own directory, or a file
    named after it at the top of internal/modules (the hand-written core
    modules such as settings.go and event.go live there)."""
    cands = []
    d = go_dir(rel)
    if d is not None:
        cands.append(sorted(d.glob("*.go")) + ancestor_tests(d))
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
    "interfaces/qt/webServices",
)

# Areas Recuerdo will port later (decided with the user, 2026-10-04):
# module path prefix -> (priority, what and why).
TESTMODE = ("medium", "test mode: classroom tests run from a server. OpenTeacher's server is in legacy/ "
            "(Flask web services on a CouchDB web database, about 800 lines): convert it to Go, or build a new one")
DEVTOOL = ("low", "developer tooling: useful later as independent command line tools")
PLANNED = {
    "javaScript": ("medium", "the in-browser version: lessons and practice in a web browser, which also makes "
                   "exercises possible that rely on HTML"),
    "interfaces/qt/testMode": TESTMODE,
    "interfaces/webServicesServer": TESTMODE,
    "logic/webDatabase": TESTMODE,
    "profileRunners/webServicesServerRunner": TESTMODE,
    "data/profileDescriptions/webServicesServer": TESTMODE,
    "logic/spellChecker": ("medium", "spell checking while entering words, with Hunspell (OpenTeacher used Enchant)"),
    "logic/interfaces/typingTutorModel": ("medium", "OpenTeacher's touch typing course, a lesson of its own kind"),
    "interfaces/qt/typingTutor": ("medium", "the touch typing course's screen and keyboard"),
    "interfaces/qt/theme": ("low", "a dark theme, as a setting"),
    "data/profileDescriptions/wordsOnly": ("low", "a setting that hides topography and media lessons "
                                           "(\"just gimme my good old OpenTeacher 2.x\")"),
    "logic/moduleGraphBuilder": DEVTOOL,
    "profileRunners/backgroundImageGenerator": DEVTOOL,
    "profileRunners/businessCardGenerator": DEVTOOL,
    "profileRunners/codeComplexity": DEVTOOL,
    "profileRunners/getTranslationAuthors": DEVTOOL,
    "profileRunners/languageCodeGuesserTableGenerator": DEVTOOL,
    "profileRunners/moduleGraph": DEVTOOL,
    "profileRunners/rosettaUpdater": DEVTOOL,
    "profileRunners/translationUpdater": DEVTOOL,
    "profileRunners/ircBot": DEVTOOL,
    "profileRunners/gtkGui": ("low", "an alternative GTK interface OpenTeacher experimented with"),
    "data/profileDescriptions/codeComplexity": DEVTOOL,
    "data/profileDescriptions/codeDocumentation": DEVTOOL,
    "data/profileDescriptions/generate": DEVTOOL,
    "data/profileDescriptions/getTranslationAuthors": DEVTOOL,
    "data/profileDescriptions/ircBot": DEVTOOL,
    "data/profileDescriptions/moduleGraph": DEVTOOL,
    "data/profileDescriptions/update": DEVTOOL,
}

# Areas whose job Recuerdo does another way: prefix -> (where, why).
COVERED_AREAS = {
    "profileRunners/packagers": ("packaging/, .github/workflows/release.yml", "Recuerdo is built by Go and "
                                 "released by CI; the Arch package is a PKGBUILD"),
    "data/profileDescriptions/package": ("packaging/, .github/workflows/release.yml", "the packaging profiles, "
                                         "with the packagers"),
}


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
    "data/profileDescriptions/all": ("cmd/recuerdo", "Recuerdo always starts with all its features, "
                                     "which is what OpenTeacher's \"all\" profile chose"),
    "logic/otxxloader": ("internal/lesson", "FileLoader reads .otwd/.ottp/.otmd zips (list.json and resources)"),
    "logic/otxxsaver": ("internal/lesson", "FileSaver writes .otwd/.ottp/.otmd zips"),
    "logic/sylkSaver": ("internal/lesson", "FileSaver.saveSYLKFile"),
    "logic/odtsaver": ("internal/modules/interfaces/qt/export", "ODT is written by Qt from the HTML export"),
    "logic/userDocumentationWrapper": ("internal/modules/interfaces/qt/gui", "Help > Getting Started shows the guide"),
    "logic/htmlGenerator/words": ("internal/lesson", "FileSaver's HTML export (also the source of PDF/ODT)"),
    "logic/htmlGenerator/topo": ("internal/lesson", "FileSaver's HTML export"),
    "logic/htmlGenerator/media": ("internal/lesson", "FileSaver's HTML export"),
    "logic/modules": ("internal/core", "Recuerdo's module manager (registration, types, default modules)"),
    "logic/execute": ("internal/core, internal/modules/execute.go", "start-up and enabling modules"),
    "logic/wordListString/parser": ("internal/lesson", "ParseWordList (\"q = a\" / tab lines with \\= escapes)"),
    "logic/wordListString/composer": ("internal/lesson", "ComposeWordList"),
    "logic/dataStore": ("internal/modules/settings.go", "the settings module's JSON file is Recuerdo's persistent store"),
    "logic/settingsFilterer": ("internal/modules/interfaces/qt/dialogs/settings", "the settings dialog lays its "
                               "settings out in fixed tabs instead of grouping them by category"),
    "logic/authors": ("internal/modules/data/openteacherAuthors", "the credits list in the About dialog"),
    "logic/languageCodeGuesser": ("internal/langcode", "Guess and Name map language names and ISO 639-1 codes "
                                  "(CLDR names via golang.org/x/text instead of Babel's tables)"),
    "logic/ocr/tesseractRecognizer": ("internal/ocr", "runs tesseract for hOCR"),
    "logic/ocr/wordListLoader": ("internal/ocr", "LoadWordList: hOCR lines to rows and columns to word pairs"),
    "logic/itemModifiers/foreignKnown": ("internal/teaching", "the Teach tab's \"Ask the answers\" (Options.AskAnswers) "
                                         "swaps questions and answers"),
    "logic/mergers/words": ("internal/lesson", "Merge, used by File > Merge Lesson"),
    "logic/testTypes/words": ("internal/teaching", "Report.MostDoneWrong, the \"word most done wrong\" fact in the "
                              "results dialog"),
    "logic/interfaces/buttonRegister": ("internal/modules/buttonregister.go", "Recuerdo's own button register"),
    "logic/interfaces/inputTypingLogic": ("internal/teaching", "Typing checks typed answers and shows corrections"),
    "logic/interfaces/lessonTracker": ("internal/modules/interfaces/qt/gui", "the GUI keeps the lesson of each tab "
                                       "(tabLessons) and the current one (currentLesson)"),
    "interfaces/qt/enterers/words": ("internal/modules/interfaces/qt/lessons/words", "the Enter tab of a word lesson"),
    "interfaces/qt/teachers/words": ("internal/modules/interfaces/qt/lessons/words", "the Teach tab of a word lesson"),
    "interfaces/qt/teachTypes/typing": ("internal/modules/interfaces/qt/lessons/words", "Typing mode of the Teach tab (internal/teaching), tested offscreen"),
    "interfaces/qt/teachTypes/shuffleAnswer": ("internal/modules/interfaces/qt/lessons/words", "Shuffle answer mode of the Teach tab, tested offscreen"),
    "interfaces/qt/teachTypes/repeatAnswer": ("internal/modules/interfaces/qt/lessons/words", "Repeat answer mode (with fading) of the Teach tab, tested offscreen"),
    "interfaces/qt/teachTypes/inMind": ("internal/modules/interfaces/qt/lessons/words", "In mind mode of the Teach tab, tested offscreen"),
    "interfaces/qt/teachTypes/hangman": ("internal/modules/interfaces/qt/lessons/words", "Hangman mode of the Teach tab, tested offscreen"),
    "interfaces/qt/startWidget": ("internal/modules/interfaces/qt/gui", "the start screen (New Lesson, Open Lesson)"),
    "interfaces/qt/testViewer": ("internal/modules/interfaces/qt/lessons/words", "the Results tab and results dialog show a session's answers"),
    "interfaces/qt/testsViewer": ("internal/modules/interfaces/qt/lessons/words", "the Results tab lists the results; charts: see progressViewer"),
    "interfaces/qt/dialogShower": ("internal/modules/interfaces/qt/gui", "the GUI shows its dialogs itself"),
    "interfaces/qt/loaderGui": ("internal/modules/interfaces/qt/gui", "File > Open and the open dialog"),
    "interfaces/qt/inputTyping": ("internal/modules/interfaces/qt/lessons/words", "the Teach tab's answer field"),
    "interfaces/qt/charsKeyboard": ("internal/modules/interfaces/qt/lessons/words", "the special characters picker (an unused, unregistered Go version was removed)"),
    "interfaces/qt/dialogs/documentation": ("internal/modules/interfaces/qt/gui", "Help > Getting Started"),
    "interfaces/qt/recentlyOpenedViewer": ("internal/modules/interfaces/qt/gui", "File > Open Recent"),
    "interfaces/qt/settingsWidgets": ("internal/modules/interfaces/qt/dialogs/settings", "the settings dialog and Teach tab build their own controls"),
    "interfaces/qt/settingsWidget/boolean": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/characterTable": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/language": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/longText": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/multiOption": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/number": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/option": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/password": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/profile": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/settingsWidget/shortText": ("internal/modules/interfaces/qt/dialogs/settings", "see settingsWidgets"),
    "interfaces/qt/dialogs/print": ("internal/modules/interfaces/qt/gui", "File > Print shows Qt's print dialog"),
    "interfaces/qt/printer": ("internal/modules/interfaces/qt/export", "Print prints the HTML export on a QPrinter"),
    "interfaces/qt/print/words": ("internal/modules/interfaces/qt/export", "Print: a word list's HTML export, named after "
                                  "the lesson, tested by printing to PDF"),
    "interfaces/qt/percentNotesViewer": ("internal/modules/interfaces/qt/charts", "GradesChart: a bar per session "
                                         "with its percentage, on the Results tab"),
    "interfaces/qt/progressViewer": ("internal/modules/interfaces/qt/charts", "TimelineChart: the last session's "
                                     "answers over time, on the Results tab"),
    "interfaces/qt/ocrGui": ("internal/modules/interfaces/qt/ocrimport", "File > Import from Picture: one dialog "
                             "instead of the wizard (straighten, crop, read with internal/ocr); tested offscreen "
                             "with a real Tesseract run"),
    "interfaces/qt/enterers/topo": ("internal/modules/interfaces/qt/lessons/topo", "Enter tab: click the map or "
                                    "type a known place; rename and remove in the list (tested offscreen)"),
    "interfaces/qt/teachers/topo": ("internal/modules/interfaces/qt/lessons/topo", "Teach tab: Place – Name and "
                                    "Name – Place on teaching.Session, results kept in the lesson (tested offscreen)"),
    "interfaces/qt/topoMaps": ("internal/modules/interfaces/qt/lessons/topo", "BundledMaps reads data/maps "
                               "(OpenTeacher's six maps with their known places, all their names); tested"),
    "interfaces/qt/print/topo": ("internal/modules/interfaces/qt/export", "Print of a topography lesson prints its "
                                 "map with the places, fitted to the page (tested by printing to PDF)"),
    "logic/savers/png": ("internal/modules/interfaces/qt/export", "Save as .png writes a topography lesson's map with "
                         "its places (MapPicture; tested), also as PDF"),
    "interfaces/qt/enterers/media": ("internal/modules/interfaces/qt/lessons/media", "Enter tab: add files (embedded) and web addresses, edit name/question/answer in a table, preview (tested offscreen)"),
    "interfaces/qt/teachers/media": ("internal/modules/interfaces/qt/lessons/media", "Teach tab: the item and its question, a typed answer on teaching.Session; results kept (tested offscreen)"),
    "interfaces/qt/mediaDisplay": ("internal/modules/interfaces/qt/lessons/media", "Preview: pictures and texts shown in place, other media opened in the system's player or browser"),
    "interfaces/qt/mediaTypes/image": ("internal/modules/interfaces/qt/lessons/media", "Kind: picture (OpenTeacher's extensions and more), shown in place"),
    "interfaces/qt/mediaTypes/text": ("internal/modules/interfaces/qt/lessons/media", "Kind: text (.txt), shown in place"),
    "interfaces/qt/mediaTypes/audio": ("internal/modules/interfaces/qt/lessons/media", "Kind: sound; played by the system's player (no QtMultimedia dependency)"),
    "interfaces/qt/mediaTypes/video": ("internal/modules/interfaces/qt/lessons/media", "Kind: video; played by the system's player (no QtMultimedia dependency)"),
    "interfaces/qt/mediaTypes/website": ("internal/modules/interfaces/qt/lessons/media", "Kind: website (http and https); opened in the web browser"),
    "interfaces/qt/mediaTypes/youtube": ("internal/modules/interfaces/qt/lessons/media", "Kind: YouTube video (watch and youtu.be links); opened in the web browser"),
    "interfaces/qt/mediaTypes/vimeo": ("internal/modules/interfaces/qt/lessons/media", "Kind: Vimeo video; opened in the web browser"),
    "interfaces/qt/mediaTypes/dailymotion": ("internal/modules/interfaces/qt/lessons/media", "Kind: Dailymotion video; opened in the web browser"),
    "interfaces/qt/print/media": ("internal/modules/interfaces/qt/export", "Print and PDF of a media lesson: a table "
                                  "of medium (pictures as thumbnails), name, question and answer (tested via pdftotext)"),
    "interfaces/textToSpeech/impl": ("internal/tts", "Speak with espeak-ng/espeak (Linux, BSD), say (macOS) or "
                                     "System.Speech (Windows) instead of pyttsx; command lines tested"),
    "interfaces/textToSpeech/providers/words": ("internal/modules/interfaces/qt/lessons/words", "Teach option "
                                                "\"Pronounce questions\" (setting kept), in the question language; tested"),
    "interfaces/textToSpeech/providers/topo": ("internal/modules/interfaces/qt/lessons/topo", "Teach option \"Pronounce "
                                               "names\": Name – Place says the place to click; tested"),
    "profileRunners/cli": ("internal/cli", "recuerdo <command>: authors, convert, merge, reverse-list, view-word-list, "
                           "ocr-word-list, new-word-list, practise-word-list (-flags for +flags); tested"),
    "profileRunners/profilesHelp": ("internal/cli", "recuerdo help lists the commands (and recuerdo -help the options)"),
    "data/profileDescriptions/cli": ("internal/cli", "the command line is reached with recuerdo <command>, not a profile"),
    "data/profileDescriptions/help": ("internal/cli", "recuerdo help"),
    "profileRunners/uiController": ("internal/modules/interfaces/qt/gui", "the GUI module connects the Qt interface to "
                                    "loading, saving, printing, dialogs and the lessons itself"),
    "logic/testTypes/topo": ("internal/modules/interfaces/qt/lessons/topo", "the Results tab shows a topography "
                             "lesson's sessions (charts); the result table model was not used by anything"),
    "logic/testTypes/media": ("internal/modules/interfaces/qt/lessons/media", "the Results tab shows a media lesson's "
                              "sessions (charts); the result table model was not used by anything"),
    "logic/translator": ("internal/i18n", "OpenTeacher's gettext translations (29 languages, merged into "
                         "data/translations by scripts/merge_translations.py) read by a small .po reader; T() with "
                         "aliases for Recuerdo's wording; the language is a setting, the system's by default"),
    "logic/friendlyTranslationNames": ("internal/i18n", "Name: each language in its own name (golang.org/x/text "
                                       "display), in the settings dialog's language choice"),
    "data/metadata": ("internal/modules/metadata.go", "Recuerdo's own metadata module (name, version, "
                      "application ID)"),
}


# Why the modules dropped one by one (beyond the areas in DROPPED) were.
DROPPED_REASONS = {
    "profileRunners/shell": "an interactive Python shell with OpenTeacher's modules loaded, for developers; "
                            "Go has no such shell, and Recuerdo's modules are used from Go code and tests",
    "data/profileDescriptions/shell": "describes the Python shell profile, dropped with it",
    "misc/testUrllibMock": "a stand-in for Python's urllib in OpenTeacher's tests; Go tests use net/http/httptest",
    "interfaces/qt/mediaTypes/liveleak": "LiveLeak closed in 2021: its video links no longer work",
    "interfaces/qt/hiddenBrowser": "a hidden web browser (an easter egg) nothing else used",
    "logic/safeHtmlChecker": "only OpenTeacher's web database (dropped) used it",
    "logic/translationIndex/builder": "OpenTeacher's translation tooling (building its translation index)",
    "logic/translationIndex/jsonWriter": "OpenTeacher's translation tooling",
    "logic/translationIndex/merger": "OpenTeacher's translation tooling",
    "logic/pyinstallerInterface": "Python packaging; Recuerdo is built by Go and released by CI",
    "logic/sourceSaver": "Python source releases; Recuerdo's source is its Git repository",
    "logic/sourceWithSetupSaver": "Python source releases with setup.py; see sourceSaver",
    "logic/ocr/cuneiformRecognizer": "Cuneiform is no longer developed; Tesseract (internal/ocr) does OCR",
    "data/profileDescriptions/selfstudy": "OpenTeacher's start-up profiles only chose which GUI modules to load "
                                          "for an audience; Recuerdo has one, smaller feature set",
    "data/profileDescriptions/studentAtHome": "as selfstudy: an audience profile",
    "data/profileDescriptions/studentAtSchool": "as selfstudy: an audience profile",
    "data/profileDescriptions/teacher": "as selfstudy: an audience profile",
}


def prefixed(rel, table):
    r = str(rel)
    for p, v in table.items():
        # as for DROPPED: a prefix of the path or of a name in it
        # ("package" covers packageArch, "javaScript" javaScriptWords)
        if r.startswith(p) or "/" + p in "/" + r:
            return v
    return None


def dropped(rel):
    r = str(rel)
    return r in DROPPED_REASONS or any(r.startswith(d) or "/" + d in "/" + r for d in DROPPED)


def rows():
    for d in legacy_modules():
        rel = d.relative_to(LEGACY)
        types, deps, pylines = module_info(d)
        st = best_status(rel)
        central = central_format(rel)
        if (central or str(rel) in COVERED) and st["status"] in ("missing", "scaffold"):
            st["status"] = "covered"
        if prefixed(rel, COVERED_AREAS):
            st["status"] = "covered"
        if dropped(rel):
            st["status"] = "dropped"
        if prefixed(rel, PLANNED):
            st["status"] = "planned"
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
