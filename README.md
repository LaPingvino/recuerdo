# OpenTeacher Go

A modern rewrite of OpenTeacher in Go, maintaining the original modular architecture while adding type safety and improved performance.

## Overview

OpenTeacher is an educational application for creating and managing language learning exercises. This Go version preserves the plugin-based architecture of the original Python application while modernizing the codebase.

## Architecture

### Modular Design
- **Interface-based modules**: Each module implements well-defined Go interfaces
- **Dependency injection**: Clean separation between interfaces and implementations  
- **Event-driven communication**: Observer pattern for inter-module communication
- **Resource management**: Modules can access local resources and configurations

### Core Components
- **Module Manager**: Handles registration, dependency resolution, and lifecycle management
- **Event System**: Thread-safe event handling with subscribe/unsubscribe patterns
- **Settings System**: JSON-based configuration with type-safe accessors
- **Execute System**: Application lifecycle and profile management

## Project Status

**Current Version**: 4.0.0-alpha  
**Phase**: Core Infrastructure Complete ✅

### Completed Features
- ✅ Complete module system with dependency resolution
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

## Migration from Python

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

OpenTeacher is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

## Documentation

- [Conversion Approach](CONVERSION_APPROACH.md) - Detailed migration strategy
- [Conversion Log](CONVERSION_LOG.md) - Development progress and decisions
- [Legacy Python Code](legacy/) - Original implementation for reference