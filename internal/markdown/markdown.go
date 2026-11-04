package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"

	"github.com/sophrone/ssg/internal/page"
)

func Parse(path string, source []byte) (*page.Page, error) {
	frontMatter, body, err := splitFrontMatter(source)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	var metadata page.FrontMatter
	if len(frontMatter) > 0 {
		if err := yaml.Unmarshal(frontMatter, &metadata); err != nil {
			return nil, fmt.Errorf("%s: invalid front matter: %w", path, err)
		}
	}

	var rendered bytes.Buffer
	if err := goldmark.Convert(body, &rendered); err != nil {
		return nil, fmt.Errorf("%s: markdown conversion failed: %w", path, err)
	}

	return &page.Page{
		SourcePath:  path,
		FrontMatter: metadata,
		Markdown:    body,
		HTML:        rendered.Bytes(),
	}, nil
}

func splitFrontMatter(source []byte) ([]byte, []byte, error) {
	text := string(source)

	if !strings.HasPrefix(text, "---\n") {
		return nil, source, nil
	}

	parts := strings.SplitN(text[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("front matter is not closed")
	}

	return []byte(parts[0]), []byte(parts[1]), nil
}
