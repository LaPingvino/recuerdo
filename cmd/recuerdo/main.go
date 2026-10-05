package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/cli"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/modules"
	datatypeicons "github.com/LaPingvino/recuerdo/internal/modules/data/dataTypeIcons"
	openteacherauthors "github.com/LaPingvino/recuerdo/internal/modules/data/openteacherAuthors"
	userdocumentation "github.com/LaPingvino/recuerdo/internal/modules/data/userDocumentation"
	resultsdialog "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/results"
	wordslesson "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words"
	"github.com/LaPingvino/recuerdo/internal/version"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	// ALL Qt imports temporarily disabled to get core system working first
	// TODO: Re-enable Qt modules incrementally once basic system is validated

	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/about"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/file"
	settingsDialog "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/settings"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessonDialogs"

	// "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/documentation" // Disabled due to build constraints
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/gui"
	qtapp "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/qtApp"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/settings"

	logicevent "github.com/LaPingvino/recuerdo/internal/modules/logic/event"

	allonce "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/allOnce"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/interval"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/smart"
	hardwords "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/hardWords"
	random "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/random_"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/reverse"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/sort"
	wordsneveransweredcorrectly "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/wordsNeverAnsweredCorrectly"
	mimicrytypefaceconverter "github.com/LaPingvino/recuerdo/internal/modules/logic/mimicryTypefaceConverter"
	notecalculatorchooser "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculatorChooser"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/american"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/dutch"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/ects"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/french"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/german"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/percents"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
	recentlyopened "github.com/LaPingvino/recuerdo/internal/modules/logic/recentlyOpened"

	reversermedia "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/media"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/words"
	// Removed duplicate auto-converted modules - using manually implemented versions instead
)

const appName = "Recuerdo"

// The GUI passes the settings module to word lessons as wordslesson.Settings.
var _ wordslesson.Settings = (*modules.SettingsModule)(nil)

// appVersion is set by release builds with -ldflags "-X main.appVersion=...".
var appVersion = "4.0.0-dev" // after OpenTeacher 3.3

// Command-line arguments
var (
	commands         = flag.String("commands", "", "Comma-separated list of commands to execute (e.g., 'show-properties,show-settings')")
	listCmds         = flag.Bool("list-commands", false, "List available commands and exit")
	helpFlag         = flag.Bool("help", false, "Show help message")
	strictValidation = flag.Bool("strict-validation", false, "Enable strict UI layout validation (fail on overlaps)")
)

func main() {
	// recuerdo <command> ...: the command line, without the GUI
	if len(os.Args) > 1 && cli.IsCommand(os.Args[1]) {
		os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}

	// Parse command-line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s %s - Language Learning Application\n\n", appName, appVersion)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options] [lesson-file]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s                              # Start normally\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s lesson.ot                    # Load lesson file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --commands=show-properties   # Execute command\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s lesson.ot --commands=show-properties  # Load file and show properties\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands (recuerdo help lists them with their arguments):\n")
		fmt.Fprintf(os.Stderr, "  authors, convert, merge, reverse-list, view-word-list, ocr-word-list,\n")
		fmt.Fprintf(os.Stderr, "  new-word-list, practise-word-list\n")
	}

	flag.Parse()

	// Handle help flag first
	if *helpFlag {
		flag.Usage()
		return
	}

	// Handle list commands
	if *listCmds {
		listAvailableCommands()
		return
	}

	// Get lesson file from positional argument
	var lessonFile string
	if flag.NArg() > 0 {
		lessonFile = flag.Arg(0)
	}

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	version.Version = appVersion
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
		// a second signal ends the process even if shutdown hangs
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		cancel()
	}()

	// Enable all modules
	fmt.Println("Enabling modules...")
	// Set strict validation mode globally if requested
	if *strictValidation {
		os.Setenv("RECUERDO_STRICT_LAYOUT", "1")
		log.Println("Strict layout validation enabled")
	}

	if err := manager.EnableAll(ctx); err != nil {
		log.Fatalf("Failed to enable modules: %v", err)
	}
	fmt.Println("All modules enabled successfully")

	// Start the main application
	if err := runApplication(ctx, manager, lessonFile, *commands); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	// Shutdown
	fmt.Println("Disabling modules...")
	if err := manager.DisableAll(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Recuerdo shutdown complete")
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

	buttonRegisterModule := modules.NewButtonRegisterModule()
	if err := manager.Register(buttonRegisterModule); err != nil {
		return fmt.Errorf("failed to register button register module: %w", err)
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

	// Register media module - DISABLED for now
	// mediaModule := media.NewMediaEntererModule()
	// if err := manager.Register(mediaModule); err != nil {
	// 	return fmt.Errorf("failed to register media module: %w", err)
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

	// Skip main module - it's a program not a library
	// mainModule := main.NewTypingTutorModule()
	// if err := manager.Register(mainModule); err != nil {
	// 	return fmt.Errorf("failed to register main module: %w", err)
	// }

	// Register datatypeicons module
	datatypeiconsModule := datatypeicons.NewDataTypeIconsModule()
	if err := manager.Register(datatypeiconsModule); err != nil {
		return fmt.Errorf("failed to register datatypeicons module: %w", err)
	}

	// Register metadata module
	metadataModule := modules.NewMetadataModule()
	if err := manager.Register(metadataModule); err != nil {
		return fmt.Errorf("failed to register metadata module: %w", err)
	}

	// Register openteacherauthors module
	openteacherauthorsModule := openteacherauthors.NewOpenTeacherAuthorsModule()
	if err := manager.Register(openteacherauthorsModule); err != nil {
		return fmt.Errorf("failed to register openteacherauthors module: %w", err)
	}

	// Register userdocumentation module
	userdocumentationModule := userdocumentation.NewUserDocumentationModule()
	if err := manager.Register(userdocumentationModule); err != nil {
		return fmt.Errorf("failed to register userdocumentation module: %w", err)
	}

	// Register event module
	logiceventModule := logicevent.NewEventModule()
	if err := manager.Register(logiceventModule); err != nil {
		return fmt.Errorf("failed to register event module: %w", err)
	}
	fmt.Printf("  ✓ Registered event module\n")

	fmt.Printf("  ✓ Registered execute module\n")

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

	// Register interval module
	intervalModule := interval.NewIntervalModule()
	if err := manager.Register(intervalModule); err != nil {
		return fmt.Errorf("failed to register interval module: %w", err)
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

	// Register mimicrytypefaceconverter module
	mimicrytypefaceconverterModule := mimicrytypefaceconverter.NewMimicryTypefaceConverterModule()
	if err := manager.Register(mimicrytypefaceconverterModule); err != nil {
		return fmt.Errorf("failed to register mimicrytypefaceconverter module: %w", err)
	}

	// Skip modulestestFiletoimport - merged into modulestest package
	// modulestestfiletoimportModule := modulestestFiletoimport.NewModulestestfiletoimportModule()
	// if err := manager.Register(modulestestfiletoimportModule); err != nil {
	// 	return fmt.Errorf("failed to register modulestestFiletoimport module: %w", err)
	// }

	// Register results dialog module
	if err := manager.Register(resultsdialog.NewResultsDialogModule()); err != nil {
		return fmt.Errorf("failed to register results dialog module: %w", err)
	}

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

	// Register percents module
	percentsModule := percents.NewPercentsNoteCalculatorModule()
	if err := manager.Register(percentsModule); err != nil {
		return fmt.Errorf("failed to register percents module: %w", err)
	}

	// Register percentscalculator module
	percentscalculatorModule := percentscalculator.NewPercentsCalculatorModule()
	if err := manager.Register(percentscalculatorModule); err != nil {
		return fmt.Errorf("failed to register percentscalculator module: %w", err)
	}

	// Register recentlyopened module
	recentlyopenedModule := recentlyopened.NewRecentlyOpenedModule()
	if err := manager.Register(recentlyopenedModule); err != nil {
		return fmt.Errorf("failed to register recentlyopened module: %w", err)
	}

	// Register the media reverser
	reversermediaModule := reversermedia.NewMediaReverserModule()
	if err := manager.Register(reversermediaModule); err != nil {
		return fmt.Errorf("failed to register media reverser module: %w", err)
	}

	// Register words module
	wordsModule := words.NewWordsReverserModule()
	if err := manager.Register(wordsModule); err != nil {
		return fmt.Errorf("failed to register words module: %w", err)
	}

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

	// Temporarily disable remaining modules that are causing conflicts
	// TODO: Re-enable these modules once import conflicts are resolved

	// Register essential interface modules required for dependencies

	fmt.Printf("  ✓ Registered buttonRegister module\n")

	fmt.Printf("  ✓ Registered lessontracker module\n")

	fmt.Println("✅ Core modules successfully registered!")
	fmt.Println("Note: Many optional modules are disabled to resolve import conflicts")

	return nil
}

func runApplication(ctx context.Context, manager *core.Manager, lessonFile, commands string) error {
	// Get the GUI module and show the main window
	guiModule, exists := manager.GetDefaultModule("ui")
	if exists {
		fmt.Println("Starting Qt GUI...")

		// Show the main window
		if guiMod, ok := guiModule.(interface{ ShowMainWindow() }); ok {
			guiMod.ShowMainWindow()

			// Load lesson file if specified
			if lessonFile != "" {
				fmt.Printf("Loading lesson file: %s\n", lessonFile)
				if err := loadLessonFile(manager, lessonFile); err != nil {
					fmt.Printf("Error loading lesson file: %v\n", err)
				}
			}

			// Execute commands if specified
			if commands != "" {
				fmt.Printf("Executing commands: %s\n", commands)
				go func() {
					// Wait a bit for GUI to fully initialize
					time.Sleep(1 * time.Second)
					executeCommands(manager, commands)
				}()
			}
		} else {
			fmt.Println("Warning: GUI module does not support ShowMainWindow()")
		}

		// Run the Qt event loop on the main thread (Qt requirement)
		if guiMod, ok := guiModule.(interface{ RunEventLoop() int }); ok {
			fmt.Println("Starting Qt event loop...")

			// Qt event loop must run on main thread - call directly.
			// A cancelled context (SIGINT/SIGTERM) asks Qt to leave it.
			loopDone := make(chan struct{})
			go func() {
				select {
				case <-ctx.Done():
					mainthread.Start(qt.QCoreApplication_Quit)
				case <-loopDone:
				}
			}()
			exitCode := guiMod.RunEventLoop()
			close(loopDone)
			fmt.Printf("Qt event loop finished with exit code: %d\n", exitCode)
			return nil
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

// listAvailableCommands lists all available commands
func listAvailableCommands() {
	fmt.Println("Available Commands:")
	fmt.Println("  show-properties   - Show lesson properties dialog")
	fmt.Println("  show-settings     - Show application settings dialog")
	fmt.Println("  show-about        - Show about dialog")
	fmt.Println("  new-lesson        - Create a new lesson")
	fmt.Println("  open-file         - Show open file dialog")
	fmt.Println("  exit              - Exit the application")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Println("  ./recuerdo sample.ot --commands=show-properties")
	fmt.Println("  ./recuerdo --commands=new-lesson,show-settings")
	fmt.Println("  ./recuerdo --list-commands")
}

// loadLessonFile loads a lesson file using the GUI module
func loadLessonFile(manager *core.Manager, lessonFile string) error {
	// Check if file exists
	if _, err := os.Stat(lessonFile); os.IsNotExist(err) {
		return fmt.Errorf("lesson file does not exist: %s", lessonFile)
	}

	// Get the GUI module
	guiModule, exists := manager.GetDefaultModule("ui")
	if !exists {
		return fmt.Errorf("no GUI module found")
	}

	// Try to load the file using the GUI module's LoadSelectedFile method
	if guiMod, ok := guiModule.(interface{ LoadSelectedFile(string) error }); ok {
		return guiMod.LoadSelectedFile(lessonFile)
	}

	return fmt.Errorf("GUI module does not support file loading")
}

// executeCommands executes a comma-separated list of commands
func executeCommands(manager *core.Manager, commands string) {
	if commands == "" {
		return
	}

	// Simple command parsing - split by comma
	cmdList := strings.Split(commands, ",")

	// Execute each command
	for _, cmd := range cmdList {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		fmt.Printf("Executing command: %s\n", cmd)
		if err := executeCommand(manager, cmd); err != nil {
			fmt.Printf("Error executing command '%s': %v\n", cmd, err)
		} else {
			fmt.Printf("Command '%s' executed successfully\n", cmd)
		}

		// Wait between commands to allow GUI updates
		time.Sleep(500 * time.Millisecond)
	}
}

// executeCommand executes a single command
func executeCommand(manager *core.Manager, command string) error {
	// Get the GUI module
	guiModule, exists := manager.GetDefaultModule("ui")
	if !exists {
		return fmt.Errorf("no GUI module found")
	}

	switch command {
	case "show-properties":
		if guiMod, ok := guiModule.(interface{ ShowPropertiesDialog() }); ok {
			guiMod.ShowPropertiesDialog()
			return nil
		}
		return fmt.Errorf("GUI module does not support ShowPropertiesDialog")

	case "show-settings":
		if guiMod, ok := guiModule.(interface{ ShowSettingsDialog() }); ok {
			guiMod.ShowSettingsDialog()
			return nil
		}
		return fmt.Errorf("GUI module does not support ShowSettingsDialog")

	case "show-about":
		if guiMod, ok := guiModule.(interface{ ShowAboutDialog() }); ok {
			guiMod.ShowAboutDialog()
			return nil
		}
		return fmt.Errorf("GUI module does not support ShowAboutDialog")

	case "new-lesson":
		if guiMod, ok := guiModule.(interface{ ShowNewLessonDialog() }); ok {
			guiMod.ShowNewLessonDialog()
			return nil
		}
		return fmt.Errorf("GUI module does not support ShowNewLessonDialog")

	case "open-file":
		if guiMod, ok := guiModule.(interface{ ShowOpenDialog() }); ok {
			guiMod.ShowOpenDialog()
			return nil
		}
		return fmt.Errorf("GUI module does not support ShowOpenDialog")

	case "exit":
		if guiMod, ok := guiModule.(interface{ Exit() }); ok {
			guiMod.Exit()
			return nil
		}
		return fmt.Errorf("GUI module does not support Exit")

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
