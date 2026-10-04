package words

import (
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
	"github.com/LaPingvino/recuerdo/internal/teaching"
)

// The word lesson's settings, shown in the settings dialog (OpenTeacher's
// noteCalculatorChooser, words TTS provider and repeatAnswer settings).
func init() {
	settingsdefs.Register(settingsdefs.Def{
		Key: NotationSetting, Category: "Results", Name: "Grades in",
		Help: "How the Results tab grades your sessions", Kind: settingsdefs.Choice,
		Choices: teaching.Notations, Default: teaching.DefaultNotation,
	})
	settingsdefs.Register(settingsdefs.Def{
		Key: PronounceSetting, Category: "Practice", Name: "Pronounce questions",
		Help: "Say each question aloud in the question language (needs espeak-ng, or the voices of macOS or Windows)",
		Kind: settingsdefs.Bool, Default: false,
	})
	settingsdefs.Register(settingsdefs.Def{
		Key: RepeatDurationSetting, Category: "Practice", Name: "Repeat answer shows the answer for",
		Help: "In the Repeat answer mode, how long the answer is shown before you type it",
		Kind: settingsdefs.Seconds, Default: float64(teaching.RepeatFadeDuration.Milliseconds()), Min: 0.5, Max: 30,
	})
}
