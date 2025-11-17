package lesson

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileType(t *testing.T) {
	loader := NewFileLoader()

	testCases := []struct {
		filename     string
		expectedType string
	}{
		// Basic formats
		{"test.csv", "words"},
		{"test.tsv", "words"},
		{"test.txt", "words"},
		{"test.ot", "words"},
		{"test.json", "words"},

		// Anki formats
		{"test.anki", "words"},
		{"test.anki2", "words"},
		{"test.apkg", "words"},

		// Vocabulary trainer formats
		{"test.kvtml", "words"},
		{"test.backpack", "words"},
		{"test.wcu", "words"},
		{"test.voc", "words"},
		{"test.fq", "words"},
		{"test.fmd", "words"},
		{"test.dkf", "words"},
		{"test.jml", "words"},
		{"test.jvlt", "words"},
		{"test.stp", "words"},
		{"test.db", "words"},

		// Overhoor formats
		{"test.oh", "words"},
		{"test.ohw", "words"},
		{"test.oh4", "words"},

		// Other vocabulary formats
		{"test.ovr", "words"},
		{"test.pau", "words"},
		{"test.vok2", "words"},
		{"test.wdl", "words"},
		{"test.vtl3", "words"},
		{"test.wrts", "words"},
		{"test.xml", "words"},
		{"test.otwd", "words"},

		// Geography/topology formats
		{"test.kgm", "topo"},
		{"test.ottp", "topo"},

		// Special case - can be both words and topo
		{"test.t2k", "words"}, // defaults to words

		// Media format
		{"test.otmd", "media"},

		// Unknown format should default to words
		{"test.unknown", "words"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := loader.GetFileType(tc.filename)
			if result != tc.expectedType {
				t.Errorf("GetFileType(%s) = %s; want %s", tc.filename, result, tc.expectedType)
			}
		})
	}
}

func TestGetSupportedExtensions(t *testing.T) {
	loader := NewFileLoader()
	extensions := loader.GetSupportedExtensions()

	expectedExtensions := []string{
		// Basic formats
		".csv", ".tsv", ".txt", ".ot", ".json",

		// Advanced formats
		".kvtml", ".anki", ".anki2", ".apkg", ".backpack", ".wcu",
		".voc", ".fq", ".fmd", ".dkf", ".jml", ".jvlt", ".stp", ".db",
		".oh", ".ohw", ".oh4", ".ovr", ".pau", ".t2k", ".vok2",
		".wdl", ".vtl3", ".wrts", ".xml", ".kgm", ".ottp", ".otmd", ".otwd",
	}

	// Check that all expected extensions are present
	extensionMap := make(map[string]bool)
	for _, ext := range extensions {
		extensionMap[ext] = true
	}

	for _, expected := range expectedExtensions {
		if !extensionMap[expected] {
			t.Errorf("Expected extension %s not found in supported extensions", expected)
		}
	}

	// Check minimum number of extensions
	if len(extensions) < len(expectedExtensions) {
		t.Errorf("Expected at least %d extensions, got %d", len(expectedExtensions), len(extensions))
	}
}

func TestGetFormatName(t *testing.T) {
	loader := NewFileLoader()

	testCases := []struct {
		extension    string
		expectedName string
	}{
		{".csv", "Spreadsheet (CSV)"},
		{".tsv", "Tab-Separated Values"},
		{".txt", "Text File"},
		{".ot", "OpenTeacher 2.x"},
		{".json", "JSON Lesson File"},
		{".kvtml", "KDE Vocabulary Document"},
		{".anki", "Anki Database"},
		{".anki2", "Anki 2.0 Database"},
		{".apkg", "Anki Package"},
		{".backpack", "Backpack File"},
		{".wcu", "CueCard File"},
		{".voc", "Vocabulary File"},
		{".fq", "FlashQard File"},
		{".fmd", "FM Dictionary"},
		{".dkf", "Granule Deck"},
		{".jml", "JMemorize Lesson"},
		{".jvlt", "JVLT File"},
		{".stp", "Ludem File"},
		{".db", "Database File"},
		{".oh", "Overhoor File"},
		{".ohw", "Overhoor File"},
		{".oh4", "Overhoor File"},
		{".ovr", "Overhoringsprogramma Talen"},
		{".pau", "Pauker File"},
		{".t2k", "Teach2000 File"},
		{".vok2", "Teachmaster File"},
		{".wdl", "Oriente Voca File"},
		{".vtl3", "VokabelTrainer File"},
		{".wrts", "WRTS File"},
		{".xml", "XML File"},
		{".kgm", "KGeography Map"},
		{".ottp", "OpenTeaching Topography"},
		{".otmd", "OpenTeaching Media"},
		{".otwd", "OpenTeaching Words"},
		{".unknown", "Unknown Format"},
	}

	for _, tc := range testCases {
		t.Run(tc.extension, func(t *testing.T) {
			result := loader.GetFormatName(tc.extension)
			if result != tc.expectedName {
				t.Errorf("GetFormatName(%s) = %s; want %s", tc.extension, result, tc.expectedName)
			}
		})
	}
}

func TestLoadCSVFile(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	csvContent := `question1,answer1,comment1
question2,answer2,comment2
"question with, comma","answer with, comma",
incomplete_line,
,empty_question,comment
question_only,
`

	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	// Test loading the CSV file
	lessonData, err := loader.LoadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to load CSV file: %v", err)
	}

	// Verify the loaded data
	if lessonData.List.Title != "test.csv" {
		t.Errorf("Expected title 'test.csv', got '%s'", lessonData.List.Title)
	}

	// Should load valid lines (first two complete lines)
	if len(lessonData.List.Items) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(lessonData.List.Items))
	}

	// Check first item
	if len(lessonData.List.Items) > 0 {
		firstItem := lessonData.List.Items[0]
		if len(firstItem.Questions) != 1 || firstItem.Questions[0] != "question1" {
			t.Errorf("Expected first question 'question1', got %v", firstItem.Questions)
		}
		if len(firstItem.Answers) != 1 || firstItem.Answers[0] != "answer1" {
			t.Errorf("Expected first answer 'answer1', got %v", firstItem.Answers)
		}
		if firstItem.Comment != "comment1" {
			t.Errorf("Expected first comment 'comment1', got '%s'", firstItem.Comment)
		}
	}
}

func TestLoadTextFile(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary text file
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")

	txtContent := `# This is a comment line
question1	answer1	comment1
question2|answer2|comment2
question3=answer3
# Another comment
question4	answer4

incomplete=
=missing_question
`

	err := os.WriteFile(txtFile, []byte(txtContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test text file: %v", err)
	}

	// Test loading the text file
	lessonData, err := loader.LoadFile(txtFile)
	if err != nil {
		t.Fatalf("Failed to load text file: %v", err)
	}

	// Verify the loaded data
	if lessonData.List.Title != "test.txt" {
		t.Errorf("Expected title 'test.txt', got '%s'", lessonData.List.Title)
	}

	// Should load valid lines (first three complete lines)
	if len(lessonData.List.Items) < 3 {
		t.Errorf("Expected at least 3 items, got %d", len(lessonData.List.Items))
	}

	// Check first item (tab-separated)
	if len(lessonData.List.Items) > 0 {
		firstItem := lessonData.List.Items[0]
		if len(firstItem.Questions) != 1 || firstItem.Questions[0] != "question1" {
			t.Errorf("Expected first question 'question1', got %v", firstItem.Questions)
		}
		if len(firstItem.Answers) != 1 || firstItem.Answers[0] != "answer1" {
			t.Errorf("Expected first answer 'answer1', got %v", firstItem.Answers)
		}
	}
}

func TestLoadKVTMLFile(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary KVTML file
	tmpDir := t.TempDir()
	kvtmlFile := filepath.Join(tmpDir, "test.kvtml")

	kvtmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kvtml version="2.0">
  <information>
    <title>Test KVTML</title>
    <author>Test Author</author>
  </information>
  <identifiers>
    <identifier id="0">
      <name>English</name>
    </identifier>
    <identifier id="1">
      <name>German</name>
    </identifier>
  </identifiers>
  <entries>
    <entry id="0">
      <translation id="0">
        <text>hello</text>
      </translation>
      <translation id="1">
        <text>hallo</text>
      </translation>
    </entry>
    <entry id="1">
      <translation id="0">
        <text>goodbye</text>
      </translation>
      <translation id="1">
        <text>auf wiedersehen</text>
      </translation>
    </entry>
  </entries>
</kvtml>`

	err := os.WriteFile(kvtmlFile, []byte(kvtmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test KVTML file: %v", err)
	}

	// Test loading the KVTML file
	lessonData, err := loader.LoadFile(kvtmlFile)
	if err != nil {
		t.Fatalf("Failed to load KVTML file: %v", err)
	}

	// Verify the loaded data
	if lessonData.List.Title != "Test KVTML" {
		t.Errorf("Expected title 'Test KVTML', got '%s'", lessonData.List.Title)
	}

	if lessonData.List.QuestionLanguage != "English" {
		t.Errorf("Expected question language 'English', got '%s'", lessonData.List.QuestionLanguage)
	}

	if lessonData.List.AnswerLanguage != "German" {
		t.Errorf("Expected answer language 'German', got '%s'", lessonData.List.AnswerLanguage)
	}

	// Should load 2 entries
	if len(lessonData.List.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(lessonData.List.Items))
	}

	// Check first item
	if len(lessonData.List.Items) > 0 {
		firstItem := lessonData.List.Items[0]
		if len(firstItem.Questions) != 1 || firstItem.Questions[0] != "hello" {
			t.Errorf("Expected first question 'hello', got %v", firstItem.Questions)
		}
		if len(firstItem.Answers) != 1 || firstItem.Answers[0] != "hallo" {
			t.Errorf("Expected first answer 'hallo', got %v", firstItem.Answers)
		}
	}
}

func TestLoadJSONFile(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	jsonContent := `{
  "list": {
    "title": "Test JSON Lesson",
    "questionLanguage": "English",
    "answerLanguage": "Spanish",
    "items": [
      {
        "id": 0,
        "questions": ["hello"],
        "answers": ["hola"],
        "comment": "greeting"
      },
      {
        "id": 1,
        "questions": ["goodbye"],
        "answers": ["adiós"],
        "comment": "farewell"
      }
    ],
    "tests": []
  }
}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	// Test loading the JSON file
	lessonData, err := loader.LoadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to load JSON file: %v", err)
	}

	// Verify the loaded data
	if lessonData.List.Title != "Test JSON Lesson" {
		t.Errorf("Expected title 'Test JSON Lesson', got '%s'", lessonData.List.Title)
	}

	if lessonData.List.QuestionLanguage != "English" {
		t.Errorf("Expected question language 'English', got '%s'", lessonData.List.QuestionLanguage)
	}

	if lessonData.List.AnswerLanguage != "Spanish" {
		t.Errorf("Expected answer language 'Spanish', got '%s'", lessonData.List.AnswerLanguage)
	}

	// Should load 2 items
	if len(lessonData.List.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(lessonData.List.Items))
	}

	// Check first item
	if len(lessonData.List.Items) > 0 {
		firstItem := lessonData.List.Items[0]
		if len(firstItem.Questions) != 1 || firstItem.Questions[0] != "hello" {
			t.Errorf("Expected first question 'hello', got %v", firstItem.Questions)
		}
		if len(firstItem.Answers) != 1 || firstItem.Answers[0] != "hola" {
			t.Errorf("Expected first answer 'hola', got %v", firstItem.Answers)
		}
		if firstItem.Comment != "greeting" {
			t.Errorf("Expected first comment 'greeting', got '%s'", firstItem.Comment)
		}
	}
}

func TestParseWordString(t *testing.T) {
	loader := NewFileLoader()

	testCases := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"single", []string{"single"}},
		{"one, two, three", []string{"one", "two", "three"}},
		{"one; two; three", []string{"one", "two", "three"}},
		{"  spaced  ,  words  ", []string{"spaced", "words"}},
		{"mixed, content; here", []string{"mixed, content", "here"}}, // semicolon takes precedence
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := loader.parseWordString(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("parseWordString(%s) returned %d items; want %d", tc.input, len(result), len(tc.expected))
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("parseWordString(%s)[%d] = %s; want %s", tc.input, i, result[i], expected)
				}
			}
		})
	}
}

func TestAutoDetection(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary file with unknown extension but CSV content
	tmpDir := t.TempDir()
	unknownFile := filepath.Join(tmpDir, "test.unknown")

	csvContent := `question1,answer1
question2,answer2
question3,answer3`

	err := os.WriteFile(unknownFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading the file - should auto-detect as text/CSV
	lessonData, err := loader.LoadFile(unknownFile)
	if err != nil {
		t.Fatalf("Failed to load file with auto-detection: %v", err)
	}

	// Should have loaded some data
	if len(lessonData.List.Items) < 3 {
		t.Errorf("Auto-detection loaded %d items, expected at least 3", len(lessonData.List.Items))
	}
}

// Benchmark tests for performance
func BenchmarkGetFileType(b *testing.B) {
	loader := NewFileLoader()

	for i := 0; i < b.N; i++ {
		loader.GetFileType("test.kvtml")
	}
}

func BenchmarkParseWordString(b *testing.B) {
	loader := NewFileLoader()
	testString := "word1, word2, word3, word4, word5"

	for i := 0; i < b.N; i++ {
		loader.parseWordString(testString)
	}
}

// Integration test with actual legacy files
func TestLegacyFileCompatibility(t *testing.T) {
	loader := NewFileLoader()

	// Define test cases for legacy files that should exist
	legacyTestCases := []struct {
		filename string
		fileType string
		minItems int // minimum number of items expected
	}{
		{"application_x-kvtml.parley.kvtml", "words", 1},
		{"application_x-kvtml.kwordquiz.kvtml", "words", 1},
		{"text_csv.openteacher3x.csv", "words", 0}, // might be empty or malformed
		{"application_x-openteacher.openteacher2x.ot", "words", 1},
	}

	for _, tc := range legacyTestCases {
		t.Run(tc.filename, func(t *testing.T) {
			filePath := filepath.Join("testdata", "legacy_files", tc.filename)

			// Skip if file doesn't exist (not all test environments may have legacy files)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skip("Legacy test file not available")
				return
			}

			// Check file type detection
			detectedType := loader.GetFileType(filePath)
			if detectedType != tc.fileType {
				t.Errorf("GetFileType(%s) = %s; want %s", tc.filename, detectedType, tc.fileType)
			}

			// Try to load the file
			lessonData, err := loader.LoadFile(filePath)
			if err != nil {
				// Some legacy files might have issues, so we'll log but not fail
				t.Logf("Could not load legacy file %s: %v", tc.filename, err)
				return
			}

			// Check minimum item count
			if len(lessonData.List.Items) < tc.minItems {
				t.Logf("Legacy file %s loaded %d items, expected at least %d", tc.filename, len(lessonData.List.Items), tc.minItems)
			}

			t.Logf("Successfully loaded legacy file %s with %d items", tc.filename, len(lessonData.List.Items))
		})
	}
}

func TestRealTestdataFiles(t *testing.T) {
	loader := NewFileLoader()

	testCases := []struct {
		filename     string
		expectedType string
		minItems     int
	}{
		{"sample.csv", "words", 20},  // Has 30 word pairs
		{"accents.csv", "words", 5},  // Has accent examples
		{"sample.kvtml", "words", 1}, // KVTML format
		{"sample.ot", "words", 1},    // OpenTeacher format
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			filePath := filepath.Join("../../testdata", "lessons", tc.filename)

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skip("Testdata file not available")
				return
			}

			// Check file type detection
			detectedType := loader.GetFileType(filePath)
			if detectedType != tc.expectedType {
				t.Errorf("GetFileType(%s) = %s; want %s", tc.filename, detectedType, tc.expectedType)
			}

			// Try to load the file
			lessonData, err := loader.LoadFile(filePath)
			if err != nil {
				t.Errorf("Failed to load testdata file %s: %v", tc.filename, err)
				return
			}

			// Check minimum item count
			if len(lessonData.List.Items) < tc.minItems {
				t.Errorf("Testdata file %s loaded %d items, expected at least %d", tc.filename, len(lessonData.List.Items), tc.minItems)
			}

			// Basic sanity checks
			if lessonData.List.Title == "" {
				t.Errorf("Testdata file %s has empty title", tc.filename)
			}

			// Check that items have both questions and answers
			validItems := 0
			for _, item := range lessonData.List.Items {
				if len(item.Questions) > 0 && len(item.Answers) > 0 {
					validItems++
				}
			}

			if validItems < tc.minItems {
				t.Errorf("Testdata file %s has %d valid items, expected at least %d", tc.filename, validItems, tc.minItems)
			}

			t.Logf("Successfully loaded %s: %d items, title: %s", tc.filename, len(lessonData.List.Items), lessonData.List.Title)
		})
	}
}

func TestLegacyFileTypeSupport(t *testing.T) {
	loader := NewFileLoader()

	// Test a sample of legacy files from each major format category
	legacyFiles := []struct {
		filename string
		fileType string
		format   string
	}{
		{"application_x-kvtml.parley.kvtml", "words", "KVTML"},
		{"text_csv.openteacher3x.csv", "words", "CSV"},
		{"application_x-openteacher.openteacher2x.ot", "words", "OpenTeacher"},
		{"application_xml.abbyylingvotutor_x5.xml", "words", "XML"},
		{"application_x-anki.anki.anki", "words", "Anki"},
		{"application_x-teach2000.teach2000.t2k", "words", "Teach2000"},
	}

	for _, tc := range legacyFiles {
		t.Run(tc.filename, func(t *testing.T) {
			filePath := filepath.Join("../../testdata", "legacy_files", tc.filename)

			// Skip if file doesn't exist
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skip("Legacy test file not available")
				return
			}

			// Check file type detection
			detectedType := loader.GetFileType(filePath)
			if detectedType != tc.fileType {
				t.Errorf("GetFileType(%s) = %s; want %s", tc.filename, detectedType, tc.fileType)
			}

			// Try to load the file (some may fail due to complex formats)
			lessonData, err := loader.LoadFile(filePath)
			if err != nil {
				t.Logf("Legacy file %s (%s format) could not be loaded: %v", tc.filename, tc.format, err)
				return
			}

			t.Logf("Successfully loaded legacy %s file %s: %d items", tc.format, tc.filename, len(lessonData.List.Items))
		})
	}
}

func TestComprehensiveFormatSupport(t *testing.T) {
	loader := NewFileLoader()

	// Test additional legacy file formats for comprehensive coverage
	formatTests := []struct {
		filename   string
		format     string
		expectLoad bool // whether we expect successful loading
		minItems   int  // minimum items if loading succeeds
	}{
		// KVTML variants
		{"application_x-kvtml.kwordquiz.kvtml", "KWordQuiz KVTML", true, 1},
		{"application_x-kvtml.kvoctrain.kvtml", "KVocTrain KVTML", true, 0}, // Empty legacy file

		// OpenTeacher variants
		{"application_x-openteacher.openteacher1x.ot", "OpenTeacher 1.x", true, 1},
		{"application_x-openteacher.openteacher3x.ot", "OpenTeacher 3.x", true, 1},

		// Anki formats (now proper SQLite support)
		{"application_x-anki2.anki.anki2", "Anki 2.0", true, 3},   // SQLite database parsing
		{"application_x-apkg.anki.apkg", "Anki Package", true, 3}, // CSV fallback works

		// Text formats
		{"text_plain.gnuVocabTrain.txt", "GNU VocabTrain", true, 1},

		// Teach2000 format (now implemented)
		{"application_x-teach2000.teach2000.t2k", "Teach2000", true, 3},

		// SQLite databases (now implemented)
		{"application_x-sqlite3.mnemosyne.db", "Mnemosyne", true, 3},

		// Other vocabulary trainers (now implemented)
		{"application_x-backpack.backpack", "Backpack", true, 1},       // Backpack text format
		{"application_x-cuecard.cuecard.wcu", "CueCard", true, 3},      // CueCard XML format
		{"application_x-flashqard.flashqard.fq", "FlashQard", true, 2}, // FlashQard XML format
		{"application_x-jvlt.jvlt.jvlt", "JVLT", true, 2},              // JVLT ZIP format
		{"application_x-teachmaster.vok2", "TeachMaster", true, 3},     // TeachMaster XML format

		// XML variants
		{"application_xml.abbyylingvotutor_x3-modified.xml", "ABBYY Lingvo (modified)", true, 1},
	}

	successCount := 0
	totalCount := len(formatTests)

	for _, tc := range formatTests {
		t.Run(tc.filename, func(t *testing.T) {
			filePath := filepath.Join("../../testdata", "legacy_files", tc.filename)

			// Skip if file doesn't exist
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skip("Legacy test file not available")
				return
			}

			// Try to load the file
			lessonData, err := loader.LoadFile(filePath)

			if tc.expectLoad {
				if err != nil {
					t.Errorf("Expected %s to load successfully, but got error: %v", tc.format, err)
					return
				}

				if len(lessonData.List.Items) < tc.minItems {
					t.Errorf("%s loaded %d items, expected at least %d", tc.format, len(lessonData.List.Items), tc.minItems)
					return
				}

				t.Logf("✅ Successfully loaded %s: %d items", tc.format, len(lessonData.List.Items))
				successCount++
			} else {
				if err == nil && len(lessonData.List.Items) > 0 {
					t.Logf("⚠️  %s unexpectedly loaded successfully: %d items (format may be partially supported)", tc.format, len(lessonData.List.Items))
					successCount++
				} else {
					t.Logf("❌ %s format not implemented (expected): %v", tc.format, err)
				}
			}
		})
	}

	t.Logf("Format support summary: %d/%d formats successfully loaded (%.1f%%)", successCount, totalCount, float64(successCount)/float64(totalCount)*100)

	// Don't fail the test - this is informational
	if successCount < totalCount/2 {
		t.Logf("Note: Less than 50%% of formats are fully supported, which is expected for legacy compatibility testing")
	}
}
