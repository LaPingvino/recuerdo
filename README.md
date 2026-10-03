# Recuerdo

Recuerdo is a language learning application that helps you create and study vocabulary lessons, geography maps, and multimedia content. It's a modern Go version of the popular OpenTeacher educational software.

## What You Can Do Right Now

### ✅ Create and Study Lessons
- **Vocabulary lessons**: Create word lists with questions and answers
- **Geography lessons**: Study maps with clickable locations
- **Media lessons**: Learn with images, audio, and video files
- **Test yourself**: Take quizzes and track your progress

### ✅ Import Your Content
- Load lessons from CSV files, text files, or other formats
- Use existing OpenTeacher lesson files (.otwd)
- Import from popular language learning platforms

### ✅ Cross-Platform Support
- Works on Linux (primary), macOS, and Windows
- Automatic system detection and configuration
- Native desktop integration with Qt

## Quick Start

1. **Download and build:**
   ```bash
   git clone https://github.com/LaPingvino/recuerdo.git
   cd recuerdo
   go build ./cmd/recuerdo
   ```

2. **Run the application:**
   ```bash
   ./recuerdo
   ```

3. **Test your system:**
   ```bash
   go build ./cmd/test-core
   ./test-core
   ```

## Features That Work Today

### Lesson Creation
- Add vocabulary word pairs
- Place locations on maps by clicking
- Attach media files (images, audio, video)
- Organize content with categories and comments

### Study Modes
- Quiz yourself on vocabulary
- Identify locations on maps
- Review multimedia content
- Track correct/incorrect answers

### File Management
- Save lessons in multiple formats
- Import from CSV, text files
- Export for sharing or backup
- Recent files list for quick access

### System Integration
- Text-to-speech for pronunciation help
- Unicode input support for international characters
- Print lessons and results
- System diagnostics and troubleshooting

## Lesson Types

### 📚 Words
Perfect for language learning:
- Question/answer pairs
- Multiple translations per word
- Comments and context notes
- Pronunciation with text-to-speech

### 🗺️ Geography (Topo)
Interactive map learning:
- Click to place locations
- Built-in maps of continents and countries
- Import your own map images
- Coordinate-based precision

### 🎬 Media
Rich multimedia lessons:
- Images with descriptions
- Audio pronunciation guides
- Video content integration
- Web links and resources

## File Formats

**Import from:**
- CSV files (comma or tab separated)
- Plain text files
- OpenTeacher files (.otwd)
- KDE Vocabulary files (.kvtml)

**Export to:**
- CSV for spreadsheets
- HTML for web viewing
- Plain text for simple sharing
- OpenTeacher format for compatibility

## System Requirements

- **Operating System**: Linux (recommended), macOS, or Windows
- **Memory**: 256MB RAM minimum
- **Storage**: 50MB available space
- **Display**: Any resolution, works with X11 or Wayland

### Optional Components
- **Qt libraries**: For full GUI experience
- **Text-to-speech**: espeak (Linux), built-in (macOS/Windows)
- **Audio**: For media lessons and pronunciation

## Getting Help

### Built-in Diagnostics
Run `./test-core` to check:
- System compatibility
- Missing dependencies
- Performance issues
- Configuration problems

### Common Issues
- **No sound**: Check text-to-speech installation
- **Import problems**: Verify file encoding (UTF-8 recommended)
- **Display issues**: Run system diagnostics for Qt setup

## Project Status

Recuerdo is actively developed and regularly updated. The core functionality is stable and ready for daily use. New features are added based on user feedback and educational needs.

**Current version**: 4.0-alpha  
**Last updated**: January 2025  
**Stability**: Good for personal use, testing welcome

## Development

Recuerdo is built with Go and uses Qt for the user interface. The modular architecture makes it easy to add new lesson types and features.

### Building from Source
```bash
# Install Go 1.23+
# Install Qt development libraries
go mod download
go build -v ./...
```

### Contributing
- Report bugs and request features via GitHub
- Submit translations for your language
- Share lesson files and templates
- Help with testing on different platforms

## License

GPL v3+ - Free and open source software (see [LICENSE](LICENSE))

Based on OpenTeacher © 2010-2023 OpenTeacher Team  
Recuerdo © 2025 Joop Kiefte

## Links

- **Project**: https://github.com/LaPingvino/recuerdo
- **Chat**: https://matrix.to/#/#recuerdo:chat.kiefte.eu
- **Original**: http://openteacher.org
- **Support**: GitHub Issues

---

*Ready to start learning? Download Recuerdo and create your first lesson!*