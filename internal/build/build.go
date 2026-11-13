package build

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	PostsDir   string
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

	postsDirName := filepath.Base(opts.PostsDir)

	var pages []*page.Page
	var posts []*page.Page

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

		if isPost(rel, postsDirName) {
			posts = append(posts, p)
		}

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

	if err := writePostsIndex(layout, posts, opts.OutputDir, config); err != nil {
		return fmt.Errorf("write posts index: %w", err)
	}

	return nil
}

func isPost(relativePath, postsDirName string) bool {
	return strings.HasPrefix(filepath.ToSlash(relativePath), postsDirName+"/")
}

func writePostsIndex(layout *template.Template, posts []*page.Page, outputDir string, config *config.Config) error {
	summaries := make([]page.PostSummary, 0, len(posts))

	for _, post := range posts {
		summaries = append(summaries, postSummary(post, outputDir))
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Date.After(summaries[j].Date)
	})

	data := struct {
		Title       string
		Description string
		Content     template.HTML
		Navigation  []page.NavigationLink
		SiteTitle   string
		Posts       []page.PostSummary
	}{
		Title:       "Posts",
		Description: "All blog posts",
		Content:     template.HTML("<h1>Posts</h1>"),
		Navigation:  config.Navigation,
		SiteTitle:   config.Title,
		Posts:       summaries,
	}

	outputPath := filepath.Join(outputDir, "posts.html")

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

	if err := file.Close(); err != nil {
		return err
	}

	fmt.Printf("built posts index -> %s\n", outputPath)
	return nil
}

func postSummary(post *page.Page, outputDir string) page.PostSummary {
	rel, err := filepath.Rel(outputDir, post.OutputPath)
	if err != nil {
		return page.PostSummary{
			Title:       post.FrontMatter.Title,
			Description: post.FrontMatter.Description,
			Date:        post.FrontMatter.Date,
		}
	}

	url := "/" + filepath.ToSlash(strings.TrimSuffix(rel, filepath.Ext(rel)))

	return page.PostSummary{
		Title:       post.FrontMatter.Title,
		Description: post.FrontMatter.Description,
		URL:         url,
		Date:        post.FrontMatter.Date,
		DateText:    post.FrontMatter.Date.Format("January 2, 2006"),
	}
}

func renderPage(layout *template.Template, p *page.Page, outputPath string, siteTitle string) error {
	data := struct {
		Title       string
		Description string
		Content     template.HTML
		Navigation  []page.NavigationLink
		SiteTitle   string
		Posts       []page.PostSummary
	}{
		Title:       p.FrontMatter.Title,
		Description: p.FrontMatter.Description,
		Content:     template.HTML(p.HTML),
		Navigation:  p.Navigation,
		SiteTitle:   siteTitle,
		Posts:       []page.PostSummary{},
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
