# Business Card Modules - Python to Go Port

This document describes the business card generation modules that have been ported from Python to Go.

## Overview

The business card generation functionality allows Recuerdo to create promotional "business card" images that can be used to promote the software. This feature was originally implemented in Python using PyQt5 for graphics rendering in OpenTeacher and has been ported to Go using the standard `image` package.

## Ported Modules

### 1. BusinessCardGenerator Module

**Location**: `internal/modules/profilerunners/businesscard/businesscard.go`
**Python Original**: `legacy/modules/org/openteacher/profileRunners/businessCardGenerator/businessCardGenerator.py`

**Key Features**:
- Generates 640x320 pixel business card images
- Combines background image with promotional text
- Saves output as PNG files
- Includes Recuerdo branding and contact information

**Dependencies**:
- `backgroundImageGenerator` - for background graphics
- `ui` - for UI integration
- `execute` - for command execution

**Current Status**: 
- ✅ Module structure ported
- ✅ Core interfaces implemented
- 🚧 Image rendering logic needs implementation
- 🚧 Text rendering needs Go font library
- 🚧 Command-line argument handling needs refinement

### 2. BackgroundImageGenerator Module

**Location**: `internal/modules/profilerunners/backgroundimage/backgroundimage.go`
**Python Original**: `legacy/modules/org/openteacher/profileRunners/backgroundImageGenerator/backgroundImageGenerator.py`

**Key Features**:
- Generates branded background images (1000x5000 pixels)
- Creates gradient backgrounds with rounded rectangles
- Renders Recuerdo logo and application name
- Supports HSV color customization based on theme

**Dependencies**:
- `metadata` - for theme colors and logo path
- `ui` - for UI integration

**Current Status**:
- ✅ Module structure ported
- ✅ Color conversion (HSV to RGB) implemented
- ✅ Basic gradient rendering implemented
- ✅ Simple line drawing implemented
- 🚧 Logo image loading needs implementation
- 🚧 Text rendering with custom fonts needs implementation
- 🚧 Rounded rectangle drawing needs improvement

### 3. GenerateBusinessCard Profile Description

**Location**: `internal/modules/data/profiledescriptions/generatebusinesscard.go`
**Python Original**: `legacy/modules/org/openteacher/data/profileDescriptions/generateBusinessCard/generateBusinessCard.py`

**Key Features**:
- Provides metadata for the business card generation profile
- Defines profile as "advanced" feature
- Conditionally activates based on BusinessCardGenerator availability

**Current Status**:
- ✅ Module structure ported
- ✅ Profile description metadata implemented
- 🚧 Module availability checking needs proper manager integration

## Implementation Notes

### Graphics Rendering Differences

The Python version used PyQt5 for all graphics operations:
- `QImage` for image creation and manipulation
- `QPainter` for drawing operations
- `QTextDocument` for rich text rendering
- `QFont` and `QFontMetrics` for typography

The Go port uses the standard library:
- `image.RGBA` for image creation
- `image/draw` for basic drawing operations
- **Missing**: Advanced text rendering (needs third-party library)
- **Missing**: Logo image loading and scaling

### Text Rendering Challenge

The original Python code renders HTML-styled text with:
```html
<div>
    <strong style='font-size: 19pt;'>Recuerdo</strong><br />
    Language learning software based on OpenTeacher<br /><br />
    
    Copyright © 2025 Joop Kiefte<br />
    Based on OpenTeacher by OpenTeacher Team<br /><br />
    
    Original OpenTeacher:<br />
    http://openteacher.org/
</div>
```

The Go version will need:
- Font loading and management
- Multi-line text layout
- Different font sizes and weights
- Color and transparency support

Potential Go libraries for text rendering:
- `golang.org/x/image/font` (basic)
- `github.com/fogleman/gg` (2D graphics)
- `github.com/llgcode/draw2d` (vector graphics)

### Module Integration

The modules follow the established Go pattern:
- Embed `core.BaseModule` for standard functionality
- Implement `core.Module` interface
- Use context for lifecycle management
- Support dependency injection through manager

## Usage

Once fully implemented, the business card generator can be used via command line:

```bash
./recuerdo generate-business-card output.png
```

This will:
1. Load the backgroundImageGenerator module
2. Generate a themed background image
3. Overlay promotional text
4. Save the result to `output.png`

## TODO Items

### High Priority
- [ ] Implement proper text rendering with Go font libraries
- [ ] Add logo image loading and scaling
- [ ] Complete image compositing functionality
- [ ] Add proper error handling and validation

### Medium Priority
- [ ] Implement rounded rectangle drawing
- [ ] Add image scaling algorithms
- [ ] Improve gradient rendering quality
- [ ] Add unit tests for all modules

### Low Priority
- [ ] Add configurable themes and colors
- [ ] Support different output formats
- [ ] Add batch processing capabilities
- [ ] Create GUI interface for business card generation

## Testing

To test the current implementation:

1. Build the project:
```bash
cd openteacher
go build ./cmd/recuerdo
```

2. The modules will be registered and can be enabled/disabled through the module manager, but full functionality requires completing the TODO items above.

## Architecture Decisions

1. **Standard Library First**: Using Go's standard `image` package instead of CGO-based solutions for better portability
2. **Modular Design**: Each component is a separate module that can be independently tested and maintained  
3. **Interface Compatibility**: Maintaining similar public interfaces to ease potential future integration
4. **Progressive Enhancement**: Basic functionality works with placeholders, advanced features can be added incrementally

This port represents the foundational structure needed for business card generation in Recuerdo Go, with the major graphics rendering work remaining to be implemented.