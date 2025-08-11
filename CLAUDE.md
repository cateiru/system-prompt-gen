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

# Test with example configuration
make test
# or
cd example && ../.bin/system-prompt-gen

# Test interactive mode
make interactive
# or
cd example && ../.bin/system-prompt-gen -i

# Clean build artifacts and generated files
make clean

# Install to system PATH
make install
```

## Architecture and Core Components

### Dual Configuration System
The tool uses a two-tier configuration system:
1. **JSON Config** (`~/.config/system-prompt-gen/config.json`) - Legacy global settings
2. **TOML Settings** (`.system_prompt/settings.toml`) - Per-project AI tool configuration

Key types in `internal/config/config.go`:
- `Config`: Main configuration with backward compatibility
- `Settings`: TOML-based tool-specific settings
- `AIToolSettings`: Individual tool settings (generate flags, paths, filenames)
- `AppSettings`: Application-level settings (language preferences)

### Generator Processing Flow
`internal/generator/generator.go` controls the core workflow:
1. Scan `.system_prompt/*.md` files (applying exclusion patterns from config)
2. Sort files alphabetically by filename
3. Merge configured headers/footers with content
4. Output to multiple targets based on TOML configuration

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

## settings.toml Configuration

Place `.system_prompt/settings.toml` in your working directory:

```toml
# Application settings
[app]
language = "ja"       # Language setting: "ja" (Japanese), "en" (English), "" (auto-detect)

[claude]
generate = true       # Set to false to disable generation, default is true
path = ""            # Unspecified defaults to current directory
file_name = ""       # Defaults to "CLAUDE.md"

[cline]
generate = true
path = ""
file_name = ""       # Defaults to ".clinerules"

[custom.toolname]    # Add custom AI tools
generate = true
path = "./custom"    # Required for custom tools
file_name = "custom.md"  # Required for custom tools
```

## CLI Usage Patterns

```bash
# Basic usage (uses .system_prompt/ in current directory)
system-prompt-gen

# Specify custom config file location
system-prompt-gen -c /path/to/config.json

# Interactive mode for preview and confirmation
system-prompt-gen -i

# Language override via environment variable
LANG=en_US.UTF-8 system-prompt-gen

# The tool expects a .system_prompt/ directory containing:
# - *.md files (prompt fragments)
# - settings.toml (optional, tool-specific configuration)
```
