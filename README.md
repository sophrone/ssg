# ssg

`ssg` is a small static site generator written in Go.

It converts Markdown files with YAML front matter into HTML pages using a shared layout template. It can also copy static assets and serve the generated site locally with live reload.

## Features

- Markdown to HTML conversion
- YAML front matter for page metadata
- Shared HTML layout template
- Static asset copying
- Draft page support
- Local development server
- Browser live reload

## Installation

Requires Go 1.24 or newer.

```bash
git clone [https://github.com/sophrone/ssg.git](https://github.com/sophrone/ssg.git)
cd ssg
go build .
```

## Usage

### Build a site

```bash
go run . build \
  -src testdata \
  -out build \
  -layout testdata/layout.html \
  -static testdata/static
```

This generates HTML files in `build/`.

### Serve a site

```bash
go run . serve \
  -src testdata \
  -out build \
  -layout testdata/layout.html \
  -static testdata/static \
  -addr :8080
```

Then open:

```text
http://localhost:8080
```

Editing a Markdown file, the layout, or a static asset rebuilds the site and reloads the browser.

## Page format

Create a Markdown file such as `index.md`:

```markdown
***
title: "The Band"
description: "Official website for the band."
draft: false
***

# Welcome

This is the band's homepage.
```

The front matter supports:

| Field | Purpose |
|---|---|
| `title` | Page title |
| `description` | Meta description |
| `date` | Optional page date |
| `draft` | Set to `true` to exclude the page from the build |

## Project layout

```text
.
├── main.go
├── internal/
│   ├── build/
│   ├── markdown/
│   └── page/
└── testdata/
    ├── index.md
    ├── about.md
    ├── layout.html
    └── static/
```

## Tests

```bash
go test ./...
```