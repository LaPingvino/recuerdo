// Package mimicrytypefaceconverter converts text typed in a "mimicry
// typeface" to Unicode. Old programs wrote Greek with fonts such as Symbol
// or Teach2000's Greek that draw Greek letters in place of Latin ones, so
// their files hold "abg" where the user saw "αβγ". Port of OpenTeacher's
// logic/mimicryTypefaceConverter.
package mimicrytypefaceconverter

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// MimicryTypefaceConverterModule is a Go port of the Python MimicryTypefaceConverterModule class
type MimicryTypefaceConverterModule struct {
	*core.BaseModule
	manager *core.Manager
	// TODO: Add module-specific fields
}

// NewMimicryTypefaceConverterModule creates a new MimicryTypefaceConverterModule instance
func NewMimicryTypefaceConverterModule() *MimicryTypefaceConverterModule {
	base := core.NewBaseModule("logic", "mimicrytypefaceconverter-module")

	return &MimicryTypefaceConverterModule{
		BaseModule: base,
	}
}

// Convert converts text in font to Unicode (see Convert).
func (mod *MimicryTypefaceConverterModule) Convert(font, text string) string {
	return Convert(font, text)
}

// greekLetters maps the Latin letters of the Symbol, Greek and TekniaGreek
// fonts to the Greek letters they draw. In OpenTeacher the three fonts
// share one table (Greek's additions also apply to Symbol), as here.
var greekLetters = map[rune]rune{
	'a': 'α', 'b': 'β', 'g': 'γ', 'd': 'δ', 'e': 'ε', 'z': 'ζ', 'h': 'η',
	'q': 'θ', 'i': 'ι', 'k': 'κ', 'l': 'λ', 'm': 'μ', 'n': 'ν', 'x': 'ξ',
	'o': 'ο', 'p': 'π', 'r': 'ρ', 's': 'σ', 't': 'τ', 'u': 'υ', 'f': 'φ',
	'c': 'χ', 'y': 'ψ', 'w': 'ω',
	'A': 'Α', 'B': 'Β', 'G': 'Γ', 'D': 'Δ', 'E': 'Ε', 'Z': 'Ζ', 'H': 'Η',
	'Q': 'Θ', 'I': 'Ι', 'K': 'Κ', 'L': 'Λ', 'M': 'Μ', 'N': 'Ν', 'X': 'Ξ',
	'O': 'Ο', 'P': 'Π', 'R': 'Ρ', 'S': 'Σ', 'T': 'Τ', 'U': 'Υ', 'F': 'Φ',
	'C': 'Χ', 'Y': 'Ψ', 'W': 'Ω',
	// added for Teach2000's Greek font
	'j': 'ς', 'v': 'ᾳ', 'J': 'ῷ', 'V': 'ῃ',
}

// Convert returns text with each letter replaced by the one it shows in
// font, if font is a known mimicry typeface (Symbol, Greek, TekniaGreek;
// case does not matter); other text is returned unchanged.
func Convert(font, text string) string {
	switch strings.ToLower(strings.TrimSpace(font)) {
	case "symbol", "greek", "tekniagreek":
	default:
		return text
	}
	return strings.Map(func(r rune) rune {
		if g, ok := greekLetters[r]; ok {
			return g
		}
		return r
	}, text)
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (mod *MimicryTypefaceConverterModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// TODO: Port Python enable logic

	fmt.Println("MimicryTypefaceConverterModule enabled")
	return nil
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (mod *MimicryTypefaceConverterModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// TODO: Port Python disable logic

	fmt.Println("MimicryTypefaceConverterModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *MimicryTypefaceConverterModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitMimicryTypefaceConverterModule creates and returns a new MimicryTypefaceConverterModule instance
// This is the Go equivalent of the Python init function
func InitMimicryTypefaceConverterModule() core.Module {
	return NewMimicryTypefaceConverterModule()
}
