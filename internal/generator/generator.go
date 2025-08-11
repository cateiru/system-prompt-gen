package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cateiru/system-prompt-gen/internal/config"
)

type Generator struct {
	config *config.Config
}

type PromptFile struct {
	Path     string
	Filename string
	Content  string
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
	for _, outputFile := range g.config.OutputFiles {
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputFile, err)
		}
	}
	return nil
}

func (g *Generator) Run() error {
	files, err := g.CollectPromptFiles()
	if err != nil {
		return fmt.Errorf("failed to collect prompt files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no prompt files found in %s", g.config.InputDir)
	}

	content := g.GeneratePrompt(files)
	
	if err := g.WriteOutputFiles(content); err != nil {
		return err
	}

	return nil
}