package lesson

import (
	"time"
)

// WordItem represents a single word/question-answer pair in a lesson
type WordItem struct {
	ID        int      `json:"id"`
	Questions []string `json:"questions"`
	Answers   []string `json:"answers"`
	Comment   string   `json:"comment,omitempty"`
	Name      string   `json:"name,omitempty"`
	// Topo-specific fields (optional)
	X *int `json:"x,omitempty"`
	Y *int `json:"y,omitempty"`
	// Media-specific fields (optional)
	Filename *string `json:"filename,omitempty"`
	Remote   *bool   `json:"remote,omitempty"`
}

// TopoItem represents a single topography item with coordinates
type TopoItem struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Questions []string `json:"questions,omitempty"`
	Answers   []string `json:"answers,omitempty"`
}

// MediaItem represents a single media item with file information
type MediaItem struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Questions []string `json:"questions"`
	Answers   []string `json:"answers"`
	Filename  string   `json:"filename,omitempty"`
	Remote    bool     `json:"remote,omitempty"`
}

// TestResult represents a single test result for an item
type TestResult struct {
	Result string     `json:"result"` // "right" or "wrong"
	ItemID int        `json:"itemId"`
	Time   *time.Time `json:"time,omitempty"`
}

// Test represents a collection of test results
type Test struct {
	Results []TestResult `json:"results"`
	Date    *time.Time   `json:"date,omitempty"`
}

// WordList represents the core lesson data structure
type WordList struct {
	Title            string     `json:"title,omitempty"`
	QuestionLanguage string     `json:"questionLanguage,omitempty"`
	AnswerLanguage   string     `json:"answerLanguage,omitempty"`
	Items            []WordItem `json:"items"`
	Tests            []Test     `json:"tests"`
}

// LessonData represents the complete lesson data as returned by loaders
type LessonData struct {
	List      WordList               `json:"list"`
	Resources map[string]interface{} `json:"resources"`
	Changed   bool                   `json:"changed,omitempty"`
}

// Lesson represents a lesson instance in the application
type Lesson struct {
	Data     LessonData
	Path     string
	DataType string // "words", "media", "topo", etc.
}

// NewWordList creates a new empty word list
func NewWordList() *WordList {
	return &WordList{
		Items: make([]WordItem, 0),
		Tests: make([]Test, 0),
	}
}

// NewLessonData creates a new lesson data structure
func NewLessonData() *LessonData {
	return &LessonData{
		List:      *NewWordList(),
		Resources: make(map[string]interface{}),
		Changed:   false,
	}
}

// NewLesson creates a new lesson instance
func NewLesson(dataType string) *Lesson {
	return &Lesson{
		Data:     *NewLessonData(),
		DataType: dataType,
	}
}

// AddWordItem adds a word item to the lesson
func (wl *WordList) AddWordItem(questions, answers []string, comment string) {
	item := WordItem{
		ID:        len(wl.Items),
		Questions: questions,
		Answers:   answers,
		Comment:   comment,
	}
	wl.Items = append(wl.Items, item)
}

// AddTopoItem adds a topography item to the lesson using extended WordItem
func (wl *WordList) AddTopoItem(name string, x, y int, questions, answers []string) {
	item := WordItem{
		ID:        len(wl.Items),
		Questions: questions,
		Answers:   answers,
		Name:      name,
		X:         &x,
		Y:         &y,
	}
	wl.Items = append(wl.Items, item)
}

// AddMediaItem adds a media item to the lesson using extended WordItem
func (wl *WordList) AddMediaItem(name string, questions, answers []string, filename string, remote bool) {
	item := WordItem{
		ID:        len(wl.Items),
		Questions: questions,
		Answers:   answers,
		Name:      name,
		Filename:  &filename,
		Remote:    &remote,
	}
	wl.Items = append(wl.Items, item)
}

// GetWordCount returns the number of word items in the lesson
func (wl *WordList) GetWordCount() int {
	return len(wl.Items)
}

// GetTestCount returns the number of tests in the lesson
func (wl *WordList) GetTestCount() int {
	return len(wl.Tests)
}

// IsTopoItem returns true if this item has topography data
func (wi *WordItem) IsTopoItem() bool {
	return wi.X != nil && wi.Y != nil
}

// IsMediaItem returns true if this item has media data
func (wi *WordItem) IsMediaItem() bool {
	return wi.Filename != nil
}

// GetTopoCoordinates returns the X,Y coordinates if this is a topo item
func (wi *WordItem) GetTopoCoordinates() (int, int, bool) {
	if wi.X != nil && wi.Y != nil {
		return *wi.X, *wi.Y, true
	}
	return 0, 0, false
}

// GetMediaInfo returns the filename and remote status if this is a media item
func (wi *WordItem) GetMediaInfo() (string, bool, bool) {
	if wi.Filename != nil {
		remote := false
		if wi.Remote != nil {
			remote = *wi.Remote
		}
		return *wi.Filename, remote, true
	}
	return "", false, false
}

// AddTestResult adds a test result to the lesson
func (wl *WordList) AddTestResult(itemID int, result string) {
	testResult := TestResult{
		Result: result,
		ItemID: itemID,
		Time:   &time.Time{},
	}

	// Add to the most recent test, or create a new one
	if len(wl.Tests) == 0 {
		wl.Tests = append(wl.Tests, Test{
			Results: []TestResult{testResult},
			Date:    &time.Time{},
		})
	} else {
		lastTest := &wl.Tests[len(wl.Tests)-1]
		lastTest.Results = append(lastTest.Results, testResult)
	}
}

// GetRightAnswersCount returns the number of correct answers for an item
func (wl *WordList) GetRightAnswersCount(itemID int) int {
	count := 0
	for _, test := range wl.Tests {
		for _, result := range test.Results {
			if result.ItemID == itemID && result.Result == "right" {
				count++
			}
		}
	}
	return count
}

// GetWrongAnswersCount returns the number of incorrect answers for an item
func (wl *WordList) GetWrongAnswersCount(itemID int) int {
	count := 0
	for _, test := range wl.Tests {
		for _, result := range test.Results {
			if result.ItemID == itemID && result.Result == "wrong" {
				count++
			}
		}
	}
	return count
}
