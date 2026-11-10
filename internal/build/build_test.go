package build_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sophrone/ssg/internal/build"
)

func TestRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	layoutPath := filepath.Join(sourceDir, "layout.html")

	layout := `<!doctype html>
<html>
  <head>
    <title>{{ .Title }}</title>
  </head>
  <body>
    <main>{{ .Content }}</main>
  </body>
</html>`

	if err := os.WriteFile(layoutPath, []byte(layout), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}

	index := `---
title: "Home"
description: "The homepage"
---

# Welcome

This is the homepage.`

	about := `---
title: "About"
description: "About page"
draft: true
---

## About

This page should not be generated.`

	if err := os.WriteFile(filepath.Join(sourceDir, "index.md"), []byte(index), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sourceDir, "about.md"), []byte(about), 0o644); err != nil {
		t.Fatalf("write about.md: %v", err)
	}

	opts := build.Options{
		SourceDir: sourceDir,
		OutputDir: outputDir,
		Layout:    layoutPath,
	}

	if err := build.Run(opts); err != nil {
		t.Fatalf("build.Run() error = %v", err)
	}

	indexHTML, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("read generated index.html: %v", err)
	}

	if !strings.Contains(string(indexHTML), "<title>Home</title>") {
		t.Errorf("index.html does not contain expected title: %q", string(indexHTML))
	}

	if !strings.Contains(string(indexHTML), "<h1>Welcome</h1>") {
		t.Errorf("index.html does not contain rendered Markdown: %q", string(indexHTML))
	}

	aboutPath := filepath.Join(outputDir, "about.html")

	if _, err := os.Stat(aboutPath); !os.IsNotExist(err) {
		t.Errorf("draft page should not be generated, but %s exists", aboutPath)
	}
}
