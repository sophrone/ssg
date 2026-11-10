package page

import "time"

type FrontMatter struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	Draft       bool      `yaml:"draft"`
}

type Page struct {
	SourcePath  string
	OutputPath  string
	FrontMatter FrontMatter
	Markdown    []byte
	HTML        []byte
}
