package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
)

type Generator struct {
	settings *config.Settings
}

type PromptFile struct {
	Path     string
	Filename string
	Content  string
}

type OutputTarget struct {
	Path     string
	ToolName string
}

func New(settings *config.Settings) *Generator {
	return &Generator{settings: settings}
}

func (g *Generator) CollectPromptFiles() ([]PromptFile, error) {
	var files []PromptFile

	err := filepath.WalkDir(
		g.settings.App.InputDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			files = append(files, PromptFile{
				Path:     path,
				Filename: d.Name(),
				Content:  string(content),
			})

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})

	return files, nil
}

func (g *Generator) GeneratePrompt(files []PromptFile) string {
	var content strings.Builder

	content.WriteString(g.settings.App.Header)

	for _, file := range files {
		content.WriteString(fmt.Sprintf("# %s\n\n", strings.TrimSuffix(file.Filename, ".md")))
		content.WriteString(file.Content)

		if !strings.HasSuffix(file.Content, "\n") {
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(g.settings.App.Footer)

	return content.String()
}

func (g *Generator) WriteOutputFiles(content string) error {
	// TOML設定を使用
	var outputs []OutputTarget

	// FIXME: for でループできるようにしたい

	// Claude
	if g.settings.Claude.Generate {
		path := g.settings.Claude.Path
		if path == "" {
			path = "."
		}
		outputs = append(outputs, OutputTarget{
			Path:     filepath.Join(path, g.settings.Claude.FileName),
			ToolName: "Claude",
		})
	}

	// Cline
	if g.settings.Cline.Generate {
		path := g.settings.Cline.Path
		if path == "" {
			path = "."
		}
		outputs = append(outputs, OutputTarget{
			Path:     filepath.Join(path, g.settings.Cline.FileName),
			ToolName: "Cline",
		})
	}

	// Custom tools
	for toolName, settings := range g.settings.Custom {
		if settings.Generate && settings.Path != "" && settings.FileName != "" {
			outputs = append(outputs, OutputTarget{
				Path:     filepath.Join(settings.Path, settings.FileName),
				ToolName: toolName,
			})
		}
	}

	// ファイル出力
	for _, target := range outputs {
		dir := filepath.Dir(target.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s", i18n.T("failed_to_create_directory", map[string]interface{}{
				"DirName": dir,
				"Error":   err,
			}))
		}

		if err := os.WriteFile(target.Path, []byte(content), 0644); err != nil {
			return fmt.Errorf("%s", i18n.T("failed_to_write_tool_file", map[string]interface{}{
				"FileName": target.Path,
				"ToolName": target.ToolName,
				"Error":    err,
			}))
		}
	}

	return nil
}

func (g *Generator) GetGeneratedTargets() []string {
	var targets []string
	if g.settings.Claude.Generate {
		path := g.settings.Claude.Path
		if path == "" {
			path = "."
		}
		targets = append(targets, filepath.Join(path, g.settings.Claude.FileName))
	}

	if g.settings.Cline.Generate {
		path := g.settings.Cline.Path
		if path == "" {
			path = "."
		}
		targets = append(targets, filepath.Join(path, g.settings.Cline.FileName))
	}

	for _, settings := range g.settings.Custom {
		if settings.Generate && settings.Path != "" && settings.FileName != "" {
			targets = append(targets, filepath.Join(settings.Path, settings.FileName))
		}
	}

	return targets
}

func (g *Generator) Run() error {
	files, err := g.CollectPromptFiles()
	if err != nil {
		return fmt.Errorf("%s", i18n.T("failed_to_collect_files", map[string]interface{}{
			"Error": err,
		}))
	}

	if len(files) == 0 {
		return fmt.Errorf("%s", i18n.T("no_prompt_files_found", map[string]interface{}{
			"InputDir": g.settings.App.InputDir,
		}))
	}

	content := g.GeneratePrompt(files)

	if err := g.WriteOutputFiles(content); err != nil {
		return err
	}

	return nil
}
