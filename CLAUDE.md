# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

system-prompt-gen is a Go CLI tool that aggregates multiple AI system prompt files from `.system_prompt/*.md` into unified prompt files for various AI tools (Claude, Cline, custom tools). It provides both command-line and interactive TUI modes.

## Build & Development Commands

```bash
# Build the project
make build
# or
go build -o .bin/system-prompt-gen .

# Test with example configuration
make test
# or
cd example && ../.bin/system-prompt-gen

# Interactive mode testing
make interactive
# or
cd example && ../.bin/system-prompt-gen -i

# Clean build artifacts and generated files
make clean

# Install to system PATH
make install
```

## Architecture & Core Components

### Configuration System (Dual Layer)
The tool uses a two-layer configuration system:
1. **JSON Config** (`~/.config/system-prompt-gen/config.json`) - Legacy global settings
2. **TOML Settings** (`.system_prompt/settings.toml`) - Per-project AI tool configuration

Key types in `internal/config/config.go`:
- `Config`: Main configuration with backwards compatibility
- `Settings`: TOML-based per-tool settings
- `AIToolSettings`: Individual tool configuration (generate flag, path, filename)

### Generator Flow
`internal/generator/generator.go` orchestrates the core workflow:
1. Scan `.system_prompt/*.md` files (excluding patterns in config)
2. Sort files alphabetically by filename
3. Merge content with configured header/footer
4. Output to multiple targets based on TOML settings

### Configuration Loading Priority
The system loads configuration in this order:
1. `LoadConfigWithSettings()` tries both JSON config and TOML settings
2. Falls back to `LoadConfig()` for JSON-only (backwards compatibility)  
3. Uses `DefaultSettings()` if no TOML file exists
4. Settings in TOML override JSON configuration for output behavior

### Output Target Resolution
When TOML settings are present, the generator:
- Checks each AI tool's `generate` flag
- Resolves paths (empty string = current directory)
- Creates directories as needed
- Supports custom tools via `[custom.toolname]` sections

### Interactive UI
`internal/ui/tui.go` provides a Bubble Tea TUI with three states:
- Loading: File collection phase
- Success: Shows preview with file count and targets
- Error: Displays errors with retry option

## Settings.toml Configuration

Place `.system_prompt/settings.toml` in your working directory:

```toml
[claude]
generate = true
path = ""           # defaults to current directory
file_name = ""      # defaults to "CLAUDE.md"

[cline] 
generate = true
path = ""
file_name = ""      # defaults to ".clinerules"

[custom.toolname]   # Add custom AI tools
generate = true
path = "./custom"   # required for custom tools
file_name = "custom.md"  # required for custom tools
```

## CLI Usage Patterns

```bash
# Basic usage (uses current directory's .system_prompt/)
system-prompt-gen

# Custom config location
system-prompt-gen -c /path/to/config.json

# Interactive mode for preview and confirmation
system-prompt-gen -i

# The tool expects .system_prompt/ directory with:
# - *.md files (prompt fragments)
# - settings.toml (optional, tool-specific config)
```