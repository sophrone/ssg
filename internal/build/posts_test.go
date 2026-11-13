package build_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sophrone/ssg/internal/build"
)

func TestPostsIndex(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	layoutPath := filepath.Join(sourceDir, "layout.html")
	configPath := filepath.Join(sourceDir, "config.yml")
	postsDir := filepath.Join(sourceDir, "posts")

	if err := os.MkdirAll(postsDir, 0o755); err != nil {
		t.Fatalf("create posts directory: %v", err)
	}

	layout := `<!doctype html>
<html>
  <head>
    <title>{{ .Title }}</title>
  </head>
  <body>
    <main>
      {{ .Content }}

      {{ if .Posts }}
      <ul>
        {{ range .Posts }}
        <li><a href="{{ .URL }}">{{ .Title }}</a></li>
        {{ end }}
      </ul>
      {{ end }}
    </main>
  </body>
</html>`

	if err := os.WriteFile(layoutPath, []byte(layout), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}

	configYAML := `title: "Test Site"
base_url: "https://example.com"

navigation:
  - title: "Home"
    url: "/"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config.yml: %v", err)
	}

	firstPost := `---
title: "First Post"
description: "The first post"
date: 2026-11-01
---

# First Post
`

	secondPost := `---
title: "Second Post"
description: "The second post"
date: 2026-11-10
---

# Second Post
`

	if err := os.WriteFile(filepath.Join(postsDir, "first-post.md"), []byte(firstPost), 0o644); err != nil {
		t.Fatalf("write first post: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postsDir, "second-post.md"), []byte(secondPost), 0o644); err != nil {
		t.Fatalf("write second post: %v", err)
	}

	opts := build.Options{
		SourceDir:  sourceDir,
		OutputDir:  outputDir,
		Layout:     layoutPath,
		StaticDir:  "",
		ConfigPath: configPath,
		PostsDir:   postsDir,
	}

	if err := build.Run(opts); err != nil {
		t.Fatalf("build.Run() error = %v", err)
	}

	postsHTML, err := os.ReadFile(filepath.Join(outputDir, "posts.html"))
	if err != nil {
		t.Fatalf("read generated posts.html: %v", err)
	}

	html := string(postsHTML)

	if !strings.Contains(html, `<a href="/posts/second-post">Second Post</a>`) {
		t.Errorf("posts.html does not contain second post link: %q", html)
	}

	if !strings.Contains(html, `<a href="/posts/first-post">First Post</a>`) {
		t.Errorf("posts.html does not contain first post link: %q", html)
	}

	secondIndex := strings.Index(html, "Second Post")
	firstIndex := strings.Index(html, "First Post")

	if secondIndex == -1 || firstIndex == -1 {
		t.Fatalf("posts.html does not contain both post titles: %q", html)
	}

	if secondIndex > firstIndex {
		t.Errorf("posts are not sorted newest-first: %q", html)
	}
}
