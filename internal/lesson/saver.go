package lesson

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FileSaver provides file saving functionality for various lesson formats
type FileSaver struct {
	// Future: add configuration options, templates, etc.
}

// NewFileSaver creates a new FileSaver instance
func NewFileSaver() *FileSaver {
	return &FileSaver{}
}

// SaveFile saves lesson data to a file in the appropriate format based on extension
func (fs *FileSaver) SaveFile(lessonData *LessonData, filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	log.Printf("[ACTION] FileSaver.SaveFile() - saving to %s format", ext)

	switch ext {
	case ".csv":
		return fs.saveCSVFile(lessonData, filePath)
	case ".ot":
		return fs.saveOpenTeacherFile(lessonData, filePath)
	case ".txt":
		return fs.saveTextFile(lessonData, filePath)
	case ".json":
		return fs.saveJSONFile(lessonData, filePath)
	case ".t2k":
		return fs.saveTeach2000File(lessonData, filePath)
	case ".kvtml":
		return fs.saveKVTMLFile(lessonData, filePath)
	case ".html":
		return fs.saveHTMLFile(lessonData, filePath)
	case ".tex":
		return fs.saveLaTeXFile(lessonData, filePath)
	case ".ottp":
		return fs.saveOpenTeachingTopoFile(lessonData, filePath)
	case ".otmd":
		return fs.saveOpenTeachingMediaFile(lessonData, filePath)
	default:
		return fmt.Errorf("unsupported save format: %s", ext)
	}
}

// saveCSVFile saves lesson data as CSV format with proper headers and encoding
func (fs *FileSaver) saveCSVFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveCSVFile() - saving CSV file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create CSV file: %v", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Determine header names
	questionHeader := lessonData.List.QuestionLanguage
	if questionHeader == "" {
		questionHeader = "Questions"
	}

	answerHeader := lessonData.List.AnswerLanguage
	if answerHeader == "" {
		answerHeader = "Answers"
	}

	// Write CSV header
	headers := []string{
		questionHeader,
		answerHeader,
		"Comment",
		"Comment After Answering",
	}

	if err := writer.Write(headers); err != nil {
		log.Printf("[ERROR] Failed to write CSV header: %v", err)
		return err
	}

	// Write lesson items
	for _, item := range lessonData.List.Items {
		// Compose questions (join multiple questions with semicolon)
		questions := strings.Join(item.Questions, "; ")

		// Compose answers (join multiple answers with semicolon)
		answers := strings.Join(item.Answers, "; ")

		// Get comment
		comment := item.Comment

		// Comment after answering (placeholder - this field exists in OpenTeacher format)
		commentAfterAnswering := ""

		record := []string{
			questions,
			answers,
			comment,
			commentAfterAnswering,
		}

		if err := writer.Write(record); err != nil {
			log.Printf("[ERROR] Failed to write CSV record: %v", err)
			return err
		}
	}

	log.Printf("[SUCCESS] FileSaver.saveCSVFile() - saved %d items to CSV file", len(lessonData.List.Items))
	return nil
}

// WordStatistics represents test statistics for a word
type WordStatistics struct {
	Right int `json:"right"`
	Wrong int `json:"wrong"`
}

// calculateWordStatistics computes test statistics for each word item
func (fs *FileSaver) calculateWordStatistics(lessonData *LessonData) map[int]WordStatistics {
	stats := make(map[int]WordStatistics)

	// Initialize stats for all items
	for _, item := range lessonData.List.Items {
		stats[item.ID] = WordStatistics{Right: 0, Wrong: 0}
	}

	// Calculate statistics from test results
	for _, test := range lessonData.List.Tests {
		for _, result := range test.Results {
			if stat, exists := stats[result.ItemID]; exists {
				if result.Result == "right" {
					stat.Right++
				} else {
					stat.Wrong++
				}
				stats[result.ItemID] = stat
			}
		}
	}

	return stats
}

// OpenTeacherXML represents the structure of an OpenTeacher .ot file
type OpenTeacherXML struct {
	XMLName          xml.Name          `xml:"root"`
	Title            string            `xml:"title"`
	QuestionLanguage string            `xml:"question_language"`
	AnswerLanguage   string            `xml:"answer_language"`
	Words            []OpenTeacherWord `xml:"word"`
}

// OpenTeacherWord represents a word entry in OpenTeacher format
type OpenTeacherWord struct {
	Known   string `xml:"known"`
	Foreign string `xml:"foreign"`
	Second  string `xml:"second,omitempty"`
	Results string `xml:"results"`
}

// saveOpenTeacherFile saves lesson data in OpenTeacher (.ot) XML format
func (fs *FileSaver) saveOpenTeacherFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveOpenTeacherFile() - saving OpenTeacher file")

	// Calculate word statistics
	stats := fs.calculateWordStatistics(lessonData)

	// Create OpenTeacher XML structure
	otXML := OpenTeacherXML{
		Title:            lessonData.List.Title,
		QuestionLanguage: lessonData.List.QuestionLanguage,
		AnswerLanguage:   lessonData.List.AnswerLanguage,
		Words:            make([]OpenTeacherWord, 0, len(lessonData.List.Items)),
	}

	// Process each word item
	for _, item := range lessonData.List.Items {
		word := OpenTeacherWord{
			Known: strings.Join(item.Questions, ", "),
		}

		// Handle answers - separate first answer as "foreign", rest as "second"
		if len(item.Answers) > 0 {
			word.Foreign = item.Answers[0]
			if len(item.Answers) > 1 {
				word.Second = strings.Join(item.Answers[1:], ", ")
			}
		} else {
			word.Foreign = "-"
		}

		// Add test results in format "wrong/total"
		stat := stats[item.ID]
		total := stat.Right + stat.Wrong
		word.Results = fmt.Sprintf("%d/%d", stat.Wrong, total)

		otXML.Words = append(otXML.Words, word)
	}

	// Create file and write XML
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create OpenTeacher file: %v", err)
		return err
	}
	defer file.Close()

	// Write XML header
	if _, err := file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"); err != nil {
		return err
	}

	// Marshal and write XML content
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "\t")
	if err := encoder.Encode(otXML); err != nil {
		log.Printf("[ERROR] Failed to write OpenTeacher XML: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveOpenTeacherFile() - saved %d items to OpenTeacher file", len(lessonData.List.Items))
	return nil
}

// saveTextFile saves lesson data in plain text format
func (fs *FileSaver) saveTextFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveTextFile() - saving text file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create text file: %v", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write metadata header
	if lessonData.List.Title != "" {
		fmt.Fprintf(writer, "%s\n\n", lessonData.List.Title)
	}

	// Write language info
	if lessonData.List.QuestionLanguage != "" && lessonData.List.AnswerLanguage != "" {
		fmt.Fprintf(writer, "%s - %s\n\n", lessonData.List.QuestionLanguage, lessonData.List.AnswerLanguage)
	}

	if len(lessonData.List.Items) == 0 {
		log.Printf("[SUCCESS] FileSaver.saveTextFile() - saved empty text file")
		return nil
	}

	// Calculate maximum question length for alignment
	maxLen := 0
	for _, item := range lessonData.List.Items {
		questionText := strings.Join(item.Questions, ", ")
		if len(questionText) > maxLen {
			maxLen = len(questionText)
		}
	}
	maxLen += 1 // Add one space
	if maxLen < 8 {
		maxLen = 8 // Minimum spacing
	}

	// Write lesson items with aligned formatting
	for _, item := range lessonData.List.Items {
		questionText := strings.Join(item.Questions, ", ")
		answerText := strings.Join(item.Answers, ", ")

		// Create spacing
		spaces := maxLen - len(questionText)
		if spaces < 1 {
			spaces = 1
		}

		fmt.Fprintf(writer, "%s%s%s\n", questionText, strings.Repeat(" ", spaces), answerText)
	}

	log.Printf("[SUCCESS] FileSaver.saveTextFile() - saved %d items to text file", len(lessonData.List.Items))
	return nil
}

// saveJSONFile saves lesson data in JSON format
func (fs *FileSaver) saveJSONFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveJSONFile() - saving JSON file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create JSON file: %v", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(lessonData); err != nil {
		log.Printf("[ERROR] Failed to write JSON: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveJSONFile() - saved %d items to JSON file", len(lessonData.List.Items))
	return nil
}

// Teach2000XML represents the structure of a Teach2000 .t2k file
type Teach2000XML struct {
	XMLName     xml.Name             `xml:"teach2000"`
	Version     string               `xml:"version"`
	Description string               `xml:"description"`
	MessageData Teach2000MessageData `xml:"message_data"`
}

// Teach2000MessageData represents the message_data section
type Teach2000MessageData struct {
	Encrypted       string                `xml:"encrypted,attr"`
	MMFilesEmbedded string                `xml:"mm_files_embedded,attr"`
	FontQuestion    string                `xml:"font_question"`
	FontAnswer      string                `xml:"font_answer"`
	Items           Teach2000Items        `xml:"items"`
	TestResults     []Teach2000TestResult `xml:"testresults>testresult"`
	MapQuizFile     string                `xml:"mapquizfile"`
}

// Teach2000Items represents the items section
type Teach2000Items struct {
	Items []Teach2000Item `xml:"item"`
}

// Teach2000Item represents a single item in Teach2000 format
type Teach2000Item struct {
	ID           string              `xml:"id,attr"`
	Questions    []Teach2000Question `xml:"questions>question"`
	Answers      Teach2000Answers    `xml:"answers"`
	Errors       int                 `xml:"errors"`
	TestCount    int                 `xml:"testcount"`
	CorrectCount int                 `xml:"correctcount"`
}

// Teach2000Question represents a question element
type Teach2000Question struct {
	ID   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

// Teach2000Answers represents the answers section
type Teach2000Answers struct {
	Type    string            `xml:"type,attr"`
	Answers []Teach2000Answer `xml:"answer"`
}

// Teach2000Answer represents an answer element
type Teach2000Answer struct {
	ID   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

// Teach2000TestResult represents a test result in Teach2000 format
type Teach2000TestResult struct {
	Score              string `xml:"score"`
	Diff               int    `xml:"diff"`
	Comment            string `xml:"comment"`
	DateTime           string `xml:"dt"`
	Duration           string `xml:"duration"`
	AnswersCorrect     int    `xml:"answerscorrect"`
	WrongOnce          int    `xml:"wrongonce"`
	WrongTwice         int    `xml:"wrongtwice"`
	WrongMoreThanTwice int    `xml:"wrongmorethantwice"`
}

// calculateNote calculates Dutch grading system note (1-10 scale)
func (fs *FileSaver) calculateNote(rightAnswers, totalAnswers int) string {
	if totalAnswers == 0 {
		return "1.00000"
	}
	note := float64(rightAnswers)/float64(totalAnswers)*9.0 + 1.0
	return fmt.Sprintf("%.5f", note)
}

// composeTeach2000DateTime formats datetime for Teach2000 format
func (fs *FileSaver) composeTeach2000DateTime(t time.Time) string {
	// Teach2000 format: YYYY-MM-DDTHH:MM:SS.fff (3 decimal places for milliseconds)
	return t.Format("2006-01-02T15:04:05.000")
}

// composeTeach2000Duration formats duration for Teach2000 format using Pascal epoch
func (fs *FileSaver) composeTeach2000Duration(duration time.Duration) string {
	// Teach2000 uses Pascal epoch (1899-12-30) for duration
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	millis := int(duration.Milliseconds()) % 1000
	return fmt.Sprintf("1899-12-30T%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

// calculateAnswersCorrect counts correct answers in a test
func (fs *FileSaver) calculateAnswersCorrect(test Test) int {
	correct := 0
	for _, result := range test.Results {
		if result.Result == "right" {
			correct++
		}
	}
	return correct
}

// calculateWrongStats calculates wrong answer frequency statistics
func (fs *FileSaver) calculateWrongStats(test Test) (wrongOnce, wrongTwice, wrongMoreThanTwice int) {
	wrongCounts := make(map[int]int)

	// Count wrong answers per item
	for _, result := range test.Results {
		if result.Result == "wrong" {
			wrongCounts[result.ItemID]++
		}
	}

	// Count frequency of wrong answers
	for _, count := range wrongCounts {
		switch count {
		case 1:
			wrongOnce++
		case 2:
			wrongTwice++
		default:
			if count > 2 {
				wrongMoreThanTwice++
			}
		}
	}

	return
}

// saveTeach2000File saves lesson data in Teach2000 (.t2k) format
func (fs *FileSaver) saveTeach2000File(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveTeach2000File() - saving Teach2000 file")

	// Calculate word statistics
	stats := fs.calculateWordStatistics(lessonData)

	// Create Teach2000 XML structure
	t2kXML := Teach2000XML{
		Version:     "831",
		Description: "Normal",
		MessageData: Teach2000MessageData{
			Encrypted:       "N",
			MMFilesEmbedded: "N",
			FontQuestion:    "Arial",
			FontAnswer:      "Lucida Sans Unicode",
			MapQuizFile:     "",
		},
	}

	// Process items
	items := make([]Teach2000Item, 0, len(lessonData.List.Items))
	for _, item := range lessonData.List.Items {
		t2kItem := Teach2000Item{
			ID: strconv.Itoa(item.ID),
		}

		// Process questions
		for i, question := range item.Questions {
			t2kItem.Questions = append(t2kItem.Questions, Teach2000Question{
				ID:   strconv.Itoa(i),
				Text: question,
			})
		}

		// Process answers
		t2kItem.Answers.Type = "0"
		for i, answer := range item.Answers {
			t2kItem.Answers.Answers = append(t2kItem.Answers.Answers, Teach2000Answer{
				ID:   strconv.Itoa(i),
				Text: answer,
			})
		}

		// Add statistics
		stat := stats[item.ID]
		t2kItem.Errors = stat.Wrong
		t2kItem.TestCount = stat.Right + stat.Wrong
		t2kItem.CorrectCount = stat.Right

		items = append(items, t2kItem)
	}
	t2kXML.MessageData.Items.Items = items

	// Process test results
	testResults := make([]Teach2000TestResult, 0, len(lessonData.List.Tests))
	for _, test := range lessonData.List.Tests {
		rightAnswers := fs.calculateAnswersCorrect(test)
		totalAnswers := len(test.Results)
		wrongOnce, wrongTwice, wrongMoreThanTwice := fs.calculateWrongStats(test)

		// Use current time if no test date is available
		testTime := time.Now()
		if test.Date != nil {
			testTime = *test.Date
		}

		// Estimate duration (default to 5 minutes if not available)
		duration := 5 * time.Minute

		testResult := Teach2000TestResult{
			Score:              fs.calculateNote(rightAnswers, totalAnswers),
			Diff:               150, // Default difficulty
			Comment:            "",
			DateTime:           fs.composeTeach2000DateTime(testTime),
			Duration:           fs.composeTeach2000Duration(duration),
			AnswersCorrect:     rightAnswers,
			WrongOnce:          wrongOnce,
			WrongTwice:         wrongTwice,
			WrongMoreThanTwice: wrongMoreThanTwice,
		}

		testResults = append(testResults, testResult)
	}
	t2kXML.MessageData.TestResults = testResults

	// Create file and write XML
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create Teach2000 file: %v", err)
		return err
	}
	defer file.Close()

	// Write XML header with comment
	header := `<?xml version="1.0" encoding="UTF-8"?>
<!--This is a Teach2000 document (http://teach2000.memtrain.com)-->
`
	if _, err := file.WriteString(header); err != nil {
		return err
	}

	// Marshal and write XML content
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "\t")
	if err := encoder.Encode(t2kXML); err != nil {
		log.Printf("[ERROR] Failed to write Teach2000 XML: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveTeach2000File() - saved %d items to Teach2000 file", len(lessonData.List.Items))
	return nil
}

// KVTMLXML represents the structure of a KVTML (.kvtml) file
type KVTMLXML struct {
	XMLName     xml.Name          `xml:"kvtml"`
	Version     string            `xml:"version,attr"`
	Information KVTMLInfo         `xml:"information"`
	Identifiers []KVTMLIdentifier `xml:"identifiers>identifier"`
	Entries     []KVTMLEntry      `xml:"entries>entry"`
	Lessons     []KVTMLLesson     `xml:"lessons>container"`
}

// KVTMLInfo represents the information section
type KVTMLInfo struct {
	Generator string `xml:"generator"`
	Title     string `xml:"title"`
	Date      string `xml:"date,omitempty"`
	Category  string `xml:"category"`
}

// KVTMLIdentifier represents a language identifier
type KVTMLIdentifier struct {
	ID     string `xml:"id,attr"`
	Name   string `xml:"name"`
	Locale string `xml:"locale"`
}

// KVTMLEntry represents a vocabulary entry
type KVTMLEntry struct {
	ID           string             `xml:"id,attr"`
	Translations []KVTMLTranslation `xml:"translation"`
}

// KVTMLTranslation represents a translation in an entry
type KVTMLTranslation struct {
	ID      string `xml:"id,attr"`
	Text    string `xml:"text"`
	Comment string `xml:"comment,omitempty"`
}

// KVTMLLesson represents a lesson container
type KVTMLLesson struct {
	Name       string             `xml:"name"`
	InPractice string             `xml:"inpractice"`
	Entries    []KVTMLLessonEntry `xml:"entry"`
}

// KVTMLLessonEntry represents an entry reference in a lesson
type KVTMLLessonEntry struct {
	ID string `xml:"id,attr"`
}

// saveKVTMLFile saves lesson data in KVTML (.kvtml) format
func (fs *FileSaver) saveKVTMLFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveKVTMLFile() - saving KVTML file")

	// Create KVTML XML structure
	kvtmlXML := KVTMLXML{
		Version: "2.0",
		Information: KVTMLInfo{
			Generator: "Recuerdo",
			Title:     lessonData.List.Title,
			Category:  "Languages",
		},
		Identifiers: []KVTMLIdentifier{
			{
				ID:     "0",
				Name:   lessonData.List.QuestionLanguage,
				Locale: "C", // Default locale
			},
			{
				ID:     "1",
				Name:   lessonData.List.AnswerLanguage,
				Locale: "C", // Default locale
			},
		},
	}

	// Add creation date if available
	if len(lessonData.List.Items) > 0 {
		kvtmlXML.Information.Date = time.Now().Format("2006-01-02")
	}

	// Process entries
	entries := make([]KVTMLEntry, 0, len(lessonData.List.Items)+1)
	for _, item := range lessonData.List.Items {
		entry := KVTMLEntry{
			ID: strconv.Itoa(item.ID),
			Translations: []KVTMLTranslation{
				{
					ID:      "0",
					Text:    strings.Join(item.Questions, ", "),
					Comment: item.Comment,
				},
				{
					ID:   "1",
					Text: strings.Join(item.Answers, ", "),
				},
			},
		}
		entries = append(entries, entry)
	}

	// Add empty entry at the end (KVTML convention)
	entries = append(entries, KVTMLEntry{
		ID: strconv.Itoa(len(lessonData.List.Items)),
		Translations: []KVTMLTranslation{
			{ID: "0"},
			{ID: "1"},
		},
	})
	kvtmlXML.Entries = entries

	// Process lessons (tests)
	lessons := make([]KVTMLLesson, 0, len(lessonData.List.Tests))
	for i, test := range lessonData.List.Tests {
		lesson := KVTMLLesson{
			Name:       fmt.Sprintf("Lesson %d", i+1),
			InPractice: "true",
		}

		// Add entries that were tested
		testedItems := make(map[int]bool)
		for _, result := range test.Results {
			if !testedItems[result.ItemID] {
				lesson.Entries = append(lesson.Entries, KVTMLLessonEntry{
					ID: strconv.Itoa(result.ItemID),
				})
				testedItems[result.ItemID] = true
			}
		}

		lessons = append(lessons, lesson)
	}
	kvtmlXML.Lessons = lessons

	// Create file and write XML
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create KVTML file: %v", err)
		return err
	}
	defer file.Close()

	// Write XML header with DOCTYPE
	header := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE kvtml PUBLIC "kvtml2.dtd" "http://edu.kde.org/kvtml/kvtml2.dtd">
`
	if _, err := file.WriteString(header); err != nil {
		return err
	}

	// Marshal and write XML content
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(kvtmlXML); err != nil {
		log.Printf("[ERROR] Failed to write KVTML XML: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveKVTMLFile() - saved %d items to KVTML file", len(lessonData.List.Items))
	return nil
}

// saveHTMLFile saves lesson data in HTML format with modern styling
func (fs *FileSaver) saveHTMLFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveHTMLFile() - saving HTML file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create HTML file: %v", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write HTML header with modern CSS styling
	fmt.Fprintf(writer, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
            color: #333;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 2px solid #eee;
        }
        .title {
            font-size: 2.5em;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        .languages {
            font-size: 1.2em;
            color: #7f8c8d;
            font-style: italic;
        }
        .vocabulary-table {
            width: 100%%;
            border-collapse: collapse;
            margin-top: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .vocabulary-table th {
            background-color: #3498db;
            color: white;
            padding: 15px;
            text-align: left;
            font-weight: 600;
        }
        .vocabulary-table td {
            padding: 12px 15px;
            border-bottom: 1px solid #eee;
        }
        .vocabulary-table tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        .vocabulary-table tr:hover {
            background-color: #e3f2fd;
        }
        .question {
            font-weight: 500;
            color: #2c3e50;
        }
        .answer {
            color: #27ae60;
        }
        .comment {
            font-style: italic;
            color: #7f8c8d;
            font-size: 0.9em;
        }
        .stats {
            margin-top: 30px;
            text-align: center;
            color: #7f8c8d;
        }
        @media (max-width: 600px) {
            body { padding: 10px; }
            .vocabulary-table th,
            .vocabulary-table td {
                padding: 8px;
                font-size: 0.9em;
            }
            .title { font-size: 2em; }
        }
        @media print {
            body { max-width: none; }
            .vocabulary-table tr:hover { background-color: transparent !important; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">%s</h1>`, lessonData.List.Title, lessonData.List.Title)

	// Add language information
	if lessonData.List.QuestionLanguage != "" && lessonData.List.AnswerLanguage != "" {
		fmt.Fprintf(writer, `
        <div class="languages">%s → %s</div>`,
			lessonData.List.QuestionLanguage, lessonData.List.AnswerLanguage)
	}

	fmt.Fprintf(writer, `
    </div>

    <table class="vocabulary-table">
        <thead>
            <tr>
                <th>%s</th>
                <th>%s</th>`,
		getColumnHeader(lessonData.List.QuestionLanguage, "Questions"),
		getColumnHeader(lessonData.List.AnswerLanguage, "Answers"))

	// Add comment column if any items have comments
	hasComments := false
	for _, item := range lessonData.List.Items {
		if item.Comment != "" {
			hasComments = true
			break
		}
	}
	if hasComments {
		fmt.Fprintf(writer, `
                <th>Comment</th>`)
	}

	fmt.Fprintf(writer, `
            </tr>
        </thead>
        <tbody>`)

	// Write vocabulary items
	for _, item := range lessonData.List.Items {
		fmt.Fprintf(writer, `
            <tr>
                <td class="question">%s</td>
                <td class="answer">%s</td>`,
			htmlEscape(strings.Join(item.Questions, ", ")),
			htmlEscape(strings.Join(item.Answers, ", ")))

		if hasComments {
			comment := ""
			if item.Comment != "" {
				comment = htmlEscape(item.Comment)
			}
			fmt.Fprintf(writer, `
                <td class="comment">%s</td>`, comment)
		}

		fmt.Fprintf(writer, `
            </tr>`)
	}

	fmt.Fprintf(writer, `
        </tbody>
    </table>

    <div class="stats">
        <p>Total vocabulary items: %d</p>
    </div>

</body>
</html>`, len(lessonData.List.Items))

	log.Printf("[SUCCESS] FileSaver.saveHTMLFile() - saved %d items to HTML file", len(lessonData.List.Items))
	return nil
}

// htmlEscape escapes HTML special characters
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// getColumnHeader returns appropriate column header
func getColumnHeader(language, fallback string) string {
	if language != "" {
		return language
	}
	return fallback
}

// saveLaTeXFile saves lesson data in LaTeX format for academic/print use
func (fs *FileSaver) saveLaTeXFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveLaTeXFile() - saving LaTeX file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create LaTeX file: %v", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write LaTeX document header
	fmt.Fprintf(writer, `\documentclass[12pt,a4paper]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{longtable}
\usepackage{booktabs}
\usepackage{geometry}
\usepackage{fancyhdr}
\usepackage{xcolor}

\geometry{margin=2.5cm}
\pagestyle{fancy}
\fancyhf{}
\fancyhead[C]{%s}
\fancyfoot[C]{\thepage}

\definecolor{headercolor}{RGB}{52, 152, 219}

\begin{document}

\title{\textbf{%s}}`,
		latexEscape(lessonData.List.Title),
		latexEscape(lessonData.List.Title))

	// Add author and language information
	if lessonData.List.QuestionLanguage != "" && lessonData.List.AnswerLanguage != "" {
		fmt.Fprintf(writer, `
\author{%s $\rightarrow$ %s}`,
			latexEscape(lessonData.List.QuestionLanguage),
			latexEscape(lessonData.List.AnswerLanguage))
	}

	fmt.Fprintf(writer, `
\date{\today}

\maketitle

\section{Vocabulary List}

This document contains %d vocabulary items for study and reference.

`, len(lessonData.List.Items))

	// Determine if we need a comment column
	hasComments := false
	for _, item := range lessonData.List.Items {
		if item.Comment != "" {
			hasComments = true
			break
		}
	}

	// Start longtable
	if hasComments {
		fmt.Fprintf(writer, `\begin{longtable}{|p{0.3\textwidth}|p{0.3\textwidth}|p{0.3\textwidth}|}
\hline
\rowcolor{headercolor!20}
\textbf{%s} & \textbf{%s} & \textbf{Comment} \\
\hline
\endhead
`,
			latexEscape(getColumnHeader(lessonData.List.QuestionLanguage, "Questions")),
			latexEscape(getColumnHeader(lessonData.List.AnswerLanguage, "Answers")))
	} else {
		fmt.Fprintf(writer, `\begin{longtable}{|p{0.45\textwidth}|p{0.45\textwidth}|}
\hline
\rowcolor{headercolor!20}
\textbf{%s} & \textbf{%s} \\
\hline
\endhead
`,
			latexEscape(getColumnHeader(lessonData.List.QuestionLanguage, "Questions")),
			latexEscape(getColumnHeader(lessonData.List.AnswerLanguage, "Answers")))
	}

	// Write vocabulary items
	for i, item := range lessonData.List.Items {
		questions := latexEscape(strings.Join(item.Questions, ", "))
		answers := latexEscape(strings.Join(item.Answers, ", "))

		if hasComments {
			comment := latexEscape(item.Comment)
			fmt.Fprintf(writer, `%s & %s & \textit{%s} \\
`, questions, answers, comment)
		} else {
			fmt.Fprintf(writer, `%s & %s \\
`, questions, answers)
		}

		// Add horizontal line between rows
		if i < len(lessonData.List.Items)-1 {
			fmt.Fprintf(writer, `\hline
`)
		}
	}

	// End table and document
	fmt.Fprintf(writer, `\hline
\end{longtable}

\vspace{1cm}

\section{Statistics}

\begin{itemize}
\item Total vocabulary items: %d
\item Document generated on: \today
\end{itemize}

\end{document}
`, len(lessonData.List.Items))

	log.Printf("[SUCCESS] FileSaver.saveLaTeXFile() - saved %d items to LaTeX file", len(lessonData.List.Items))
	return nil
}

// latexEscape escapes LaTeX special characters
func latexEscape(s string) string {
	// Handle backslash first to avoid double-escaping
	s = strings.ReplaceAll(s, `\`, `\textbackslash{}`)

	// Then handle other special characters
	s = strings.ReplaceAll(s, `{`, `\{`)
	s = strings.ReplaceAll(s, `}`, `\}`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, `&`, `\&`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `#`, `\#`)
	s = strings.ReplaceAll(s, `^`, `\textasciicircum{}`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	s = strings.ReplaceAll(s, `~`, `\textasciitilde{}`)

	return s
}

// GetSupportedSaveExtensions returns a list of supported save file extensions
func (fs *FileSaver) GetSupportedSaveExtensions() []string {
	return []string{
		".csv",
		".ot",    // OpenTeacher format
		".txt",   // Plain text
		".json",  // JSON format
		".t2k",   // Teach2000 format
		".kvtml", // KDE Vocabulary Document
		".html",  // HTML export
		".tex",   // LaTeX export
		// Future formats to be implemented:
		// ".xml",   // Generic XML
		// ".pdf",   // PDF export (requires additional libraries)
	}
}

// GetSaveFormatName returns a human-readable name for a file extension
func (fs *FileSaver) GetSaveFormatName(ext string) string {
	switch strings.ToLower(ext) {
	case ".csv":
		return "Comma-Separated Values (Spreadsheet)"
	case ".ot":
		return "OpenTeacher 2.x Format"
	case ".t2k":
		return "Teach2000 Format"
	case ".kvtml":
		return "KDE Vocabulary Document"
	case ".xml":
		return "XML Document"
	case ".json":
		return "JSON Document"
	case ".txt":
		return "Plain Text File"
	case ".html":
		return "HTML Document"
	case ".pdf":
		return "PDF Document"
	case ".tex":
		return "LaTeX Document"
	default:
		return "Unknown Format"
	}
}

// GetSaveFilter returns a Qt-style file filter string for save dialogs
func (fs *FileSaver) GetSaveFilter() string {
	filters := []string{}

	for _, ext := range fs.GetSupportedSaveExtensions() {
		formatName := fs.GetSaveFormatName(ext)
		filter := fmt.Sprintf("%s (*%s)", formatName, ext)
		filters = append(filters, filter)
	}

	return strings.Join(filters, ";;")
}

// ValidateLessonData performs basic validation on lesson data before saving
func (fs *FileSaver) ValidateLessonData(lessonData *LessonData) error {
	if lessonData == nil {
		return fmt.Errorf("lesson data is nil")
	}

	if len(lessonData.List.Items) == 0 {
		return fmt.Errorf("lesson contains no items to save")
	}

	// Check for valid items
	validItems := 0
	for _, item := range lessonData.List.Items {
		if len(item.Questions) > 0 && len(item.Answers) > 0 {
			hasValidQuestion := false
			hasValidAnswer := false

			for _, q := range item.Questions {
				if strings.TrimSpace(q) != "" {
					hasValidQuestion = true
					break
				}
			}

			for _, a := range item.Answers {
				if strings.TrimSpace(a) != "" {
					hasValidAnswer = true
					break
				}
			}

			if hasValidQuestion && hasValidAnswer {
				validItems++
			}
		}
	}

	if validItems == 0 {
		return fmt.Errorf("lesson contains no valid question-answer pairs")
	}

	return nil
}

// SaveWithValidation saves lesson data with validation
func (fs *FileSaver) SaveWithValidation(lessonData *LessonData, filePath string) error {
	// Validate lesson data first
	if err := fs.ValidateLessonData(lessonData); err != nil {
		return fmt.Errorf("lesson validation failed: %w", err)
	}

	// Save the file
	return fs.SaveFile(lessonData, filePath)
}

// GetDefaultFilename generates a default filename based on lesson title
func (fs *FileSaver) GetDefaultFilename(lessonData *LessonData, ext string) string {
	if lessonData == nil || lessonData.List.Title == "" {
		return fmt.Sprintf("lesson%s", ext)
	}

	// Clean the title for use as filename
	title := lessonData.List.Title

	// Replace invalid filename characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		title = strings.ReplaceAll(title, char, "_")
	}

	// Remove leading/trailing spaces and dots
	title = strings.Trim(title, " .")

	// Limit length to reasonable filename size
	if len(title) > 50 {
		title = title[:50]
	}

	if title == "" {
		title = "lesson"
	}

	return fmt.Sprintf("%s%s", title, ext)
}

// saveOpenTeachingTopoFile saves lesson data as OpenTeaching Topography (.ottp) format
func (fs *FileSaver) saveOpenTeachingTopoFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveOpenTeachingTopoFile() - saving OpenTeaching Topo file")

	// Create the JSON structure for OpenTeacher format
	otData := map[string]interface{}{
		"file-format-version": "3.1",
		"items":               make([]map[string]interface{}, 0),
		"tests":               make([]interface{}, 0),
	}

	// Convert items to OpenTeacher format
	for _, item := range lessonData.List.Items {
		if x, y, hasCoords := item.GetTopoCoordinates(); hasCoords {
			otItem := map[string]interface{}{
				"id":   item.ID,
				"name": item.Name,
				"x":    x,
				"y":    y,
			}
			if item.Name == "" && len(item.Questions) > 0 {
				otItem["name"] = item.Questions[0]
			}
			otData["items"] = append(otData["items"].([]map[string]interface{}), otItem)
		}
	}

	// Create ZIP file
	zipFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create OTTP file: %v", err)
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add list.json to ZIP
	jsonWriter, err := zipWriter.Create("list.json")
	if err != nil {
		log.Printf("[ERROR] Failed to create list.json in ZIP: %v", err)
		return err
	}

	jsonData, err := json.MarshalIndent(otData, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal topo JSON: %v", err)
		return err
	}

	_, err = jsonWriter.Write(jsonData)
	if err != nil {
		log.Printf("[ERROR] Failed to write JSON to ZIP: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveOpenTeachingTopoFile() - saved %d topo items", len(otData["items"].([]map[string]interface{})))
	return nil
}

// saveOpenTeachingMediaFile saves lesson data as OpenTeaching Media (.otmd) format
func (fs *FileSaver) saveOpenTeachingMediaFile(lessonData *LessonData, filePath string) error {
	log.Printf("[ACTION] FileSaver.saveOpenTeachingMediaFile() - saving OpenTeaching Media file")

	// Create the JSON structure for OpenTeacher format
	otData := map[string]interface{}{
		"file-format-version": "3.1",
		"items":               make([]map[string]interface{}, 0),
		"tests":               make([]interface{}, 0),
	}

	// Convert items to OpenTeacher format
	for _, item := range lessonData.List.Items {
		if filename, remote, hasMedia := item.GetMediaInfo(); hasMedia || item.Name != "" {
			otItem := map[string]interface{}{
				"id":     item.ID,
				"name":   item.Name,
				"remote": remote,
			}

			if hasMedia {
				otItem["filename"] = filename
			}

			// Add questions and answers
			if len(item.Questions) > 0 {
				otItem["question"] = item.Questions[0]
			} else {
				otItem["question"] = ""
			}

			if len(item.Answers) > 0 {
				otItem["answer"] = item.Answers[0]
			} else {
				otItem["answer"] = ""
			}

			otData["items"] = append(otData["items"].([]map[string]interface{}), otItem)
		}
	}

	// Create ZIP file
	zipFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create OTMD file: %v", err)
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add list.json to ZIP
	jsonWriter, err := zipWriter.Create("list.json")
	if err != nil {
		log.Printf("[ERROR] Failed to create list.json in ZIP: %v", err)
		return err
	}

	jsonData, err := json.MarshalIndent(otData, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal media JSON: %v", err)
		return err
	}

	_, err = jsonWriter.Write(jsonData)
	if err != nil {
		log.Printf("[ERROR] Failed to write JSON to ZIP: %v", err)
		return err
	}

	log.Printf("[SUCCESS] FileSaver.saveOpenTeachingMediaFile() - saved %d media items", len(otData["items"].([]map[string]interface{})))
	return nil
}
