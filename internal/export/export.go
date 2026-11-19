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
	Input       string
	Output      string
	Format      string   // "4:5", "9:16", "16:9", "1:1"
	Background  string   // "#rrggbb"
	Scale       float64  // percent, 20..130 (as in project.js)
	X, Y        float64  // percent of canvas, -100..100
	TrimStart   float64  // seconds from the start of the source
	TrimEnd     *float64 // nil includes the rest of the source
	Clips       []Clip   // when set, supersedes single-clip trim
	Speed       float64  // zero defaults to normal playback
	NoOverwrite bool     // publish only if no destination exists
	FFmpeg      string   // optional path, default "ffmpeg"
	FFprobe     string   // optional path, default "ffprobe"
}

// Info is what the export learned about the source.
type Info struct {
	HasAudio bool
	FPS      float64
	Duration float64 // zero if unknown
}

func (o Options) ValidateLayout() error {
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
		if !finite(v) {
			return errors.New("ongeldige waarde")
		}
	}
	if o.Scale < 20 || o.Scale > 130 || o.X < -100 || o.X > 100 || o.Y < -100 || o.Y > 100 {
		return errors.New("ongeldige schaal of positie")
	}
	return o.validateTimeline()
}

func (o *Options) normalize() error {
	if err := o.ValidateLayout(); err != nil {
		return err
	}
	if o.Format == "" {
		o.Format = "4:5"
	}
	if o.Speed == 0 {
		o.Speed = 1
	}
	if len(o.Clips) > 0 {
		o.TrimStart, o.TrimEnd = 0, nil
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
	return layoutGraph(o, fps, "[0:v:0]"), nil
}

func layoutGraph(o Options, fps float64, input string) string {
	W, H := canvases[o.Format][0], canvases[o.Format][1]
	s := o.Scale / 100
	// Fit inside the canvas (like object-fit: contain), then apply the scale.
	// Dimensions are forced even for yuv420p. Commas are escaped for ffmpeg.
	scale := fmt.Sprintf("scale=w='max(2,trunc(iw*min(%d/iw\\,%d/ih)*%s/2)*2)':h='max(2,trunc(ih*min(%d/iw\\,%d/ih)*%s/2)*2)'",
		W, H, f(s), W, H, f(s))
	ox := fmt.Sprintf("(main_w-overlay_w)/2+main_w*%s/100", f(o.X))
	oy := fmt.Sprintf("(main_h-overlay_h)/2+main_h*%s/100", f(o.Y))
	bg := "0x" + strings.TrimPrefix(o.Background, "#")
	timing := ""
	if o.Speed != 1 {
		timing = "setpts=(PTS-STARTPTS)/" + f(o.Speed) + ","
	}
	return fmt.Sprintf("color=c=%s:s=%dx%d:r=%s[bg];%s%ssetsar=1,%s[v];[bg][v]overlay=x='%s':y='%s':shortest=1:eof_action=endall,format=yuv420p[out]",
		bg, W, H, f(fps), input, timing, scale, ox, oy)
}

func f(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// Probe reads audio presence, frame rate and duration of the source.
func Probe(ctx context.Context, ffprobe, input string) (Info, error) {
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	out, err := exec.CommandContext(ctx, ffprobe, "-v", "error", "-protocol_whitelist", "file,pipe", "-show_streams", "-show_format", "-of", "json", input).Output()
	if err != nil {
		return Info{}, fmt.Errorf("ffprobe mislukt: %w", err)
	}
	var data struct {
		Streams []struct {
			Type string `json:"codec_type"`
			Avg  string `json:"avg_frame_rate"`
			R    string `json:"r_frame_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return Info{}, err
	}
	duration, _ := strconv.ParseFloat(data.Format.Duration, 64)
	info := Info{FPS: 30, Duration: duration}
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
	if err := o.normalize(); err != nil {
		return nil, err
	}
	audio := "0:a:0"
	graph := layoutGraph(o, info.FPS, "[0:v:0]")
	if len(o.Clips) > 0 {
		graph = o.timelineGraph(info.HasAudio) + ";" + layoutGraph(o, info.FPS, "[timeline]")
		audio = "[timelineaudio]"
	}
	if info.HasAudio && o.Speed != 1 {
		if len(o.Clips) == 0 {
			audio = "[0:a:0]"
		}
		graph += ";" + audio + "atempo=" + f(o.Speed) + "[speedaudio]"
		audio = "[speedaudio]"
	}
	args := []string{"-hide_banner", "-v", "error", "-y"}
	if o.TrimStart > 0 {
		args = append(args, "-ss", f(o.TrimStart))
	}
	args = append(args, "-protocol_whitelist", "file,pipe", "-i", o.Input)
	if o.TrimEnd != nil {
		args = append(args, "-t", f((*o.TrimEnd-o.TrimStart)/o.Speed))
	}
	args = append(args, "-filter_complex", graph, "-map", "[out]")
	if info.HasAudio {
		args = append(args, "-map", audio)
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
	input, err := os.Stat(o.Input)
	if err != nil {
		return Info{}, fmt.Errorf("invoer niet leesbaar: %w", err)
	}
	if o.NoOverwrite {
		if _, err := os.Lstat(o.Output); err == nil {
			return Info{}, fmt.Errorf("uitvoer bestaat al: %w", os.ErrExist)
		} else if !os.IsNotExist(err) {
			return Info{}, err
		}
	}
	if output, err := os.Stat(o.Output); err == nil && os.SameFile(input, output) {
		return Info{}, errors.New("uitvoer mag niet gelijk zijn aan invoer")
	}
	info, err := Probe(ctx, o.FFprobe, o.Input)
	if err != nil {
		return Info{}, err
	}
	if info.Duration > 0 && (o.TrimStart >= info.Duration ||
		(o.TrimEnd != nil && *o.TrimEnd > info.Duration+0.05)) {
		return Info{}, errors.New("trimgrenzen vallen buiten de bronvideo")
	}
	if info.Duration > 0 && len(o.Clips) > 0 && o.Clips[len(o.Clips)-1].End > info.Duration+0.05 {
		return Info{}, errors.New("clip valt buiten de bronvideo")
	}
	dir, err := os.MkdirTemp(filepath.Dir(o.Output), ".videolab-export-")
	if err != nil {
		return Info{}, fmt.Errorf("tijdelijke uitvoer maken: %w", err)
	}
	defer os.RemoveAll(dir)
	tmp := filepath.Join(dir, "export.mp4")
	args, err := Args(o, info, tmp)
	if err != nil {
		return Info{}, err
	}
	cmd := exec.CommandContext(ctx, o.FFmpeg, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return Info{}, fmt.Errorf("ffmpeg mislukt: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if o.NoOverwrite {
		err = os.Link(tmp, o.Output)
	} else {
		err = os.Rename(tmp, o.Output)
	}
	if err != nil {
		return Info{}, err
	}
	return info, nil
}
