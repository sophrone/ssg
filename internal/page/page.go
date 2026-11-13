package page

import "time"

type FrontMatter struct {
	Title              string    `yaml:"title"`
	Description        string    `yaml:"description"`
	Date               time.Time `yaml:"date"`
	Draft              bool      `yaml:"draft"`
	HideFromNavigation bool      `yaml:"hide_from_navigation"`
}

type Page struct {
	SourcePath  string
	OutputPath  string
	FrontMatter FrontMatter
	Markdown    []byte
	HTML        []byte
	Navigation  []NavigationLink
}

type NavigationLink struct {
	Title string
	URL   string
}

type PostSummary struct {
	Title       string
	Description string
	URL         string
	Date        time.Time
	DateText    string
}
