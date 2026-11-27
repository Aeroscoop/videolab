// Command videolab-mcp serves local VideoLab tools over MCP stdio.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/Aeroscoop/videolab/internal/export"
	"github.com/Aeroscoop/videolab/internal/project"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type projectInput struct {
	Project string `json:"project" jsonschema:"Path to a VideoLab JSON project relative to the configured folder."`
}

type probeInput struct {
	Input string `json:"input" jsonschema:"Path to a video relative to the configured folder."`
}

type renderInput struct {
	Project string `json:"project" jsonschema:"Path to a VideoLab JSON project relative to the configured folder."`
	Input   string `json:"input" jsonschema:"Path to the source video relative to the configured folder."`
	Output  string `json:"output" jsonschema:"Path to a new MP4 file relative to the configured folder."`
}

type projectResult struct {
	Format string  `json:"format"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Clips  int     `json:"clips"`
	Speed  float64 `json:"speed"`
}

type videoResult struct {
	Duration float64 `json:"durationSeconds"`
	FPS      float64 `json:"fps"`
	HasAudio bool    `json:"hasAudio"`
}

type renderResult struct {
	Output string `json:"output"`
}

type workspace string

func main() {
	root := flag.String("root", "", "Folder containing the VideoLab projects and videos to work with")
	flag.Parse()
	if *root == "" {
		log.Fatal("use -root to specify your VideoLab project folder")
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		log.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		log.Fatal(err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		log.Fatal(err)
	}
	if !info.IsDir() {
		log.Fatal("-root must be a folder")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := newServer(workspace(resolved)).Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func newServer(w workspace) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "videolab", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name: "validate_project", Description: "Validate a saved VideoLab project and report its export settings.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, w.validate)
	mcp.AddTool(server, &mcp.Tool{
		Name: "probe_video", Description: "Read duration, frame rate and audio availability from a local video.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, w.probe)
	falseHint := false
	mcp.AddTool(server, &mcp.Tool{
		Name: "export_mp4", Description: "Render a local video with a saved VideoLab project to a new MP4. Existing files are preserved.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &falseHint, OpenWorldHint: &falseHint},
	}, w.render)
	return server
}

func (w workspace) resolve(name string) (string, error) {
	if !filepath.IsLocal(name) {
		return "", errors.New("use a relative path inside the configured folder")
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(string(w), name))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(string(w), resolved)
	if err != nil || !filepath.IsLocal(rel) {
		return "", errors.New("path resolves outside the configured folder")
	}
	return resolved, nil
}

func (w workspace) file(name string) (string, error) {
	path, err := w.resolve(name)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%q is not a regular file", name)
	}
	return path, nil
}

func (w workspace) readProject(name string) (project.Project, error) {
	path, err := w.file(name)
	if err != nil {
		return project.Project{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return project.Project{}, err
	}
	defer f.Close()
	const maxProjectSize = 1 << 20
	data, err := io.ReadAll(io.LimitReader(f, maxProjectSize+1))
	if err != nil {
		return project.Project{}, err
	}
	if len(data) > maxProjectSize {
		return project.Project{}, errors.New("project JSON exceeds 1 MiB")
	}
	return project.Decode(data)
}

func (w workspace) validate(_ context.Context, _ *mcp.CallToolRequest, in projectInput) (*mcp.CallToolResult, projectResult, error) {
	p, err := w.readProject(in.Project)
	if err != nil {
		return nil, projectResult{}, err
	}
	width, height, _ := export.Canvas(p.Format)
	return nil, projectResult{
		Format: p.Format, Width: width, Height: height,
		Clips: max(1, len(p.Clips)), Speed: p.Speed,
	}, nil
}

func (w workspace) probe(ctx context.Context, _ *mcp.CallToolRequest, in probeInput) (*mcp.CallToolResult, videoResult, error) {
	path, err := w.file(in.Input)
	if err != nil {
		return nil, videoResult{}, err
	}
	info, err := export.Probe(ctx, "", path)
	return nil, videoResult{Duration: info.Duration, FPS: info.FPS, HasAudio: info.HasAudio}, err
}

func (w workspace) render(ctx context.Context, _ *mcp.CallToolRequest, in renderInput) (*mcp.CallToolResult, renderResult, error) {
	p, err := w.readProject(in.Project)
	if err != nil {
		return nil, renderResult{}, err
	}
	if !filepath.IsLocal(in.Output) || !strings.EqualFold(filepath.Ext(in.Output), ".mp4") {
		return nil, renderResult{}, errors.New("choose a new relative output path ending in .mp4")
	}
	parent, err := w.resolve(filepath.Dir(in.Output))
	if err != nil {
		return nil, renderResult{}, err
	}
	options, err := p.ExportOptions()
	if err != nil {
		return nil, renderResult{}, err
	}
	options.Input, err = w.file(in.Input)
	if err != nil {
		return nil, renderResult{}, err
	}
	options.Output = filepath.Join(parent, filepath.Base(in.Output))
	options.NoOverwrite = true
	if _, err := export.Export(ctx, options); err != nil {
		return nil, renderResult{}, err
	}
	return nil, renderResult{Output: filepath.Clean(in.Output)}, nil
}
