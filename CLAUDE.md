# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

system-prompt-gen is a Go CLI tool that aggregates multiple AI system prompt files from `.system_prompt/*.md` into unified prompt files for various AI tools (Claude, Cline, custom tools). It provides both command-line execution and interactive TUI modes with full internationalization support.

## Build and Development Commands

```bash
# Build the project
make build
# or
go build -o .bin/system-prompt-gen .

# Test commands (use these for development)
make test-unit      # Run unit tests (go test -v ./...)
make test-coverage  # Run tests with coverage report
make test-verbose   # Run tests with race detection

# Integration testing with example configuration
make test          # cd example && ../.bin/system-prompt-gen
make interactive   # cd example && ../.bin/system-prompt-gen -i

# Clean build artifacts and generated files
make clean

# Install to system PATH
make install
```

## Architecture and Core Components

### Configuration System
The tool primarily uses TOML-based configuration:
1. **TOML Settings** (`.system_prompt/settings.toml`) - Main configuration for AI tool settings and app preferences
2. **JSON Config** (legacy) - Maintained for backward compatibility only

The TOML system is the preferred and actively developed configuration method.

Key types in `internal/config/config.go`:
- `Config`: Main configuration with backward compatibility
- `Settings`: TOML-based tool-specific settings
- `AIToolSettings`: Individual tool settings (generate flags, paths, filenames, exclude patterns)
- `AppSettings`: Application-level settings (headers, footers, directories)

### Generator Processing Flow
`internal/generator/generator.go` controls the core workflow:
1. For each enabled tool, collect `.system_prompt/*.md` files (applying tool-specific exclusion patterns)
2. Sort files alphabetically by filename
3. Merge configured headers/footers with content
4. Generate tool-specific output files based on TOML configuration

Key functions:
- `CollectPromptFiles()`: Collects all `.md` files (legacy function)
- `CollectPromptFilesForTool()`: Collects files for specific tool with exclude patterns applied
- `WriteOutputFilesWithExcludes()`: Main output function that processes each tool independently

### Configuration Loading Priority
The system loads configuration in this order:
1. `LoadConfigWithSettings()` attempts both JSON and TOML configurations
2. Falls back to `LoadConfig()` (JSON only, backward compatibility)
3. Uses `DefaultSettings()` if TOML file doesn't exist
4. TOML settings override JSON configuration output behavior

### Output Target Resolution
When TOML settings exist, the generator:
- Checks each AI tool's `generate` flag
- Resolves paths (empty string = current directory)
- Creates directories as needed
- Supports custom tools via `[custom.toolname]` sections

### Interactive UI
`internal/ui/tui.go` provides a Bubble Tea TUI with three states:
- Loading: File collection phase
- Success: File count and target preview display
- Error: Error display with retry option

### Internationalization System
`internal/i18n/i18n.go` provides comprehensive i18n support:
- Uses `github.com/nicksnyder/go-i18n/v2` with embedded JSON translation files
- Language detection: settings.toml → LANG environment variable → fallback (ja → en)
- All user-facing messages (CLI, TUI, errors) are localized
- Translation files: `internal/i18n/locales/{en,ja}.json`
- Use `i18n.T()` function for translating messages throughout the codebase
- Error messages use localized templates with `i18n.T("message_key", map[string]interface{}{"Key": value})`

## settings.toml Configuration

Place `.system_prompt/settings.toml` in your working directory:

```toml
# Application settings
[app]
header = "Custom header content"    # Optional header for all generated files
footer = "Custom footer content"    # Optional footer for all generated files

[tools.claude]
generate = true       # Set to false to disable generation, default is true
dir_name = ""         # Directory name (empty = current directory)
file_name = ""        # File name (empty = default: "CLAUDE.md")
exclude = ["003_*.md", "temp*.md"]  # Exclude patterns for files (optional)

[tools.cline]
generate = true
dir_name = ""
file_name = ""        # Defaults to ".clinerules"
exclude = ["001_*.md"]              # Tool-specific exclude patterns

[tools.github_copilot]
generate = false      # Built-in support for GitHub Copilot instructions
dir_name = ".github"  # Default: .github/copilot-instructions.md
file_name = "copilot-instructions.md"

[tools.custom_tool]   # Add custom AI tools
generate = true
dir_name = "./custom" # Required for custom tools
file_name = "custom.md"  # Required for custom tools
include = ["public_*.md", "common_*.md"]  # Include only public and common files
exclude = ["private*.md"]           # Exclude sensitive files from custom tools
```

### Include/Exclude Patterns

Each tool can define `include` and `exclude` patterns to filter files from `.system_prompt/`:

#### Include Patterns
- `include = ["pattern1", "pattern2"]` - Include only files matching these patterns
- If undefined, all files are included by default
- Uses shell-style glob patterns (`*`, `?`, `[...]`)
- Patterns are matched against relative paths from `.system_prompt/` directory
- Common patterns: `"01-*.md"`, `"public_*.md"`, `"*"` (all files)

#### Exclude Patterns  
- `exclude = ["pattern1", "pattern2"]` - Exclude files matching these patterns
- **Exclude takes priority** - files matching both include and exclude patterns are excluded
- Uses `filepath.Match` for pattern matching (shell-style patterns)
- Common patterns: `"003_*.md"`, `"temp*.md"`, `"private*.md"`, `"draft_*.md"`

#### Processing Order
1. If `include` is defined, only files matching include patterns are considered
2. If `include` is undefined, all files are considered
3. Files matching `exclude` patterns are then removed (exclude takes priority)
4. Each tool processes only the remaining files

## CLI Usage Patterns

```bash
# Basic usage (uses .system_prompt/ in current directory)
system-prompt-gen

# Specify custom settings file location
system-prompt-gen -s /path/to/settings.toml

# Interactive mode for preview and confirmation
system-prompt-gen -i

# Non-interactive mode (recommended for automated scripts and CI/CD)
system-prompt-gen -i=false

# Language setting via command line flag
system-prompt-gen --language ja
system-prompt-gen -l en

# Language override via environment variable (still supported)
LANG=en_US.UTF-8 system-prompt-gen

# The tool expects a .system_prompt/ directory containing:
# - *.md files (prompt fragments)
# - settings.toml (optional, tool-specific configuration)
```

### Important Usage Notes

- **TTY Auto-Detection**: The tool automatically detects non-TTY environments (CI/CD, Claude Code, pipes, etc.) and falls back to non-interactive mode, even when `-i=true` is specified. This eliminates the need to manually specify `-i=false` in automated environments.
- **Manual override**: Use `-i=false` to explicitly force non-interactive mode when needed
- **Example directory**: The `example/` directory contains a complete working setup with exclude patterns demonstration
- **Test your configuration**: Run `cd example && ../.bin/system-prompt-gen -i=false` to see exclude patterns in action

### TTY Auto-Fallback Behavior

The application includes automatic TTY detection (implemented in `cmd/root.go:70-74`):
- When interactive mode is enabled (`-i=true`) but no TTY is detected, the tool automatically switches to non-interactive mode
- A localized message is displayed to inform the user of the automatic fallback
- This works seamlessly in environments like Claude Code, CI/CD pipelines, and shell pipes
