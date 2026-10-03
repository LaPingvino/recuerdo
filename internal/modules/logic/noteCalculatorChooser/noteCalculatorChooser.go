// Package notecalculatorchooser picks the grade notation the user wants.
// Port of OpenTeacher's logic/noteCalculatorChooser.
package notecalculatorchooser

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/american"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/dutch"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/ects"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/french"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/german"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/percents"
)

// SettingKey is the setting holding the chosen notation (its display name),
// named as in OpenTeacher.
const SettingKey = "org.openteacher.noteCalculatorChooser.noteCalculator"

// Default is the notation used when none is chosen.
const Default = "Dutch"

var calculators = []notecalculators.Calculator{
	american.NewAmericanNoteCalculatorModule(),
	dutch.NewDutchNoteCalculatorModule(),
	ects.NewECTSNoteCalculatorModule(),
	french.NewFrenchNoteCalculatorModule(),
	german.NewGermanNoteCalculatorModule(),
	percents.NewPercentsNoteCalculatorModule(),
}

// Names lists the notations, sorted by name (as OpenTeacher offers them).
func Names() []string {
	names := make([]string, len(calculators))
	for i, c := range calculators {
		names[i] = c.DisplayName()
	}
	return names
}

// Choose returns the calculator with the given display name, or the
// default one.
func Choose(name string) notecalculators.Calculator {
	var fallback notecalculators.Calculator
	for _, c := range calculators {
		switch c.DisplayName() {
		case name:
			return c
		case Default:
			fallback = c
		}
	}
	return fallback
}

// NoteCalculatorChooserModule offers Choose as an OpenTeacher
// "noteCalculatorChooser" module.
type NoteCalculatorChooserModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewNoteCalculatorChooserModule creates the module.
func NewNoteCalculatorChooserModule() *NoteCalculatorChooserModule {
	return &NoteCalculatorChooserModule{BaseModule: core.NewBaseModule("noteCalculatorChooser", "note-calculator-chooser")}
}

// NoteCalculator returns the calculator for the chosen notation.
func (mod *NoteCalculatorChooserModule) NoteCalculator(name string) notecalculators.Calculator {
	return Choose(name)
}

func (mod *NoteCalculatorChooserModule) Enable(ctx context.Context) error {
	return mod.BaseModule.Enable(ctx)
}
func (mod *NoteCalculatorChooserModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *NoteCalculatorChooserModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitNoteCalculatorChooserModule creates and returns the module.
func InitNoteCalculatorChooserModule() core.Module { return NewNoteCalculatorChooserModule() }
