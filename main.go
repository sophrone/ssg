package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/sophrone/ssg/internal/build"
)

type reloadServer struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newReloadServer() *reloadServer {
	return &reloadServer{
		clients: make(map[chan struct{}]struct{}),
	}
}

func (s *reloadServer) broadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		select {
		case client <- struct{}{}:
		default:
		}
	}
}

func (s *reloadServer) handleReload(w http.ResponseWriter, r *http.Request) {
	events := make(chan struct{})
	defer close(events)

	s.mu.Lock()
	s.clients[events] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, events)
		s.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-events:
			fmt.Fprint(w, "event: message\n")
			fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
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
	static := flags.String("static", "static", "static assets directory")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	opts := build.Options{
		SourceDir: *source,
		OutputDir: *output,
		Layout:    *layout,
		StaticDir: *static,
	}

	if err := build.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "build failed:", err)
		os.Exit(1)
	}
}

func serveCmd(args []string) {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)

	source := flags.String("src", ".", "source directory containing Markdown files")
	output := flags.String("out", "build", "directory to serve")
	layout := flags.String("layout", "layout.html", "HTML layout template")
	static := flags.String("static", "static", "static assets directory")
	addr := flags.String("addr", ":8080", "HTTP listen address")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	opts := build.Options{
		SourceDir: *source,
		OutputDir: *output,
		Layout:    *layout,
		StaticDir: *static,
	}

	if err := build.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "initial build failed:", err)
		os.Exit(1)
	}

	reload := newReloadServer()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(os.Stderr, "create watcher failed:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	watchPaths := []string{
		*source,
		*static,
		filepath.Dir(*layout),
	}

	for _, path := range watchPaths {
		if err := watcher.Add(path); err != nil {
			fmt.Fprintf(os.Stderr, "watch %s failed: %v\n", path, err)
			os.Exit(1)
		}
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					if err := build.Run(opts); err != nil {
						fmt.Fprintln(os.Stderr, "rebuild failed:", err)
						continue
					}

					fmt.Println("site rebuilt")
					reload.broadcast()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				fmt.Fprintln(os.Stderr, "watcher error:", err)
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir(*output)))
	http.HandleFunc("/__reload", reload.handleReload)

	fmt.Printf("serving %s at http://localhost%s\n", *output, *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`ssg is a static site generator.

Usage:
  ssg build [flags]   Build Markdown files into HTML
  ssg serve [flags]   Serve the generated site and rebuild on changes

Build flags:
  -src string      Source directory containing Markdown files
  -out string      Output directory
  -layout string   HTML layout template
  -static string   Static assets directory

Serve flags:
  -src string      Source directory containing Markdown files
  -out string      Directory to serve
  -layout string   HTML layout template
  -static string   Static assets directory
  -addr string     HTTP listen address`)
}
