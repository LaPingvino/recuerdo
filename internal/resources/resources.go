// Package resources finds Recuerdo's data files (data/maps,
// data/character_sets.json) wherever Recuerdo is installed.
package resources

import (
	"os"
	"path/filepath"
)

// marker is a file every resource directory has.
const marker = "data/maps"

// Dir is the directory holding data/: $RECUERDO_DATA if set,
// else the first of these that has them: next to the executable (Windows
// install folder, portable zip), ../share/recuerdo (Linux packages and the
// AppImage), ../Resources (macOS app bundle), the working directory
// (running from the source tree).
func Dir() string {
	if d := os.Getenv("RECUERDO_DATA"); d != "" {
		return d
	}
	for _, d := range candidates() {
		if st, err := os.Stat(filepath.Join(d, marker)); err == nil && st.IsDir() {
			return d
		}
	}
	return "."
}

func candidates() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		bin := filepath.Dir(exe)
		dirs = append(dirs,
			bin,
			filepath.Join(bin, "..", "share", "recuerdo"),
			filepath.Join(bin, "..", "Resources"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}
	return dirs
}
