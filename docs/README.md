# Recuerdo Documentation

Welcome to the Recuerdo project documentation. This directory contains comprehensive information about the project's architecture, conversion process, and module system.

## Overview

Recuerdo is a modern language learning application written in Go, evolved from the OpenTeacher Python project. It features a modular architecture with Qt-based user interfaces and comprehensive system detection capabilities.

## Documentation Structure

### Conversion Documentation
- [`conversion/`](./conversion/) - Complete documentation of the Python-to-Go conversion process
  - [`automated-summary.md`](./conversion/automated-summary.md) - Statistics and overview of automated conversion
  - [`complete-summary.md`](./conversion/complete-summary.md) - Final conversion results and achievements
  - [`approach.md`](./conversion/approach.md) - Technical approach and methodology
  - [`log.md`](./conversion/log.md) - Detailed conversion log
  - [`log.json`](./conversion/log.json) - Machine-readable conversion data

### Module Documentation
- [`modules/`](./modules/) - Documentation for specific modules
  - [`business-card.md`](./modules/business-card.md) - Business card generation module

## Key Features

- **System Detection**: Automatic detection of display servers (X11/Wayland), Qt backends, and input methods
- **Unicode Character Picker**: Fallback input system for international characters
- **Modular Architecture**: Plugin-based system with hot-swappable modules
- **Cross-Platform**: Support for Linux desktop environments (GNOME, KDE, etc.)

## Getting Started

1. Build the test application:
   ```bash
   go build ./cmd/test-core
   ```

2. Run the system detection:
   ```bash
   ./test-core
   ```

## Architecture

Recuerdo follows a clean modular architecture:

- **Core System** (`internal/core/`) - Module management and lifecycle
- **Modules** (`internal/modules/`) - Feature implementations
  - `system/` - System detection and diagnostics
  - `interfaces/qt/` - Qt-based user interfaces  
  - `logic/` - Business logic modules
  - `data/` - Data handling modules
- **Legacy** (`legacy/`) - Original Python codebase for reference

## System Requirements

- Go 1.19 or later
- Qt5/Qt6 development libraries
- Linux with X11 or Wayland support

## Contributing

When adding new modules or features:

1. Follow the existing module structure in `internal/modules/`
2. Implement the `core.Module` interface
3. Add proper registration in the module registry
4. Include comprehensive system detection if UI-related
5. Update documentation accordingly

## Support

For keyboard input issues on Wayland systems, Recuerdo includes:
- Automatic backend detection
- Input method diagnostics  
- Unicode character picker as fallback
- System-specific troubleshooting guidance

## License

This project maintains compatibility with the original OpenTeacher licensing.