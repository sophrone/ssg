package markdown_test

import (
	"strings"
	"testing"

	"github.com/sophrone/ssg/internal/markdown"
)

func TestParse(t *testing.T) {
	source := []byte(`---
title: "Home"
description: "A homepage"
---

# Hello

This is **bold**.
`)

	p, err := markdown.Parse("index.md", source)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if p.FrontMatter.Title != "Home" {
		t.Errorf("Title = %q, want %q", p.FrontMatter.Title, "Home")
	}

	if p.FrontMatter.Description != "A homepage" {
		t.Errorf("Description = %q, want %q", p.FrontMatter.Description, "A homepage")
	}

	if !strings.Contains(string(p.HTML), "<h1>Hello</h1>") {
		t.Errorf("HTML does not contain rendered heading: %q", string(p.HTML))
	}

	if !strings.Contains(string(p.HTML), "<strong>bold</strong>") {
		t.Errorf("HTML does not contain rendered bold text: %q", string(p.HTML))
	}
}

func TestParseWithoutFrontMatter(t *testing.T) {
	source := []byte("# Hello\n\nNo metadata here.")

	p, err := markdown.Parse("plain.md", source)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if p.FrontMatter.Title != "" {
		t.Errorf("Title = %q, want empty string", p.FrontMatter.Title)
	}

	if !strings.Contains(string(p.HTML), "<h1>Hello</h1>") {
		t.Errorf("HTML does not contain rendered heading: %q", string(p.HTML))
	}
}

func TestParseUnclosedFrontMatter(t *testing.T) {
	source := []byte(`---
title: "Broken"

# Hello
`)

	if _, err := markdown.Parse("broken.md", source); err == nil {
		t.Fatal("Parse() expected error for unclosed front matter")
	}
}
