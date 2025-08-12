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

func (g *Generator) CollectPromptFilesForTool(toolName string, toolSettings config.AIToolSettings) ([]PromptFile, error) {
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

			relPath, err := filepath.Rel(g.settings.App.InputDir, path)
			if err != nil {
				return err
			}

			// Include パターンのチェック（未定義の場合は全てを含める）
			if len(toolSettings.Include) > 0 {
				includeMatched := false
				for _, pattern := range toolSettings.Include {
					if matched, _ := filepath.Match(pattern, relPath); matched {
						includeMatched = true
						break
					}
				}
				if !includeMatched {
					return nil
				}
			}

			// Exclude パターンのチェック（Include より優先）
			for _, pattern := range toolSettings.Exclude {
				if matched, _ := filepath.Match(pattern, relPath); matched {
					return nil
				}
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

	for name, tool := range g.settings.Tools {
		paths := []string{
			g.settings.App.OutputDir,
		}
		if tool.DirName != "" {
			paths = append(paths, string(tool.DirName))
		}
		paths = append(paths, string(tool.FileName))

		outputs = append(outputs, OutputTarget{
			Path:     filepath.Join(paths...),
			ToolName: name,
		})
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

func (g *Generator) WriteOutputFilesWithExcludes() error {
	for name, tool := range g.settings.Tools {
		files, err := g.CollectPromptFilesForTool(name, tool)
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

		paths := []string{
			g.settings.App.OutputDir,
		}
		if tool.DirName != "" {
			paths = append(paths, string(tool.DirName))
		}
		paths = append(paths, string(tool.FileName))

		outputPath := filepath.Join(paths...)
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s", i18n.T("failed_to_create_directory", map[string]interface{}{
				"DirName": dir,
				"Error":   err,
			}))
		}

		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("%s", i18n.T("failed_to_write_tool_file", map[string]interface{}{
				"FileName": outputPath,
				"ToolName": name,
				"Error":    err,
			}))
		}
	}

	return nil
}

func (g *Generator) GetGeneratedTargets() []string {
	var targets []string

	for _, tool := range g.settings.Tools {
		paths := []string{
			g.settings.App.OutputDir,
		}
		if tool.DirName != "" {
			paths = append(paths, string(tool.DirName))
		}
		paths = append(paths, string(tool.FileName))

		targets = append(targets, filepath.Join(paths...))
	}

	return targets
}

func (g *Generator) Run() error {
	return g.WriteOutputFilesWithExcludes()
}
