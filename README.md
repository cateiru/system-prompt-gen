# system-prompt-gen

[日本語](./README_ja.md) | **English**

A Go CLI tool that aggregates multiple AI system prompt files from `.system_prompt/*.md` into unified prompt files for various AI tools (Claude, Cline, custom tools). It provides both command-line execution and interactive TUI modes with full internationalization support.

## Features

- 🚀 Aggregates `.system_prompt/*.md` files to generate files for various AI tools
- 🎛️ Choose between interactive TUI mode and command-line mode
- 🌍 Full internationalization support (Japanese & English)
- ⚙️ Flexible configuration management with TOML configuration files
- 🔧 Support for custom AI tools
- 🚫 Tool-specific file exclusion patterns
- 🎨 Beautiful TUI using Bubble Tea

## Installation

### Install from Binary

Download the latest version from the Releases page.

### Build from Source

```bash
git clone https://github.com/yourusername/system-prompt-gen
cd system-prompt-gen
make build
```

## Usage

### Basic Usage

```bash
# Use .system_prompt/ in current directory
system-prompt-gen

# Specify custom settings file location
system-prompt-gen -s /path/to/settings.toml

# Run in interactive mode
system-prompt-gen -i

# Specify language
system-prompt-gen --language ja
system-prompt-gen -l en
```

### Directory Structure

The tool expects the following directory structure:

```txt
your-project/
├── .system_prompt/
│   ├── settings.toml      # Configuration file (optional)
│   ├── 01-base.md         # Prompt file
│   ├── 02-context.md      # Prompt file
│   └── 03-rules.md        # Prompt file
├── CLAUDE.md              # Generated file
└── .clinerules            # Generated file
```

## Configuration File

Place your configuration file at `.system_prompt/settings.toml`:

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
exclude = ["private*.md"]           # Exclude sensitive files from custom tools
```

### Exclude Patterns

Each tool can define `exclude` patterns to filter out specific files from `.system_prompt/`:
- Uses shell-style glob patterns (`*`, `?`, `[...]`)
- Patterns are matched against relative paths from `.system_prompt/` directory  
- Common patterns: `"003_*.md"`, `"temp*.md"`, `"private*.md"`, `"draft_*.md"`
- Each tool processes only the files not excluded by its patterns

## Development

### Build and Test Commands

```bash
# Build the project
make build

# Test commands (for development)
make test-unit      # Run unit tests
make test-coverage  # Run tests with coverage report
make test-verbose   # Run tests with race detection

# Integration testing
make test          # Integration test with example configuration
make interactive   # Integration test with interactive mode

# Clean build artifacts and generated files
make clean

# Install to system PATH
make install
```

### Architecture

#### Configuration System

Primarily uses TOML-based configuration:

1. **TOML Settings** (`.system_prompt/settings.toml`) - Main configuration for AI tool settings and app preferences
2. **JSON Config** (legacy) - Maintained for backward compatibility only

#### Generator Processing Flow

`internal/generator/generator.go` controls the core workflow:

1. For each enabled tool, collect `.system_prompt/*.md` files (applying tool-specific exclusion patterns)
2. Sort files alphabetically by filename
3. Merge configured headers/footers with content
4. Generate tool-specific output files based on TOML configuration

#### Internationalization System

`internal/i18n/i18n.go` provides comprehensive i18n support:

- Uses `github.com/nicksnyder/go-i18n/v2` with embedded JSON translation files
- Language detection: settings.toml → LANG environment variable → fallback (ja → en)
- All user-facing messages (CLI, TUI, errors) are localized

## Supported AI Tools

- **Claude** - Prompt files for Anthropic Claude
- **Cline** - Rule files for Cline VS Code extension
- **GitHub Copilot** - Instruction files for GitHub Copilot
- **Custom Tools** - Custom files for any AI tool

## License

MIT License

## Contributing

Pull requests and issues are welcome. Please follow the existing code style before contributing.

## Support

Please submit bug reports and feature requests through [Issues](https://github.com/yourusername/system-prompt-gen/issues).
