package lesson

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSaver_SaveCSVFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "Dutch",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hallo"},
					Comment:   "greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye", "bye"},
					Answers:   []string{"dag", "tot ziens"},
					Comment:   "",
				},
				{
					ID:        2,
					Questions: []string{"thank you"},
					Answers:   []string{"dank je"},
					Comment:   "polite expression",
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.csv")

	// Create file saver and save
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save CSV file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("CSV file was not created")
	}

	// Read and verify file contents
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Verify structure
	if len(records) != 4 { // header + 3 items
		t.Errorf("Expected 4 records (header + 3 items), got %d", len(records))
	}

	// Verify header
	expectedHeader := []string{"English", "Dutch", "Comment", "Comment After Answering"}
	if len(records) > 0 {
		header := records[0]
		if !equalStringSlices(header, expectedHeader) {
			t.Errorf("Header mismatch. Expected %v, got %v", expectedHeader, header)
		}
	}

	// Verify first item
	if len(records) > 1 {
		firstItem := records[1]
		expectedFirst := []string{"hello", "hallo", "greeting", ""}
		if !equalStringSlices(firstItem, expectedFirst) {
			t.Errorf("First item mismatch. Expected %v, got %v", expectedFirst, firstItem)
		}
	}

	// Verify second item (multiple questions/answers)
	if len(records) > 2 {
		secondItem := records[2]
		expectedSecond := []string{"goodbye; bye", "dag; tot ziens", "", ""}
		if !equalStringSlices(secondItem, expectedSecond) {
			t.Errorf("Second item mismatch. Expected %v, got %v", expectedSecond, secondItem)
		}
	}

	// Verify third item
	if len(records) > 3 {
		thirdItem := records[3]
		expectedThird := []string{"thank you", "dank je", "polite expression", ""}
		if !equalStringSlices(thirdItem, expectedThird) {
			t.Errorf("Third item mismatch. Expected %v, got %v", expectedThird, thirdItem)
		}
	}
}

func TestFileSaver_SaveCSVFileWithDefaultHeaders(t *testing.T) {
	// Create test lesson data without language specification
	lessonData := &LessonData{
		List: WordList{
			Title: "No Language Lesson",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"test"},
					Answers:   []string{"prueba"},
					Comment:   "",
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "default_headers.csv")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save CSV file: %v", err)
	}

	// Read and verify headers use defaults
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(records) < 1 {
		t.Fatal("No header found in CSV")
	}

	expectedHeader := []string{"Questions", "Answers", "Comment", "Comment After Answering"}
	if !equalStringSlices(records[0], expectedHeader) {
		t.Errorf("Default header mismatch. Expected %v, got %v", expectedHeader, records[0])
	}
}

func TestFileSaver_ValidateLessonData(t *testing.T) {
	saver := NewFileSaver()

	// Test nil lesson data
	err := saver.ValidateLessonData(nil)
	if err == nil {
		t.Error("Expected error for nil lesson data")
	}

	// Test empty lesson
	emptyLesson := &LessonData{
		List: WordList{
			Items: []WordItem{},
		},
	}
	err = saver.ValidateLessonData(emptyLesson)
	if err == nil {
		t.Error("Expected error for empty lesson")
	}

	// Test lesson with no valid items
	invalidLesson := &LessonData{
		List: WordList{
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{""},
					Answers:   []string{""},
				},
				{
					ID:        1,
					Questions: []string{},
					Answers:   []string{},
				},
			},
		},
	}
	err = saver.ValidateLessonData(invalidLesson)
	if err == nil {
		t.Error("Expected error for lesson with no valid items")
	}

	// Test valid lesson
	validLesson := &LessonData{
		List: WordList{
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hola"},
				},
			},
		},
	}
	err = saver.ValidateLessonData(validLesson)
	if err != nil {
		t.Errorf("Unexpected error for valid lesson: %v", err)
	}
}

func TestFileSaver_GetDefaultFilename(t *testing.T) {
	saver := NewFileSaver()

	tests := []struct {
		name        string
		lessonData  *LessonData
		ext         string
		expected    string
		description string
	}{
		{
			name:       "nil lesson data",
			lessonData: nil,
			ext:        ".csv",
			expected:   "lesson.csv",
		},
		{
			name: "empty title",
			lessonData: &LessonData{
				List: WordList{Title: ""},
			},
			ext:      ".csv",
			expected: "lesson.csv",
		},
		{
			name: "normal title",
			lessonData: &LessonData{
				List: WordList{Title: "Spanish Basics"},
			},
			ext:      ".csv",
			expected: "Spanish Basics.csv",
		},
		{
			name: "title with invalid chars",
			lessonData: &LessonData{
				List: WordList{Title: "Test/Lesson\\With:Invalid*Chars?"},
			},
			ext:      ".csv",
			expected: "Test_Lesson_With_Invalid_Chars_.csv",
		},
		{
			name: "very long title",
			lessonData: &LessonData{
				List: WordList{Title: "This is a very long lesson title that should be truncated to a reasonable length for filename purposes"},
			},
			ext:      ".csv",
			expected: "This is a very long lesson title that should be tr.csv",
		},
		{
			name: "title with spaces and dots",
			lessonData: &LessonData{
				List: WordList{Title: "  ..Test Lesson..  "},
			},
			ext:      ".csv",
			expected: "Test Lesson.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := saver.GetDefaultFilename(tt.lessonData, tt.ext)
			if result != tt.expected {
				t.Errorf("GetDefaultFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileSaver_GetSupportedExtensions(t *testing.T) {
	saver := NewFileSaver()
	extensions := saver.GetSupportedSaveExtensions()

	// Should have at least CSV
	if len(extensions) == 0 {
		t.Error("No supported extensions returned")
	}

	// CSV should be supported
	found := false
	for _, ext := range extensions {
		if ext == ".csv" {
			found = true
			break
		}
	}
	if !found {
		t.Error("CSV extension not found in supported extensions")
	}
}

func TestFileSaver_GetSaveFormatName(t *testing.T) {
	saver := NewFileSaver()

	tests := []struct {
		ext      string
		expected string
	}{
		{".csv", "Comma-Separated Values (Spreadsheet)"},
		{".CSV", "Comma-Separated Values (Spreadsheet)"}, // case insensitive
		{".unknown", "Unknown Format"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := saver.GetSaveFormatName(tt.ext)
			if result != tt.expected {
				t.Errorf("GetSaveFormatName(%s) = %v, want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestFileSaver_GetSaveFilter(t *testing.T) {
	saver := NewFileSaver()
	filter := saver.GetSaveFilter()

	// Should contain CSV format
	if !strings.Contains(filter, "Comma-Separated Values") {
		t.Error("Save filter should contain CSV format description")
	}

	if !strings.Contains(filter, "*.csv") {
		t.Error("Save filter should contain CSV extension pattern")
	}

	// Should use ;; as separator (Qt style)
	if strings.Contains(filter, ";;") || len(saver.GetSupportedSaveExtensions()) <= 1 {
		// OK - either has multiple formats with ;; separator, or only one format
	} else {
		t.Error("Multiple formats should be separated by ;;")
	}
}

func TestFileSaver_UnsupportedFormat(t *testing.T) {
	lessonData := &LessonData{
		List: WordList{
			Items: []WordItem{
				{Questions: []string{"test"}, Answers: []string{"prueba"}},
			},
		},
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.unsupported")

	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported save format") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

func TestFileSaver_SaveWithValidation(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "validated.csv")

	saver := NewFileSaver()

	// Test with invalid data
	invalidData := &LessonData{
		List: WordList{Items: []WordItem{}},
	}
	err := saver.SaveWithValidation(invalidData, testFile)
	if err == nil {
		t.Error("Expected validation error for empty lesson")
	}

	// Test with valid data
	validData := &LessonData{
		List: WordList{
			Items: []WordItem{
				{Questions: []string{"hello"}, Answers: []string{"hola"}},
			},
		},
	}
	err = saver.SaveWithValidation(validData, testFile)
	if err != nil {
		t.Errorf("Unexpected error with valid data: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File was not created after successful validation")
	}
}

func TestFileSaver_SaveOpenTeacherFile(t *testing.T) {
	// Create test lesson data with test results
	lessonData := &LessonData{
		List: WordList{
			Title:            "OpenTeacher Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "Spanish",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hola", "saludos"},
					Comment:   "greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye", "bye"},
					Answers:   []string{"adiós"},
					Comment:   "",
				},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 0, Result: "wrong"},
						{ItemID: 1, Result: "right"},
					},
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.ot")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save OpenTeacher file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("OpenTeacher file was not created")
	}

	// Read and parse XML
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved OpenTeacher file: %v", err)
	}
	defer file.Close()

	var otXML OpenTeacherXML
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&otXML); err != nil {
		t.Fatalf("Failed to parse OpenTeacher XML: %v", err)
	}

	// Verify XML structure
	if otXML.Title != "OpenTeacher Test Lesson" {
		t.Errorf("Title mismatch. Expected 'OpenTeacher Test Lesson', got '%s'", otXML.Title)
	}

	if otXML.QuestionLanguage != "English" {
		t.Errorf("QuestionLanguage mismatch. Expected 'English', got '%s'", otXML.QuestionLanguage)
	}

	if otXML.AnswerLanguage != "Spanish" {
		t.Errorf("AnswerLanguage mismatch. Expected 'Spanish', got '%s'", otXML.AnswerLanguage)
	}

	if len(otXML.Words) != 2 {
		t.Errorf("Expected 2 words, got %d", len(otXML.Words))
	}

	// Verify first word
	if len(otXML.Words) > 0 {
		word := otXML.Words[0]
		if word.Known != "hello" {
			t.Errorf("First word 'known' mismatch. Expected 'hello', got '%s'", word.Known)
		}
		if word.Foreign != "hola" {
			t.Errorf("First word 'foreign' mismatch. Expected 'hola', got '%s'", word.Foreign)
		}
		if word.Second != "saludos" {
			t.Errorf("First word 'second' mismatch. Expected 'saludos', got '%s'", word.Second)
		}
		if word.Results != "1/2" { // 1 wrong, 2 total
			t.Errorf("First word 'results' mismatch. Expected '1/2', got '%s'", word.Results)
		}
	}

	// Verify second word
	if len(otXML.Words) > 1 {
		word := otXML.Words[1]
		if word.Known != "goodbye, bye" {
			t.Errorf("Second word 'known' mismatch. Expected 'goodbye, bye', got '%s'", word.Known)
		}
		if word.Foreign != "adiós" {
			t.Errorf("Second word 'foreign' mismatch. Expected 'adiós', got '%s'", word.Foreign)
		}
		if word.Results != "0/1" { // 0 wrong, 1 total
			t.Errorf("Second word 'results' mismatch. Expected '0/1', got '%s'", word.Results)
		}
	}
}

func TestFileSaver_SaveTextFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "Text Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "German",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hallo"},
				},
				{
					ID:        1,
					Questions: []string{"goodbye"},
					Answers:   []string{"auf wiedersehen", "tschüss"},
				},
				{
					ID:        2,
					Questions: []string{"very long question text"},
					Answers:   []string{"answer"},
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.txt")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save text file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Text file was not created")
	}

	// Read file contents
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved text file: %v", err)
	}

	text := string(content)

	// Verify content structure
	if !strings.Contains(text, "Text Test Lesson") {
		t.Error("File should contain lesson title")
	}

	if !strings.Contains(text, "English - German") {
		t.Error("File should contain language information")
	}

	if !strings.Contains(text, "hello") {
		t.Error("File should contain question text")
	}

	if !strings.Contains(text, "hallo") {
		t.Error("File should contain answer text")
	}

	if !strings.Contains(text, "auf wiedersehen, tschüss") {
		t.Error("File should contain multiple answers joined with comma")
	}

	// Verify alignment (longer questions should have proper spacing)
	lines := strings.Split(text, "\n")
	var contentLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Text Test Lesson") && !strings.Contains(line, "English - German") {
			contentLines = append(contentLines, line)
		}
	}

	if len(contentLines) != 3 {
		t.Errorf("Expected 3 content lines, got %d", len(contentLines))
	}
}

func TestFileSaver_SaveJSONFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "JSON Test Lesson",
			QuestionLanguage: "French",
			AnswerLanguage:   "Italian",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"bonjour"},
					Answers:   []string{"ciao", "buongiorno"},
					Comment:   "greeting",
				},
				{
					ID:        1,
					Questions: []string{"au revoir"},
					Answers:   []string{"arrivederci"},
					Comment:   "",
				},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 1, Result: "wrong"},
					},
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.json")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save JSON file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("JSON file was not created")
	}

	// Read and parse JSON
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved JSON file: %v", err)
	}
	defer file.Close()

	var parsedData LessonData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&parsedData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify JSON structure
	if parsedData.List.Title != "JSON Test Lesson" {
		t.Errorf("Title mismatch. Expected 'JSON Test Lesson', got '%s'", parsedData.List.Title)
	}

	if parsedData.List.QuestionLanguage != "French" {
		t.Errorf("QuestionLanguage mismatch. Expected 'French', got '%s'", parsedData.List.QuestionLanguage)
	}

	if parsedData.List.AnswerLanguage != "Italian" {
		t.Errorf("AnswerLanguage mismatch. Expected 'Italian', got '%s'", parsedData.List.AnswerLanguage)
	}

	if len(parsedData.List.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(parsedData.List.Items))
	}

	// Verify first item
	if len(parsedData.List.Items) > 0 {
		item := parsedData.List.Items[0]
		if len(item.Questions) != 1 || item.Questions[0] != "bonjour" {
			t.Errorf("First item questions mismatch. Expected ['bonjour'], got %v", item.Questions)
		}
		if len(item.Answers) != 2 || item.Answers[0] != "ciao" || item.Answers[1] != "buongiorno" {
			t.Errorf("First item answers mismatch. Expected ['ciao', 'buongiorno'], got %v", item.Answers)
		}
		if item.Comment != "greeting" {
			t.Errorf("First item comment mismatch. Expected 'greeting', got '%s'", item.Comment)
		}
	}

	// Verify test data is preserved
	if len(parsedData.List.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(parsedData.List.Tests))
	}

	if len(parsedData.List.Tests) > 0 && len(parsedData.List.Tests[0].Results) != 2 {
		t.Errorf("Expected 2 test results, got %d", len(parsedData.List.Tests[0].Results))
	}
}

func TestFileSaver_SaveTextFileEmpty(t *testing.T) {
	// Test empty lesson
	lessonData := &LessonData{
		List: WordList{
			Title: "Empty Lesson",
			Items: []WordItem{},
		},
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.txt")

	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save empty text file: %v", err)
	}

	// Read file contents
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read empty text file: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Empty Lesson") {
		t.Error("Empty file should still contain title")
	}
}

func TestFileSaver_CalculateWordStatistics(t *testing.T) {
	lessonData := &LessonData{
		List: WordList{
			Items: []WordItem{
				{ID: 0}, {ID: 1}, {ID: 2},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 0, Result: "right"},
						{ItemID: 0, Result: "wrong"},
						{ItemID: 1, Result: "wrong"},
						{ItemID: 1, Result: "wrong"},
						{ItemID: 2, Result: "right"},
					},
				},
			},
		},
	}

	saver := NewFileSaver()
	stats := saver.calculateWordStatistics(lessonData)

	// Verify statistics
	if len(stats) != 3 {
		t.Errorf("Expected stats for 3 items, got %d", len(stats))
	}

	// Item 0: 2 right, 1 wrong
	if stat, exists := stats[0]; !exists || stat.Right != 2 || stat.Wrong != 1 {
		t.Errorf("Item 0 stats incorrect. Expected Right=2, Wrong=1, got Right=%d, Wrong=%d", stat.Right, stat.Wrong)
	}

	// Item 1: 0 right, 2 wrong
	if stat, exists := stats[1]; !exists || stat.Right != 0 || stat.Wrong != 2 {
		t.Errorf("Item 1 stats incorrect. Expected Right=0, Wrong=2, got Right=%d, Wrong=%d", stat.Right, stat.Wrong)
	}

	// Item 2: 1 right, 0 wrong
	if stat, exists := stats[2]; !exists || stat.Right != 1 || stat.Wrong != 0 {
		t.Errorf("Item 2 stats incorrect. Expected Right=1, Wrong=0, got Right=%d, Wrong=%d", stat.Right, stat.Wrong)
	}
}

func TestFileSaver_SaveTeach2000File(t *testing.T) {
	// Create test lesson data with detailed test results
	lessonData := &LessonData{
		List: WordList{
			Title:            "Teach2000 Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "Dutch",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello", "hi"},
					Answers:   []string{"hallo"},
					Comment:   "greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye"},
					Answers:   []string{"dag", "tot ziens"},
					Comment:   "",
				},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 0, Result: "wrong"},
						{ItemID: 0, Result: "wrong"},
						{ItemID: 1, Result: "right"},
						{ItemID: 1, Result: "wrong"},
					},
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.t2k")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save Teach2000 file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Teach2000 file was not created")
	}

	// Read and parse XML
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved Teach2000 file: %v", err)
	}
	defer file.Close()

	var t2kXML Teach2000XML
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&t2kXML); err != nil {
		t.Fatalf("Failed to parse Teach2000 XML: %v", err)
	}

	// Verify XML structure
	if t2kXML.Version != "831" {
		t.Errorf("Version mismatch. Expected '831', got '%s'", t2kXML.Version)
	}

	if t2kXML.Description != "Normal" {
		t.Errorf("Description mismatch. Expected 'Normal', got '%s'", t2kXML.Description)
	}

	if len(t2kXML.MessageData.Items.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(t2kXML.MessageData.Items.Items))
	}

	// Verify first item statistics
	if len(t2kXML.MessageData.Items.Items) > 0 {
		item := t2kXML.MessageData.Items.Items[0]
		if item.ID != "0" {
			t.Errorf("First item ID mismatch. Expected '0', got '%s'", item.ID)
		}
		if item.Errors != 2 {
			t.Errorf("First item errors mismatch. Expected 2, got %d", item.Errors)
		}
		if item.TestCount != 3 {
			t.Errorf("First item test count mismatch. Expected 3, got %d", item.TestCount)
		}
		if item.CorrectCount != 1 {
			t.Errorf("First item correct count mismatch. Expected 1, got %d", item.CorrectCount)
		}
		if len(item.Questions) != 2 {
			t.Errorf("First item should have 2 questions, got %d", len(item.Questions))
		}
	}

	// Verify test results
	if len(t2kXML.MessageData.TestResults) != 1 {
		t.Errorf("Expected 1 test result, got %d", len(t2kXML.MessageData.TestResults))
	}

	if len(t2kXML.MessageData.TestResults) > 0 {
		testResult := t2kXML.MessageData.TestResults[0]
		if testResult.AnswersCorrect != 2 {
			t.Errorf("Test result answers correct mismatch. Expected 2, got %d", testResult.AnswersCorrect)
		}
		if testResult.WrongOnce != 1 { // Item 1 wrong once
			t.Errorf("Test result wrong once mismatch. Expected 1, got %d", testResult.WrongOnce)
		}
		if testResult.WrongTwice != 1 { // Item 0 wrong twice
			t.Errorf("Test result wrong twice mismatch. Expected 1, got %d", testResult.WrongTwice)
		}
		// Check that score is calculated (should be between 1 and 10)
		if testResult.Score == "" {
			t.Error("Test result score should not be empty")
		}
	}
}

func TestFileSaver_SaveKVTMLFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "KVTML Test Lesson",
			QuestionLanguage: "German",
			AnswerLanguage:   "French",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hallo"},
					Answers:   []string{"bonjour", "salut"},
					Comment:   "greeting",
				},
				{
					ID:        1,
					Questions: []string{"tschüss"},
					Answers:   []string{"au revoir"},
					Comment:   "",
				},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 1, Result: "wrong"},
					},
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.kvtml")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save KVTML file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("KVTML file was not created")
	}

	// Read and parse XML
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved KVTML file: %v", err)
	}
	defer file.Close()

	var kvtmlXML KVTMLXML
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&kvtmlXML); err != nil {
		t.Fatalf("Failed to parse KVTML XML: %v", err)
	}

	// Verify XML structure
	if kvtmlXML.Version != "2.0" {
		t.Errorf("Version mismatch. Expected '2.0', got '%s'", kvtmlXML.Version)
	}

	if kvtmlXML.Information.Title != "KVTML Test Lesson" {
		t.Errorf("Title mismatch. Expected 'KVTML Test Lesson', got '%s'", kvtmlXML.Information.Title)
	}

	if kvtmlXML.Information.Generator != "Recuerdo" {
		t.Errorf("Generator mismatch. Expected 'Recuerdo', got '%s'", kvtmlXML.Information.Generator)
	}

	if len(kvtmlXML.Identifiers) != 2 {
		t.Errorf("Expected 2 identifiers, got %d", len(kvtmlXML.Identifiers))
	}

	// Verify identifiers
	if len(kvtmlXML.Identifiers) >= 2 {
		if kvtmlXML.Identifiers[0].Name != "German" {
			t.Errorf("First identifier name mismatch. Expected 'German', got '%s'", kvtmlXML.Identifiers[0].Name)
		}
		if kvtmlXML.Identifiers[1].Name != "French" {
			t.Errorf("Second identifier name mismatch. Expected 'French', got '%s'", kvtmlXML.Identifiers[1].Name)
		}
	}

	// Verify entries (should be 2 items + 1 empty entry)
	if len(kvtmlXML.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(kvtmlXML.Entries))
	}

	// Verify first entry
	if len(kvtmlXML.Entries) > 0 {
		entry := kvtmlXML.Entries[0]
		if entry.ID != "0" {
			t.Errorf("First entry ID mismatch. Expected '0', got '%s'", entry.ID)
		}
		if len(entry.Translations) != 2 {
			t.Errorf("First entry should have 2 translations, got %d", len(entry.Translations))
		}
		if len(entry.Translations) >= 2 {
			if entry.Translations[0].Text != "hallo" {
				t.Errorf("First entry question mismatch. Expected 'hallo', got '%s'", entry.Translations[0].Text)
			}
			if entry.Translations[1].Text != "bonjour, salut" {
				t.Errorf("First entry answer mismatch. Expected 'bonjour, salut', got '%s'", entry.Translations[1].Text)
			}
			if entry.Translations[0].Comment != "greeting" {
				t.Errorf("First entry comment mismatch. Expected 'greeting', got '%s'", entry.Translations[0].Comment)
			}
		}
	}

	// Verify lessons
	if len(kvtmlXML.Lessons) != 1 {
		t.Errorf("Expected 1 lesson, got %d", len(kvtmlXML.Lessons))
	}

	if len(kvtmlXML.Lessons) > 0 {
		lesson := kvtmlXML.Lessons[0]
		if lesson.Name != "Lesson 1" {
			t.Errorf("Lesson name mismatch. Expected 'Lesson 1', got '%s'", lesson.Name)
		}
		if len(lesson.Entries) != 2 {
			t.Errorf("Expected 2 lesson entries, got %d", len(lesson.Entries))
		}
	}
}

func TestFileSaver_CalculateNote(t *testing.T) {
	saver := NewFileSaver()

	tests := []struct {
		right, total int
		expectedNote string
	}{
		{0, 0, "1.00000"},  // No answers
		{0, 1, "1.00000"},  // All wrong
		{1, 1, "10.00000"}, // All right
		{1, 2, "5.50000"},  // Half right: 1/2 * 9 + 1 = 5.5
		{3, 4, "7.75000"},  // 3/4 right: 3/4 * 9 + 1 = 7.75
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d/%d", tt.right, tt.total), func(t *testing.T) {
			result := saver.calculateNote(tt.right, tt.total)
			if result != tt.expectedNote {
				t.Errorf("calculateNote(%d, %d) = %s, want %s", tt.right, tt.total, result, tt.expectedNote)
			}
		})
	}
}

func TestFileSaver_CalculateWrongStats(t *testing.T) {
	saver := NewFileSaver()

	test := Test{
		Results: []TestResult{
			{ItemID: 0, Result: "wrong"}, // Item 0 wrong once
			{ItemID: 1, Result: "wrong"}, // Item 1 wrong once
			{ItemID: 1, Result: "wrong"}, // Item 1 wrong twice total
			{ItemID: 2, Result: "wrong"}, // Item 2 wrong once
			{ItemID: 2, Result: "wrong"}, // Item 2 wrong twice
			{ItemID: 2, Result: "wrong"}, // Item 2 wrong three times total
			{ItemID: 3, Result: "right"}, // Item 3 right (should be ignored)
		},
	}

	wrongOnce, wrongTwice, wrongMoreThanTwice := saver.calculateWrongStats(test)

	if wrongOnce != 1 { // Only item 0 wrong once
		t.Errorf("Wrong once count mismatch. Expected 1, got %d", wrongOnce)
	}
	if wrongTwice != 1 { // Only item 1 wrong twice
		t.Errorf("Wrong twice count mismatch. Expected 1, got %d", wrongTwice)
	}
	if wrongMoreThanTwice != 1 { // Only item 2 wrong more than twice
		t.Errorf("Wrong more than twice count mismatch. Expected 1, got %d", wrongMoreThanTwice)
	}
}

func TestFileSaver_SaveHTMLFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "HTML Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "Spanish",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hola"},
					Comment:   "common greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye"},
					Answers:   []string{"adiós", "hasta luego"},
					Comment:   "",
				},
				{
					ID:        2,
					Questions: []string{"<special> & \"chars\""},
					Answers:   []string{"<especial> & \"caracteres\""},
					Comment:   "test HTML escaping",
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.html")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save HTML file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("HTML file was not created")
	}

	// Read file contents
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved HTML file: %v", err)
	}

	html := string(content)

	// Verify HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("File should contain HTML5 DOCTYPE")
	}

	if !strings.Contains(html, "HTML Test Lesson") {
		t.Error("File should contain lesson title")
	}

	if !strings.Contains(html, "English → Spanish") {
		t.Error("File should contain language information")
	}

	if !strings.Contains(html, "hola") {
		t.Error("File should contain answer text")
	}

	if !strings.Contains(html, "common greeting") {
		t.Error("File should contain comment text")
	}

	// Verify HTML escaping
	if !strings.Contains(html, "&lt;special&gt; &amp; &quot;chars&quot;") {
		t.Error("File should properly escape HTML special characters")
	}

	if !strings.Contains(html, "&lt;especial&gt; &amp; &quot;caracteres&quot;") {
		t.Error("File should properly escape HTML special characters in answers")
	}

	// Verify CSS styling is present
	if !strings.Contains(html, "vocabulary-table") {
		t.Error("File should contain CSS styling")
	}

	if !strings.Contains(html, "Total vocabulary items: 3") {
		t.Error("File should contain item count")
	}

	// Verify responsive design meta tag
	if !strings.Contains(html, "viewport") {
		t.Error("File should contain viewport meta tag for responsive design")
	}
}

func TestFileSaver_SaveLaTeXFile(t *testing.T) {
	// Create test lesson data
	lessonData := &LessonData{
		List: WordList{
			Title:            "LaTeX Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "German",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello"},
					Answers:   []string{"hallo"},
					Comment:   "common greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye"},
					Answers:   []string{"auf wiedersehen", "tschüss"},
					Comment:   "",
				},
				{
					ID:        2,
					Questions: []string{"special & chars: $#%^_~{}"},
					Answers:   []string{"spezielle & zeichen: $#%^_~{}"},
					Comment:   "test LaTeX escaping & formatting",
				},
			},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_lesson.tex")

	// Save file
	saver := NewFileSaver()
	err := saver.SaveFile(lessonData, testFile)
	if err != nil {
		t.Fatalf("Failed to save LaTeX file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("LaTeX file was not created")
	}

	// Read file contents
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved LaTeX file: %v", err)
	}

	latex := string(content)

	// Verify LaTeX document structure
	if !strings.Contains(latex, `\documentclass[12pt,a4paper]{article}`) {
		t.Error("File should contain LaTeX document class")
	}

	if !strings.Contains(latex, `\usepackage[utf8]{inputenc}`) {
		t.Error("File should contain UTF-8 input encoding")
	}

	if !strings.Contains(latex, `\begin{document}`) {
		t.Error("File should contain document begin")
	}

	if !strings.Contains(latex, `\end{document}`) {
		t.Error("File should contain document end")
	}

	// Verify title
	if !strings.Contains(latex, "LaTeX Test Lesson") {
		t.Error("File should contain lesson title")
	}

	// Verify language information
	if !strings.Contains(latex, `English $\rightarrow$ German`) {
		t.Error("File should contain language information with LaTeX arrow")
	}

	// Verify table structure
	if !strings.Contains(latex, `\begin{longtable}`) {
		t.Error("File should contain longtable environment")
	}

	if !strings.Contains(latex, `\end{longtable}`) {
		t.Error("File should contain longtable end")
	}

	// Verify content
	if !strings.Contains(latex, "hallo") {
		t.Error("File should contain answer text")
	}

	if !strings.Contains(latex, `\textit{common greeting}`) {
		t.Error("File should contain comment in italic")
	}

	// Verify LaTeX escaping
	if !strings.Contains(latex, `spezielle \& zeichen: \$\#\%\textasciicircum{}\_\textasciitilde{}\{\}`) {
		t.Error("File should properly escape LaTeX special characters")
	}

	// Verify statistics section
	if !strings.Contains(latex, `Total vocabulary items: 3`) {
		t.Error("File should contain item count in statistics")
	}

	// Verify table headers
	if !strings.Contains(latex, `\textbf{English}`) {
		t.Error("File should contain bold question language header")
	}

	if !strings.Contains(latex, `\textbf{German}`) {
		t.Error("File should contain bold answer language header")
	}

	if !strings.Contains(latex, `\textbf{Comment}`) {
		t.Error("File should contain bold comment header")
	}
}

func TestFileSaver_AllFormatsIntegration(t *testing.T) {
	// Create comprehensive test lesson data with all features
	lessonData := &LessonData{
		List: WordList{
			Title:            "Complete Export Test Lesson",
			QuestionLanguage: "English",
			AnswerLanguage:   "Spanish",
			Items: []WordItem{
				{
					ID:        0,
					Questions: []string{"hello", "hi"},
					Answers:   []string{"hola", "saludos"},
					Comment:   "common greeting",
				},
				{
					ID:        1,
					Questions: []string{"goodbye"},
					Answers:   []string{"adiós", "hasta luego"},
					Comment:   "farewell expression",
				},
				{
					ID:        2,
					Questions: []string{"thank you"},
					Answers:   []string{"gracias"},
					Comment:   "",
				},
				{
					ID:        3,
					Questions: []string{"special & <chars> \"test\""},
					Answers:   []string{"especial & <caracteres> \"prueba\""},
					Comment:   "testing escaping & formatting",
				},
			},
			Tests: []Test{
				{
					Results: []TestResult{
						{ItemID: 0, Result: "right"},
						{ItemID: 0, Result: "wrong"},
						{ItemID: 1, Result: "right"},
						{ItemID: 1, Result: "right"},
						{ItemID: 2, Result: "wrong"},
						{ItemID: 3, Result: "right"},
					},
				},
			},
		},
	}

	tempDir := t.TempDir()
	saver := NewFileSaver()

	// Test all supported formats
	formats := []struct {
		ext  string
		desc string
	}{
		{".csv", "CSV Spreadsheet"},
		{".ot", "OpenTeacher XML"},
		{".txt", "Plain Text"},
		{".json", "JSON Data"},
		{".t2k", "Teach2000 XML"},
		{".kvtml", "KVTML Vocabulary"},
		{".html", "HTML Web Page"},
		{".tex", "LaTeX Document"},
	}

	for _, format := range formats {
		t.Run(format.desc, func(t *testing.T) {
			testFile := filepath.Join(tempDir, fmt.Sprintf("test%s", format.ext))

			// Save file
			err := saver.SaveFile(lessonData, testFile)
			if err != nil {
				t.Fatalf("Failed to save %s file: %v", format.desc, err)
			}

			// Verify file exists
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Fatalf("%s file was not created", format.desc)
			}

			// Verify file has content
			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("Failed to get file info for %s: %v", format.desc, err)
			}
			if info.Size() == 0 {
				t.Fatalf("%s file is empty", format.desc)
			}

			// Read and verify basic content
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read %s file: %v", format.desc, err)
			}

			contentStr := string(content)

			// Verify vocabulary content is present (all formats should have this)
			if !strings.Contains(contentStr, "hello") && !strings.Contains(contentStr, "hola") {
				t.Errorf("%s file should contain vocabulary content", format.desc)
			}

			// Verify lesson title is present in formats that include it
			// Note: CSV and Teach2000 don't include lesson titles by design
			if format.ext != ".csv" && format.ext != ".t2k" {
				if !strings.Contains(contentStr, "Complete Export Test") &&
					!strings.Contains(contentStr, "Complete_Export_Test") {
					t.Errorf("%s file should contain lesson title", format.desc)
				}
			}

			t.Logf("✅ Successfully exported %s format (%d bytes)", format.desc, len(content))
		})
	}

	t.Logf("🎉 All %d export formats working correctly!", len(formats))
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
