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
	config *config.Config
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

func New(cfg *config.Config) *Generator {
	return &Generator{config: cfg}
}

func (g *Generator) CollectPromptFiles() ([]PromptFile, error) {
	var files []PromptFile

	err := filepath.WalkDir(g.config.InputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		if g.shouldExclude(path) {
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
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})

	return files, nil
}

func (g *Generator) shouldExclude(path string) bool {
	for _, exclude := range g.config.ExcludeFiles {
		if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

func (g *Generator) GeneratePrompt(files []PromptFile) string {
	var content strings.Builder

	content.WriteString(g.config.Header)

	for _, file := range files {
		content.WriteString(fmt.Sprintf("# %s\n\n", strings.TrimSuffix(file.Filename, ".md")))
		content.WriteString(file.Content)

		if !strings.HasSuffix(file.Content, "\n") {
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(g.config.Footer)

	return content.String()
}

func (g *Generator) WriteOutputFiles(content string) error {
	if g.config.Settings == nil {
		// 従来の方式でフォールバック
		for _, outputFile := range g.config.OutputFiles {
			if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
				return fmt.Errorf("%s", i18n.T("failed_to_write_file", map[string]interface{}{
					"FileName": outputFile,
					"Error":    err,
				}))
			}
		}
		return nil
	}

	// TOML設定を使用
	var outputs []OutputTarget

	// Claude
	if g.config.Settings.Claude.Generate {
		path := g.config.Settings.Claude.Path
		if path == "" {
			path = "."
		}
		outputs = append(outputs, OutputTarget{
			Path:     filepath.Join(path, g.config.Settings.Claude.FileName),
			ToolName: "Claude",
		})
	}

	// Cline
	if g.config.Settings.Cline.Generate {
		path := g.config.Settings.Cline.Path
		if path == "" {
			path = "."
		}
		outputs = append(outputs, OutputTarget{
			Path:     filepath.Join(path, g.config.Settings.Cline.FileName),
			ToolName: "Cline",
		})
	}

	// Custom tools
	for toolName, settings := range g.config.Settings.Custom {
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
	if g.config.Settings == nil {
		return g.config.OutputFiles
	}

	var targets []string
	if g.config.Settings.Claude.Generate {
		path := g.config.Settings.Claude.Path
		if path == "" {
			path = "."
		}
		targets = append(targets, filepath.Join(path, g.config.Settings.Claude.FileName))
	}

	if g.config.Settings.Cline.Generate {
		path := g.config.Settings.Cline.Path
		if path == "" {
			path = "."
		}
		targets = append(targets, filepath.Join(path, g.config.Settings.Cline.FileName))
	}

	for _, settings := range g.config.Settings.Custom {
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
			"InputDir": g.config.InputDir,
		}))
	}

	content := g.GeneratePrompt(files)

	if err := g.WriteOutputFiles(content); err != nil {
		return err
	}

	return nil
}
