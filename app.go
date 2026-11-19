package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Aeroscoop/videolab/internal/export"
	"github.com/Aeroscoop/videolab/internal/project"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	mu            sync.RWMutex
	sourcePath    string
	sourceName    string
	sourceTempDir string
	sourceURL     string
	previewOrigin string
	previewServer *http.Server
	previewError  error
}

type VideoSource struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ExportSettings struct {
	SourceURL string          `json:"sourceURL"`
	Project   project.Project `json:"project"`
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// WebKit's custom asset scheme does not handle media seeking. Serve only
	// the current file at an unguessable URL on the local loopback interface.
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		a.previewError = err
		return
	}
	a.previewOrigin = "http://" + listener.Addr().String()
	capture, err := fs.Sub(assets, "frontend/capture")
	if err != nil {
		listener.Close()
		a.previewError = err
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", a.servePreview)
	captureFiles := http.StripPrefix("/capture/", http.FileServerFS(capture))
	mux.HandleFunc("/capture/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; media-src 'self' blob:; frame-ancestors 'none'")
		w.Header().Set("Permissions-Policy", "display-capture=(self)")
		captureFiles.ServeHTTP(w, r)
	})
	a.previewServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go a.previewServer.Serve(listener)
}

func (a *App) shutdown(context.Context) {
	a.ClearVideo()
	if a.previewServer != nil {
		a.previewServer.Close()
	}
}
func (a *App) Version() string { return "phase-5-dev" }

func (a *App) OpenCapturePage() error {
	if a.previewError != nil {
		return fmt.Errorf("opnamepagina starten: %w", a.previewError)
	}
	runtime.BrowserOpenURL(a.ctx, a.previewOrigin+"/capture/")
	return nil
}

func (a *App) OpenVideo() (*VideoSource, error) {
	if a.previewError != nil {
		return nil, fmt.Errorf("preview starten: %w", a.previewError)
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Video importeren",
		Filters: []runtime.FileFilter{
			{DisplayName: "Videobestanden", Pattern: "*.mp4;*.mov;*.m4v;*.webm;*.mkv;*.avi;*.MP4;*.MOV;*.M4V;*.WEBM;*.MKV;*.AVI"},
			{DisplayName: "Alle bestanden", Pattern: "*"},
		},
	})
	if err != nil || path == "" {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("video openen: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("kies een videobestand")
	}
	source, scratch, err := a.prepareVideo(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sourceTempDir != "" {
		os.RemoveAll(a.sourceTempDir)
	}
	a.sourcePath, a.sourceName, a.sourceTempDir = source, info.Name(), scratch
	a.sourceURL = a.previewOrigin + "/preview/" + rand.Text()
	return &VideoSource{
		Name: info.Name(),
		Size: info.Size(),
		Type: mime.TypeByExtension(strings.ToLower(filepath.Ext(path))),
		URL:  a.sourceURL,
	}, nil
}

// Normalize browser recordings for the Linux WebKit media backend. The
// browser file stays intact and both preview and export use the same copy.
func (a *App) prepareVideo(path string) (source, scratch string, err error) {
	if !strings.EqualFold(filepath.Ext(path), ".webm") {
		return path, "", nil
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", "", errors.New("installeer FFmpeg om deze WebM-opname te importeren")
	}
	dir, err := os.MkdirTemp("", "videolab-webm-")
	if err != nil {
		return "", "", err
	}
	source = filepath.Join(dir, "recording.mp4")
	cmd := exec.CommandContext(a.ctx, ffmpeg, "-v", "error", "-protocol_whitelist", "file,pipe",
		"-i", path, "-map", "0:v:0", "-map", "0:a:0?",
		"-vf", "fps=30,pad=ceil(iw/2)*2:ceil(ih/2)*2", "-c:v", "libx264", "-preset", "veryfast",
		"-crf", "18", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", "-movflags", "+faststart", source)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		return "", "", fmt.Errorf("WebM voorbereiden voor de preview: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return source, dir, nil
}

func (a *App) ClearVideo() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sourceTempDir != "" {
		os.RemoveAll(a.sourceTempDir)
	}
	a.sourcePath, a.sourceName, a.sourceURL, a.sourceTempDir = "", "", "", ""
}

func (a *App) sourceFor(url string) (path, name string) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if url == a.sourceURL {
		return a.sourcePath, a.sourceName
	}
	return "", ""
}

func (a *App) servePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path, _ := a.sourceFor(a.previewOrigin + r.URL.Path)
	if path == "" {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (a *App) ExportMP4(settings ExportSettings) (string, error) {
	options, err := settings.Project.ExportOptions()
	if err != nil {
		return "", err
	}
	source, name := a.sourceFor(settings.SourceURL)
	if source == "" {
		return "", errors.New("importeer eerst een video")
	}
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			return "", fmt.Errorf("%s ontbreekt. Installeer FFmpeg om MP4 te exporteren", tool)
		}
	}
	output, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "MP4 exporteren",
		DefaultFilename:      strings.TrimSuffix(name, filepath.Ext(name)) + "-export.mp4",
		CanCreateDirectories: true,
		Filters:              []runtime.FileFilter{{DisplayName: "MP4-video", Pattern: "*.mp4;*.MP4"}},
	})
	if err != nil || output == "" {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(output), ".mp4") {
		return "", errors.New("gebruik een bestandsnaam die eindigt op .mp4")
	}
	options.Input, options.Output = source, output
	if _, err := export.Export(a.ctx, options); err != nil {
		return "", err
	}
	return filepath.Base(output), nil
}
