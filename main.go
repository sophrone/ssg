package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/sophrone/ssg/internal/build"
)

func main() {
	if len(os.Args) < 2 {
		usage()

	}

	switch os.Args[1] {
	case "build":
		buildCmd(os.Args[2:])
	case "serve":
		serveCmd(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func buildCmd(args []string) {
	flags := flag.NewFlagSet("build", flag.ExitOnError)

	source := flags.String("src", ".", "source directory containing Markdown files")
	output := flags.String("out", "build", "output directory")
	layout := flags.String("layout", "layout.html", "HTML layout template")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	opts := build.Options{
		SourceDir: *source,
		OutputDir: *output,
		Layout:    *layout,
	}

	if err := build.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "build failed:", err)
		os.Exit(1)
	}
}

func serveCmd(args []string) {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)

	output := flags.String("out", "build", "directory to serve")
	addr := flags.String("addr", ":8080", "HTTP listen address")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	fmt.Printf("serving %s at http://localhost%s\n", *output, *addr)

	if err := http.ListenAndServe(*addr, http.FileServer(http.Dir(*output))); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`ssg is a static site generator.

Usage:
  ssg build [flags]   Build Markdown files into HTML
  ssg serve [flags]   Serve the generated site

Build flags:
  -src string     Source directory containing Markdown files
  -out string     Output directory
  -layout string  HTML layout template

Serve flags:
  -out string   Directory to serve
  -addr string  HTTP listen address`)
}
