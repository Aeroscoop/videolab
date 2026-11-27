// Package project defines the JSON project consumed by VideoLab and its MCP server.
package project

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Aeroscoop/videolab/internal/export"
)

type Media struct {
	Name string `json:"name"`
	Size int64  `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

type Project struct {
	Version    int           `json:"version"`
	Name       string        `json:"name"`
	Format     string        `json:"format"`
	Background string        `json:"background"`
	Scale      float64       `json:"scale"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	Safe       bool          `json:"safe"`
	Media      *Media        `json:"media"`
	TrimStart  float64       `json:"trimStart"`
	TrimEnd    *float64      `json:"trimEnd"`
	Clips      []export.Clip `json:"clips"`
	Speed      float64       `json:"speed"`
	Overlay    string        `json:"overlay"`
}

// Decode supplies defaults for optional fields in previously saved projects.
func Decode(data []byte) (Project, error) {
	p := Project{
		Name: "Nieuw project", Background: "#a9c9b6",
		Scale: 80, Safe: true, Clips: []export.Clip{}, Speed: 1, Overlay: "general",
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return Project{}, fmt.Errorf("project lezen: %w", err)
	}
	if _, err := p.ExportOptions(); err != nil {
		return Project{}, err
	}
	return p, nil
}

func (p Project) ExportOptions() (export.Options, error) {
	_, _, supported := export.Canvas(p.Format)
	if p.Version != 1 || !supported {
		return export.Options{}, errors.New("ongeldig VideoLab-project")
	}
	if p.Speed == 0 {
		return export.Options{}, errors.New("ongeldige afspeelsnelheid")
	}
	switch p.Overlay {
	case "general", "youtube-vertical", "youtube-tablet", "youtube-wide", "youtube-desktop", "tiktok-feed", "linkedin":
	default:
		return export.Options{}, errors.New("ongeldige richtzone")
	}
	options := export.Options{
		Format: p.Format, Background: p.Background, Scale: p.Scale, X: p.X, Y: p.Y,
		TrimStart: p.TrimStart, TrimEnd: p.TrimEnd, Clips: p.Clips, Speed: p.Speed,
	}
	return options, options.ValidateLayout()
}
