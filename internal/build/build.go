package build

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sophrone/ssg/internal/markdown"
	"github.com/sophrone/ssg/internal/page"
)

type Options struct {
	SourceDir string
	OutputDir string
	Layout    string
}

func Run(opts Options) error {
	layout, err := template.ParseFiles(opts.Layout)
	if err != nil {
		return fmt.Errorf("load layout: %w", err)
	}

	if err := os.RemoveAll(opts.OutputDir); err != nil {
		return fmt.Errorf("clean output directory: %w", err)
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	return filepath.WalkDir(opts.SourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		p, err := markdown.Parse(path, source)
		if err != nil {
			return err
		}

		if p.FrontMatter.Draft {
			return nil
		}

		rel, err := filepath.Rel(opts.SourceDir, path)
		if err != nil {
			return err
		}

		outputName := strings.TrimSuffix(rel, filepath.Ext(rel)) + ".html"
		outputPath := filepath.Join(opts.OutputDir, outputName)

		if err := renderPage(layout, p, outputPath); err != nil {
			return err
		}

		fmt.Printf("built %s -> %s\n", path, outputPath)
		return nil
	})
}

func renderPage(layout *template.Template, p *page.Page, outputPath string) error {
	data := struct {
		Title       string
		Description string
		Content     template.HTML
	}{
		Title:       p.FrontMatter.Title,
		Description: p.FrontMatter.Description,
		Content:     template.HTML(p.HTML),
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return layout.Execute(file, data)
}
