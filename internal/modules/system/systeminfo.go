package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// SystemInfoModule provides system diagnostics at startup
type SystemInfoModule struct {
	*core.BaseModule
	info *SystemInfo
}

// SystemInfo holds information about the current display system
type SystemInfo struct {
	SessionType    string
	DisplayServer  string
	QtBackend      string
	WaylandDisplay string
	X11Display     string
	Desktop        string
	Compositor     string
	InputMethod    string
	KeyboardLayout string
	IsXWayland     bool
}

// NewSystemInfoModule creates a new system info module
func NewSystemInfoModule() *SystemInfoModule {
	base := core.NewBaseModule("system", "systeminfo")

	return &SystemInfoModule{
		BaseModule: base,
	}
}

// Enable activates the module and prints system information
func (mod *SystemInfoModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Detect and print system information
	mod.info = DetectSystem()
	mod.PrintStartupInfo()

	return nil
}

// Disable deactivates the module
func (mod *SystemInfoModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// DetectSystem analyzes the current system configuration
func DetectSystem() *SystemInfo {
	info := &SystemInfo{}

	// Basic environment detection
	info.SessionType = getEnvOrDefault("XDG_SESSION_TYPE", "unknown")
	info.WaylandDisplay = getEnvOrDefault("WAYLAND_DISPLAY", "")
	info.X11Display = getEnvOrDefault("DISPLAY", "")
	info.Desktop = getEnvOrDefault("XDG_CURRENT_DESKTOP", "unknown")
	info.InputMethod = getEnvOrDefault("QT_IM_MODULE", "")

	// Determine display server
	info.DisplayServer = determineDisplayServer(info)

	// Detect Qt backend
	info.QtBackend = detectQtBackend()

	// Detect compositor (Wayland only)
	if info.SessionType == "wayland" {
		info.Compositor = detectWaylandCompositor()
	}

	// Check if running under XWayland
	info.IsXWayland = info.SessionType == "wayland" && info.X11Display != ""

	// Get keyboard layout
	info.KeyboardLayout = detectKeyboardLayout()

	return info
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func determineDisplayServer(info *SystemInfo) string {
	if info.SessionType == "wayland" {
		if info.IsXWayland {
			return "wayland+xwayland"
		}
		return "wayland"
	} else if info.X11Display != "" {
		return "x11"
	}
	return "unknown"
}

func detectQtBackend() string {
	// Check explicit Qt platform setting
	if platform := os.Getenv("QT_QPA_PLATFORM"); platform != "" {
		return platform
	}

	// Fallback to educated guess based on session
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType == "wayland" {
		if os.Getenv("DISPLAY") != "" {
			return "xcb (auto-detected via XWayland)"
		}
		return "wayland (auto-detected)"
	}
	return "xcb (auto-detected)"
}

func detectWaylandCompositor() string {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")

	switch strings.ToLower(desktop) {
	case "gnome":
		return "gnome-shell (mutter)"
	case "kde":
		return "kwin"
	case "sway":
		return "sway"
	case "weston":
		return "weston"
	default:
		if session := os.Getenv("XDG_SESSION_DESKTOP"); session != "" {
			return fmt.Sprintf("unknown (%s)", session)
		}
		return "unknown"
	}
}

func detectKeyboardLayout() string {
	cmd := exec.Command("setxkbmap", "-query")
	output, err := cmd.Output()
	if err != nil {
		return "unknown (setxkbmap failed)"
	}

	lines := strings.Split(string(output), "\n")
	layout := "unknown"
	variant := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "layout:") {
			layout = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.HasPrefix(line, "variant:") {
			variant = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	if variant != "" {
		return fmt.Sprintf("%s (%s)", layout, variant)
	}
	return layout
}

// PrintStartupInfo prints system information at application startup
func (mod *SystemInfoModule) PrintStartupInfo() {
	info := mod.info

	fmt.Println("🚀 Recuerdo Starting...")
	fmt.Println("========================")
	fmt.Printf("Session:      %s\n", info.SessionType)
	fmt.Printf("Display:      %s\n", info.DisplayServer)
	fmt.Printf("Qt Backend:   %s\n", info.QtBackend)
	fmt.Printf("Desktop:      %s\n", info.Desktop)
	fmt.Printf("Keyboard:     %s\n", info.KeyboardLayout)
	fmt.Printf("Input Method: %s\n", getInputMethodStatus(info.InputMethod))

	if info.IsXWayland {
		fmt.Printf("XWayland:     ✓ Active\n")
	}

	// Show potential issues
	issues := mod.getCompatibilityIssues()
	if len(issues) > 0 {
		fmt.Println("\n⚠️  Potential Input Issues:")
		for _, issue := range issues {
			fmt.Printf("   • %s\n", issue)
		}
		fmt.Println("💡 Use Unicode character picker if keyboard input fails")
	} else {
		fmt.Println("✅ Configuration looks good for keyboard input")
	}

	// Show testing commands
	fmt.Println("\n🧪 Testing Different Backends:")
	commands := mod.getTestingCommands()
	for _, cmd := range commands {
		fmt.Printf("   %s\n", cmd)
	}

	fmt.Println("")
}

// getCompatibilityIssues returns a list of potential compatibility issues
func (mod *SystemInfoModule) getCompatibilityIssues() []string {
	var issues []string
	info := mod.info

	if info.IsXWayland && info.QtBackend != "wayland" {
		issues = append(issues, "Qt using XWayland backend - may cause input issues")
	}

	if info.InputMethod == "" {
		issues = append(issues, "No input method configured")
	}

	if info.SessionType == "wayland" && strings.Contains(info.QtBackend, "xcb") {
		issues = append(issues, "Qt using X11 backend on Wayland session")
	}

	return issues
}

func getInputMethodStatus(inputMethod string) string {
	if inputMethod == "" {
		return "none"
	}
	return inputMethod
}

// getTestingCommands returns commands for testing different configurations
func (mod *SystemInfoModule) getTestingCommands() []string {
	info := mod.info
	commands := []string{}

	if info.SessionType == "wayland" {
		commands = append(commands, "QT_QPA_PLATFORM=wayland go run your_app.go  # Force Wayland")
		commands = append(commands, "QT_QPA_PLATFORM=xcb go run your_app.go      # Force X11")
	} else {
		commands = append(commands, "QT_QPA_PLATFORM=xcb go run your_app.go      # Force X11")
	}

	commands = append(commands, "QT_IM_MODULE=ibus go run your_app.go         # Test with IBus")
	commands = append(commands, "QT_IM_MODULE= go run your_app.go             # Test without IM")

	return commands
}

// GetSystemInfo returns the detected system information
func (mod *SystemInfoModule) GetSystemInfo() *SystemInfo {
	return mod.info
}

// InitSystemInfoModule creates and returns a new SystemInfoModule instance
func InitSystemInfoModule() core.Module {
	return NewSystemInfoModule()
}
