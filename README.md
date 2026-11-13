# ssg

`ssg` is a small static site generator written in Go.

It converts Markdown files with YAML front matter into HTML pages using a shared layout template. It can copy static assets, generate blog posts and a posts index, serve the site locally, and reload the browser when content changes.

## Features

- Markdown to HTML conversion
- YAML front matter for page metadata
- Shared HTML layout template
- Static asset copying
- Draft page support
- Clean URLs
- Custom 404 page
- Site configuration
- Configurable navigation
- Blog posts with dates
- Generated posts index
- Posts sorted newest-first
- Local development server
- Browser live reload
- Tests for Markdown parsing, site building, and posts indexing

## Requirements

- Go 1.24 or newer

## Installation

```bash
git clone [https://github.com/sophrone/ssg.git](https://github.com/sophrone/ssg.git)
cd ssg
go build .
```

## Usage

### Build a site

```bash
make build
```

This is equivalent to:

```bash
go run . build \
  -src testdata \
  -out build \
  -layout testdata/layout.html \
  -static testdata/static \
  -config config.yml \
  -posts testdata/posts
```

The generated site is written to `build/`.

### Serve a site

```bash
make serve
```

Then open:

```text
http://localhost:8080
```

Editing a Markdown file, the layout, or a static asset rebuilds the site and reloads the browser.

## Site configuration

Create `config.yml` in the project root:

```yaml
title: "The Band"
base_url: "[https://example.com](https://example.com)"

navigation:
  - title: "The Band"
    url: "/"
  - title: "About"
    url: "/about"
  - title: "Posts"
    url: "/posts"
```

## Page format

Create a Markdown file such as `testdata/index.md`:

```markdown
***
title: "The Band"
description: "Official website for the band."
draft: false
***

# Welcome

This is the band's homepage.
```

### Front matter

| Field | Purpose |
|---|---|
| `title` | Page title |
| `description` | Meta description |
| `date` | Optional page date |
| `draft` | Set to `true` to exclude the page from the build |
| `hide_from_navigation` | Set to `true` to keep a page out of the navigation menu |

## Blog posts

Create Markdown files inside `testdata/posts/`:

```markdown
***
title: "First Post"
description: "Our first blog post."
date: 2026-11-01
***

# First Post

This is the first post.
```

The generator creates individual post pages and a `posts.html` index sorted newest-first.

## Project layout

```text
.
├── main.go
├── config.yml
├── Makefile
├── internal/
│   ├── build/
│   ├── config/
│   ├── markdown/
│   └── page/
└── testdata/
    ├── index.md
    ├── about.md
    ├── 404.md
    ├── layout.html
    ├── posts/
    └── static/
```

## Development

Format code:

```bash
make fmt
```

Run tests:

```bash
make test
```

Build the sample site:

```bash
make build
```

Remove generated output:

```bash
make clean
```