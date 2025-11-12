package logging

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	SUCCESS
	WARNING
	ERROR
	ACTION
	EVENT
	STUB
	LEGACY_REMINDER
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case SUCCESS:
		return "SUCCESS"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case ACTION:
		return "ACTION"
	case EVENT:
		return "EVENT"
	case STUB:
		return "STUB"
	case LEGACY_REMINDER:
		return "LEGACY"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for the log level
func (l LogLevel) Color() string {
	switch l {
	case DEBUG:
		return "\033[90m" // Dark gray
	case INFO:
		return "\033[36m" // Cyan
	case SUCCESS:
		return "\033[32m" // Green
	case WARNING:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	case ACTION:
		return "\033[35m" // Magenta
	case EVENT:
		return "\033[34m" // Blue
	case STUB:
		return "\033[93m" // Bright yellow
	case LEGACY_REMINDER:
		return "\033[96m" // Bright cyan
	default:
		return "\033[0m" // Reset
	}
}

// Logger is the main logging interface
type Logger struct {
	module      string
	enableColor bool
	minLevel    LogLevel
}

// Global logger instance
var globalLogger *Logger

// init initializes the global logger
func init() {
	globalLogger = NewLogger("OpenTeacher")
}

// NewLogger creates a new logger instance for a specific module
func NewLogger(module string) *Logger {
	return &Logger{
		module:      module,
		enableColor: true, // Enable by default, can be disabled for files
		minLevel:    DEBUG,
	}
}

// SetMinLevel sets the minimum log level to display
func (l *Logger) SetMinLevel(level LogLevel) {
	l.minLevel = level
}

// SetColorEnabled enables or disables colored output
func (l *Logger) SetColorEnabled(enabled bool) {
	l.enableColor = enabled
}

// formatMessage formats a log message with timestamp, level, module, and caller info
func (l *Logger) formatMessage(level LogLevel, message string) string {
	now := time.Now().Format("2006/01/02 15:04:05")

	// Get caller information
	_, file, line, ok := runtime.Caller(3)
	var caller string
	if ok {
		filename := filepath.Base(file)
		caller = fmt.Sprintf("%s:%d", filename, line)
	} else {
		caller = "unknown:0"
	}

	levelStr := level.String()
	if l.enableColor {
		levelStr = level.Color() + "[" + level.String() + "]" + "\033[0m"
	} else {
		levelStr = "[" + level.String() + "]"
	}

	return fmt.Sprintf("%s %s %s.%s - %s",
		now, levelStr, l.module, caller, message)
}

// log is the core logging function
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	message := fmt.Sprintf(format, args...)
	formatted := l.formatMessage(level, message)

	log.Println(formatted)

	// For legacy reminders, also suggest checking Python code
	if level == LEGACY_REMINDER {
		suggestion := l.formatMessage(INFO, "💡 SUGGESTION: Check legacy/modules/org/openteacher/ for Python implementation")
		log.Println(suggestion)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	l.log(SUCCESS, format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(WARNING, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Action logs an action being performed
func (l *Logger) Action(format string, args ...interface{}) {
	l.log(ACTION, format, args...)
}

// Event logs an event occurrence
func (l *Logger) Event(format string, args ...interface{}) {
	l.log(EVENT, format, args...)
}

// Stub logs when hitting a stub/unimplemented function
func (l *Logger) Stub(functionName, suggestedLegacyPath string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf("STUB: %s() - %s", functionName, message)
	l.log(STUB, fullMessage)

	if suggestedLegacyPath != "" {
		l.log(LEGACY_REMINDER, "Check implementation in: %s", suggestedLegacyPath)
	}
}

// LegacyReminder logs a reminder to check legacy Python code
func (l *Logger) LegacyReminder(component, pythonPath, reason string) {
	l.log(LEGACY_REMINDER, "%s needs implementation - check %s (%s)",
		component, pythonPath, reason)
}

// DeadEnd logs when reaching a dead end that needs legacy code consultation
func (l *Logger) DeadEnd(component, issue, suggestedPythonPath string) {
	l.log(ERROR, "DEAD END: %s - %s", component, issue)
	l.log(LEGACY_REMINDER, "⚠️  IMPLEMENTATION NEEDED: Check %s for Python reference", suggestedPythonPath)
}

// Module creates a new logger instance for a specific module
func (l *Logger) Module(moduleName string) *Logger {
	return NewLogger(moduleName)
}

// Global logging functions that use the global logger

// Debug logs a debug message using global logger
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

// Info logs an info message using global logger
func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

// Success logs a success message using global logger
func Success(format string, args ...interface{}) {
	globalLogger.Success(format, args...)
}

// Warning logs a warning message using global logger
func Warning(format string, args ...interface{}) {
	globalLogger.Warning(format, args...)
}

// Error logs an error message using global logger
func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

// Action logs an action using global logger
func Action(format string, args ...interface{}) {
	globalLogger.Action(format, args...)
}

// Event logs an event using global logger
func Event(format string, args ...interface{}) {
	globalLogger.Event(format, args...)
}

// Stub logs a stub function using global logger
func Stub(functionName, suggestedLegacyPath string, format string, args ...interface{}) {
	globalLogger.Stub(functionName, suggestedLegacyPath, format, args...)
}

// LegacyReminder logs a legacy reminder using global logger
func LegacyReminder(component, pythonPath, reason string) {
	globalLogger.LegacyReminder(component, pythonPath, reason)
}

// DeadEnd logs a dead end using global logger
func DeadEnd(component, issue, suggestedPythonPath string) {
	globalLogger.DeadEnd(component, issue, suggestedPythonPath)
}

// GetModuleLogger returns a logger instance for a specific module
func GetModuleLogger(moduleName string) *Logger {
	return NewLogger(moduleName)
}

// SetGlobalMinLevel sets the minimum log level for the global logger
func SetGlobalMinLevel(level LogLevel) {
	globalLogger.SetMinLevel(level)
}

// SetGlobalColorEnabled sets color output for the global logger
func SetGlobalColorEnabled(enabled bool) {
	globalLogger.SetColorEnabled(enabled)
}

// Helper functions for common patterns

// FileOperation logs file operations with context
func FileOperation(operation, filePath string, err error) {
	if err != nil {
		Error("File %s failed for %s: %v", operation, filePath, err)
	} else {
		Success("File %s successful for %s", operation, filePath)
	}
}

// ModuleLifecycle logs module enable/disable events
func ModuleLifecycle(moduleName, action string, err error) {
	if err != nil {
		Error("Module %s %s failed: %v", moduleName, action, err)
	} else {
		Success("Module %s %s successful", moduleName, action)
	}
}

// GuiEvent logs GUI events with user interaction context
func GuiEvent(eventType, component string, details ...interface{}) {
	if len(details) > 0 {
		Event("%s in %s: %v", eventType, component, details[0])
	} else {
		Event("%s in %s", eventType, component)
	}
}

// Performance logs performance-related messages
func Performance(operation string, duration time.Duration, details string) {
	if duration > time.Second {
		Warning("SLOW: %s took %v - %s", operation, duration, details)
	} else {
		Debug("Performance: %s took %v - %s", operation, duration, details)
	}
}

// TODO logs TODO items with priority
func TODO(priority string, description, legacyReference string) {
	Warning("TODO [%s]: %s", priority, description)
	if legacyReference != "" {
		LegacyReminder("TODO item", legacyReference, "implementation needed")
	}
}
