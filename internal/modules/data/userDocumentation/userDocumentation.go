// Package userdocumentation finds Recuerdo's user documentation: a
// getting started guide in data/documentation (with screenshots), after
// OpenTeacher's quick start (data/userDocumentation).
package userdocumentation

import (
	"os"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/resources"
)

// Dir is the directory with the documentation and its pictures.
func Dir() string { return filepath.Join(resources.Dir(), "data", "documentation") }

// GettingStarted returns the getting started guide as an HTML fragment;
// its pictures are relative to Dir.
func GettingStarted() (string, error) {
	b, err := os.ReadFile(filepath.Join(Dir(), "getting-started.html"))
	return string(b), err
}

// UserDocumentationModule offers the documentation as an OpenTeacher
// "userDocumentation" module.
type UserDocumentationModule struct {
	*core.BaseModule
}

// NewUserDocumentationModule creates the module.
func NewUserDocumentationModule() *UserDocumentationModule {
	return &UserDocumentationModule{BaseModule: core.NewBaseModule("data", "userdocumentation-module")}
}

// GetHTML returns the getting started guide.
func (mod *UserDocumentationModule) GetHTML() (string, error) { return GettingStarted() }

// InitUserDocumentationModule creates the module.
func InitUserDocumentationModule() core.Module { return NewUserDocumentationModule() }
