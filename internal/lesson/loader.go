package lesson

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// FileLoader provides file loading functionality for various lesson formats
type FileLoader struct{}

// NewFileLoader creates a new file loader instance
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// LoadFile loads a lesson file and returns lesson data
func (fl *FileLoader) LoadFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.LoadFile() - loading file: %s", filePath)

	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv", ".tsv":
		return fl.loadCSV(filePath)
	case ".txt":
		return fl.loadTextFile(filePath)
	case ".ot", ".otwd":
		return fl.loadOpenTeacherFile(filePath)
	case ".json":
		return fl.loadJSONFile(filePath)
	case ".kvtml":
		return fl.loadKVTMLFile(filePath)
	case ".anki", ".anki2", ".db":
		return fl.loadSQLiteFile(filePath)
	case ".t2k":
		return fl.loadTeach2000File(filePath)
	case ".jvlt":
		return fl.loadJVLTFile(filePath)
	case ".fq":
		return fl.loadFlashQardFile(filePath)
	case ".vok2":
		return fl.loadTeachMasterFile(filePath)
	case ".wcu":
		return fl.loadCueCardFile(filePath)
	case ".backpack":
		return fl.loadBackpackFile(filePath)
	case ".xml":
		return fl.loadXMLFile(filePath)
	case ".kgm":
		return fl.loadKGeographyMapFile(filePath)
	case ".ottp":
		return fl.loadOpenTeachingTopoFile(filePath)
	case ".otmd":
		return fl.loadOpenTeachingMediaFile(filePath)
	default:
		// Try to auto-detect format by content
		return fl.loadAutoDetect(filePath)
	}
}

// GetFileType determines the lesson type for a given file path
func (fl *FileLoader) GetFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv", ".tsv", ".txt", ".ot", ".json":
		return "words"
	case ".anki", ".anki2", ".apkg", ".backpack", ".wcu", ".voc", ".fq", ".fmd":
		return "words"
	case ".dkf", ".jml", ".jvlt", ".kvtml", ".stp", ".db", ".oh", ".ohw", ".oh4":
		return "words"
	case ".ovr", ".pau", ".vok2", ".wdl", ".vtl3", ".wrts", ".xml", ".otwd":
		return "words"
	case ".kgm", ".ottp":
		return "topo"
	case ".t2k":
		return "words" // Can be both words and topo, defaulting to words
	case ".otmd":
		return "media"
	default:
		return "words" // Default to words for unknown formats
	}
}

// loadCSV loads CSV or TSV files
func (fl *FileLoader) loadCSV(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadCSV() - parsing CSV file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open CSV file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Determine delimiter
	delimiter := ','
	if strings.HasSuffix(strings.ToLower(filePath), ".tsv") {
		delimiter = '\t'
	}

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	itemID := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[WARNING] Error reading CSV line: %v", err)
			continue
		}

		if len(record) < 2 {
			continue // Skip lines with insufficient data
		}

		// Parse questions and answers (may be comma-separated within cells)
		questions := fl.parseWordString(strings.TrimSpace(record[0]))
		answers := fl.parseWordString(strings.TrimSpace(record[1]))

		comment := ""
		if len(record) > 2 {
			comment = strings.TrimSpace(record[2])
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        itemID,
				Questions: questions,
				Answers:   answers,
				Comment:   comment,
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadCSV() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadTextFile loads simple text files with word pairs
func (fl *FileLoader) loadTextFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadTextFile() - parsing text file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open text file: %v", err)
		return nil, err
	}
	defer file.Close()

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	itemID := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Try different separators
		var questions, answers []string
		var comment string

		if strings.Contains(line, "\t") {
			// Tab-separated
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(parts[1])
				if len(parts) > 2 {
					comment = strings.Join(parts[2:], " ")
				}
			}
		} else if strings.Contains(line, "|") {
			// Pipe-separated
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(parts[1])
				if len(parts) > 2 {
					comment = strings.Join(parts[2:], " ")
				}
			}
		} else if strings.Contains(line, "=") {
			// Equals-separated
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(strings.Join(parts[1:], "="))
			}
		} else if strings.Contains(line, ":") {
			// Colon-separated (GNU VocabTrain format)
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(strings.Join(parts[1:], ":"))
			}
		} else {
			// Try to find other patterns or skip
			continue
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        itemID,
				Questions: questions,
				Answers:   answers,
				Comment:   comment,
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Error reading text file: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] FileLoader.loadTextFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadOpenTeacherFile loads OpenTeacher .ot XML files
func (fl *FileLoader) loadOpenTeacherFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadOpenTeacherFile() - parsing .ot XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open .ot file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Simple XML structure for .ot files
	type OTWord struct {
		Known   string `xml:"known"`
		Foreign string `xml:"foreign"`
		Second  string `xml:"second"`
		Results string `xml:"results"`
	}

	type OTRoot struct {
		XMLName          xml.Name `xml:"openteaching-words-file"`
		Title            string   `xml:"title"`
		QuestionLanguage string   `xml:"question_language"`
		AnswerLanguage   string   `xml:"answer_language"`
		Words            []OTWord `xml:"word"`
	}

	// OpenTeacher 2.x format uses <root> instead of <openteaching-words-file>
	type OTRoot2x struct {
		XMLName          xml.Name `xml:"root"`
		Title            string   `xml:"title"`
		QuestionLanguage string   `xml:"question_language"`
		AnswerLanguage   string   `xml:"answer_language"`
		Words            []OTWord `xml:"word"`
	}

	var root OTRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		// Try OpenTeacher 2.x format with <root> element
		file.Seek(0, 0) // Reset file position
		var root2x OTRoot2x
		decoder2x := xml.NewDecoder(file)
		if err2 := decoder2x.Decode(&root2x); err2 != nil {
			log.Printf("[ERROR] Failed to parse .ot XML in both 3.x and 2.x formats: %v, %v", err, err2)
			return nil, err
		}
		// Convert 2.x format to 3.x format structure
		root.Title = root2x.Title
		root.QuestionLanguage = root2x.QuestionLanguage
		root.AnswerLanguage = root2x.AnswerLanguage
		root.Words = root2x.Words
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Title
	lessonData.List.QuestionLanguage = root.QuestionLanguage
	lessonData.List.AnswerLanguage = root.AnswerLanguage

	// Create a test for storing results
	test := Test{Results: []TestResult{}}

	for i, word := range root.Words {
		questions := fl.parseWordString(word.Known)

		// Handle foreign and second fields
		foreignText := word.Foreign
		if word.Second != "" {
			foreignText += ", " + word.Second
		}
		answers := fl.parseWordString(foreignText)

		item := WordItem{
			ID:        i,
			Questions: questions,
			Answers:   answers,
			Comment:   "",
		}
		lessonData.List.Items = append(lessonData.List.Items, item)

		// Parse results (format: "wrong/total")
		if word.Results != "" {
			parts := strings.Split(word.Results, "/")
			if len(parts) == 2 {
				if wrong, err := strconv.Atoi(parts[0]); err == nil {
					if total, err := strconv.Atoi(parts[1]); err == nil {
						right := total - wrong

						// Add right results
						for j := 0; j < right; j++ {
							test.Results = append(test.Results, TestResult{
								Result: "right",
								ItemID: i,
							})
						}

						// Add wrong results
						for j := 0; j < wrong; j++ {
							test.Results = append(test.Results, TestResult{
								Result: "wrong",
								ItemID: i,
							})
						}
					}
				}
			}
		}
	}

	// Only add test if it has results
	if len(test.Results) > 0 {
		lessonData.List.Tests = append(lessonData.List.Tests, test)
	}

	log.Printf("[SUCCESS] FileLoader.loadOpenTeacherFile() - loaded %d word pairs with %d test results",
		len(lessonData.List.Items), len(test.Results))
	return lessonData, nil
}

// loadJSONFile loads JSON-formatted lesson files
func (fl *FileLoader) loadJSONFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadJSONFile() - parsing JSON file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open JSON file: %v", err)
		return nil, err
	}
	defer file.Close()

	var lessonData LessonData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&lessonData); err != nil {
		log.Printf("[ERROR] Failed to parse JSON: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] FileLoader.loadJSONFile() - loaded %d word pairs", len(lessonData.List.Items))
	return &lessonData, nil
}

// loadAutoDetect attempts to auto-detect file format and load accordingly
func (fl *FileLoader) loadAutoDetect(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadAutoDetect() - attempting to auto-detect format")

	// Try different loaders in order of likelihood
	loaders := []func(string) (*LessonData, error){
		fl.loadCSV,
		fl.loadTextFile,
		fl.loadJSONFile,
	}

	var lastErr error
	for _, loader := range loaders {
		if data, err := loader(filePath); err == nil {
			return data, nil
		} else {
			lastErr = err
		}
	}

	log.Printf("[ERROR] FileLoader.loadAutoDetect() - all format detection failed")
	return nil, fmt.Errorf("unable to detect file format: %v", lastErr)
}

// parseWordString parses a string containing potentially multiple words/phrases
// This mimics the Python wordsStringParser functionality
func (fl *FileLoader) parseWordString(input string) []string {
	if input == "" {
		return []string{}
	}

	input = strings.TrimSpace(input)

	// Handle multiple separators
	var words []string

	// First try semicolon separation (most explicit)
	if strings.Contains(input, ";") {
		parts := strings.Split(input, ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				words = append(words, trimmed)
			}
		}
	} else if strings.Contains(input, ",") {
		// Try comma separation
		parts := strings.Split(input, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				words = append(words, trimmed)
			}
		}
	} else {
		// Single word/phrase
		words = []string{input}
	}

	return words
}

// loadKVTMLFile loads KVTML (KDE Vocabulary Document) files
func (fl *FileLoader) loadKVTMLFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadKVTMLFile() - parsing KVTML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open KVTML file: %v", err)
		return nil, err
	}
	defer file.Close()

	// KVTML XML structure
	type KVTMLTranslation struct {
		ID   string `xml:"id,attr"`
		Text string `xml:"text"`
	}

	type KVTMLEntry struct {
		ID           string             `xml:"id,attr"`
		Translations []KVTMLTranslation `xml:"translation"`
	}

	type KVTMLIdentifier struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name"`
	}

	type KVTMLInformation struct {
		Title   string `xml:"title"`
		Author  string `xml:"author"`
		Comment string `xml:"comment"`
	}

	type KVTMLRoot struct {
		XMLName     xml.Name          `xml:"kvtml"`
		Version     string            `xml:"version,attr"`
		Information KVTMLInformation  `xml:"information"`
		Identifiers []KVTMLIdentifier `xml:"identifiers>identifier"`
		Entries     []KVTMLEntry      `xml:"entries>entry"`
	}

	var root KVTMLRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse KVTML XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Information.Title
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	// Set language names if available
	if len(root.Identifiers) >= 2 {
		lessonData.List.QuestionLanguage = root.Identifiers[0].Name
		lessonData.List.AnswerLanguage = root.Identifiers[1].Name
	}

	// Process entries
	for i, entry := range root.Entries {
		var questions, answers []string

		// Find question and answer translations
		for _, translation := range entry.Translations {
			if translation.ID == "0" && translation.Text != "" {
				questions = fl.parseWordString(translation.Text)
			} else if translation.ID == "1" && translation.Text != "" {
				answers = fl.parseWordString(translation.Text)
			}
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        i,
				Questions: questions,
				Answers:   answers,
				Comment:   "",
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadKVTMLFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadAnkiFile loads Anki database files
// loadSQLiteFile loads SQLite-based files (.anki, .anki2, .db for Mnemosyne)
func (fl *FileLoader) loadSQLiteFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadSQLiteFile() - parsing SQLite database file")

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite database: %v", err)
		return nil, err
	}
	defer db.Close()

	// Check if this is an Anki database (has notes and cards tables)
	if fl.isAnkiDatabase(db) {
		return fl.loadAnkiDatabase(db, filePath)
	}

	// Check if this is a Mnemosyne database (has cards table)
	if fl.isMnemosyseDatabase(db) {
		return fl.loadMnemosyseDatabase(db, filePath)
	}

	log.Printf("[WARNING] Unknown SQLite database format")
	return fl.loadGenericSQLiteDatabase(db, filePath)
}

// isAnkiDatabase checks if the database has Anki-specific tables
func (fl *FileLoader) isAnkiDatabase(db *sql.DB) bool {
	// Check for Anki-specific combination: must have both fields AND cards tables
	// Or notes table (Anki 2.x)
	var hasFields, hasCards, hasNotes bool

	query := `SELECT name FROM sqlite_master WHERE type='table' AND name IN ('fields', 'cards', 'notes')`
	rows, err := db.Query(query)
	if err != nil {
		return false
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		switch tableName {
		case "fields":
			hasFields = true
		case "cards":
			hasCards = true
		case "notes":
			hasNotes = true
		}
	}

	// Anki 1.x has fields+cards, Anki 2.x has notes+cards
	return (hasFields && hasCards) || (hasNotes && hasCards)
}

// isMnemosyseDatabase checks if the database has Mnemosyne-specific structure
func (fl *FileLoader) isMnemosyseDatabase(db *sql.DB) bool {
	// Mnemosyne has facts and data_for_fact tables, but NOT fields table
	var hasFacts, hasDataForFact, hasFields bool

	query := `SELECT name FROM sqlite_master WHERE type='table' AND name IN ('facts', 'data_for_fact', 'fields')`
	rows, err := db.Query(query)
	if err != nil {
		return false
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		switch tableName {
		case "facts":
			hasFacts = true
		case "data_for_fact":
			hasDataForFact = true
		case "fields":
			hasFields = true
		}
	}

	// Mnemosyne has facts+data_for_fact but NOT fields (which Anki has)
	return hasFacts && hasDataForFact && !hasFields
}

// loadAnkiDatabase loads an Anki SQLite database
func (fl *FileLoader) loadAnkiDatabase(db *sql.DB, filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadAnkiDatabase() - parsing Anki SQLite database")

	lessonData := NewLessonData()
	lessonData.List.Title = fl.extractAnkiDeckName(db, filepath.Base(filePath))

	// Check if this is Anki 2.x (has notes table)
	var hasNotes bool
	checkQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name='notes'`
	checkRows, _ := db.Query(checkQuery)
	if checkRows != nil {
		hasNotes = checkRows.Next()
		checkRows.Close()
	}

	var rows *sql.Rows
	var err error

	if hasNotes {
		// Anki 2.x format
		query := `SELECT DISTINCT n.flds FROM notes n JOIN cards c ON n.id = c.nid WHERE c.queue != -1 LIMIT 1000`
		rows, err = db.Query(query)
	} else {
		// Anki 1.x format - get question/answer pairs from fields table
		query := `
			SELECT
				q.value as question,
				a.value as answer
			FROM facts fa
			JOIN fields q ON fa.id = q.factId AND q.ordinal = 0
			JOIN fields a ON fa.id = a.factId AND a.ordinal = 1
			WHERE q.value != '' AND a.value != ''
			LIMIT 1000`
		rows, err = db.Query(query)
	}
	if err != nil {
		log.Printf("[ERROR] Failed to query Anki database: %v", err)
		return lessonData, nil // Return empty lesson rather than error
	}
	defer rows.Close()

	itemID := 0
	for rows.Next() {
		if hasNotes {
			// Anki 2.x format - fields are tab-separated
			var fields string
			if err := rows.Scan(&fields); err != nil {
				log.Printf("[WARNING] Error scanning Anki 2.x row: %v", err)
				continue
			}

			// Split fields on \x1f separator (Anki 2.x) or tab
			fieldList := strings.Split(fields, "\x1f")
			if len(fieldList) < 2 {
				fieldList = strings.Split(fields, "\t")
			}

			if len(fieldList) >= 2 {
				cleanQuestion := fl.stripHTMLTags(strings.TrimSpace(fieldList[0]))
				cleanAnswer := fl.stripHTMLTags(strings.TrimSpace(fieldList[1]))

				if len(cleanQuestion) > 0 && len(cleanAnswer) > 0 {
					item := WordItem{
						ID:        itemID,
						Questions: []string{cleanQuestion},
						Answers:   []string{cleanAnswer},
						Comment:   "",
					}
					lessonData.List.Items = append(lessonData.List.Items, item)
					itemID++
				}
			}
		} else {
			// Anki 1.x format - separate question/answer fields
			var question, answer string
			if err := rows.Scan(&question, &answer); err != nil {
				log.Printf("[WARNING] Error scanning Anki 1.x row: %v", err)
				continue
			}

			cleanQuestion := fl.stripHTMLTags(strings.TrimSpace(question))
			cleanAnswer := fl.stripHTMLTags(strings.TrimSpace(answer))

			if len(cleanQuestion) > 0 && len(cleanAnswer) > 0 {
				item := WordItem{
					ID:        itemID,
					Questions: []string{cleanQuestion},
					Answers:   []string{cleanAnswer},
					Comment:   "",
				}
				lessonData.List.Items = append(lessonData.List.Items, item)
				itemID++
			}
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadAnkiDatabase() - loaded %d word pairs from Anki database", len(lessonData.List.Items))
	return lessonData, nil
}

// extractAnkiDeckName extracts the deck name from Anki database
func (fl *FileLoader) extractAnkiDeckName(db *sql.DB, fallbackName string) string {
	// Check if this is Anki 2.x (has col table with JSON data)
	var colData string
	query := `SELECT decks FROM col LIMIT 1`
	err := db.QueryRow(query).Scan(&colData)
	if err == nil && colData != "" {
		// Parse JSON to extract deck names
		// Simple extraction - look for "name": "deckname" pattern
		start := strings.Index(colData, `"name": "`)
		if start != -1 {
			start += 9 // length of `"name": "`
			end := strings.Index(colData[start:], `"`)
			if end != -1 && start+end < len(colData) {
				deckName := colData[start : start+end]
				if deckName != "Default" && deckName != "Standaard" && deckName != "" {
					log.Printf("[SUCCESS] FileLoader.extractAnkiDeckName() - found deck name: %s", deckName)
					return deckName
				}
			}
		}
	}

	// Check if this is Anki 1.x (has decks table)
	var description string
	query = `SELECT description FROM decks WHERE description != '' LIMIT 1`
	err = db.QueryRow(query).Scan(&description)
	if err == nil && description != "" {
		log.Printf("[SUCCESS] FileLoader.extractAnkiDeckName() - found deck description: %s", description)
		return description
	}

	// Fallback to filename
	log.Printf("[WARNING] FileLoader.extractAnkiDeckName() - using fallback name: %s", fallbackName)
	return fallbackName
}

// loadMnemosyseDatabase loads a Mnemosyne SQLite database
func (fl *FileLoader) loadMnemosyseDatabase(db *sql.DB, filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadMnemosyseDatabase() - parsing Mnemosyne SQLite database")

	lessonData := NewLessonData()
	lessonData.List.Title = fl.extractMnemosyneDeckName(db, filepath.Base(filePath))

	// Query Mnemosyne database - try different schema versions
	query := `
		SELECT
			COALESCE(q.value, '') as question,
			COALESCE(a.value, '') as answer,
			'' as tags
		FROM facts f
		LEFT JOIN data_for_fact q ON f._id = q._fact_id AND q.key = 'f'
		LEFT JOIN data_for_fact a ON f._id = a._fact_id AND a.key = 'b'
		WHERE q.value IS NOT NULL AND a.value IS NOT NULL
		LIMIT 1000`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[ERROR] Failed to query Mnemosyne database: %v", err)
		return lessonData, nil // Return empty lesson rather than error
	}
	defer rows.Close()

	itemID := 0
	for rows.Next() {
		var question, answer, tags string
		if err := rows.Scan(&question, &answer, &tags); err != nil {
			log.Printf("[WARNING] Error scanning Mnemosyne row: %v", err)
			continue
		}

		if len(strings.TrimSpace(question)) > 0 && len(strings.TrimSpace(answer)) > 0 {
			item := WordItem{
				ID:        itemID,
				Questions: []string{strings.TrimSpace(question)},
				Answers:   []string{strings.TrimSpace(answer)},
				Comment:   strings.TrimSpace(tags),
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadMnemosyseDatabase() - loaded %d word pairs from Mnemosyne database", len(lessonData.List.Items))
	return lessonData, nil
}

// extractMnemosyneDeckName extracts the deck name from Mnemosyne database
func (fl *FileLoader) extractMnemosyneDeckName(db *sql.DB, fallbackName string) string {
	// Try to get deck name from categories table if it exists
	var name string
	query := `SELECT name FROM categories WHERE _id = 1 LIMIT 1`
	err := db.QueryRow(query).Scan(&name)
	if err == nil && name != "" {
		log.Printf("[SUCCESS] FileLoader.extractMnemosyneDeckName() - found category name: %s", name)
		return name
	}

	// Try to get name from any configuration or metadata if available
	query = `SELECT value FROM global_variables WHERE key = 'name' OR key = 'title' LIMIT 1`
	err = db.QueryRow(query).Scan(&name)
	if err == nil && name != "" {
		log.Printf("[SUCCESS] FileLoader.extractMnemosyneDeckName() - found deck name: %s", name)
		return name
	}

	// Fallback to filename
	log.Printf("[WARNING] FileLoader.extractMnemosyneDeckName() - using fallback name: %s", fallbackName)
	return fallbackName
}

// loadGenericSQLiteDatabase attempts to load any SQLite database with basic heuristics
func (fl *FileLoader) loadGenericSQLiteDatabase(db *sql.DB, filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadGenericSQLiteDatabase() - parsing generic SQLite database")

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	log.Printf("[WARNING] Generic SQLite parsing not implemented - trying CSV fallback")
	db.Close()
	return fl.loadAutoDetect(filePath)
}

// stripHTMLTags removes basic HTML tags from text
func (fl *FileLoader) stripHTMLTags(text string) string {
	// Simple HTML tag removal
	result := text
	// Remove common HTML tags
	replacements := [][]string{
		{"<br>", " "},
		{"<br/>", " "},
		{"<div>", ""},
		{"</div>", ""},
		{"<b>", ""},
		{"</b>", ""},
		{"<i>", ""},
		{"</i>", ""},
		{"<u>", ""},
		{"</u>", ""},
	}

	for _, replacement := range replacements {
		result = strings.ReplaceAll(result, replacement[0], replacement[1])
	}

	// Remove any remaining tags with a simple regex-like approach
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}

	return strings.TrimSpace(result)
}

// loadXMLFile loads XML files (including ABBYY format)
func (fl *FileLoader) loadXMLFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadXMLFile() - parsing XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open XML file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Basic XML structure for simple word lists
	type XMLWord struct {
		XMLName xml.Name `xml:"word"`
		Known   string   `xml:"known"`
		Foreign string   `xml:"foreign"`
		Text    string   `xml:",chardata"`
	}

	type XMLRoot struct {
		XMLName xml.Name  `xml:"root"`
		Title   string    `xml:"title"`
		Words   []XMLWord `xml:"word"`
	}

	// Try to parse as simple XML word list
	var root XMLRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		// If XML parsing fails, try as text file
		log.Printf("[WARNING] XML parsing failed, trying as text file: %v", err)
		return fl.loadTextFile(filePath)
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Title
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	for i, word := range root.Words {
		var questions, answers []string

		if word.Known != "" {
			questions = fl.parseWordString(word.Known)
		}
		if word.Foreign != "" {
			answers = fl.parseWordString(word.Foreign)
		}
		if word.Text != "" && len(questions) == 0 {
			questions = fl.parseWordString(word.Text)
		}

		if len(questions) > 0 || len(answers) > 0 {
			item := WordItem{
				ID:        i,
				Questions: questions,
				Answers:   answers,
				Comment:   "",
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadXMLFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadTeach2000File parses Teach2000 (.t2k) XML files
func (fl *FileLoader) loadTeach2000File(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadTeach2000File() - parsing Teach2000 XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open Teach2000 file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Teach2000 XML structure
	type Teach2000Question struct {
		ID   string `xml:"id,attr"`
		Text string `xml:",chardata"`
	}

	type Teach2000Answer struct {
		ID   string `xml:"id,attr"`
		Text string `xml:",chardata"`
	}

	type Teach2000Questions struct {
		Questions []Teach2000Question `xml:"question"`
	}

	type Teach2000Answers struct {
		Type    string            `xml:"type,attr"`
		Answers []Teach2000Answer `xml:"answer"`
	}

	type Teach2000Item struct {
		ID           string             `xml:"id,attr"`
		Questions    Teach2000Questions `xml:"questions"`
		Answers      Teach2000Answers   `xml:"answers"`
		Errors       int                `xml:"errors"`
		TestCount    int                `xml:"testcount"`
		CorrectCount int                `xml:"correctcount"`
	}

	type Teach2000Items struct {
		Items []Teach2000Item `xml:"item"`
	}

	type Teach2000MessageData struct {
		Items Teach2000Items `xml:"items"`
	}

	type Teach2000Root struct {
		XMLName     xml.Name             `xml:"teach2000"`
		Version     string               `xml:"version"`
		Description string               `xml:"description"`
		MessageData Teach2000MessageData `xml:"message_data"`
	}

	var root Teach2000Root
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse Teach2000 XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Description
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	itemID := 0
	for _, item := range root.MessageData.Items.Items {
		if len(item.Questions.Questions) > 0 && len(item.Answers.Answers) > 0 {
			// Extract question texts
			var questions []string
			for _, q := range item.Questions.Questions {
				if strings.TrimSpace(q.Text) != "" {
					questions = append(questions, strings.TrimSpace(q.Text))
				}
			}

			// Extract answer texts
			var answers []string
			for _, a := range item.Answers.Answers {
				if strings.TrimSpace(a.Text) != "" {
					answers = append(answers, strings.TrimSpace(a.Text))
				}
			}

			if len(questions) > 0 && len(answers) > 0 {
				wordItem := WordItem{
					ID:        itemID,
					Questions: questions,
					Answers:   answers,
					Comment:   fmt.Sprintf("TestCount: %d, Correct: %d, Errors: %d", item.TestCount, item.CorrectCount, item.Errors),
				}
				lessonData.List.Items = append(lessonData.List.Items, wordItem)
				itemID++
			}
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadTeach2000File() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadJVLTFile parses JVLT (.jvlt) ZIP files containing XML data
func (fl *FileLoader) loadJVLTFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadJVLTFile() - parsing JVLT ZIP file")

	// Open ZIP file
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open JVLT ZIP file: %v", err)
		return nil, err
	}
	defer reader.Close()

	// Look for the dict.xml file inside the ZIP
	var dictFile *zip.File
	for _, file := range reader.File {
		if file.Name == "dict.xml" || file.Name == "dictionary.xml" || strings.HasSuffix(file.Name, ".xml") {
			dictFile = file
			break
		}
	}

	if dictFile == nil {
		log.Printf("[ERROR] No XML file found in JVLT ZIP")
		return nil, fmt.Errorf("no dictionary XML file found in JVLT archive")
	}

	// Open and read the XML file
	xmlReader, err := dictFile.Open()
	if err != nil {
		log.Printf("[ERROR] Failed to open XML file in ZIP: %v", err)
		return nil, err
	}
	defer xmlReader.Close()

	// JVLT XML structure
	type JVLTSense struct {
		Trans string `xml:"trans"`
		Def   string `xml:"def"`
	}

	type JVLTEntry struct {
		ID    string      `xml:"id,attr"`
		Orth  string      `xml:"orth"`
		Pron  []string    `xml:"pron"`
		Sense []JVLTSense `xml:"sense"`
	}

	type JVLTDict struct {
		XMLName xml.Name    `xml:"dictionary"`
		Entries []JVLTEntry `xml:"entry"`
	}

	var dict JVLTDict
	decoder := xml.NewDecoder(xmlReader)
	if err := decoder.Decode(&dict); err != nil {
		log.Printf("[ERROR] Failed to parse JVLT XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	itemID := 0
	for _, entry := range dict.Entries {
		question := strings.TrimSpace(entry.Orth)
		var answer string

		// Get translation from first sense
		if len(entry.Sense) > 0 {
			answer = strings.TrimSpace(entry.Sense[0].Trans)
			if answer == "" {
				answer = strings.TrimSpace(entry.Sense[0].Def)
			}
		}

		if question != "" && answer != "" {
			lessonData.List.Items = append(lessonData.List.Items, WordItem{
				ID:        itemID,
				Questions: []string{question},
				Answers:   []string{answer},
			})
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadJVLTFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadFlashQardFile parses FlashQard (.fq) XML files
func (fl *FileLoader) loadFlashQardFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadFlashQardFile() - parsing FlashQard XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open FlashQard file: %v", err)
		return nil, err
	}
	defer file.Close()

	// FlashQard XML structure
	type FlashQardSideDocument struct {
		HTML string `xml:"html"`
	}

	type FlashQardCard struct {
		Type              string                `xml:"type,attr"`
		FrontSideDocument FlashQardSideDocument `xml:"frontsidedocument"`
		BackSideDocument  FlashQardSideDocument `xml:"backsidedocument"`
	}

	type FlashQardStage struct {
		Card FlashQardCard `xml:"card"`
	}

	type FlashQardBox struct {
		Name   string           `xml:"name,attr"`
		Stages []FlashQardStage `xml:"stage"`
	}

	type FlashQardRoot struct {
		XMLName xml.Name     `xml:"flashcards"`
		Version string       `xml:"flashqardversion,attr"`
		Author  string       `xml:"author,attr"`
		Comment string       `xml:"comment,attr"`
		Box     FlashQardBox `xml:"box"`
	}

	var root FlashQardRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse FlashQard XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Box.Name
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	itemID := 0
	for _, stage := range root.Box.Stages {
		card := stage.Card
		question := strings.TrimSpace(card.FrontSideDocument.HTML)
		answer := strings.TrimSpace(card.BackSideDocument.HTML)

		if question != "" && answer != "" {
			lessonData.List.Items = append(lessonData.List.Items, WordItem{
				ID:        itemID,
				Questions: []string{question},
				Answers:   []string{answer},
			})
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadFlashQardFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadTeachMasterFile parses TeachMaster (.vok2) XML files
func (fl *FileLoader) loadTeachMasterFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadTeachMasterFile() - parsing TeachMaster XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open TeachMaster file: %v", err)
		return nil, err
	}
	defer file.Close()

	// TeachMaster XML structure
	type TeachMasterHeader struct {
		Title  string `xml:"titel"`
		Author string `xml:"autor"`
	}

	type TeachMasterEntry struct {
		SprEins   string `xml:"spreins"` // Language 1
		SprZwei   string `xml:"sprzwei"` // Language 2
		Synonym   string `xml:"synonym"`
		Bemerkung string `xml:"bemerkung"` // Comment
	}

	type TeachMasterRoot struct {
		XMLName xml.Name           `xml:"teachmaster"`
		Header  TeachMasterHeader  `xml:"header"`
		Entries []TeachMasterEntry `xml:"vokabelsatz"`
	}

	var root TeachMasterRoot
	decoder := xml.NewDecoder(file)

	// Add charset reader for ISO-8859-1 encoding
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "ISO-8859-1":
			return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
		default:
			return nil, fmt.Errorf("unsupported charset: %s", charset)
		}
	}

	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse TeachMaster XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Header.Title
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	itemID := 0
	for _, entry := range root.Entries {
		question := strings.TrimSpace(entry.SprEins)
		answer := strings.TrimSpace(entry.SprZwei)

		if question != "" && answer != "" {
			item := WordItem{
				ID:        itemID,
				Questions: []string{question},
				Answers:   []string{answer},
			}

			// Add comment if available
			if entry.Bemerkung != "" {
				item.Comment = strings.TrimSpace(entry.Bemerkung)
			}

			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadTeachMasterFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadCueCardFile parses CueCard (.wcu) XML files
func (fl *FileLoader) loadCueCardFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadCueCardFile() - parsing CueCard XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open CueCard file: %v", err)
		return nil, err
	}
	defer file.Close()

	// CueCard XML structure
	type CueCard struct {
		Question        string `xml:"Question,attr"`
		Answer          string `xml:"Answer,attr"`
		History         string `xml:"History,attr"`
		QuestionPicture string `xml:"QuestionPicture,attr"`
		AnswerPicture   string `xml:"AnswerPicture,attr"`
	}

	type CueCardRoot struct {
		XMLName xml.Name  `xml:"CueCards"`
		Version string    `xml:"Version,attr"`
		Cards   []CueCard `xml:"Card"`
	}

	var root CueCardRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse CueCard XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	itemID := 0
	for _, card := range root.Cards {
		question := strings.TrimSpace(card.Question)
		answer := strings.TrimSpace(card.Answer)

		if question != "" && answer != "" {
			item := map[string]interface{}{
				"id":       itemID,
				"question": question,
				"answer":   answer,
			}

			// Add media information if available
			if card.QuestionPicture != "" {
				item["questionPicture"] = card.QuestionPicture
			}
			if card.AnswerPicture != "" {
				item["answerPicture"] = card.AnswerPicture
			}
			if card.History != "" {
				item["history"] = card.History
			}

			lessonData.List.Items = append(lessonData.List.Items, WordItem{
				ID:        itemID,
				Questions: []string{question},
				Answers:   []string{answer},
			})
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadCueCardFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadBackpackFile parses Backpack (.backpack) text files
func (fl *FileLoader) loadBackpackFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadBackpackFile() - parsing Backpack text file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open Backpack file: %v", err)
		return nil, err
	}
	defer file.Close()

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	itemID := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to find a separator - Backpack format seems to use various separators
		// Based on the sample, it looks like question and answer are concatenated
		// Let's try to split on common patterns or use a simple heuristic

		// Look for Greek/Latin characters followed by lowercase text as a pattern
		var question, answer string

		// Simple heuristic: if line contains mixed scripts, try to split
		if len(line) > 10 {
			// Try to find where Greek/special chars end and Latin chars begin
			splitPos := -1
			for i, r := range line {
				// Look for transition from Greek/special to Latin lowercase
				if i > 0 && ((r >= 'a' && r <= 'z') || r == ' ') {
					// Check if previous character was Greek or special
					prevR := rune(line[i-1])
					if prevR > 127 || (prevR >= 'A' && prevR <= 'Z') {
						splitPos = i
						break
					}
				}
			}

			if splitPos > 0 && splitPos < len(line)-1 {
				question = strings.TrimSpace(line[:splitPos])
				answer = strings.TrimSpace(line[splitPos:])
			}
		}

		// Fallback: try common separators
		if question == "" || answer == "" {
			separators := []string{"\t", "  ", " - ", " = ", ":"}
			for _, sep := range separators {
				if strings.Contains(line, sep) {
					parts := strings.SplitN(line, sep, 2)
					if len(parts) == 2 {
						question = strings.TrimSpace(parts[0])
						answer = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}

		// If still no split found, skip this line or treat as single word
		if question == "" || answer == "" {
			log.Printf("[WARNING] Could not parse Backpack line: %s", line)
			continue
		}

		if question != "" && answer != "" {
			lessonData.List.Items = append(lessonData.List.Items, WordItem{
				ID:        itemID,
				Questions: []string{question},
				Answers:   []string{answer},
			})
			itemID++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Error reading Backpack file: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] FileLoader.loadBackpackFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// GetSupportedExtensions returns a list of supported file extensions
func (fl *FileLoader) GetSupportedExtensions() []string {
	return []string{
		".csv", ".tsv", ".txt", ".ot", ".json", ".kvtml", ".anki", ".anki2",
		".apkg", ".backpack", ".wcu", ".voc", ".fq", ".fmd", ".dkf", ".jml",
		".jvlt", ".stp", ".db", ".oh", ".ohw", ".oh4", ".ovr", ".pau",
		".t2k", ".vok2", ".wdl", ".vtl3", ".wrts", ".xml", ".kgm", ".ottp",
		".otmd", ".otwd",
	}
}

// GetFormatName returns a human-readable name for a file extension
func (fl *FileLoader) GetFormatName(ext string) string {
	switch strings.ToLower(ext) {
	case ".csv":
		return "Spreadsheet (CSV)"
	case ".tsv":
		return "Tab-Separated Values"
	case ".txt":
		return "Text File"
	case ".ot":
		return "OpenTeacher 2.x"
	case ".json":
		return "JSON Lesson File"
	case ".kvtml":
		return "KDE Vocabulary Document"
	case ".anki":
		return "Anki Database"
	case ".anki2":
		return "Anki 2.0 Database"
	case ".apkg":
		return "Anki Package"
	case ".backpack":
		return "Backpack File"
	case ".wcu":
		return "CueCard File"
	case ".voc":
		return "Vocabulary File"
	case ".fq":
		return "FlashQard File"
	case ".fmd":
		return "FM Dictionary"
	case ".dkf":
		return "Granule Deck"
	case ".jml":
		return "JMemorize Lesson"
	case ".jvlt":
		return "JVLT File"
	case ".stp":
		return "Ludem File"
	case ".db":
		return "Database File"
	case ".oh", ".ohw", ".oh4":
		return "Overhoor File"
	case ".ovr":
		return "Overhoringsprogramma Talen"
	case ".pau":
		return "Pauker File"
	case ".t2k":
		return "Teach2000 File"
	case ".vok2":
		return "TeachMaster File"
	case ".wdl":
		return "Oriente Voca File"
	case ".vtl3":
		return "VokabelTrainer File"
	case ".wrts":
		return "WRTS File"
	case ".xml":
		return "XML File"
	case ".kgm":
		return "KGeography Map"
	case ".ottp":
		return "OpenTeaching Topography"
	case ".otmd":
		return "OpenTeaching Media"
	case ".otwd":
		return "OpenTeaching Words"
	default:
		return "Unknown Format"
	}
}

// loadKGeographyMapFile parses KGeography Map (.kgm) files
func (fl *FileLoader) loadKGeographyMapFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadKGeographyMapFile() - parsing KGeography Map file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open KGeography file: %v", err)
		return nil, err
	}
	defer file.Close()

	// KGeography maps are XML files with place information
	type KGMDivision struct {
		Name         string `xml:"name"`
		Capital      string `xml:"capital,omitempty"`
		Flag         string `xml:"flag,omitempty"`
		BlurFlag     string `xml:"blurFlag,omitempty"`
		Color        string `xml:"color,omitempty"`
		FalseCapital string `xml:"falseCapital,omitempty"`
	}

	type KGMMap struct {
		XMLName   xml.Name      `xml:"map"`
		Name      string        `xml:"name"`
		File      string        `xml:"file"`
		Divisions []KGMDivision `xml:"division"`
	}

	var kgmMap KGMMap
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&kgmMap); err != nil {
		log.Printf("[ERROR] Failed to parse KGeography XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = kgmMap.Name
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	itemID := 0
	for _, division := range kgmMap.Divisions {
		if division.Name != "" {
			// For topography, the "question" is the location and "answer" is the name
			item := map[string]interface{}{
				"id":   itemID,
				"name": division.Name,
			}

			if division.Capital != "" {
				item["capital"] = division.Capital
			}
			if division.Flag != "" {
				item["flag"] = division.Flag
			}
			if division.Color != "" {
				item["color"] = division.Color
			}

			// For the word list format, use name as both question and answer
			// This will be properly handled by the topo test type module
			lessonData.List.Items = append(lessonData.List.Items, WordItem{
				ID:        itemID,
				Questions: []string{division.Name},
				Answers:   []string{division.Name},
			})
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadKGeographyMapFile() - loaded %d places", len(lessonData.List.Items))
	return lessonData, nil
}

// loadOpenTeachingTopoFile parses OpenTeaching Topography (.ottp) ZIP files
func (fl *FileLoader) loadOpenTeachingTopoFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadOpenTeachingTopoFile() - parsing OpenTeaching Topography ZIP file")

	// Open ZIP file
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open OpenTeaching Topo ZIP file: %v", err)
		return nil, err
	}
	defer reader.Close()

	// Look for list.json file inside the ZIP
	var listFile *zip.File
	for _, file := range reader.File {
		if file.Name == "list.json" {
			listFile = file
			break
		}
	}

	if listFile == nil {
		log.Printf("[ERROR] No list.json file found in OpenTeaching Topo ZIP")
		return nil, fmt.Errorf("no list.json file found in OpenTeaching Topography archive")
	}

	// Open and read the JSON file
	jsonReader, err := listFile.Open()
	if err != nil {
		log.Printf("[ERROR] Failed to open list.json file in ZIP: %v", err)
		return nil, err
	}
	defer jsonReader.Close()

	// Read JSON content
	jsonData, err := io.ReadAll(jsonReader)
	if err != nil {
		log.Printf("[ERROR] Failed to read list.json content: %v", err)
		return nil, err
	}

	// Parse OpenTeacher format JSON
	var otData map[string]interface{}
	if err := json.Unmarshal(jsonData, &otData); err != nil {
		log.Printf("[ERROR] Failed to parse OpenTeaching Topo JSON: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()

	// Extract title
	if title, ok := otData["title"].(string); ok {
		lessonData.List.Title = title
	} else {
		// Generate better fallback title based on content
		itemCount := 0
		if items, ok := otData["items"].([]interface{}); ok {
			itemCount = len(items)
		}

		if itemCount > 0 {
			lessonData.List.Title = fmt.Sprintf("Topography Lesson (%d places)", itemCount)
		} else {
			lessonData.List.Title = "Topography Lesson"
		}
	}

	// Extract topo items with coordinates
	itemID := 0
	if items, ok := otData["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				name := ""
				x := 0
				y := 0

				if n, exists := itemMap["name"]; exists {
					name = fmt.Sprintf("%v", n)
				}
				if xVal, exists := itemMap["x"]; exists {
					if xFloat, ok := xVal.(float64); ok {
						x = int(xFloat)
					}
				}
				if yVal, exists := itemMap["y"]; exists {
					if yFloat, ok := yVal.(float64); ok {
						y = int(yFloat)
					}
				}

				if name != "" {
					lessonData.List.Items = append(lessonData.List.Items, WordItem{
						ID:        itemID,
						Name:      name,
						Questions: []string{name},
						Answers:   []string{name},
						X:         &x,
						Y:         &y,
					})
					itemID++
				}
			}
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadOpenTeachingTopoFile() - loaded %d places", len(lessonData.List.Items))
	return lessonData, nil
}

// loadOpenTeachingMediaFile parses OpenTeaching Media (.otmd) ZIP files
func (fl *FileLoader) loadOpenTeachingMediaFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadOpenTeachingMediaFile() - parsing OpenTeaching Media ZIP file")

	// Open ZIP file
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open OpenTeaching Media ZIP file: %v", err)
		return nil, err
	}
	defer reader.Close()

	// Look for list.json file inside the ZIP
	var listFile *zip.File
	for _, file := range reader.File {
		if file.Name == "list.json" {
			listFile = file
			break
		}
	}

	if listFile == nil {
		log.Printf("[ERROR] No list.json file found in OpenTeaching Media ZIP")
		return nil, fmt.Errorf("no list.json file found in OpenTeaching Media archive")
	}

	// Open and read the JSON file
	jsonReader, err := listFile.Open()
	if err != nil {
		log.Printf("[ERROR] Failed to open list.json file in ZIP: %v", err)
		return nil, err
	}
	defer jsonReader.Close()

	// Read JSON content
	jsonData, err := io.ReadAll(jsonReader)
	if err != nil {
		log.Printf("[ERROR] Failed to read list.json content: %v", err)
		return nil, err
	}

	// Parse OpenTeacher format JSON
	var otData map[string]interface{}
	if err := json.Unmarshal(jsonData, &otData); err != nil {
		log.Printf("[ERROR] Failed to parse OpenTeaching Media JSON: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()

	// Extract title
	if title, ok := otData["title"].(string); ok {
		lessonData.List.Title = title
	} else {
		// Generate better fallback title based on content
		itemCount := 0
		if items, ok := otData["items"].([]interface{}); ok {
			itemCount = len(items)
		}

		if itemCount > 0 {
			lessonData.List.Title = fmt.Sprintf("Media Lesson (%d items)", itemCount)
		} else {
			lessonData.List.Title = "Media Lesson"
		}
	}

	// Extract media items with file information
	itemID := 0
	if items, ok := otData["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				name := ""
				question := ""
				answer := ""
				filename := ""
				remote := false

				if n, exists := itemMap["name"]; exists {
					name = fmt.Sprintf("%v", n)
				}
				if q, exists := itemMap["question"]; exists {
					question = fmt.Sprintf("%v", q)
				}
				if a, exists := itemMap["answer"]; exists {
					answer = fmt.Sprintf("%v", a)
				}
				if f, exists := itemMap["filename"]; exists {
					filename = fmt.Sprintf("%v", f)
				}
				if r, exists := itemMap["remote"]; exists {
					if rBool, ok := r.(bool); ok {
						remote = rBool
					}
				}

				// Use question and answer if available, otherwise fall back to name
				questions := []string{name}
				answers := []string{name}
				if question != "" && answer != "" {
					questions = []string{question}
					answers = []string{answer}
				}

				if name != "" {
					lessonData.List.Items = append(lessonData.List.Items, WordItem{
						ID:        itemID,
						Name:      name,
						Questions: questions,
						Answers:   answers,
						Filename:  &filename,
						Remote:    &remote,
					})
					itemID++
				}
			}
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadOpenTeachingMediaFile() - loaded %d media items", len(lessonData.List.Items))
	return lessonData, nil
}
