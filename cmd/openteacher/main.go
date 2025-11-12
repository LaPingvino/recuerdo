package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/LaPingvino/openteacher/internal/modules"
	"github.com/LaPingvino/openteacher/internal/modules/data/chars/cyrillic"
	"github.com/LaPingvino/openteacher/internal/modules/data/chars/greek"
	"github.com/LaPingvino/openteacher/internal/modules/data/chars/symbols"
	datatypeicons "github.com/LaPingvino/openteacher/internal/modules/data/dataTypeIcons"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/africa"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/asia"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/europe"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/latinamerica"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/usa"
	"github.com/LaPingvino/openteacher/internal/modules/data/maps/world"
	"github.com/LaPingvino/openteacher/internal/modules/data/metadata"
	openteacherauthors "github.com/LaPingvino/openteacher/internal/modules/data/openteacherAuthors"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/all"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/cli"
	codecomplexity "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/codeComplexity"
	codedocumentation "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/codeDocumentation"
	generatebusinesscard "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/generateBusinessCard"
	generatelanguagecodeguessertable "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/generateLanguageCodeGuesserTable"
	generatemobile "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/generateMobile"
	generateweb "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/generateWeb"
	generatewebsite "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/generateWebsite"
	gettranslationauthors "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/getTranslationAuthors"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/help"
	ircbot "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/ircBot"
	modulegraph "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/moduleGraph"
	packagearch "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageArch"
	packagedebian "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageDebian"
	packagemac "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageMac"
	packagerpm "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageRpm"
	packagesource "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageSource"
	packagesourcewithsetup "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageSourceWithSetup"
	packagewindowsmsi "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageWindowsMsi"
	packagewindowsportable "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/packageWindowsPortable"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/selfstudy"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/shell"
	studentathome "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/studentAtHome"
	studentatschool "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/studentAtSchool"
	"github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/teacher"
	updaterosetta "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/updateRosetta"
	updatetranslations "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/updateTranslations"
	profilewebservicesserver "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/webServicesServer"
	wordsonly "github.com/LaPingvino/openteacher/internal/modules/data/profileDescriptions/wordsOnly"
	"github.com/LaPingvino/openteacher/internal/modules/data/profiledescriptions"
	userdocumentation "github.com/LaPingvino/openteacher/internal/modules/data/userDocumentation"

	// ALL Qt imports temporarily disabled to get core system working first
	// TODO: Re-enable Qt modules incrementally once basic system is validated

	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/dialogs/about"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/dialogs/file"
	settingsDialog "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/dialogs/settings"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/lessonDialogs"

	// "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/dialogs/documentation" // Disabled due to build constraints
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/gui"
	loadergui "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/loaderGui"
	mediadisplay "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaDisplay"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/audio"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/dailymotion"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/image"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/liveleak"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/text"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/video"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/vimeo"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/website"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/mediaTypes/youtube"
	ocrgui "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/ocrGui"
	percentnotesviewer "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/percentNotesViewer"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/printer"
	progressviewer "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/progressViewer"
	qtapp "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/qtApp"
	recentlyopenedviewer "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/recentlyOpenedViewer"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/boolean"
	charactertable "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/characterTable"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/language"
	longtext "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/longText"
	multioption "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/multiOption"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/number"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/option"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/password"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/profile"
	shorttext "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidget/shortText"
	settingswidgets "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/settingsWidgets"
	startwidget "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/startWidget"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/teachTypes/hangman"
	inmind "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/teachTypes/inMind"
	repeatanswer "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/teachTypes/repeatAnswer"
	shuffleanswer "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/teachTypes/shuffleAnswer"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/teachTypes/typing"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/testMode/menu"
	studentsview "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/testMode/studentsView"
	teacherpanel "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/testMode/teacherPanel"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/theme"
	topomaps "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/topoMaps"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/typingTutor/keyboard"
	"github.com/LaPingvino/openteacher/internal/modules/logic/authors"
	"github.com/LaPingvino/openteacher/internal/modules/logic/settings"

	logicevent "github.com/LaPingvino/openteacher/internal/modules/logic/event"
	"github.com/LaPingvino/openteacher/internal/modules/logic/execute"
	friendlytranslationnames "github.com/LaPingvino/openteacher/internal/modules/logic/friendlyTranslationNames"
	javascriptwords "github.com/LaPingvino/openteacher/internal/modules/logic/htmlGenerator/javaScriptWords"

	htmltopo "github.com/LaPingvino/openteacher/internal/modules/logic/htmlGenerator/topo"
	htmlwords "github.com/LaPingvino/openteacher/internal/modules/logic/htmlGenerator/words"

	buttonregister "github.com/LaPingvino/openteacher/internal/modules/logic/interfaces/buttonRegister"
	inputtypinglogic "github.com/LaPingvino/openteacher/internal/modules/logic/interfaces/inputTypingLogic"
	javascriptinputtypinglogic "github.com/LaPingvino/openteacher/internal/modules/logic/interfaces/javaScriptInputTypingLogic"
	lessontracker "github.com/LaPingvino/openteacher/internal/modules/logic/interfaces/lessonTracker"
	typingtutormodel "github.com/LaPingvino/openteacher/internal/modules/logic/interfaces/typingTutorModel"
	foreignknown "github.com/LaPingvino/openteacher/internal/modules/logic/itemModifiers/foreignKnown"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/bisect"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/evaluator"
	jsevent "github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/event"
	lessontype "github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/lessonType"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/libraries/jquery"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/libraries/tmpl"
	jsmap "github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/map"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/sum"
	"github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/translator"
	weblogicgenerator "github.com/LaPingvino/openteacher/internal/modules/logic/javaScript/webLogicGenerator"
	javascriptpercentscalculator "github.com/LaPingvino/openteacher/internal/modules/logic/javaScriptPercentsCalculator"
	languagecodeguesser "github.com/LaPingvino/openteacher/internal/modules/logic/languageCodeGuesser"
	allonce "github.com/LaPingvino/openteacher/internal/modules/logic/lessonTypes/allOnce"
	"github.com/LaPingvino/openteacher/internal/modules/logic/lessonTypes/smart"
	hardwords "github.com/LaPingvino/openteacher/internal/modules/logic/listModifiers/hardWords"
	random "github.com/LaPingvino/openteacher/internal/modules/logic/listModifiers/random_"
	"github.com/LaPingvino/openteacher/internal/modules/logic/listModifiers/reverse"
	"github.com/LaPingvino/openteacher/internal/modules/logic/listModifiers/sort"
	wordsneveransweredcorrectly "github.com/LaPingvino/openteacher/internal/modules/logic/listModifiers/wordsNeverAnsweredCorrectly"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loader"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/abbyy"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/anki"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/anki2"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/apkg"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/backpack"
	csv "github.com/LaPingvino/openteacher/internal/modules/logic/loaders/csv_"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/cuecard"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/domingo"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/flashqard"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/fmd"
	gnuvocabtrain "github.com/LaPingvino/openteacher/internal/modules/logic/loaders/gnuVocabTrain"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/granule"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/jml"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/jvlt"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/kgm"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/kvtml"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/ludem"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/mnemosyne"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/ot"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/otmd"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/ottp"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/otwd"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/overhoor"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/ovr"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/pauker"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/t2k"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/teachmaster"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/voca"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/vocabularium"
	vokabeltrainer "github.com/LaPingvino/openteacher/internal/modules/logic/loaders/vokabelTrainer"
	vtraintxt "github.com/LaPingvino/openteacher/internal/modules/logic/loaders/vtrainTxt"
	"github.com/LaPingvino/openteacher/internal/modules/logic/loaders/wrts"
	mergerwords "github.com/LaPingvino/openteacher/internal/modules/logic/mergers/words"
	mimicrytypefaceconverter "github.com/LaPingvino/openteacher/internal/modules/logic/mimicryTypefaceConverter"
	logicmodules "github.com/LaPingvino/openteacher/internal/modules/logic/modules"
	notecalculatorchooser "github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculatorChooser"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/american"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/dutch"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/ects"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/french"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/german"
	"github.com/LaPingvino/openteacher/internal/modules/logic/noteCalculators/javaScript/percents"
	cuneiformrecognizer "github.com/LaPingvino/openteacher/internal/modules/logic/ocr/cuneiformRecognizer"
	tesseractrecognizer "github.com/LaPingvino/openteacher/internal/modules/logic/ocr/tesseractRecognizer"
	wordlistloader "github.com/LaPingvino/openteacher/internal/modules/logic/ocr/wordListLoader"
	"github.com/LaPingvino/openteacher/internal/modules/logic/odtsaver"
	"github.com/LaPingvino/openteacher/internal/modules/logic/otxxloader"
	"github.com/LaPingvino/openteacher/internal/modules/logic/otxxsaver"
	percentscalculator "github.com/LaPingvino/openteacher/internal/modules/logic/percentsCalculator"
	pyinstallerinterface "github.com/LaPingvino/openteacher/internal/modules/logic/pyinstallerInterface"
	recentlyopened "github.com/LaPingvino/openteacher/internal/modules/logic/recentlyOpened"

	"github.com/LaPingvino/openteacher/internal/modules/logic/reversers/words"
	safehtmlchecker "github.com/LaPingvino/openteacher/internal/modules/logic/safeHtmlChecker"
	"github.com/LaPingvino/openteacher/internal/modules/logic/saver"
	"github.com/LaPingvino/openteacher/internal/modules/logic/savers/latex"
	libreofficeformats "github.com/LaPingvino/openteacher/internal/modules/logic/savers/libreofficeFormats"
	mediahtml "github.com/LaPingvino/openteacher/internal/modules/logic/savers/mediaHtml"
	"github.com/LaPingvino/openteacher/internal/modules/logic/savers/odt"
	"github.com/LaPingvino/openteacher/internal/modules/logic/savers/pdf"
	"github.com/LaPingvino/openteacher/internal/modules/logic/savers/png"
	"github.com/LaPingvino/openteacher/internal/modules/logic/savers/sylk"
	topohtml "github.com/LaPingvino/openteacher/internal/modules/logic/savers/topoHtml"
	wordshtml "github.com/LaPingvino/openteacher/internal/modules/logic/savers/wordsHtml"

	testtypesmedia "github.com/LaPingvino/openteacher/internal/modules/logic/testTypes/media"
	testtypestopo "github.com/LaPingvino/openteacher/internal/modules/logic/testTypes/topo"
	testtypeswords "github.com/LaPingvino/openteacher/internal/modules/logic/testTypes/words"

	logictranslator "github.com/LaPingvino/openteacher/internal/modules/logic/translator"
	// Removed duplicate auto-converted modules - using manually implemented versions instead
)

const (
	appName    = "OpenTeacher"
	appVersion = "4.0.0-alpha"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("%s %s - Starting...\n", appName, appVersion)

	// Create module manager
	manager := core.NewManager()

	// Register all modules
	if err := registerAllModules(manager); err != nil {
		log.Fatalf("Failed to register modules: %v", err)
	}

	fmt.Printf("Registered %d modules of %d types\n", manager.ModuleCount(), manager.TypeCount())
	fmt.Printf("Available module types: %v\n", manager.ListTypes())

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Println("Shutting down gracefully...")
		cancel()
	}()

	// Enable all modules
	fmt.Println("Enabling modules...")
	if err := manager.EnableAll(ctx); err != nil {
		log.Fatalf("Failed to enable modules: %v", err)
	}

	fmt.Println("All modules enabled successfully")

	// Start the main application
	if err := runApplication(ctx, manager); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	// Shutdown
	fmt.Println("Disabling modules...")
	if err := manager.DisableAll(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("OpenTeacher shutdown complete")
}

func registerAllModules(manager *core.Manager) error {
	// Create and register essential modules
	executeModule := modules.NewExecuteModule()
	if err := manager.Register(executeModule); err != nil {
		return fmt.Errorf("failed to register execute module: %w", err)
	}

	eventModule := modules.NewEventModule()
	if err := manager.Register(eventModule); err != nil {
		return fmt.Errorf("failed to register event module: %w", err)
	}

	settingsModule := modules.NewSettingsModule()
	if err := manager.Register(settingsModule); err != nil {
		return fmt.Errorf("failed to register settings module: %w", err)
	}

	// Skip duplicate modules.NewMetadataModule - using real metadata module instead

	buttonRegisterModule := modules.NewButtonRegisterModule()
	if err := manager.Register(buttonRegisterModule); err != nil {
		return fmt.Errorf("failed to register button register module: %w", err)
	}

	startWidgetModule := startwidget.InitStartwidgetModule()
	if err := manager.Register(startWidgetModule); err != nil {
		return fmt.Errorf("failed to register start widget module: %w", err)
	}

	// Register qtApp module first (required by GUI module)
	qtappModule := qtapp.NewQtAppModule()
	if err := manager.Register(qtappModule); err != nil {
		return fmt.Errorf("failed to register qtapp module: %w", err)
	}

	// Register the real Qt GUI module instead of stub
	guiModule := gui.InitGuiModule()
	if err := manager.Register(guiModule); err != nil {
		return fmt.Errorf("failed to register gui module: %w", err)
	}

	// Register dialog modules
	fileDialogModule := file.NewFileDialogModule()
	if err := manager.Register(fileDialogModule); err != nil {
		return fmt.Errorf("failed to register file dialog module: %w", err)
	}

	aboutDialogModule := about.NewAboutDialogModule()
	if err := manager.Register(aboutDialogModule); err != nil {
		return fmt.Errorf("failed to register about dialog module: %w", err)
	}

	settingsDialogModule := settingsDialog.NewSettingsDialogModule()
	if err := manager.Register(settingsDialogModule); err != nil {
		return fmt.Errorf("failed to register settings dialog module: %w", err)
	}

	lessonDialogsModule := lessonDialogs.NewLessonDialogsModule()
	if err := manager.Register(lessonDialogsModule); err != nil {
		return fmt.Errorf("failed to register lesson dialogs module: %w", err)
	}

	// Temporarily disable business card and background image modules to test core system
	// TODO: Re-enable once Qt UI system is properly integrated
	// businessCardModule := businesscard.Init()
	// if err := manager.Register(businessCardModule); err != nil {
	// 	return fmt.Errorf("failed to register business card module: %w", err)
	// }

	// backgroundImageModule := backgroundimage.Init()
	// if err := manager.Register(backgroundImageModule); err != nil {
	// 	return fmt.Errorf("failed to register background image module: %w", err)
	// }

	profileDescModule := profiledescriptions.Init()
	if err := manager.Register(profileDescModule); err != nil {
		return fmt.Errorf("failed to register profile description module: %w", err)
	}

	// qtApp module is now registered above with GUI module

	// Register charskeyboard module - DISABLED for now
	// charskeyboardModule := charskeyboard.NewCharsKeyboardModule()
	// if err := manager.Register(charskeyboardModule); err != nil {
	// 	return fmt.Errorf("failed to register charskeyboard module: %w", err)
	// }

	// Register dialogshower module - DISABLED for now
	// dialogshowerModule := dialogshower.NewDialogShowerModule()
	// if err := manager.Register(dialogshowerModule); err != nil {
	// 	return fmt.Errorf("failed to register dialogshower module: %w", err)
	// }

	// Register about module - DISABLED for now
	// aboutModule := about.NewAboutDialogModule()
	// if err := manager.Register(aboutModule); err != nil {
	// 	return fmt.Errorf("failed to register about module: %w", err)
	// }

	// Register documentation module - disabled due to build constraints
	// documentationModule := documentation.NewDocumentationModule()
	// if err := manager.Register(documentationModule); err != nil {
	//	return fmt.Errorf("failed to register documentation module: %w", err)
	// }

	// Register file module - DISABLED for now
	// fileModule := file.NewFileDialogModule()
	// if err := manager.Register(fileModule); err != nil {
	// 	return fmt.Errorf("failed to register file module: %w", err)
	// }

	// Register print module - DISABLED for now
	// printModule := print.NewPrintDialogModule()
	// if err := manager.Register(printModule); err != nil {
	// 	return fmt.Errorf("failed to register print module: %w", err)
	// }

	// Register results module - DISABLED for now
	// resultsModule := results.NewResultsDialogModule()
	// if err := manager.Register(resultsModule); err != nil {
	// 	return fmt.Errorf("failed to register results module: %w", err)
	// }

	// Register settings module - DISABLED for now
	// settingsModule := settings.NewSettingsDialogModule()
	// if err := manager.Register(settingsModule); err != nil {
	// 	return fmt.Errorf("failed to register settings module: %w", err)
	// }

	// Register media module - DISABLED for now
	// mediaModule := media.NewMediaEntererModule()
	// if err := manager.Register(mediaModule); err != nil {
	// 	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// Register plaintextwords module - DISABLED for now
	// plaintextwordsModule := plaintextwords.NewPlainTextWordsEntererModule()
	// if err := manager.Register(plaintextwordsModule); err != nil {
	// 	return fmt.Errorf("failed to register plaintextwords module: %w", err)
	// }

	// Register topo module - DISABLED for now
	// topoModule := topo.NewTopoEntererModule()
	// if err := manager.Register(topoModule); err != nil {
	// 	return fmt.Errorf("failed to register topo module: %w", err)
	// }

	// Temporarily disable problematic Qt modules to get basic system working
	// TODO: Re-enable once Qt API issues are resolved

	// Register words module - DISABLED for now
	// wordsModule := words.NewWordsEntererModule()
	// if err := manager.Register(wordsModule); err != nil {
	// 	return fmt.Errorf("failed to register words module: %w", err)
	// }

	// Register gui module - DISABLED for now
	// guiModule := gui.NewGuiModule()
	// if err := manager.Register(guiModule); err != nil {
	// 	return fmt.Errorf("failed to register gui module: %w", err)
	// }

	// Register gui ui module (second function from gui package) - DISABLED for now
	// guiuiModule := gui.NewLessonTabWidget()
	// if err := manager.Register(guiuiModule); err != nil {
	// 	return fmt.Errorf("failed to register gui ui module: %w", err)
	// }

	// Register hiddenbrowser module - DISABLED for now
	// hiddenbrowserModule := hiddenbrowser.NewHiddenBrowserModule()
	// if err := manager.Register(hiddenbrowserModule); err != nil {
	// 	return fmt.Errorf("failed to register hiddenbrowser module: %w", err)
	// }

	// Register inputtyping module - DISABLED for now
	// inputtypingModule := inputtyping.NewInputTypingModule()
	// if err := manager.Register(inputtypingModule); err != nil {
	// 	return fmt.Errorf("failed to register inputtyping module: %w", err)
	// }

	// Register lessondialogs module - DISABLED for now
	// lessondialogsModule := lessondialogs.NewLessonDialogsModule()
	// if err := manager.Register(lessondialogsModule); err != nil {
	// 	return fmt.Errorf("failed to register lessondialogs module: %w", err)
	// }

	// Register media module
	mediaModule := testtypesmedia.NewMediaTestTypeModule()
	if err := manager.Register(mediaModule); err != nil {
		return fmt.Errorf("failed to register media module: %w", err)
	}

	// Register topo module
	teachtopoModule := testtypestopo.NewTopoTestTypeModule()
	if err := manager.Register(teachtopoModule); err != nil {
		return fmt.Errorf("failed to register topo module: %w", err)
	}

	// Register words module
	teachwordsModule := testtypeswords.NewWordsTestTypeModule()
	if err := manager.Register(teachwordsModule); err != nil {
		return fmt.Errorf("failed to register words module: %w", err)
	}

	// Register loadergui module
	loaderguiModule := loadergui.NewLoaderGuiModule()
	if err := manager.Register(loaderguiModule); err != nil {
		return fmt.Errorf("failed to register loadergui module: %w", err)
	}

	// Register mediadisplay module
	mediadisplayModule := mediadisplay.NewMediaDisplayModule()
	if err := manager.Register(mediadisplayModule); err != nil {
		return fmt.Errorf("failed to register mediadisplay module: %w", err)
	}

	// Register audio module
	audioModule := audio.NewMediaTypeModule()
	if err := manager.Register(audioModule); err != nil {
		return fmt.Errorf("failed to register audio module: %w", err)
	}

	// Register dailymotion module
	dailymotionModule := dailymotion.NewMediaTypeModule()
	if err := manager.Register(dailymotionModule); err != nil {
		return fmt.Errorf("failed to register dailymotion module: %w", err)
	}

	// Register image module
	imageModule := image.NewMediaTypeModule()
	if err := manager.Register(imageModule); err != nil {
		return fmt.Errorf("failed to register image module: %w", err)
	}

	// Register liveleak module
	liveleakModule := liveleak.NewMediaTypeModule()
	if err := manager.Register(liveleakModule); err != nil {
		return fmt.Errorf("failed to register liveleak module: %w", err)
	}

	// Register text module
	textModule := text.NewMediaTypeModule()
	if err := manager.Register(textModule); err != nil {
		return fmt.Errorf("failed to register text module: %w", err)
	}

	// Register video module
	videoModule := video.NewMediaTypeModule()
	if err := manager.Register(videoModule); err != nil {
		return fmt.Errorf("failed to register video module: %w", err)
	}

	// Register vimeo module
	vimeoModule := vimeo.NewMediaTypeModule()
	if err := manager.Register(vimeoModule); err != nil {
		return fmt.Errorf("failed to register vimeo module: %w", err)
	}

	// Register website module
	websiteModule := website.NewMediaTypeModule()
	if err := manager.Register(websiteModule); err != nil {
		return fmt.Errorf("failed to register website module: %w", err)
	}

	// Register youtube module
	youtubeModule := youtube.NewMediaTypeModule()
	if err := manager.Register(youtubeModule); err != nil {
		return fmt.Errorf("failed to register youtube module: %w", err)
	}

	// Register ocrgui module
	ocrguiModule := ocrgui.NewOcrGuiModule()
	if err := manager.Register(ocrguiModule); err != nil {
		return fmt.Errorf("failed to register ocrgui module: %w", err)
	}

	// Register percentnotesviewer module
	percentnotesviewerModule := percentnotesviewer.NewPercentNotesViewerModule()
	if err := manager.Register(percentnotesviewerModule); err != nil {
		return fmt.Errorf("failed to register percentnotesviewer module: %w", err)
	}

	// Register media module - DISABLED (duplicate module name conflict)
	// printmediaModule := testtypesmedia.NewMediaTestTypeModule()
	// if err := manager.Register(printmediaModule); err != nil {
	//	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// Register topo module - DISABLED (duplicate module name conflict)
	// printtopoModule := testtypestopo.NewTopoTestTypeModule()
	// if err := manager.Register(printtopoModule); err != nil {
	//	return fmt.Errorf("failed to register topo module: %w", err)
	// }

	// Register words module - DISABLED (duplicate module name conflict)
	// printwordsModule := testtypeswords.NewWordsTestTypeModule()
	// if err := manager.Register(printwordsModule); err != nil {
	//	return fmt.Errorf("failed to register words module: %w", err)
	// }

	// Register printer module
	printerModule := printer.NewPrinterModule()
	if err := manager.Register(printerModule); err != nil {
		return fmt.Errorf("failed to register printer module: %w", err)
	}

	// Register progressviewer module
	progressviewerModule := progressviewer.NewProgressViewerModule()
	if err := manager.Register(progressviewerModule); err != nil {
		return fmt.Errorf("failed to register progressviewer module: %w", err)
	}

	// Register recentlyopenedviewer module
	recentlyopenedviewerModule := recentlyopenedviewer.NewRecentlyOpenedViewerModule()
	if err := manager.Register(recentlyopenedviewerModule); err != nil {
		return fmt.Errorf("failed to register recentlyopenedviewer module: %w", err)
	}

	// Register boolean module
	booleanModule := boolean.NewSettingsWidgetModule()
	if err := manager.Register(booleanModule); err != nil {
		return fmt.Errorf("failed to register boolean module: %w", err)
	}

	// Register charactertable module
	charactertableModule := charactertable.NewSettingsWidgetModule()
	if err := manager.Register(charactertableModule); err != nil {
		return fmt.Errorf("failed to register charactertable module: %w", err)
	}

	// Register language module
	languageModule := language.NewSettingsWidgetModule()
	if err := manager.Register(languageModule); err != nil {
		return fmt.Errorf("failed to register language module: %w", err)
	}

	// Register longtext module
	longtextModule := longtext.NewSettingsWidgetModule()
	if err := manager.Register(longtextModule); err != nil {
		return fmt.Errorf("failed to register longtext module: %w", err)
	}

	// Register multioption module
	multioptionModule := multioption.NewSettingsWidgetModule()
	if err := manager.Register(multioptionModule); err != nil {
		return fmt.Errorf("failed to register multioption module: %w", err)
	}

	// Register number module
	numberModule := number.NewSettingsWidgetModule()
	if err := manager.Register(numberModule); err != nil {
		return fmt.Errorf("failed to register number module: %w", err)
	}

	// Register option module
	optionModule := option.NewSettingsWidgetModule()
	if err := manager.Register(optionModule); err != nil {
		return fmt.Errorf("failed to register option module: %w", err)
	}

	// Register password module
	passwordModule := password.NewSettingsWidgetModule()
	if err := manager.Register(passwordModule); err != nil {
		return fmt.Errorf("failed to register password module: %w", err)
	}

	// Register profile module
	profileModule := profile.NewSettingsWidgetModule()
	if err := manager.Register(profileModule); err != nil {
		return fmt.Errorf("failed to register profile module: %w", err)
	}

	// Register shorttext module
	shorttextModule := shorttext.NewSettingsWidgetModule()
	if err := manager.Register(shorttextModule); err != nil {
		return fmt.Errorf("failed to register shorttext module: %w", err)
	}

	// Register settingswidgets module
	settingswidgetsModule := settingswidgets.NewSettingsWidgetsModule()
	if err := manager.Register(settingswidgetsModule); err != nil {
		return fmt.Errorf("failed to register settingswidgets module: %w", err)
	}

	// Register startwidget module
	startwidgetModule := startwidget.NewStartwidgetModule()
	if err := manager.Register(startwidgetModule); err != nil {
		return fmt.Errorf("failed to register startwidget module: %w", err)
	}

	// Register hangman graphics module (from hangman package)
	hangmangraphicsModule := hangman.NewHangmanGraphics()
	if err := manager.Register(hangmangraphicsModule); err != nil {
		return fmt.Errorf("failed to register hangman graphics module: %w", err)
	}

	// Register hangman module
	hangmanModule := hangman.NewTypingTeachTypeModule()
	if err := manager.Register(hangmanModule); err != nil {
		return fmt.Errorf("failed to register hangman module: %w", err)
	}

	// Register inmind module
	inmindModule := inmind.NewInMindTeachTypeModule()
	if err := manager.Register(inmindModule); err != nil {
		return fmt.Errorf("failed to register inmind module: %w", err)
	}

	// Register repeatanswer module
	repeatanswerModule := repeatanswer.NewRepeatAnswerTeachTypeModule()
	if err := manager.Register(repeatanswerModule); err != nil {
		return fmt.Errorf("failed to register repeatanswer module: %w", err)
	}

	// Register shuffleanswer module
	shuffleanswerModule := shuffleanswer.NewShuffleAnswerTeachTypeModule()
	if err := manager.Register(shuffleanswerModule); err != nil {
		return fmt.Errorf("failed to register shuffleanswer module: %w", err)
	}

	// Register typing module
	typingModule := typing.NewTypingTeachTypeModule()
	if err := manager.Register(typingModule); err != nil {
		return fmt.Errorf("failed to register typing module: %w", err)
	}

	// Register media module - DISABLED (duplicate module name conflict)
	// teachermediaModule := testtypesmedia.NewMediaTestTypeModule()
	// if err := manager.Register(teachermediaModule); err != nil {
	//	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// Register topo module - DISABLED (duplicate module name conflict)
	// teachertopoModule := testtypestopo.NewTopoTestTypeModule()
	// if err := manager.Register(teachertopoModule); err != nil {
	//	return fmt.Errorf("failed to register topo module: %w", err)
	// }

	// Register words module - DISABLED (duplicate module name conflict)
	// teacherwordsModule := testtypeswords.NewWordsTestTypeModule()
	// if err := manager.Register(teacherwordsModule); err != nil {
	//	return fmt.Errorf("failed to register words module: %w", err)
	// }

	// Register menu module
	menuModule := menu.NewTestMenuModule()
	if err := manager.Register(menuModule); err != nil {
		return fmt.Errorf("failed to register menu module: %w", err)
	}

	// Register studentsview module
	studentsviewModule := studentsview.NewTestModeStudentsViewModule()
	if err := manager.Register(studentsviewModule); err != nil {
		return fmt.Errorf("failed to register studentsview module: %w", err)
	}

	// Register teacher module - DISABLED for now (module doesn't exist)
	// teacherModule := teacher.NewTestModeTeacherModule()
	// if err := manager.Register(teacherModule); err != nil {
	// 	return fmt.Errorf("failed to register teacher module: %w", err)
	// }

	// Register teacherpanel module
	teacherpanelModule := teacherpanel.NewTestModeTeacherPanelModule()
	if err := manager.Register(teacherpanelModule); err != nil {
		return fmt.Errorf("failed to register teacherpanel module: %w", err)
	}

	// Register theme module
	themeModule := theme.NewThemeModule()
	if err := manager.Register(themeModule); err != nil {
		return fmt.Errorf("failed to register theme module: %w", err)
	}

	// Register topomaps module
	topomapsModule := topomaps.NewTopoMapsModule()
	if err := manager.Register(topomapsModule); err != nil {
		return fmt.Errorf("failed to register topomaps module: %w", err)
	}

	// Register keyboard module
	keyboardModule := keyboard.NewTypingTutorKeyboardModule()
	if err := manager.Register(keyboardModule); err != nil {
		return fmt.Errorf("failed to register keyboard module: %w", err)
	}

	// Skip main module - it's a program not a library
	// mainModule := main.NewTypingTutorModule()
	// if err := manager.Register(mainModule); err != nil {
	// 	return fmt.Errorf("failed to register main module: %w", err)
	// }

	// Register translator module
	logictranslatorModule := logictranslator.NewTranslatorModule()
	if err := manager.Register(logictranslatorModule); err != nil {
		return fmt.Errorf("failed to register translator module: %w", err)
	}

	// Register cyrillic module
	cyrillicModule := cyrillic.NewCyrillicModule()
	if err := manager.Register(cyrillicModule); err != nil {
		return fmt.Errorf("failed to register cyrillic module: %w", err)
	}

	// Register greek module
	greekModule := greek.NewGreekModule()
	if err := manager.Register(greekModule); err != nil {
		return fmt.Errorf("failed to register greek module: %w", err)
	}

	// Register symbols module
	symbolsModule := symbols.NewSymbolsModule()
	if err := manager.Register(symbolsModule); err != nil {
		return fmt.Errorf("failed to register symbols module: %w", err)
	}

	// Register datatypeicons module
	datatypeiconsModule := datatypeicons.NewDataTypeIconsModule()
	if err := manager.Register(datatypeiconsModule); err != nil {
		return fmt.Errorf("failed to register datatypeicons module: %w", err)
	}

	// Register africa module
	africaModule := africa.NewMapModule()
	if err := manager.Register(africaModule); err != nil {
		return fmt.Errorf("failed to register africa module: %w", err)
	}

	// Register asia module
	asiaModule := asia.NewMapModule()
	if err := manager.Register(asiaModule); err != nil {
		return fmt.Errorf("failed to register asia module: %w", err)
	}

	// Register europe module
	europeModule := europe.NewMapModule()
	if err := manager.Register(europeModule); err != nil {
		return fmt.Errorf("failed to register europe module: %w", err)
	}

	// Register latinamerica module
	latinamericaModule := latinamerica.NewMapModule()
	if err := manager.Register(latinamericaModule); err != nil {
		return fmt.Errorf("failed to register latinamerica module: %w", err)
	}

	// Register usa module
	usaModule := usa.NewMapModule()
	if err := manager.Register(usaModule); err != nil {
		return fmt.Errorf("failed to register usa module: %w", err)
	}

	// Register world module
	worldModule := world.NewMapModule()
	if err := manager.Register(worldModule); err != nil {
		return fmt.Errorf("failed to register world module: %w", err)
	}

	// Register metadata module
	metadataModule := metadata.NewMetadataModule()
	if err := manager.Register(metadataModule); err != nil {
		return fmt.Errorf("failed to register metadata module: %w", err)
	}

	// Register openteacherauthors module
	openteacherauthorsModule := openteacherauthors.NewOpenTeacherAuthorsModule()
	if err := manager.Register(openteacherauthorsModule); err != nil {
		return fmt.Errorf("failed to register openteacherauthors module: %w", err)
	}

	// Register all module
	allModule := all.NewProfileDescriptionModule()
	if err := manager.Register(allModule); err != nil {
		return fmt.Errorf("failed to register all module: %w", err)
	}

	// Register cli module
	cliModule := cli.NewProfileDescriptionModule()
	if err := manager.Register(cliModule); err != nil {
		return fmt.Errorf("failed to register cli module: %w", err)
	}

	// Register codecomplexity module
	codecomplexityModule := codecomplexity.NewProfileDescriptionModule()
	if err := manager.Register(codecomplexityModule); err != nil {
		return fmt.Errorf("failed to register codecomplexity module: %w", err)
	}

	// Register codedocumentation module
	codedocumentationModule := codedocumentation.NewProfileDescriptionModule()
	if err := manager.Register(codedocumentationModule); err != nil {
		return fmt.Errorf("failed to register codedocumentation module: %w", err)
	}

	// Register generatebusinesscard module
	generatebusinesscardModule := generatebusinesscard.NewProfileDescriptionModule()
	if err := manager.Register(generatebusinesscardModule); err != nil {
		return fmt.Errorf("failed to register generatebusinesscard module: %w", err)
	}

	// Register generatelanguagecodeguessertable module
	generatelanguagecodeguessertableModule := generatelanguagecodeguessertable.NewProfileDescriptionModule()
	if err := manager.Register(generatelanguagecodeguessertableModule); err != nil {
		return fmt.Errorf("failed to register generatelanguagecodeguessertable module: %w", err)
	}

	// Register generatemobile module
	generatemobileModule := generatemobile.NewProfileDescriptionModule()
	if err := manager.Register(generatemobileModule); err != nil {
		return fmt.Errorf("failed to register generatemobile module: %w", err)
	}

	// Register generateweb module
	generatewebModule := generateweb.NewProfileDescriptionModule()
	if err := manager.Register(generatewebModule); err != nil {
		return fmt.Errorf("failed to register generateweb module: %w", err)
	}

	// Register generatewebsite module
	generatewebsiteModule := generatewebsite.NewProfileDescriptionModule()
	if err := manager.Register(generatewebsiteModule); err != nil {
		return fmt.Errorf("failed to register generatewebsite module: %w", err)
	}

	// Register gettranslationauthors module
	gettranslationauthorsModule := gettranslationauthors.NewProfileDescriptionModule()
	if err := manager.Register(gettranslationauthorsModule); err != nil {
		return fmt.Errorf("failed to register gettranslationauthors module: %w", err)
	}

	// Register help module
	helpModule := help.NewProfileDescriptionModule()
	if err := manager.Register(helpModule); err != nil {
		return fmt.Errorf("failed to register help module: %w", err)
	}

	// Register ircbot module
	ircbotModule := ircbot.NewProfileDescriptionModule()
	if err := manager.Register(ircbotModule); err != nil {
		return fmt.Errorf("failed to register ircbot module: %w", err)
	}

	// Register modulegraph module
	modulegraphModule := modulegraph.NewProfileDescriptionModule()
	if err := manager.Register(modulegraphModule); err != nil {
		return fmt.Errorf("failed to register modulegraph module: %w", err)
	}

	// Register packagearch module
	packagearchModule := packagearch.NewProfileDescriptionModule()
	if err := manager.Register(packagearchModule); err != nil {
		return fmt.Errorf("failed to register packagearch module: %w", err)
	}

	// Register packagedebian module
	packagedebianModule := packagedebian.NewProfileDescriptionModule()
	if err := manager.Register(packagedebianModule); err != nil {
		return fmt.Errorf("failed to register packagedebian module: %w", err)
	}

	// Register packagemac module
	packagemacModule := packagemac.NewProfileDescriptionModule()
	if err := manager.Register(packagemacModule); err != nil {
		return fmt.Errorf("failed to register packagemac module: %w", err)
	}

	// Register packagerpm module
	packagerpmModule := packagerpm.NewProfileDescriptionModule()
	if err := manager.Register(packagerpmModule); err != nil {
		return fmt.Errorf("failed to register packagerpm module: %w", err)
	}

	// Register packagesource module
	packagesourceModule := packagesource.NewProfileDescriptionModule()
	if err := manager.Register(packagesourceModule); err != nil {
		return fmt.Errorf("failed to register packagesource module: %w", err)
	}

	// Register packagesourcewithsetup module
	packagesourcewithsetupModule := packagesourcewithsetup.NewProfileDescriptionModule()
	if err := manager.Register(packagesourcewithsetupModule); err != nil {
		return fmt.Errorf("failed to register packagesourcewithsetup module: %w", err)
	}

	// Register packagewindowsmsi module
	packagewindowsmsiModule := packagewindowsmsi.NewProfileDescriptionModule()
	if err := manager.Register(packagewindowsmsiModule); err != nil {
		return fmt.Errorf("failed to register packagewindowsmsi module: %w", err)
	}

	// Register packagewindowsportable module
	packagewindowsportableModule := packagewindowsportable.NewProfileDescriptionModule()
	if err := manager.Register(packagewindowsportableModule); err != nil {
		return fmt.Errorf("failed to register packagewindowsportable module: %w", err)
	}

	// Register selfstudy module
	selfstudyModule := selfstudy.NewProfileDescriptionModule()
	if err := manager.Register(selfstudyModule); err != nil {
		return fmt.Errorf("failed to register selfstudy module: %w", err)
	}

	// Register shell module
	shellModule := shell.NewProfileDescriptionModule()
	if err := manager.Register(shellModule); err != nil {
		return fmt.Errorf("failed to register shell module: %w", err)
	}

	// Register studentathome module
	studentathomeModule := studentathome.NewProfileDescriptionModule()
	if err := manager.Register(studentathomeModule); err != nil {
		return fmt.Errorf("failed to register studentathome module: %w", err)
	}

	// Register studentatschool module
	studentatschoolModule := studentatschool.NewProfileDescriptionModule()
	if err := manager.Register(studentatschoolModule); err != nil {
		return fmt.Errorf("failed to register studentatschool module: %w", err)
	}

	// Register teacher module
	teacherModule := teacher.NewProfileDescriptionModule()
	if err := manager.Register(teacherModule); err != nil {
		return fmt.Errorf("failed to register teacher module: %w", err)
	}

	// Register updaterosetta module
	updaterosettaModule := updaterosetta.NewProfileDescriptionModule()
	if err := manager.Register(updaterosettaModule); err != nil {
		return fmt.Errorf("failed to register updaterosetta module: %w", err)
	}

	// Register updatetranslations module
	updatetranslationsModule := updatetranslations.NewProfileDescriptionModule()
	if err := manager.Register(updatetranslationsModule); err != nil {
		return fmt.Errorf("failed to register updatetranslations module: %w", err)
	}

	// Register webservicesserver module
	profilewebservicesserverModule := profilewebservicesserver.NewProfileDescriptionModule()
	if err := manager.Register(profilewebservicesserverModule); err != nil {
		return fmt.Errorf("failed to register webservicesserver module: %w", err)
	}

	// Register wordsonly module
	wordsonlyModule := wordsonly.NewProfileDescriptionModule()
	if err := manager.Register(wordsonlyModule); err != nil {
		return fmt.Errorf("failed to register wordsonly module: %w", err)
	}

	// Register userdocumentation module
	userdocumentationModule := userdocumentation.NewUserDocumentationModule()
	if err := manager.Register(userdocumentationModule); err != nil {
		return fmt.Errorf("failed to register userdocumentation module: %w", err)
	}

	// Register authors module
	authorsModule := authors.NewAuthorsModule()
	if err := manager.Register(authorsModule); err != nil {
		return fmt.Errorf("failed to register authors module: %w", err)
	}

	// Register event module
	logiceventModule := logicevent.NewEventModule()
	if err := manager.Register(logiceventModule); err != nil {
		return fmt.Errorf("failed to register event module: %w", err)
	}
	fmt.Printf("  ✓ Registered event module\n")

	// Register execute module
	logicExecuteModule := execute.NewExecuteModule()
	if err := manager.Register(logicExecuteModule); err != nil {
		return fmt.Errorf("failed to register execute module: %w", err)
	}
	fmt.Printf("  ✓ Registered execute module\n")

	// Register friendlytranslationnames module
	friendlytranslationnamesModule := friendlytranslationnames.NewFriendlyTranslationNamesModule()
	if err := manager.Register(friendlytranslationnamesModule); err != nil {
		return fmt.Errorf("failed to register friendlytranslationnames module: %w", err)
	}

	// Register javascriptwords module
	javascriptwordsModule := javascriptwords.NewWordsHtmlGeneratorModule()
	if err := manager.Register(javascriptwordsModule); err != nil {
		return fmt.Errorf("failed to register javascriptwords module: %w", err)
	}

	// Register media module - DISABLED (duplicate module name conflict)
	// htmlgenmediaModule := testtypesmedia.NewMediaTestTypeModule()
	// if err := manager.Register(htmlgenmediaModule); err != nil {
	//	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// Register topo module
	htmltopoModule := htmltopo.NewTopoHtmlGeneratorModule()
	if err := manager.Register(htmltopoModule); err != nil {
		return fmt.Errorf("failed to register topo module: %w", err)
	}

	// Register words module
	htmlwordsModule := htmlwords.NewWordsHtmlGeneratorModule()
	if err := manager.Register(htmlwordsModule); err != nil {
		return fmt.Errorf("failed to register words module: %w", err)
	}

	// Register foreignknown module
	foreignknownModule := foreignknown.NewForeignKnownModule()
	if err := manager.Register(foreignknownModule); err != nil {
		return fmt.Errorf("failed to register foreignknown module: %w", err)
	}

	// Register bisect module
	bisectModule := bisect.NewJSBisectModule()
	if err := manager.Register(bisectModule); err != nil {
		return fmt.Errorf("failed to register bisect module: %w", err)
	}

	// Register evaluator module
	evaluatorModule := evaluator.NewJSEvaluatorModule()
	if err := manager.Register(evaluatorModule); err != nil {
		return fmt.Errorf("failed to register evaluator module: %w", err)
	}

	// Register event module
	jsEventModule := jsevent.NewJavascriptEventModule()
	if err := manager.Register(jsEventModule); err != nil {
		return fmt.Errorf("failed to register event module: %w", err)
	}

	// Register lessontype module
	lessontypeModule := lessontype.NewJavascriptLessonTypeModule()
	if err := manager.Register(lessontypeModule); err != nil {
		return fmt.Errorf("failed to register lessontype module: %w", err)
	}

	// Register jquery module
	jqueryModule := jquery.NewJSLibModule()
	if err := manager.Register(jqueryModule); err != nil {
		return fmt.Errorf("failed to register jquery module: %w", err)
	}

	// Register tmpl module
	tmplModule := tmpl.NewJSLibModule()
	if err := manager.Register(tmplModule); err != nil {
		return fmt.Errorf("failed to register tmpl module: %w", err)
	}

	// Register map module
	mapModule := jsmap.NewJSMapModule()
	if err := manager.Register(mapModule); err != nil {
		return fmt.Errorf("failed to register map module: %w", err)
	}

	// Register sum module
	sumModule := sum.NewJSSumModule()
	if err := manager.Register(sumModule); err != nil {
		return fmt.Errorf("failed to register sum module: %w", err)
	}

	// Register translator module
	translatorModule := translator.NewJSTranslatorModule()
	if err := manager.Register(translatorModule); err != nil {
		return fmt.Errorf("failed to register translator module: %w", err)
	}

	// Register weblogicgenerator module
	weblogicgeneratorModule := weblogicgenerator.NewWebLogicGeneratorModule()
	if err := manager.Register(weblogicgeneratorModule); err != nil {
		return fmt.Errorf("failed to register weblogicgenerator module: %w", err)
	}

	// Register javascriptpercentscalculator module
	javascriptpercentscalculatorModule := javascriptpercentscalculator.NewPercentsCalculatorModule()
	if err := manager.Register(javascriptpercentscalculatorModule); err != nil {
		return fmt.Errorf("failed to register javascriptpercentscalculator module: %w", err)
	}

	// Register languagecodeguesser module
	languagecodeguesserModule := languagecodeguesser.NewLanguageCodeGuesserModule()
	if err := manager.Register(languagecodeguesserModule); err != nil {
		return fmt.Errorf("failed to register languagecodeguesser module: %w", err)
	}

	// Skip languagecodeguesserTables - merged into languagecodeguesser package
	// languagecodeguessertablesModule := languagecodeguesserTables.NewLanguagecodeguessertablesModule()
	// if err := manager.Register(languagecodeguessertablesModule); err != nil {
	// 	return fmt.Errorf("failed to register languagecodeguesserTables module: %w", err)
	// }

	// Register allonce module
	allonceModule := allonce.NewAllOnceModule()
	if err := manager.Register(allonceModule); err != nil {
		return fmt.Errorf("failed to register allonce module: %w", err)
	}

	// Register smart module
	smartModule := smart.NewSmartModule()
	if err := manager.Register(smartModule); err != nil {
		return fmt.Errorf("failed to register smart module: %w", err)
	}

	// Register hardwords module
	hardwordsModule := hardwords.NewHardWordsModule()
	if err := manager.Register(hardwordsModule); err != nil {
		return fmt.Errorf("failed to register hardwords module: %w", err)
	}

	// Register random module
	randomModule := random.NewRandomModule()
	if err := manager.Register(randomModule); err != nil {
		return fmt.Errorf("failed to register random module: %w", err)
	}

	// Register reverse module
	reverseModule := reverse.NewReverseModule()
	if err := manager.Register(reverseModule); err != nil {
		return fmt.Errorf("failed to register reverse module: %w", err)
	}

	// Register sort module
	sortModule := sort.NewSortModule()
	if err := manager.Register(sortModule); err != nil {
		return fmt.Errorf("failed to register sort module: %w", err)
	}

	// Register wordsneveransweredcorrectly module
	wordsneveransweredcorrectlyModule := wordsneveransweredcorrectly.NewWordsNeverAnsweredCorrectlyModule()
	if err := manager.Register(wordsneveransweredcorrectlyModule); err != nil {
		return fmt.Errorf("failed to register wordsneveransweredcorrectly module: %w", err)
	}

	// Register loader module
	loaderModule := loader.NewLoaderModule()
	if err := manager.Register(loaderModule); err != nil {
		return fmt.Errorf("failed to register loader module: %w", err)
	}

	// Register abbyy module
	abbyyModule := abbyy.NewAbbyyLoaderModule()
	if err := manager.Register(abbyyModule); err != nil {
		return fmt.Errorf("failed to register abbyy module: %w", err)
	}

	// Register anki module
	ankiModule := anki.NewAnkiLoaderModule()
	if err := manager.Register(ankiModule); err != nil {
		return fmt.Errorf("failed to register anki module: %w", err)
	}

	// Register anki2 module
	anki2Module := anki2.NewAnki2LoaderModule()
	if err := manager.Register(anki2Module); err != nil {
		return fmt.Errorf("failed to register anki2 module: %w", err)
	}

	// Register apkg module
	apkgModule := apkg.NewAnkiApkgLoaderModule()
	if err := manager.Register(apkgModule); err != nil {
		return fmt.Errorf("failed to register apkg module: %w", err)
	}

	// Register backpack module
	backpackModule := backpack.NewBackpackLoaderModule()
	if err := manager.Register(backpackModule); err != nil {
		return fmt.Errorf("failed to register backpack module: %w", err)
	}

	// Register csv module
	csvModule := csv.NewCsvLoaderModule()
	if err := manager.Register(csvModule); err != nil {
		return fmt.Errorf("failed to register csv module: %w", err)
	}

	// Register cuecard module
	cuecardModule := cuecard.NewCueCardLoaderModule()
	if err := manager.Register(cuecardModule); err != nil {
		return fmt.Errorf("failed to register cuecard module: %w", err)
	}

	// Register domingo module
	domingoModule := domingo.NewDomingoLoaderModule()
	if err := manager.Register(domingoModule); err != nil {
		return fmt.Errorf("failed to register domingo module: %w", err)
	}

	// Register flashqard module
	flashqardModule := flashqard.NewFlashQardLoaderModule()
	if err := manager.Register(flashqardModule); err != nil {
		return fmt.Errorf("failed to register flashqard module: %w", err)
	}

	// Register fmd module
	fmdModule := fmd.NewFmdLoaderModule()
	if err := manager.Register(fmdModule); err != nil {
		return fmt.Errorf("failed to register fmd module: %w", err)
	}

	// Register gnuvocabtrain module
	gnuvocabtrainModule := gnuvocabtrain.NewGnuVocabTrainLoaderModule()
	if err := manager.Register(gnuvocabtrainModule); err != nil {
		return fmt.Errorf("failed to register gnuvocabtrain module: %w", err)
	}

	// Register granule module
	granuleModule := granule.NewGranuleLoaderModule()
	if err := manager.Register(granuleModule); err != nil {
		return fmt.Errorf("failed to register granule module: %w", err)
	}

	// Register jml module
	jmlModule := jml.NewJMemorizeLessonLoaderModule()
	if err := manager.Register(jmlModule); err != nil {
		return fmt.Errorf("failed to register jml module: %w", err)
	}

	// Register jvlt module
	jvltModule := jvlt.NewJvltLoaderModule()
	if err := manager.Register(jvltModule); err != nil {
		return fmt.Errorf("failed to register jvlt module: %w", err)
	}

	// Register kgm module
	kgmModule := kgm.NewKGeographyMapLoaderModule()
	if err := manager.Register(kgmModule); err != nil {
		return fmt.Errorf("failed to register kgm module: %w", err)
	}

	// Register kvtml module
	kvtmlModule := kvtml.NewKvtmlLoaderModule()
	if err := manager.Register(kvtmlModule); err != nil {
		return fmt.Errorf("failed to register kvtml module: %w", err)
	}

	// Register ludem module
	ludemModule := ludem.NewLudemLoaderModule()
	if err := manager.Register(ludemModule); err != nil {
		return fmt.Errorf("failed to register ludem module: %w", err)
	}

	// Register mnemosyne module
	mnemosyneModule := mnemosyne.NewMnemosyneLoaderModule()
	if err := manager.Register(mnemosyneModule); err != nil {
		return fmt.Errorf("failed to register mnemosyne module: %w", err)
	}

	// Register ot module
	otModule := ot.NewOpenTeacherLoaderModule()
	if err := manager.Register(otModule); err != nil {
		return fmt.Errorf("failed to register ot module: %w", err)
	}

	// Register otmd module
	otmdModule := otmd.NewOpenTeachingMediaLoaderModule()
	if err := manager.Register(otmdModule); err != nil {
		return fmt.Errorf("failed to register otmd module: %w", err)
	}

	// Register ottp module
	ottpModule := ottp.NewOpenTeachingTopoLoaderModule()
	if err := manager.Register(ottpModule); err != nil {
		return fmt.Errorf("failed to register ottp module: %w", err)
	}

	// Register otwd module
	otwdModule := otwd.NewOpenTeachingWordsLoaderModule()
	if err := manager.Register(otwdModule); err != nil {
		return fmt.Errorf("failed to register otwd module: %w", err)
	}

	// Register overhoor module
	overhoorModule := overhoor.NewOverhoorLoaderModule()
	if err := manager.Register(overhoorModule); err != nil {
		return fmt.Errorf("failed to register overhoor module: %w", err)
	}

	// Register ovr module
	ovrModule := ovr.NewOverhoringsprogrammaTalenLoaderModule()
	if err := manager.Register(ovrModule); err != nil {
		return fmt.Errorf("failed to register ovr module: %w", err)
	}

	// Register pauker module
	paukerModule := pauker.NewPaukerLoaderModule()
	if err := manager.Register(paukerModule); err != nil {
		return fmt.Errorf("failed to register pauker module: %w", err)
	}

	// Register t2k module
	t2kModule := t2k.NewTeach2000LoaderModule()
	if err := manager.Register(t2kModule); err != nil {
		return fmt.Errorf("failed to register t2k module: %w", err)
	}

	// Register teachmaster module
	teachmasterModule := teachmaster.NewTeachmasterLoaderModule()
	if err := manager.Register(teachmasterModule); err != nil {
		return fmt.Errorf("failed to register teachmaster module: %w", err)
	}

	// Register voca module
	vocaModule := voca.NewVocaLoaderModule()
	if err := manager.Register(vocaModule); err != nil {
		return fmt.Errorf("failed to register voca module: %w", err)
	}

	// Register vocabularium module
	vocabulariumModule := vocabularium.NewVocabulariumLoaderModule()
	if err := manager.Register(vocabulariumModule); err != nil {
		return fmt.Errorf("failed to register vocabularium module: %w", err)
	}

	// Register vokabeltrainer module
	vokabeltrainerModule := vokabeltrainer.NewVokabelTrainerLoaderModule()
	if err := manager.Register(vokabeltrainerModule); err != nil {
		return fmt.Errorf("failed to register vokabeltrainer module: %w", err)
	}

	// Register vtraintxt module
	vtraintxtModule := vtraintxt.NewVTrainTxtLoaderModule()
	if err := manager.Register(vtraintxtModule); err != nil {
		return fmt.Errorf("failed to register vtraintxt module: %w", err)
	}

	// Register wrts module
	wrtsModule := wrts.NewWrtsLoaderModule()
	if err := manager.Register(wrtsModule); err != nil {
		return fmt.Errorf("failed to register wrts module: %w", err)
	}

	// Register words module
	mergerwordsModule := mergerwords.NewWordsMergerModule()
	if err := manager.Register(mergerwordsModule); err != nil {
		return fmt.Errorf("failed to register words module: %w", err)
	}

	// Register mimicrytypefaceconverter module
	mimicrytypefaceconverterModule := mimicrytypefaceconverter.NewMimicryTypefaceConverterModule()
	if err := manager.Register(mimicrytypefaceconverterModule); err != nil {
		return fmt.Errorf("failed to register mimicrytypefaceconverter module: %w", err)
	}

	// Register modules module
	logicModulesModule := logicmodules.NewModulesModule()
	if err := manager.Register(logicModulesModule); err != nil {
		return fmt.Errorf("failed to register modules module: %w", err)
	}

	// Skip modulestestFiletoimport - merged into modulestest package
	// modulestestfiletoimportModule := modulestestFiletoimport.NewModulestestfiletoimportModule()
	// if err := manager.Register(modulestestfiletoimportModule); err != nil {
	// 	return fmt.Errorf("failed to register modulestestFiletoimport module: %w", err)
	// }

	// Register notecalculatorchooser module
	notecalculatorchooserModule := notecalculatorchooser.NewNoteCalculatorChooserModule()
	if err := manager.Register(notecalculatorchooserModule); err != nil {
		return fmt.Errorf("failed to register notecalculatorchooser module: %w", err)
	}

	// Register american module
	americanModule := american.NewAmericanNoteCalculatorModule()
	if err := manager.Register(americanModule); err != nil {
		return fmt.Errorf("failed to register american module: %w", err)
	}

	// Register dutch module
	dutchModule := dutch.NewDutchNoteCalculatorModule()
	if err := manager.Register(dutchModule); err != nil {
		return fmt.Errorf("failed to register dutch module: %w", err)
	}

	// Register ects module
	ectsModule := ects.NewECTSNoteCalculatorModule()
	if err := manager.Register(ectsModule); err != nil {
		return fmt.Errorf("failed to register ects module: %w", err)
	}

	// Register french module
	frenchModule := french.NewFrenchNoteCalculatorModule()
	if err := manager.Register(frenchModule); err != nil {
		return fmt.Errorf("failed to register french module: %w", err)
	}

	// Register german module
	germanModule := german.NewGermanNoteCalculatorModule()
	if err := manager.Register(germanModule); err != nil {
		return fmt.Errorf("failed to register german module: %w", err)
	}

	// Register american module
	noteCalcAmericanModule := american.NewAmericanNoteCalculatorModule()
	if err := manager.Register(noteCalcAmericanModule); err != nil {
		return fmt.Errorf("failed to register american module: %w", err)
	}

	// Register dutch module
	noteCalcDutchModule := dutch.NewDutchNoteCalculatorModule()
	if err := manager.Register(noteCalcDutchModule); err != nil {
		return fmt.Errorf("failed to register dutch module: %w", err)
	}

	// Register ects module
	noteCalcEctsModule := ects.NewECTSNoteCalculatorModule()
	if err := manager.Register(noteCalcEctsModule); err != nil {
		return fmt.Errorf("failed to register ects module: %w", err)
	}

	// Register french module
	noteCalcFrenchModule := french.NewFrenchNoteCalculatorModule()
	if err := manager.Register(noteCalcFrenchModule); err != nil {
		return fmt.Errorf("failed to register french module: %w", err)
	}

	// Register german module
	noteCalcGermanModule := german.NewGermanNoteCalculatorModule()
	if err := manager.Register(noteCalcGermanModule); err != nil {
		return fmt.Errorf("failed to register german module: %w", err)
	}

	// Register percents module
	percentsModule := percents.NewPercentsNoteCalculatorModule()
	if err := manager.Register(percentsModule); err != nil {
		return fmt.Errorf("failed to register percents module: %w", err)
	}

	// Register percents module
	noteCalcPercentsModule := percents.NewPercentsNoteCalculatorModule()
	if err := manager.Register(noteCalcPercentsModule); err != nil {
		return fmt.Errorf("failed to register percents module: %w", err)
	}

	// Register cuneiformrecognizer module
	cuneiformrecognizerModule := cuneiformrecognizer.NewCuneiformOCRModule()
	if err := manager.Register(cuneiformrecognizerModule); err != nil {
		return fmt.Errorf("failed to register cuneiformrecognizer module: %w", err)
	}

	// Register tesseractrecognizer module
	tesseractrecognizerModule := tesseractrecognizer.NewTesseractOCRModule()
	if err := manager.Register(tesseractrecognizerModule); err != nil {
		return fmt.Errorf("failed to register tesseractrecognizer module: %w", err)
	}

	// Register wordlistloader module
	wordlistloaderModule := wordlistloader.NewOcrWordListLoaderModule()
	if err := manager.Register(wordlistloaderModule); err != nil {
		return fmt.Errorf("failed to register wordlistloader module: %w", err)
	}

	// Register odtsaver module
	odtsaverModule := odtsaver.NewOdtSaverModule()
	if err := manager.Register(odtsaverModule); err != nil {
		return fmt.Errorf("failed to register odtsaver module: %w", err)
	}

	// Register otxxloader module
	otxxloaderModule := otxxloader.NewOtxxLoaderModule()
	if err := manager.Register(otxxloaderModule); err != nil {
		return fmt.Errorf("failed to register otxxloader module: %w", err)
	}

	// Register otxxsaver module
	otxxsaverModule := otxxsaver.NewOtxxSaverModule()
	if err := manager.Register(otxxsaverModule); err != nil {
		return fmt.Errorf("failed to register otxxsaver module: %w", err)
	}

	// Register percentscalculator module
	percentscalculatorModule := percentscalculator.NewPercentsCalculatorModule()
	if err := manager.Register(percentscalculatorModule); err != nil {
		return fmt.Errorf("failed to register percentscalculator module: %w", err)
	}

	// Register pyinstallerinterface module
	pyinstallerinterfaceModule := pyinstallerinterface.NewPyinstallerInterfaceModule()
	if err := manager.Register(pyinstallerinterfaceModule); err != nil {
		return fmt.Errorf("failed to register pyinstallerinterface module: %w", err)
	}

	// Register recentlyopened module
	recentlyopenedModule := recentlyopened.NewRecentlyOpenedModule()
	if err := manager.Register(recentlyopenedModule); err != nil {
		return fmt.Errorf("failed to register recentlyopened module: %w", err)
	}

	// Register media module - DISABLED (import removed)
	// reversermediaModule := reversermedia.NewMediaReverserModule()
	// if err := manager.Register(reversermediaModule); err != nil {
	//	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// Register words module
	wordsModule := words.NewWordsReverserModule()
	if err := manager.Register(wordsModule); err != nil {
		return fmt.Errorf("failed to register words module: %w", err)
	}

	// Register safehtmlchecker module
	safehtmlcheckerModule := safehtmlchecker.NewSafeHtmlCheckerModule()
	if err := manager.Register(safehtmlcheckerModule); err != nil {
		return fmt.Errorf("failed to register safehtmlchecker module: %w", err)
	}

	// Register saver module
	saverModule := saver.NewSaverModule()
	if err := manager.Register(saverModule); err != nil {
		return fmt.Errorf("failed to register saver module: %w", err)
	}

	// Register csv module - DISABLED (module doesn't exist)
	// csvModule := csv.NewCsvSaverModule()
	// if err := manager.Register(csvModule); err != nil {
	//	return fmt.Errorf("failed to register csv module: %w", err)
	// }

	// Register kvtml module - DISABLED (module doesn't exist)
	// kvtmlModule := kvtml.NewKvtmlSaverModule()
	// if err := manager.Register(kvtmlModule); err != nil {
	//	return fmt.Errorf("failed to register kvtml module: %w", err)
	// }

	// Register latex module
	latexModule := latex.NewLaTeXSaverModule()
	if err := manager.Register(latexModule); err != nil {
		return fmt.Errorf("failed to register latex module: %w", err)
	}

	// Register libreofficeformats module
	libreofficeformatsModule := libreofficeformats.NewLibreofficeFormatsSaverModule()
	if err := manager.Register(libreofficeformatsModule); err != nil {
		return fmt.Errorf("failed to register libreofficeformats module: %w", err)
	}

	// Register mediahtml module
	mediahtmlModule := mediahtml.NewHtmlSaverModule()
	if err := manager.Register(mediahtmlModule); err != nil {
		return fmt.Errorf("failed to register mediahtml module: %w", err)
	}

	// Register odt module
	odtModule := odt.NewOdtSaverModule()
	if err := manager.Register(odtModule); err != nil {
		return fmt.Errorf("failed to register odt module: %w", err)
	}

	// Register ot module - DISABLED (module doesn't exist)
	// otModule := ot.NewOpenTeacherSaverModule()
	// if err := manager.Register(otModule); err != nil {
	//	return fmt.Errorf("failed to register ot module: %w", err)
	// }

	// Register ottp module - DISABLED (module doesn't exist)
	// ottpModule := ottp.NewOpenTeachingTopoSaverModule()
	// if err := manager.Register(ottpModule); err != nil {
	//	return fmt.Errorf("failed to register ottp module: %w", err)
	// }

	// Register otwd module - DISABLED (module doesn't exist)
	// otwdModule := otwd.NewOpenTeachingWordsSaverModule()
	// if err := manager.Register(otwdModule); err != nil {
	//	return fmt.Errorf("failed to register otwd module: %w", err)
	// }

	// Register pdf module
	pdfModule := pdf.NewPdfSaverModule()
	if err := manager.Register(pdfModule); err != nil {
		return fmt.Errorf("failed to register pdf module: %w", err)
	}

	// Register png module
	pngModule := png.NewPngSaverModule()
	if err := manager.Register(pngModule); err != nil {
		return fmt.Errorf("failed to register png module: %w", err)
	}

	// Register sylk module
	sylkModule := sylk.NewSylkSaverModule()
	if err := manager.Register(sylkModule); err != nil {
		return fmt.Errorf("failed to register sylk module: %w", err)
	}

	// Register t2k module - DISABLED (module doesn't exist)
	// t2kModule := t2k.NewTeach2000SaverModule()
	// if err := manager.Register(t2kModule); err != nil {
	//	return fmt.Errorf("failed to register t2k module: %w", err)
	// }

	// Register topohtml module
	topohtmlModule := topohtml.NewHtmlSaverModule()
	if err := manager.Register(topohtmlModule); err != nil {
		return fmt.Errorf("failed to register topohtml module: %w", err)
	}

	// Register wordshtml module
	wordshtmlModule := wordshtml.NewHtmlSaverModule()
	if err := manager.Register(wordshtmlModule); err != nil {
		return fmt.Errorf("failed to register wordshtml module: %w", err)
	}

	// Register wrts module - DISABLED (module doesn't exist)
	// wrtsModule := wrts.NewWrtsSaverModule()
	// if err := manager.Register(wrtsModule); err != nil {
	//	return fmt.Errorf("failed to register wrts module: %w", err)
	// }

	// Register settings module
	// Register settings module (priority 1600)
	logicSettingsModule := settings.NewSettingsModule()
	if err := manager.Register(logicSettingsModule); err != nil {
		return fmt.Errorf("failed to register settings module: %w", err)
	}
	fmt.Printf("  ✓ Registered settings module\n")

	// Register settingsfilterer module - DISABLED (causes conflicts)
	// settingsfiltererModule := settingsfilterer.NewSettingsFiltererModule()
	// if err := manager.Register(settingsfiltererModule); err != nil {
	//	return fmt.Errorf("failed to register settingsfilterer module: %w", err)
	// }

	// Register sourcewithsetupsaver module - DISABLED (causes conflicts)
	// sourcewithsetupsaverModule := sourcewithsetupsaver.NewSourceWithSetupSaverModule()
	// if err := manager.Register(sourcewithsetupsaverModule); err != nil {
	//	return fmt.Errorf("failed to register sourcewithsetupsaver module: %w", err)
	// }

	// Register spellchecker module - DISABLED (causes conflicts)
	// spellcheckerModule := spellchecker.NewSpellCheckModule()
	// if err := manager.Register(spellcheckerModule); err != nil {
	//	return fmt.Errorf("failed to register spellchecker module: %w", err)
	// }

	// Register sylksaver module - DISABLED (causes conflicts)
	// sylksaverModule := sylksaver.NewSylkSaverModule()
	// if err := manager.Register(sylksaverModule); err != nil {
	//	return fmt.Errorf("failed to register sylksaver module: %w", err)
	// }

	// Register test type modules - DISABLED (causes conflicts)
	// testtypesmediaModule := testtypesmedia.NewMediaTestTypeModule()
	// if err := manager.Register(testtypesmediaModule); err != nil {
	//	return fmt.Errorf("failed to register media module: %w", err)
	// }

	// testtypestopoModule := testtypestopo.NewTopoTestTypeModule()
	// if err := manager.Register(testtypestopoModule); err != nil {
	//	return fmt.Errorf("failed to register topo module: %w", err)
	// }

	// testtypeswordsModule := testtypeswords.NewWordsTestTypeModule()
	// if err := manager.Register(testtypeswordsModule); err != nil {
	//	return fmt.Errorf("failed to register words module: %w", err)
	// }

	// Temporarily disable remaining modules that are causing conflicts
	// TODO: Re-enable these modules once import conflicts are resolved

	// Register essential interface modules required for dependencies

	// Register buttonRegister module
	logicButtonRegisterModule := buttonregister.NewButtonRegisterModule()
	if err := manager.Register(logicButtonRegisterModule); err != nil {
		return fmt.Errorf("failed to register buttonRegister module: %w", err)
	}
	fmt.Printf("  ✓ Registered buttonRegister module\n")

	// Register inputtypinglogic module
	inputtypinglogicModule := inputtypinglogic.NewInputTypingLogicModule()
	if err := manager.Register(inputtypinglogicModule); err != nil {
		return fmt.Errorf("failed to register inputtypinglogic module: %w", err)
	}

	// Register javascriptinputtypinglogic module
	javascriptinputtypinglogicModule := javascriptinputtypinglogic.NewJSInputTypingLogicModule()
	if err := manager.Register(javascriptinputtypinglogicModule); err != nil {
		return fmt.Errorf("failed to register javascriptinputtypinglogic module: %w", err)
	}

	// Register lessontracker module
	lessontrackerModule := lessontracker.NewLessonTrackerModule()
	if err := manager.Register(lessontrackerModule); err != nil {
		return fmt.Errorf("failed to register lessontracker module: %w", err)
	}
	fmt.Printf("  ✓ Registered lessontracker module\n")

	// Register typingtutormodel module
	typingtutormodelModule := typingtutormodel.NewTypingTutorModelModule()
	if err := manager.Register(typingtutormodelModule); err != nil {
		return fmt.Errorf("failed to register typingtutormodel module: %w", err)
	}

	fmt.Println("✅ Core modules successfully registered!")
	fmt.Println("Note: Many optional modules are disabled to resolve import conflicts")

	return nil
	return nil
}

func runApplication(ctx context.Context, manager *core.Manager) error {
	// Get the GUI module and show the main window
	guiModule, exists := manager.GetDefaultModule("ui")
	if exists {
		fmt.Println("Starting Qt GUI...")

		// Show the main window
		if guiMod, ok := guiModule.(interface{ ShowMainWindow() }); ok {
			guiMod.ShowMainWindow()
		} else {
			fmt.Println("Warning: GUI module does not support ShowMainWindow()")
		}

		// Run the Qt event loop in a goroutine to allow cancellation
		if guiMod, ok := guiModule.(interface{ RunEventLoop() int }); ok {
			fmt.Println("Starting Qt event loop...")

			// Run Qt event loop in separate goroutine to handle context cancellation
			exitCodeChan := make(chan int, 1)
			go func() {
				exitCode := guiMod.RunEventLoop()
				exitCodeChan <- exitCode
			}()

			// Wait for either event loop to finish or context cancellation
			select {
			case exitCode := <-exitCodeChan:
				fmt.Printf("Qt event loop finished with exit code: %d\n", exitCode)
				return nil
			case <-ctx.Done():
				fmt.Println("Context cancelled, shutting down Qt GUI...")
				return ctx.Err()
			}
		} else {
			fmt.Println("Warning: GUI module does not support RunEventLoop()")
		}
	}

	// Fallback to execute module if no GUI or GUI doesn't work
	fmt.Println("GUI not available or not functional, falling back to execute module...")
	executeModule, exists := manager.GetDefaultModule("execute")
	if !exists {
		return fmt.Errorf("no execute or ui module found")
	}

	execMod, ok := executeModule.(core.ExecuteModule)
	if !ok {
		return fmt.Errorf("execute module does not implement ExecuteModule interface")
	}

	fmt.Println("Starting main application loop...")

	// Start running - this will block until context is cancelled
	if err := execMod.StartRunning(ctx); err != nil {
		return fmt.Errorf("execute module failed: %w", err)
	}

	return nil
}
