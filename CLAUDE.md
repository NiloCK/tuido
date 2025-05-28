# CLAUDE.md - AI Assistant Guide for Tuido

## Project Overview

Tuido is a terminal user interface (TUI) application written in Go for managing todo items in the [x]it! format. It provides an opinionated, efficient interface for browsing and managing tasks, inspired by Getting Things Done methodology and productivity best practices.

## Key Features

- **File Discovery**: Recursively searches directories for [x]it! compatible items in `.xit`, `.md`, and `.txt` files
- **Interactive TUI**: Built with Charm's Bubble Tea framework for rich terminal interfaces
- **Task Management**: Create, edit, update, and organize todo items with multiple states
- **Time Management**: Built-in Pomodoro timer and time tracking capabilities
- **Smart Organization**: Progressive snoozing, search/filtering, importance levels, and intelligent sorting
- **Configuration**: Flexible configuration system using `.tuido` files
- **Cross-platform**: Distributed as native binaries and npm package

## Architecture Overview

### Directory Structure

```
tuido/
├── main.go              # Application entry point
├── tui/                 # Terminal user interface package
│   ├── tui.go          # Main TUI logic and models
│   ├── init.go         # Initialization
│   ├── update.go       # Event handling and updates
│   ├── view.go         # Rendering and display
│   ├── config.go       # Configuration management
│   ├── nag.go          # Friction/deterrence system
│   └── peek.go         # File preview functionality
├── tuido/              # Core business logic package
│   ├── tuido.go        # Item management and file operations
│   ├── time.go         # Date/time utilities and shorthands
│   └── tuido_test.go   # Unit tests
├── utils/              # Utility functions
│   ├── utils.go        # General utilities
│   └── versioning.go   # Version checking and updates
├── tuido-npm/          # NPM package distribution
└── .tuido              # Project configuration file
```

### Core Components

#### 1. Main Entry Point (`main.go`)
Simple entry point that calls `tui.Run()` to start the application.

#### 2. TUI Package (`tui/`)
Implements the terminal interface using Bubble Tea framework:
- **Models**: Application state management with different modes (navigation, filter, edit, help, pomo, nag, peek)
- **Updates**: Event handling for keyboard input and timer events
- **Views**: Rendering logic for different interface states
- **Configuration**: Manages `.tuido` config files and settings

#### 3. Tuido Package (`tuido/`)
Core business logic for todo item management:
- **Item struct**: Represents individual todo items with file location, line number, and content
- **Status management**: Handles item states (open, ongoing, checked, obsolete)
- **File operations**: Reading from and writing to source files
- **Tag system**: Supports metadata tags like `#due`, `#active`, `#repeat`, etc.
- **Time utilities**: Date/time parsing and shorthand expansion

#### 4. Utils Package (`utils/`)
General utilities including version checking and terminal capabilities.

## Key Data Structures

### Item
```go
type Item struct {
    file string  // Source file path
    line int     // Line number in source file
    raw  string  // Raw text content
}
```

### Status Types
- `Open`: `[ ]` - Noted but not begun
- `Ongoing`: `[@]` - In progress
- `Checked`: `[x]` - Completed
- `Obsolete`: `[~]` - No longer necessary

### Configuration
```go
type config struct {
    extensions        []string // File extensions to parse
    writeto          string   // Where to write new items
    frictionThreshold int     // Deterrence threshold
}
```

## Configuration System

### Global Configuration
- Default location: `$HOME/.config/tuido.conf` (Linux) or `$HOME/AppData` (Windows)
- Default write location: `$HOME/.tuido/`

### Local Configuration (`.tuido` files)
- Can be placed in any directory to customize behavior for that subtree
- Supports `extensions`, `writeto`, and `frictionThreshold` settings
- Project uses: `extensions=go,md`, `writeto=readme.md`, `frictionThreshold=3`

## File Parsing and Management

### Supported Formats
- **[x]it! format**: Standard todo format with status boxes
- **Markdown**: Supports bulleted lists with todo items
- **Code comments**: Parses `//` style comments for todo items
- **Flexible whitespace**: Handles leading whitespace and indentation

### File Operations
- Reads items from multiple file types simultaneously
- Preserves file structure when updating items
- Respects `.gitignore` patterns to avoid parsing build artifacts
- Uses line-based editing to maintain file integrity

## Time and Date Features

### Shorthand System
Format: `[item] [shorthand][timespan]`
- `d` = due date
- `a` = active after (snooze)
- `r` = recurring
- `e` = estimate

Timespans: `m` (minute), `h` (hour), `d` (day), `w` (week), `M` (month), `y` (year)

Examples:
- `fix bug d2w` → `fix bug #due=2024-01-15`
- `call mom r1w` → `call mom #repeat=1w`
- `clean garage e3h` → `clean garage #estimate=3h`

### Tag System
- `#due=YYYY-MM-DD`: Due dates for sorting and prioritization
- `#active=YYYY-MM-DD`: Hide item until specified date
- `#repeat=Xd/w/M/y`: Auto-reschedule when completed
- `#estimate=Xh`: Time estimates
- `#spent=X.XX`: Tracked time in minutes
- `#zzz=N`: Snooze count for progressive snoozing

## UI Modes and Navigation

### Modes
- **Navigation**: Browse and select items
- **Filter**: Search/filter items by text
- **Edit**: Modify item text
- **Help**: Show keyboard shortcuts
- **Pomo**: Pomodoro timer mode
- **Nag**: Friction screen for adding too many items
- **Peek**: Preview item context in source file

### Key Bindings
- `?`: Help
- `n`: New item
- `Space`: Mark as open
- `x/X`: Mark as done
- `s/~`: Mark as obsolete
- `a/@`: Mark as ongoing
- `e`: Edit item
- `p`: Pomodoro session
- `z`: Snooze item
- `!/1`: Increase/decrease importance
- `Tab`: Switch between pending/done
- `/`: Filter items
- `q`: Quit

## Development Guidelines

### Dependencies
- **Bubble Tea**: TUI framework for terminal interfaces
- **Lipgloss**: Styling and layout for terminal output
- **Chroma**: Syntax highlighting for file previews
- **go-git**: Git repository operations and .gitignore support

### Testing
- Unit tests in `tuido_test.go`
- Integration testing via GitHub Actions
- VHS (Video Recording) for demo generation

### Building and Distribution
- **GoReleaser**: Multi-platform binary builds
- **NPM Package**: Node.js distribution wrapper
- **GitHub Actions**: Automated CI/CD pipeline

### Code Style
- Standard Go formatting with `go fmt`
- Extensive use of receiver methods on structs
- Error handling with explicit error returns
- File operations with proper cleanup (defer statements)

## Common Development Tasks

### Adding New Features
1. Define data structures in `tuido/` package
2. Implement business logic and file operations
3. Add UI components in `tui/` package
4. Update key bindings and help text
5. Add tests and documentation

### Modifying File Parsing
- Update `IsTuido()` function for new formats
- Modify `trim()` function for comment styles
- Add language-specific parsing rules

### Extending Configuration
- Add new fields to `config` struct
- Update `parseConfig()` function
- Add default values in `runConfig`

### Performance Considerations
- File operations are synchronous but efficient
- Large directories with many files may impact startup time
- Color calculations for tags happen once at startup
- Progressive snoozing uses Fibonacci sequence for exponential backoff

## Troubleshooting Common Issues

### File Permission Issues
- Ensure write permissions for target directories
- Check that `.tuido` directories are accessible

### Performance Issues
- Use `.gitignore` to exclude large directories
- Limit file extensions in `.tuido` config files
- Consider excluding binary file types

### Parsing Issues
- Verify [x]it! format compliance
- Check for unsupported comment styles in code files
- Ensure proper whitespace in todo items

This documentation provides a comprehensive foundation for understanding and working with the Tuido codebase. The project demonstrates excellent Go practices, thoughtful UX design, and solid architectural patterns for terminal applications.