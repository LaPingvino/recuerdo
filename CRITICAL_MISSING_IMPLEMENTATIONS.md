# Critical Missing Implementations - Detailed Analysis

## Overview
Based on comprehensive analysis of the recuerdo codebase, this document identifies specific missing implementations that are critical for full OpenTeacher compatibility.

## 1. SQLite-Based Formats (High Priority)

### Anki 2.0 Database (.anki2)
**Status**: 🔴 **STUB ONLY** - Returns empty lesson
**File Type**: SQLite 3.x database
**Real Example**: `testdata/legacy_files/application_x-anki2.anki.anki2`

**Current Implementation**:
```go
func (fl *FileLoader) loadAnkiFile(filePath string) (*LessonData, error) {
    // Placeholder implementation - would need SQLite support
    lessonData := NewLessonData()
    lessonData.List.Title = filepath.Base(filePath)
    log.Printf("[WARNING] Anki format not fully implemented yet")
    return lessonData, nil  // ❌ RETURNS EMPTY LESSON!
}
```

**Required Implementation**:
```sql
-- Anki database schema (simplified)
SELECT 
    f.flds as fields,  -- Tab-separated question/answer
    n.mod as modified,
    n.tags as tags
FROM notes n 
JOIN cards c ON n.id = c.nid
JOIN col co ON co.id = 1
JOIN field_models f ON f.id = n.mid
WHERE c.queue != -1;  -- Not suspended
```

**Impact**: Anki is the most popular SRS platform worldwide. Missing support blocks major user adoption.

### Mnemosyne Database (.db)
**Status**: 🔴 **COMPLETELY MISSING** - No loader exists
**File Type**: SQLite 3.x database  
**Real Example**: `testdata/legacy_files/application_x-sqlite3.mnemosyne.db`

**Required Implementation**:
```sql
-- Mnemosyne database schema
SELECT 
    cards.question,
    cards.answer, 
    cards.tags,
    cards.grade,
    cards.easiness,
    cards.acq_reps,
    cards.ret_reps
FROM cards 
WHERE active = 1;
```

**Impact**: Popular open-source SRS with dedicated user base.

## 2. XML-Based Formats (Medium Priority)

### Teach2000 Format (.t2k)
**Status**: 🟡 **FAILS PARSING** - Treated as CSV, returns 0 items
**File Type**: XML with UTF-8 encoding
**Real Example**: `testdata/legacy_files/application_x-teach2000.teach2000.t2k`

**Sample Content**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<teach2000>
    <version>852</version>
    <description>Normal</description>
    <message_data>
        <items>
            <item id="0">
                <questions>
                    <question id="0">een</question>
                </questions>
                <answers type="0">
                    <answer id="0">one</answer>
                </answers>
                <errors>0</errors>
                <testcount>4</testcount>
                <correctcount>4</correctcount>
            </item>
        </items>
    </message_data>
</teach2000>
```

**Required Go Structs**:
```go
type Teach2000Root struct {
    XMLName     xml.Name `xml:"teach2000"`
    Version     string   `xml:"version"`
    Description string   `xml:"description"`
    MessageData Teach2000MessageData `xml:"message_data"`
}

type Teach2000MessageData struct {
    Items Teach2000Items `xml:"items"`
}

type Teach2000Items struct {
    Items []Teach2000Item `xml:"item"`
}

type Teach2000Item struct {
    ID        string `xml:"id,attr"`
    Questions struct {
        Questions []string `xml:"question"`
    } `xml:"questions"`
    Answers struct {
        Type    string   `xml:"type,attr"`
        Answers []string `xml:"answer"`
    } `xml:"answers"`
    TestCount    int `xml:"testcount"`
    CorrectCount int `xml:"correctcount"`
}
```

### JMemorize Lesson (.jml)
**Status**: 🔴 **COMPLETELY MISSING** - No loader exists
**File Type**: Compressed XML (ZIP archive containing lesson.xml)
**Real Example**: `testdata/legacy_files/application_x-jmemorizelesson.jmemorize.jml`

**File Structure**:
```
jmemorize.jml (ZIP archive)
├── lesson.xml (main lesson data)
├── images/ (optional image resources)
└── sounds/ (optional audio resources)
```

**Required Implementation**:
```go
func (fl *FileLoader) loadJMemorizeFile(filePath string) (*LessonData, error) {
    // Open ZIP archive
    reader, err := zip.OpenReader(filePath)
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    
    // Find lesson.xml
    for _, f := range reader.File {
        if f.Name == "lesson.xml" {
            rc, err := f.Open()
            if err != nil {
                return nil, err
            }
            defer rc.Close()
            
            // Parse XML content
            return fl.parseJMemorizeXML(rc)
        }
    }
    
    return nil, fmt.Errorf("lesson.xml not found in JMemorize file")
}
```

## 3. Lesson Type Modules (Critical Functionality Gaps)

### Topography Lesson Module
**Status**: 🔴 **EMPTY STUB** - Zero functionality
**Python Original**: 580 lines of complex functionality
**Go Current**: 67 lines of TODOs

**Missing Core Functionality**:
1. **Map Loading and Display**
   ```python
   # Python original
   def resources(self):
       fd, screenshotPath = tempfile.mkstemp()
       os.close(fd)
       self._tempFiles.add(screenshotPath)
       
       screenshot = self.enterWidget.enterMap.getScreenshot()
       screenshot.save(screenshotPath, "PNG")
       
       return {
           "mapPath": self.enterWidget.mapChooser.currentMap["mapPath"],
           "mapScreenshot": screenshotPath,
       }
   ```

2. **Geographic Teaching Logic**
   ```python
   # Python original
   def tabChanged(self):
       lessonDialogsModule = self._modules.default("active", type="lessonDialogs")
       lessonDialogsModule.onTabChanged(
           self.fileTab, 
           self.enterWidget, 
           self.teachWidget, 
           lambda: self.teachWidget.initiateLesson(
               self.enterWidget.list, 
               self.enterWidget.mapChooser.currentMap["mapPath"]
           )
       )
   ```

3. **Resource Management**
   ```python
   # Python original  
   def _removeTempFiles(self):
       for file in self._tempFiles:
           os.remove(file)
   ```

**Current Go Stub**:
```go
func (mod *TeachTopoLessonModule) Createlesson() {
    // TODO: Port Python method logic
}
```

### Media Lesson Module  
**Status**: 🔴 **EMPTY STUB** - Zero functionality
**Python Original**: 300+ lines with multimedia support
**Go Current**: 67 lines of TODOs

**Missing Core Functionality**:
1. **Media Type Validation**
   ```python
   # Python original
   modules = self._modules.sort("active", type="mediaType")
   for item in list["items"]:
       itemSupported = False
       for module in modules:
           if module.supports(item["filename"]):
               itemSupported = True
               break
       if not itemSupported:
           QtWidgets.QMessageBox.critical(
               self.enterWidget, 
               _("Unsupported media"), 
               _("This type of media isn't supported on your computer")
           )
           self.stop()
           return
   ```

2. **Media Player Integration**
   ```python
   # Python original
   def stop(self):
       # Stop media playing
       self.enterWidget.mediaDisplay.stop()
       self.teachWidget.mediaDisplay.stop()
       self.stopped.send()
       return True
   ```

**Current Go Stub**:
```go
func (mod *MediaLessonModule) Createlesson() {
    // TODO: Port Python method logic
}
```

## 4. Qt Interface Compilation Issues

### File Dialog Module - Import Broken
**Error**:
```
internal/modules/interfaces/qt/dialogs/file/file_test.go:16:12: 
module.GetType undefined (type *FileDialogModule has no field or method GetType)
```

**Impact**: Users cannot import lesson files through the GUI

### Character Keyboard - miqt API Incompatibility  
**Error**:
```
internal/modules/interfaces/qt/charsKeyboard/charsKeyboard.go:50:34: 
too many arguments in call to qt.NewQWidget
have (nil, number) want (*qt.QWidget)
```

**Impact**: Special character input non-functional

### Dialog Components - Method Changes
**Error**:
```
internal/modules/interfaces/qt/dialogShower/dialogShower.go:90:9: 
msgBox.SetDefaultButton2 undefined
```

**Impact**: Confirmation dialogs and user interaction broken

## 5. Implementation Priority Matrix

### Phase 1 (Immediate - High Impact/Low Effort)
| Format | Users | Effort | Status |
|--------|-------|---------|---------|
| Teach2000 XML | High | Low | Add XML parser |
| CSV separator fix | High | Low | Add colon support |
| Test validation | Critical | Low | Check item counts |

### Phase 2 (Short Term - High Impact/Medium Effort)  
| Format | Users | Effort | Status |
|--------|-------|---------|---------|
| Anki SQLite | Very High | Medium | Add database parsing |
| Mnemosyne SQLite | Medium | Medium | Add database parsing |
| JMemorize ZIP | Medium | Medium | Add ZIP + XML parsing |

### Phase 3 (Long Term - Medium Impact/High Effort)
| Component | Users | Effort | Status |
|-----------|-------|---------|---------|
| Topo Module | Medium | High | Port 580 lines Python |
| Media Module | Medium | High | Port 300+ lines Python |
| Qt Interface Fix | All | High | Fix miqt migration |

## 6. Data Validation Gaps

### Current Test Issue
Tests pass but don't validate actual data extraction:

```go
// This test "passes" but loads 0 items!
{"application_x-anki2.anki.anki2", "Anki 2.0", false, 0},   
```

### Required Test Enhancement
```go
func TestRealDataExtraction(t *testing.T) {
    tests := []struct {
        file           string
        expectedItems  int
        expectedFirst  string
    }{
        {"anki.anki2", 5, "hello"},
        {"mnemosyne.db", 10, "bonjour"},  
        {"teach2000.t2k", 3, "een"},
    }
    
    for _, tc := range tests {
        data, err := loader.LoadFile(tc.file)
        require.NoError(t, err)
        assert.Equal(t, tc.expectedItems, len(data.List.Items))
        if len(data.List.Items) > 0 {
            assert.Contains(t, data.List.Items[0].Questions[0], tc.expectedFirst)
        }
    }
}
```

## 7. Business Impact

### User Experience
- **File Import**: 58% of formats fail silently
- **Lesson Types**: Only "words" lessons work (33% of lesson types)
- **Teaching**: No topography or media teaching possible
- **Migration**: Users cannot import their existing lesson libraries

### Adoption Blockers
1. Anki users cannot migrate (largest SRS user base)
2. Educators cannot use topography lessons
3. Media-based learning unsupported
4. GUI import functionality broken

## 8. Recommended Immediate Actions

### Week 1: Critical Fixes
1. Fix Teach2000 XML parser (affects test files)
2. Add proper test validation (check actual item counts)
3. Implement basic Anki SQLite support

### Week 2: Core Functionality  
1. Add Mnemosyne database support
2. Fix Qt file dialog for imports
3. Begin topo module porting

### Week 3: Polish and Testing
1. Complete JMemorize ZIP support
2. Fix remaining Qt interface issues
3. Comprehensive real-file testing

This analysis reveals that while the test suite passes, the actual user-facing functionality has critical gaps that block real-world usage and migration from OpenTeacher.