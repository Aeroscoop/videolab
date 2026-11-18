// Package export renders one clip to an MP4 using the same layout rules as the
// preview in frontend/js/app.js: the clip is fitted (contain) inside the canvas,
// scaled by Scale percent around its centre, then shifted by X/Y percent of the
// canvas width/height, on top of a solid background colour. Anything outside
// the canvas is cropped. It depends only on the standard library and ffmpeg.
package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Canvas sizes per project format (matches the preview aspect ratios).
var canvases = map[string][2]int{
	"4:5":  {1080, 1350},
	"9:16": {1080, 1920},
	"16:9": {1920, 1080},
	"1:1":  {1080, 1080},
}

var colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Options describes one export. Zero Format means "4:5".
type Options struct {
	Input      string
	Output     string
	Format     string  // "4:5", "9:16", "16:9", "1:1"
	Background string  // "#rrggbb"
	Scale      float64 // percent, 20..130 (as in project.js)
	X, Y       float64 // percent of canvas, -100..100
	FFmpeg     string  // optional path, default "ffmpeg"
	FFprobe    string  // optional path, default "ffprobe"
}

// Info is what the export learned about the source.
type Info struct {
	HasAudio bool
	FPS      float64
}

func (o *Options) normalize() error {
	if o.Format == "" {
		o.Format = "4:5"
	}
	if _, ok := canvases[o.Format]; !ok {
		return fmt.Errorf("ongeldig formaat %q", o.Format)
	}
	if !colorRe.MatchString(o.Background) {
		return errors.New("ongeldige kleur")
	}
	for _, v := range []float64{o.Scale, o.X, o.Y} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("ongeldige waarde")
		}
	}
	if o.Scale < 20 || o.Scale > 130 || o.X < -100 || o.X > 100 || o.Y < -100 || o.Y > 100 {
		return errors.New("ongeldige schaal of positie")
	}
	if o.Input == "" || o.Output == "" {
		return errors.New("invoer en uitvoer zijn verplicht")
	}
	if filepath.Clean(o.Input) == filepath.Clean(o.Output) {
		return errors.New("uitvoer mag niet gelijk zijn aan invoer")
	}
	if o.FFmpeg == "" {
		o.FFmpeg = "ffmpeg"
	}
	if o.FFprobe == "" {
		o.FFprobe = "ffprobe"
	}
	return nil
}

// Canvas returns the output size for a format.
func Canvas(format string) (w, h int, ok bool) {
	c, ok := canvases[format]
	return c[0], c[1], ok
}

// FilterGraph builds the filtergraph. hasAudio is not needed here; fps is the
// frame rate of the background layer.
func FilterGraph(o Options, fps float64) (string, error) {
	if err := o.normalize(); err != nil {
		return "", err
	}
	W, H := canvases[o.Format][0], canvases[o.Format][1]
	s := o.Scale / 100
	// Fit inside the canvas (like object-fit: contain), then apply the scale.
	// Dimensions are forced even for yuv420p. Commas are escaped for ffmpeg.
	scale := fmt.Sprintf("scale=w='max(2,trunc(iw*min(%d/iw\\,%d/ih)*%s/2)*2)':h='max(2,trunc(ih*min(%d/iw\\,%d/ih)*%s/2)*2)'",
		W, H, f(s), W, H, f(s))
	ox := fmt.Sprintf("(main_w-overlay_w)/2+main_w*%s/100", f(o.X))
	oy := fmt.Sprintf("(main_h-overlay_h)/2+main_h*%s/100", f(o.Y))
	bg := "0x" + strings.TrimPrefix(o.Background, "#")
	return fmt.Sprintf("color=c=%s:s=%dx%d:r=%s[bg];[0:v:0]setsar=1,%s[v];[bg][v]overlay=x='%s':y='%s':shortest=1:eof_action=endall,format=yuv420p[out]",
		bg, W, H, f(fps), scale, ox, oy), nil
}

func f(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// Probe reads audio presence and frame rate of the source.
func Probe(ctx context.Context, ffprobe, input string) (Info, error) {
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	out, err := exec.CommandContext(ctx, ffprobe, "-v", "error", "-show_streams", "-of", "json", input).Output()
	if err != nil {
		return Info{}, fmt.Errorf("ffprobe mislukt: %w", err)
	}
	var data struct {
		Streams []struct {
			Type string `json:"codec_type"`
			Avg  string `json:"avg_frame_rate"`
			R    string `json:"r_frame_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return Info{}, err
	}
	info := Info{FPS: 30}
	hasVideo := false
	for _, s := range data.Streams {
		switch s.Type {
		case "audio":
			info.HasAudio = true
		case "video":
			if !hasVideo {
				hasVideo = true
				if v := parseRate(s.Avg); v > 0 {
					info.FPS = v
				} else if v := parseRate(s.R); v > 0 {
					info.FPS = v
				}
			}
		}
	}
	if !hasVideo {
		return Info{}, errors.New("geen videostream in invoer")
	}
	if info.FPS > 60 {
		info.FPS = 60
	}
	return info, nil
}

func parseRate(s string) float64 {
	a, b, ok := strings.Cut(s, "/")
	if !ok {
		return 0
	}
	n, e1 := strconv.ParseFloat(a, 64)
	d, e2 := strconv.ParseFloat(b, 64)
	if e1 != nil || e2 != nil || d == 0 {
		return 0
	}
	return n / d
}

// Args builds the ffmpeg arguments for a given source description.
func Args(o Options, info Info, tmpOut string) ([]string, error) {
	graph, err := FilterGraph(o, info.FPS)
	if err != nil {
		return nil, err
	}
	args := []string{"-hide_banner", "-v", "error", "-y", "-i", o.Input,
		"-filter_complex", graph, "-map", "[out]"}
	if info.HasAudio {
		args = append(args, "-map", "0:a:0")
	}
	args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "18", "-pix_fmt", "yuv420p")
	if info.HasAudio {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	} else {
		args = append(args, "-an")
	}
	args = append(args, "-movflags", "+faststart", "-f", "mp4", tmpOut)
	return args, nil
}

// Export renders the clip. It writes to a temporary file next to Output and
// renames it on success, so a failed export never leaves a broken Output.
func Export(ctx context.Context, o Options) (Info, error) {
	if err := o.normalize(); err != nil {
		return Info{}, err
	}
	if _, err := os.Stat(o.Input); err != nil {
		return Info{}, fmt.Errorf("invoer niet leesbaar: %w", err)
	}
	info, err := Probe(ctx, o.FFprobe, o.Input)
	if err != nil {
		return Info{}, err
	}
	tmp := o.Output + ".part"
	args, err := Args(o, info, tmp)
	if err != nil {
		return Info{}, err
	}
	cmd := exec.CommandContext(ctx, o.FFmpeg, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmp)
		return Info{}, fmt.Errorf("ffmpeg mislukt: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if err := os.Rename(tmp, o.Output); err != nil {
		os.Remove(tmp)
		return Info{}, err
	}
	return info, nil
}
