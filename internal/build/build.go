package build

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sophrone/ssg/internal/config"
	"github.com/sophrone/ssg/internal/markdown"
	"github.com/sophrone/ssg/internal/page"
)

type Options struct {
	SourceDir  string
	OutputDir  string
	Layout     string
	StaticDir  string
	ConfigPath string
}

func Run(opts Options) error {
	config, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	layout, err := template.ParseFiles(opts.Layout)
	if err != nil {
		return fmt.Errorf("load layout: %w", err)
	}

	var pages []*page.Page

	// Collect every page before rendering any of them.
	err = filepath.WalkDir(opts.SourceDir, func(path string, entry fs.DirEntry, err error) error {
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
			return fmt.Errorf("get relative path for %s: %w", path, err)
		}

		outputName := strings.TrimSuffix(rel, filepath.Ext(rel)) + ".html"
		p.OutputPath = filepath.Join(opts.OutputDir, outputName)

		pages = append(pages, p)

		return nil
	})
	if err != nil {
		return fmt.Errorf("collect pages: %w", err)
	}

	if err := os.RemoveAll(opts.OutputDir); err != nil {
		return fmt.Errorf("clean output directory: %w", err)
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := copyStatic(opts.StaticDir, opts.OutputDir); err != nil {
		return fmt.Errorf("copy static assets: %w", err)
	}

	// Every page receives the same complete navigation.
	for _, p := range pages {
		p.Navigation = config.Navigation

		if err := renderPage(layout, p, p.OutputPath, config.Title); err != nil {
			return fmt.Errorf("render %s: %w", p.SourcePath, err)
		}

		fmt.Printf("built %s -> %s\n", p.SourcePath, p.OutputPath)
	}

	return nil
}

func renderPage(layout *template.Template, p *page.Page, outputPath string, siteTitle string) error {
	data := struct {
		Title       string
		Description string
		Content     template.HTML
		Navigation  []page.NavigationLink
		SiteTitle   string
	}{
		Title:       p.FrontMatter.Title,
		Description: p.FrontMatter.Description,
		Content:     template.HTML(p.HTML),
		Navigation:  p.Navigation,
		SiteTitle:   siteTitle,
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	if err := layout.Execute(file, data); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}

func navigationLink(p *page.Page, outputName string) page.NavigationLink {
	url := "/" + filepath.ToSlash(outputName)

	switch {
	case url == "/index.html":
		url = "/"
	case strings.HasSuffix(url, "/index.html"):
		url = strings.TrimSuffix(url, "index.html")
	default:
		url = strings.TrimSuffix(url, ".html")
	}

	return page.NavigationLink{
		Title: p.FrontMatter.Title,
		URL:   url,
	}
}

func copyStatic(sourceDir, outputDir string) error {
	if sourceDir == "" {
		return nil
	}

	info, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("check static directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", sourceDir)
	}

	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destination := filepath.Join(outputDir, rel)

		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read static file %s: %w", path, err)
		}

		if err := os.WriteFile(destination, source, 0o644); err != nil {
			return fmt.Errorf("write static file %s: %w", destination, err)
		}

		fmt.Printf("copied %s -> %s\n", path, destination)
		return nil
	})
}
