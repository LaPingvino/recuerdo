# Recuerdo

A modern language learning application in Go, evolved from OpenTeacher. Recuerdo features a clean modular architecture with comprehensive system detection and robust Unicode input handling.

## Overview

Recuerdo is an educational application for creating and managing language learning exercises. This Go version modernizes the original Python OpenTeacher codebase while maintaining its powerful plugin-based architecture and adding enhanced cross-platform support.

## Quick Start

```bash
# Build and run the test application
go build ./cmd/test-core
./test-core
```

The system will automatically detect your display server configuration and provide diagnostics for potential input issues.

## Key Features

- **Smart System Detection**: Automatic detection of X11/Wayland, Qt backends, and input methods
- **Unicode Character Picker**: Fallback input system for international keyboards  
- **Modular Architecture**: Clean plugin-based design with hot-swappable modules
- **Cross-Platform Support**: Native integration with GNOME, KDE, and other Linux desktops
- **Input Diagnostics**: Comprehensive troubleshooting for keyboard input issues

## Architecture

### Directory Structure
```
recuerdo/
├── cmd/           # Entry points and applications
├── internal/      # Core application code
│   ├── core/      # Module system and lifecycle management
│   ├── modules/   # Feature modules
│   └── system/    # System detection and utilities
├── data/          # Application data files
├── docs/          # Comprehensive documentation
├── legacy/        # Original Python codebase (reference)
└── scripts/       # Build and development scripts
```

### Core Components
- **Module Manager**: Registration, dependency resolution, and lifecycle management
- **System Detection**: Automatic environment detection and diagnostics
- **Event System**: Thread-safe inter-module communication
- **Settings System**: Type-safe configuration management

## Documentation

- **[Complete Documentation](./docs/)** - Comprehensive project documentation
- **[Module Documentation](./docs/modules/)** - Specific module implementations
- **[Conversion History](./docs/conversion/)** - Python-to-Go conversion documentation

## Project Status

**Current Version**: 4.0.0-alpha  
**Status**: Core Infrastructure Complete ✅

### Completed Features
- ✅ Complete modular system with dependency resolution
- ✅ System detection and Qt backend diagnostics
- ✅ Unicode character picker for input fallback
- ✅ Thread-safe event handling system
- ✅ Settings persistence and management
- ✅ Application lifecycle management
- ✅ Comprehensive test suite (36 test cases)
- ✅ Graceful shutdown handling

### In Progress
- 🔄 Qt GUI framework integration
- 🔄 File format loaders/savers
- 🔄 Lesson management system

### Planned Features
- 📋 Complete GUI with Qt5/Qt6 bindings
- 📋 Import/export for various file formats
- 📋 Teaching modes and exercise types
- 📋 Progress tracking and statistics
- 📋 Multi-language support

## Building and Running

### Prerequisites
- Go 1.21 or later
- Qt5/Qt6 development libraries (for future GUI features)

### Build
```bash
go build ./cmd/openteacher
```

### Run
```bash
./openteacher
```

### Test
```bash
go test ./...
```

## Development

### Project Structure
```
├── cmd/openteacher/          # Main application entry point
├── internal/
│   ├── core/                 # Core module interfaces and manager
│   └── modules/              # Module implementations
├── legacy/                   # Original Python code (reference)
├── pkg/                      # Public APIs (future)
└── docs/                     # Documentation
```

### Key Interfaces

**Module Interface**
```go
type Module interface {
    Type() string
    Name() string
    Requires() []string
    Uses() []string
    Priority() int
    Enable(ctx context.Context) error
    Disable(ctx context.Context) error
    IsActive() bool
}
```

**Event System**
```go
type EventModule interface {
    Module
    CreateEvent(name string) Event
    Subscribe(eventName string, handler EventHandler) error
    Unsubscribe(eventName string, handler EventHandler) error
}
```

### Adding New Modules

1. Implement the `core.Module` interface
2. Add any specialized interfaces (e.g., `core.ExecuteModule`)
3. Register the module in `cmd/openteacher/main.go`
4. Write comprehensive tests

Example:
```go
type MyModule struct {
    *core.BaseModule
}

func NewMyModule() *MyModule {
    base := core.NewBaseModule("my-type", "my-module")
    base.SetRequires("event", "settings")
    return &MyModule{BaseModule: base}
}

func (m *MyModule) Enable(ctx context.Context) error {
    // Module initialization
    return m.BaseModule.Enable(ctx)
}
```

## Testing Philosophy

- **Test-first development**: Write tests before implementation
- **Interface testing**: Test behavior through interfaces
- **Concurrent safety**: All tests verify thread safety
- **Error conditions**: Comprehensive error case coverage

## Based on OpenTeacher

This project maintains architectural compatibility with the original Python OpenTeacher:

- **Module types**: Same module type system (`execute`, `event`, `settings`, etc.)
- **Dependencies**: Same dependency relationships between modules  
- **Resource handling**: Compatible resource path management
- **Configuration**: Settings use similar key-value structure

## Contributing

1. Follow Go conventions and formatting (`go fmt`, `go vet`)
2. Write tests for all new functionality
3. Update documentation for interface changes
4. Ensure thread safety for concurrent operations
5. Follow the existing error handling patterns

## License

Recuerdo is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

Copyright © 2025 Joop Kiefte
Based on OpenTeacher © 2010-2023 OpenTeacher Team

## Documentation

- [Conversion Approach](CONVERSION_APPROACH.md) - Detailed migration strategy
- [Conversion Log](CONVERSION_LOG.md) - Development progress and decisions
- [Legacy Python Code](legacy/) - Original implementation for reference