package lesson

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	mimicry "github.com/LaPingvino/recuerdo/internal/modules/logic/mimicryTypefaceConverter"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// Voca (Oriente) word lists, .wdl: a big-endian binary format in versions
// 4.0 and 3.0 ("VOCAWRDL" files) and Vocatude 1.0. Ported from
// OpenTeacher's loader, which is based on inspecting files: only the
// languages and the words are read, everything else is skipped.

var errVocaShort = errors.New("Voca file ends early")

type vocaReader struct {
	data []byte
	pos  int
	err  error
}

func (r *vocaReader) next(n int) []byte {
	if r.err != nil || n < 0 || r.pos+n > len(r.data) {
		r.err = errVocaShort
		// zeros for the reader to go on with, never as many as the
		// (broken) file asks: n can be anything, 2^62 crashed the app
		return make([]byte, min(max(n, 0), 8))
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *vocaReader) skip(n int)   { r.next(n) }
func (r *vocaReader) int32() int   { return int(int32(binary.BigEndian.Uint32(r.next(4)))) }
func (r *vocaReader) int64() int64 { return int64(binary.BigEndian.Uint64(r.next(8))) }
func (r *vocaReader) byte() byte   { return r.next(1)[0] }
func (r *vocaReader) bool() bool   { return r.byte() != 0 }
func (r *vocaReader) atEnd() bool  { return r.pos >= len(r.data) }
func (r *vocaReader) str() string  { return string(r.next(r.int32())) }
func (r *vocaReader) skipStr()     { r.skip(r.int32()) }
func (r *vocaReader) skipStrs(n int) {
	for i := 0; i < n && r.err == nil; i++ {
		r.skipStr()
	}
}

// cstr reads a null-terminated string.
func (r *vocaReader) cstr() []byte {
	i := bytes.IndexByte(r.data[min(r.pos, len(r.data)):], 0)
	if i < 0 {
		r.err = errVocaShort
		return nil
	}
	s := r.data[r.pos : r.pos+i]
	r.pos += i + 1
	return s
}

// skipEmbeddedFile skips a sound or picture: a size, then if there is one
// its extension and data.
func (r *vocaReader) skipEmbeddedFile() {
	if size := r.int64(); size > 0 {
		r.skipStr()
		r.skip(int(size))
	}
}

// loadVoca loads a Voca or Vocatude word list.
func (fl *FileLoader) loadVoca(path string) (*LessonData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := &vocaReader{data: raw}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	if len(raw) >= 10 && string(raw[:8]) == "VOCAWRDL" {
		switch v := [2]byte{raw[8], raw[9]}; v {
		case [2]byte{4, 0}, [2]byte{3, 0}:
			fl.readVoca(r, data, v[0] == 4)
		default:
			return nil, fmt.Errorf("unknown Voca version %d.%d", v[0], v[1])
		}
	} else {
		fl.readVocatude(r, data)
	}
	if r.err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, r.err)
	}
	return data, nil
}

// readVoca reads Voca 3.0 and 4.0 files.
func (fl *FileLoader) readVoca(r *vocaReader, data *LessonData, v4 bool) {
	r.skip(10) // magic and version
	// the foreign (answer) language comes first, then the reference
	// (question) language, each with its font name and size
	data.List.AnswerLanguage = r.str()
	answerFont := r.str()
	r.skip(4)
	data.List.QuestionLanguage = r.str()
	questionFont := r.str()
	r.skip(4)

	// phonetic font and size, foreign characters
	r.skipStr()
	r.skip(4)
	r.skipStr()
	// parts of speech, each with grammar categories
	for i, n := 0, r.int32(); i < n && r.err == nil; i++ {
		r.skipStr()
		r.skipStrs(r.int32())
	}
	// upload information
	if r.bool() {
		r.skipStr()
		for i := 0; i < 2; i++ {
			if r.bool() {
				r.skip(3) // ISO language code
			}
		}
		r.skipStrs(2)
	}
	if v4 {
		r.skip(12) // mastered score, chance and exam count
	}
	// exercise types and their configurations
	r.skip(4)
	for i, n := 0, r.int32(); i < n && r.err == nil; i++ {
		r.skip(1)
		r.skipStrs(2)
		for j, m := 0, r.int32(); j < m && r.err == nil; j++ {
			r.skip(1)
			r.skipStr()
			r.skipStrs(r.int32()) // question fields
			for k, a := 0, r.int32(); k < a && r.err == nil; k++ {
				r.skip(1)
				r.skipStr()
			}
			r.skipStrs(r.int32()) // info fields
		}
	}

	for i, n := 0, r.int32(); i < n && r.err == nil; i++ {
		r.skip(4) // sequence
		answer := r.str()
		r.skip(1)
		question := r.str()
		r.skip(1)
		for j := 0; j < 2; j++ { // context, phonetic
			r.skipStr()
			r.skip(1)
		}
		r.skipStr() // part of speech
		for g, gn := 0, r.int32(); g < gn && r.err == nil; g++ {
			r.skipStrs(2) // grammar category and value
		}
		r.skip(1)
		r.skipEmbeddedFile() // sound
		r.skip(1)
		if v4 {
			r.skipEmbeddedFile() // picture
			r.skip(1)
		}
		r.skipStrs(2)
		fl.addVocaItem(data, questionFont, question, answerFont, answer)
	}
}

// vocatudeHeader starts every Vocatude 1.0 file.
var vocatudeHeader, _ = hex.DecodeString("2ef6264f25bd2c59038f59aa1e3446076c4a44cf2800")

// readVocatude reads Vocatude 1.0 files: null-terminated strings in the
// character set each language names.
func (fl *FileLoader) readVocatude(r *vocaReader, data *LessonData) {
	if !bytes.Equal(r.next(len(vocatudeHeader)), vocatudeHeader) {
		r.err = fmt.Errorf("not a Voca or Vocatude file")
		return
	}
	answerLang := r.cstr()
	answerEnc := vocaCharset(r.byte())
	answerFont := string(r.cstr())
	r.skip(1)
	questionLang := r.cstr()
	questionEnc := vocaCharset(r.byte())
	questionFont := string(r.cstr())
	r.skip(1)
	r.cstr()
	r.skip(2) // 0d 00
	data.List.AnswerLanguage = decodeVoca(answerEnc, answerLang)
	data.List.QuestionLanguage = decodeVoca(questionEnc, questionLang)

	for !r.atEnd() && r.err == nil {
		answer := decodeVoca(answerEnc, r.cstr())
		question := decodeVoca(questionEnc, r.cstr())
		for i := 0; i < 3; i++ {
			r.cstr()
		}
		r.skip(r.int32()) // sound
		r.skip(8)
		r.skip(2) // 0d 00
		if r.err == nil {
			fl.addVocaItem(data, questionFont, question, answerFont, answer)
		}
	}
}

func (fl *FileLoader) addVocaItem(data *LessonData, questionFont, question, answerFont, answer string) {
	q := fl.parseWordString(mimicry.Convert(questionFont, question))
	a := fl.parseWordString(mimicry.Convert(answerFont, answer))
	if len(q) > 0 || len(a) > 0 {
		data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
	}
}

// vocaCharset maps a Windows character set code to its encoding.
func vocaCharset(code byte) encoding.Encoding {
	switch code {
	case 128:
		return japanese.ShiftJIS
	case 129:
		return korean.EUCKR
	case 134:
		return simplifiedchinese.GBK
	case 136:
		return traditionalchinese.Big5
	case 177:
		return charmap.Windows1255
	case 178:
		return charmap.Windows1256
	case 161:
		return charmap.Windows1253
	case 162:
		return charmap.Windows1254
	case 163:
		return charmap.Windows1258
	case 222:
		return charmap.Windows874
	case 238:
		return charmap.Windows1250
	case 204:
		return charmap.Windows1251
	case 186:
		return charmap.Windows1257
	}
	// 0 and 1 (Western), and Johab (130), for which Go has no decoder
	return charmap.Windows1252
}

func decodeVoca(enc encoding.Encoding, b []byte) string {
	s, err := enc.NewDecoder().Bytes(b)
	if err != nil {
		return string(b)
	}
	return string(s)
}
