// Package cli is Recuerdo's command line: OpenTeacher's cli profile
// (authors, convert, merge, reverse-list, view-word-list, ocr-word-list,
// new-word-list, practise-word-list), with ordinary -flags where
// OpenTeacher used +flags.
package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	openteacherauthors "github.com/LaPingvino/recuerdo/internal/modules/data/openteacherAuthors"
	mediareverser "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/media"
	wordsreverser "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/words"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
	"github.com/LaPingvino/recuerdo/internal/ocr"
	"github.com/LaPingvino/recuerdo/internal/teaching"
)

type command struct {
	usage, help string
	run         func(c *ctx, args []string) error
}

type ctx struct {
	in       *bufio.Reader
	out, err io.Writer
}

var commands map[string]command

func init() {
	commands = map[string]command{
		"authors":            {"[-c role]", "show OpenTeacher's authors", authors},
		"convert":            {"[-f format] files...", "convert word, topography and media files (default format: otwd)", convert},
		"merge":              {"output base others...", "merge files of the same kind into output", merge},
		"reverse-list":       {"input output", "swap the questions and answers of a list", reverseList},
		"view-word-list":     {"[-p list|title|question-lang|answer-lang] files...", "show a word list", viewWordList},
		"ocr-word-list":      {"picture output", "read a word list from a scan or photo (needs Tesseract)", ocrWordList},
		"new-word-list":      {"[-t title] [-q lang] [-a lang] output [input|-]", "make a word list from lines like \"question = answer\" (default: standard input)", newWordList},
		"practise-word-list": {"[-l lesson type] file", "practise a word list in the terminal", practise},
	}
}

// IsCommand reports whether name is one of the commands (or help).
func IsCommand(name string) bool {
	_, ok := commands[name]
	return ok || name == "help"
}

// Run runs the command line args (the command and its arguments) and
// returns the exit status.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	c := &ctx{in: bufio.NewReader(stdin), out: stdout, err: stderr}
	if len(args) == 0 || args[0] == "help" {
		Help(stdout)
		return 0
	}
	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(stderr, "recuerdo: unknown command %q\n", args[0])
		Help(stderr)
		return 2
	}
	// the lesson package logs what it does: not for a command line
	logOut := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(logOut)
	if err := cmd.run(c, args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "recuerdo %s: %v\n", args[0], err)
		return 1
	}
	return 0
}

// Help lists the commands.
func Help(w io.Writer) {
	fmt.Fprintln(w, "Usage: recuerdo [lesson file]        start the program")
	fmt.Fprintln(w, "       recuerdo <command> [arguments]")
	fmt.Fprintln(w, "\nCommands:")
	for _, name := range []string{"authors", "convert", "merge", "reverse-list", "view-word-list", "ocr-word-list", "new-word-list", "practise-word-list"} {
		c := commands[name]
		fmt.Fprintf(w, "  %s %s\n      %s\n", name, c.usage, c.help)
	}
}

func flags(c *ctx, name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.err)
	return fs
}

func need(fs *flag.FlagSet, n int, what string) error {
	if fs.NArg() < n {
		return fmt.Errorf("need %s (see recuerdo help)", what)
	}
	return nil
}

func load(path string) (*lesson.LessonData, error) { return lesson.NewFileLoader().LoadFile(path) }

func save(data *lesson.LessonData, path string) error {
	return lesson.NewFileSaver().SaveFile(data, path)
}

func authors(c *ctx, args []string) error {
	fs := flags(c, "authors")
	role := fs.String("c", "", "only show authors in this role")
	if err := fs.Parse(args); err != nil {
		return err
	}
	roles, names := openteacherauthors.ByRole()
	for _, r := range roles {
		if *role != "" && !strings.EqualFold(r, *role) {
			continue
		}
		fmt.Fprintf(c.out, "%s:\n", r)
		for _, n := range names[r] {
			fmt.Fprintf(c.out, "  %s\n", n)
		}
	}
	return nil
}

func convert(c *ctx, args []string) error {
	fs := flags(c, "convert")
	format := fs.String("f", "otwd", "output format (file extension)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 1, "files to convert"); err != nil {
		return err
	}
	ext := "." + strings.TrimPrefix(*format, ".")
	for _, in := range fs.Args() {
		data, err := load(in)
		if err != nil {
			return fmt.Errorf("%s: %w", in, err)
		}
		out := strings.TrimSuffix(in, filepath.Ext(in)) + ext
		if out == in {
			return fmt.Errorf("%s is already a %s file", in, ext)
		}
		if err := save(data, out); err != nil {
			return fmt.Errorf("%s: %w", out, err)
		}
		fmt.Fprintf(c.out, "%s -> %s\n", in, out)
	}
	return nil
}

func merge(c *ctx, args []string) error {
	fs := flags(c, "merge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 3, "an output file, a base file and files to merge"); err != nil {
		return err
	}
	base, err := load(fs.Arg(1))
	if err != nil {
		return err
	}
	for _, p := range fs.Args()[2:] {
		other, err := load(p)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		lesson.Merge(&base.List, other.List)
	}
	return save(base, fs.Arg(0))
}

func reverseList(c *ctx, args []string) error {
	fs := flags(c, "reverse-list")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 2, "an input and an output file"); err != nil {
		return err
	}
	data, err := load(fs.Arg(0))
	if err != nil {
		return err
	}
	for _, it := range data.List.Items {
		if it.IsTopoItem() {
			return errors.New("a topography lesson has no questions and answers to swap")
		}
		if it.IsMediaItem() {
			mediareverser.Reverse(&data.List)
			return save(data, fs.Arg(1))
		}
	}
	wordsreverser.Reverse(&data.List)
	return save(data, fs.Arg(1))
}

func viewWordList(c *ctx, args []string) error {
	fs := flags(c, "view-word-list")
	part := fs.String("p", "list", "what to show: list, title, question-lang or answer-lang")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 1, "files to show"); err != nil {
		return err
	}
	for _, p := range fs.Args() {
		data, err := load(p)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		switch *part {
		case "title":
			fmt.Fprintln(c.out, data.List.Title)
		case "question-lang":
			fmt.Fprintln(c.out, data.List.QuestionLanguage)
		case "answer-lang":
			fmt.Fprintln(c.out, data.List.AnswerLanguage)
		case "list":
			width := 0
			qs := make([]string, len(data.List.Items))
			for i, it := range data.List.Items {
				qs[i] = composer.Compose(checker.StoredAnswers(it.Questions))
				width = max(width, len([]rune(qs[i])))
			}
			for i, it := range data.List.Items {
				q := qs[i] + strings.Repeat(" ", width-len([]rune(qs[i])))
				fmt.Fprintf(c.out, "%s  %s\n", q, composer.Compose(checker.StoredAnswers(it.Answers)))
			}
		default:
			return fmt.Errorf("unknown part %q: use list, title, question-lang or answer-lang", *part)
		}
	}
	return nil
}

func ocrWordList(c *ctx, args []string) error {
	fs := flags(c, "ocr-word-list")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 2, "a picture and an output file"); err != nil {
		return err
	}
	items, err := ocr.LoadWordList(fs.Arg(0))
	if err != nil {
		return err
	}
	data := lesson.NewLessonData()
	data.List.Title = strings.TrimSuffix(filepath.Base(fs.Arg(0)), filepath.Ext(fs.Arg(0)))
	data.List.Items = items
	fmt.Fprintf(c.out, "%d word pairs read\n", len(items))
	return save(data, fs.Arg(1))
}

func newWordList(c *ctx, args []string) error {
	fs := flags(c, "new-word-list")
	title := fs.String("t", "", "title")
	qlang := fs.String("q", "", "question language")
	alang := fs.String("a", "", "answer language")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 1, "an output file"); err != nil {
		return err
	}
	var text []byte
	var err error
	if fs.NArg() < 2 || fs.Arg(1) == "-" {
		text, err = io.ReadAll(c.in)
	} else {
		text, err = os.ReadFile(fs.Arg(1))
	}
	if err != nil {
		return err
	}
	items, err := lesson.ParseWordList(string(text), true)
	if err != nil {
		return err
	}
	data := lesson.NewLessonData()
	data.List.Title, data.List.QuestionLanguage, data.List.AnswerLanguage = *title, *qlang, *alang
	data.List.Items = items
	return save(data, fs.Arg(0))
}

func practise(c *ctx, args []string) error {
	fs := flags(c, "practise-word-list")
	lessonType := fs.String("l", teaching.LessonTypes[0], "lesson type: "+strings.Join(teaching.LessonTypes, ", "))
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := need(fs, 1, "a word list"); err != nil {
		return err
	}
	data, err := load(fs.Arg(0))
	if err != nil {
		return err
	}
	if len(data.List.Items) == 0 {
		return errors.New("the list has no words")
	}
	s := teaching.New(data.List, teaching.Options{LessonType: *lessonType})
	s.Start()
	fmt.Fprintln(c.out, "Type the answer and press Enter; an empty line stops.")
	for !s.Done() {
		item, _, ok := s.Current()
		if !ok {
			break
		}
		asked, total := s.Progress()
		fmt.Fprintf(c.out, "[%d/%d] %s: ", asked+1, total, composer.Compose(checker.StoredAnswers(item.Questions)))
		line, err := c.in.ReadString('\n')
		if line = strings.TrimSpace(line); line == "" {
			fmt.Fprintln(c.out)
			break
		}
		a := s.Answer(line)
		if a.Right {
			fmt.Fprintln(c.out, "Right!")
		} else {
			fmt.Fprintf(c.out, "Wrong, the answer is: %s\n", a.Correct)
		}
		s.Next()
		if err != nil {
			break
		}
	}
	// (the results are not written into the file: a word list in a
	// simple format would lose its layout)
	right, answered := s.Score()
	fmt.Fprintf(c.out, "%d of %d right.\n", right, answered)
	return nil
}
