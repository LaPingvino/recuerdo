# File Format Support in Recuerdo

This document provides a comprehensive overview of all file formats supported by Recuerdo, based on the original OpenTeacher codebase and our Go implementation.

## Lesson Types

Recuerdo supports three main lesson types:

- **words** - Vocabulary/language learning lessons
- **topo** - Geography/topography lessons  
- **media** - Media-based lessons

## File Format Support Matrix

### ✅ Fully Implemented and Working

| Extension | Format Name | Type | Original Loader | Status |
|-----------|-------------|------|-----------------|---------|
| `.csv` | Spreadsheet (CSV) | words | csv_ | ✅ Working |
| `.tsv` | Tab-Separated Values | words | csv_ | ✅ Working |
| `.json` | JSON Lesson File | words | - | ✅ Working |
| `.ot` | OpenTeacher 2.x/3.x | words | ot | ✅ Working |
| `.kvtml` | KDE Vocabulary Document | words | kvtml | ✅ Working |
| `.xml` | XML File (ABBYY Lingvo) | words | abbyy | ✅ Working |
| `.otwd` | OpenTeaching Words | words | otwd | ✅ Working |
| `.kgm` | KGeography Map | topo | kgm | ✅ Working |
| `.ottp` | OpenTeaching Topography | topo | ottp | ✅ Working |
| `.otmd` | OpenTeaching Media | media | otmd | ✅ Working |

### ⚠️ Partially Working (Auto-detection fallback)

| Extension | Format Name | Type | Original Loader | Status |
|-----------|-------------|------|-----------------|---------|
| `.anki2` | Anki 2.0 Database | words | anki2 | ⚠️ Fallback to CSV |
| `.apkg` | Anki Package | words | apkg | ⚠️ Fallback to CSV |
| `.backpack` | Backpack File | words | backpack | ⚠️ Fallback to CSV |
| `.t2k` | Teach2000 File | words | t2k | ⚠️ Fallback to text |

### ❌ Not Yet Implemented

| Extension | Format Name | Type | Original Loader | Status |
|-----------|-------------|------|-----------------|---------|
| `.anki` | Anki Database | words | anki | ❌ Stub only |
| `.wcu` | CueCard File | words | cuecard | ❌ Not implemented |
| `.voc` | Vocabulary File | words | domingo/vocabularium | ❌ Not implemented |
| `.fq` | FlashQard File | words | flashqard | ❌ Not implemented |
| `.fmd` | FM Dictionary | words | fmd | ❌ Not implemented |
| `.dkf` | Granule Deck | words | granule | ❌ Not implemented |
| `.jml` | JMemorize Lesson | words | jml | ❌ Not implemented |
| `.jvlt` | JVLT File | words | jvlt | ❌ Not implemented |
| `.stp` | Ludem File | words | ludem | ❌ Not implemented |
| `.db` | Database File (Mnemosyne) | words | mnemosyne | ❌ Not implemented |
| `.oh` | Overhoor File | words | overhoor | ❌ Not implemented |
| `.ohw` | Overhoor File | words | overhoor | ❌ Not implemented |
| `.oh4` | Overhoor File | words | overhoor | ❌ Not implemented |
| `.ovr` | Overhoringsprogramma Talen | words | ovr | ❌ Not implemented |
| `.pau` | Pauker File | words | pauker | ❌ Not implemented |
| `.vok2` | Teachmaster File | words | teachmaster | ❌ Not implemented |
| `.wdl` | Oriente Voca File | words | voca | ❌ Not implemented |
| `.vtl3` | VokabelTrainer File | words | vokabelTrainer | ❌ Not implemented |
| `.wrts` | WRTS File | words | wrts | ❌ Not implemented |

### 📝 Text Format Support

| Format | Separators | Status |
|--------|------------|---------|
| Basic text | Tab, pipe `|`, equals `=` | ✅ Working |
| GNU VocabTrain | Colon `:` | ❌ Needs colon support |
| VTrain | Various | ❌ Format not analyzed |

## Implementation Status Summary

- **Total formats in original OpenTeacher**: 35+ formats
- **Fully working in Recuerdo**: 10 formats (29%)
- **Partially working**: 4 formats (11%)
- **Not implemented**: 21+ formats (60%)

## Priority Implementation List

Based on usage and complexity, the following formats should be prioritized:

### High Priority (Common formats)
1. **GNU VocabTrain text** - Add colon separator support
2. **Anki/Anki2** - Popular flashcard format
3. **Mnemosyne (.db)** - SQLite database format

### Medium Priority (Niche but useful)
4. **JMemorize (.jml)** - XML-based format
5. **Pauker (.pau)** - XML-based format
6. **JVLT (.jvlt)** - XML-based format

### Low Priority (Legacy/Obsolete)
7. **Teach2000 (.t2k)** - Old Windows format
8. **Various European trainers** (.voc, .wcu, .fq, .fmd, .dkf, etc.)

## Current Loader Architecture

The Go implementation uses a modular loader system:

```go
// File type detection
func (fl *FileLoader) GetFileType(filePath string) string

// Format-specific loaders
func (fl *FileLoader) loadCSV(filePath string) (*LessonData, error)
func (fl *FileLoader) loadKVTMLFile(filePath string) (*LessonData, error)
func (fl *FileLoader) loadOpenTeacherFile(filePath string) (*LessonData, error)
// ... etc
```

## Auto-Detection System

When a file extension is unknown or ambiguous, the loader tries formats in order:

1. **CSV** - Comma/tab separated values
2. **Text** - Various text separators (tab, pipe, equals)
3. **JSON** - JSON lesson format

This fallback system explains why some complex binary formats (like Anki) appear to "work" - they're being parsed as text/CSV with limited success.

## Testing Coverage

- ✅ **Unit tests** for all working formats
- ✅ **Real file tests** using testdata samples
- ✅ **Legacy file tests** for compatibility
- ⚠️ **Missing tests** for unimplemented formats

## Migration Notes

When migrating from the Python OpenTeacher to Go Recuerdo:

1. **CSV/TSV files** - Perfect compatibility
2. **OpenTeacher (.ot) files** - Full compatibility (supports both 2.x and 3.x formats)
3. **KVTML files** - Good compatibility
4. **Complex binary formats** - May need manual conversion

## Adding New Format Support

To add support for a new format:

1. Add extension to `GetFileType()` switch statement
2. Add format name to `GetFormatName()` 
3. Implement `load[Format]File()` method
4. Add to auto-detection chain if needed
5. Add comprehensive tests
6. Update this documentation

## References

- Original OpenTeacher loaders: `legacy/modules/org/openteacher/logic/loaders/`
- Current Go implementation: `internal/lesson/loader.go`
- Test coverage: `internal/lesson/loader_test.go`
