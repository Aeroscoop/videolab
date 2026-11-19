package export

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

const userClip = "/home/videolab/testmedia/test-4x5.mp4"

func need(t *testing.T) {
	t.Helper()
	for _, b := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(b); err != nil {
			t.Skip(b + " ontbreekt")
		}
	}
}

// makeClip creates a synthetic 1080x1350 clip with audio (or without).
func makeClip(t *testing.T, audio bool) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "src.mp4")
	args := []string{"-v", "error", "-y", "-f", "lavfi", "-i", "testsrc2=size=1080x1350:rate=30:duration=2"}
	if audio {
		args = append(args, "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:duration=2", "-c:a", "aac")
	}
	args = append(args, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-shortest", p)
	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("clip maken: %v %s", err, out)
	}
	return p
}

type probed struct {
	W, H     int
	Duration float64
	Video    string
	Audio    string
	NAudio   int
}

func probe(t *testing.T, p string) probed {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", p).Output()
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Streams []struct {
			Type, Codec string `json:"-"`
			CodecType   string `json:"codec_type"`
			CodecName   string `json:"codec_name"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			PixFmt      string `json:"pix_fmt"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &d); err != nil {
		t.Fatal(err)
	}
	var r probed
	r.Duration, _ = strconv.ParseFloat(d.Format.Duration, 64)
	for _, s := range d.Streams {
		if s.CodecType == "video" {
			r.W, r.H, r.Video = s.Width, s.Height, s.CodecName
		}
		if s.CodecType == "audio" {
			r.Audio = s.CodecName
			r.NAudio++
		}
	}
	return r
}

// pixel returns the RGB value at (x,y) of the first frame.
func pixel(t *testing.T, p string, x, y int) [3]int {
	t.Helper()
	out, err := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vf",
		"format=rgb24,crop=1:1:"+strconv.Itoa(x)+":"+strconv.Itoa(y), "-frames:v", "1", "-f", "rawvideo", "-pix_fmt", "rgb24", "-").Output()
	if err != nil || len(out) != 3 {
		t.Fatalf("pixel: %v len=%d", err, len(out))
	}
	return [3]int{int(out[0]), int(out[1]), int(out[2])}
}

func near(a, b [3]int, tol int) bool {
	for i := range a {
		if math.Abs(float64(a[i]-b[i])) > float64(tol) {
			return false
		}
	}
	return true
}

func run(t *testing.T, src string, scale, x, y float64) (string, Info) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out.mp4")
	info, err := Export(context.Background(), Options{Input: src, Output: out, Format: "4:5", Background: "#804020", Scale: scale, X: x, Y: y})
	if err != nil {
		t.Fatal(err)
	}
	return out, info
}

var red = [3]int{128, 64, 32} // achtergrond #804020, komt niet voor in testsrc2

func TestExportWithAudio_FullScale(t *testing.T) {
	need(t)
	out, info := run(t, makeClip(t, true), 100, 0, 0)
	if !info.HasAudio {
		t.Fatal("audio niet gedetecteerd")
	}
	p := probe(t, out)
	if p.W != 1080 || p.H != 1350 || p.Video != "h264" || p.Audio != "aac" || p.NAudio != 1 {
		t.Fatalf("onverwacht: %+v", p)
	}
	if math.Abs(p.Duration-2) > 0.2 {
		t.Fatalf("duur %v", p.Duration)
	}
	// Bij schaal 100 vult een 4:5-bron het hele kader: geen achtergrond in de hoek.
	if near(pixel(t, out, 4, 4), red, 40) {
		t.Fatal("hoek is achtergrondkleur bij schaal 100")
	}
}

func TestExportWithoutAudio(t *testing.T) {
	need(t)
	out, info := run(t, makeClip(t, false), 100, 0, 0)
	if info.HasAudio {
		t.Fatal("audio onterecht gedetecteerd")
	}
	p := probe(t, out)
	if p.W != 1080 || p.H != 1350 || p.Video != "h264" || p.NAudio != 0 {
		t.Fatalf("onverwacht: %+v", p)
	}
}

func TestSmallerScaleAndShift(t *testing.T) {
	need(t)
	src := makeClip(t, true)
	// Schaal 50: kader 540x675, gecentreerd: x 270..810, y 337..1012.
	small, _ := run(t, src, 50, 0, 0)
	if p := probe(t, small); p.W != 1080 || p.H != 1350 || p.NAudio != 1 {
		t.Fatalf("onverwacht: %+v", p)
	}
	if !near(pixel(t, small, 10, 10), red, 20) {
		t.Fatal("hoek moet achtergrond zijn bij schaal 50")
	}
	if near(pixel(t, small, 300, 675), red, 30) {
		t.Fatal("(300,675) moet bronbeeld zijn bij schaal 50, x=0")
	}
	// X=+30 schuift 324 px naar rechts: x 594..1134. (300,675) wordt achtergrond.
	moved, _ := run(t, src, 50, 30, 0)
	if !near(pixel(t, moved, 300, 675), red, 20) {
		t.Fatal("(300,675) moet achtergrond zijn na X=+30")
	}
	if near(pixel(t, moved, 900, 675), red, 30) {
		t.Fatal("(900,675) moet bronbeeld zijn na X=+30")
	}
	// Y=-30 schuift 405 px omhoog: y -68..607. (540,900) wordt achtergrond.
	up, _ := run(t, src, 50, 0, -30)
	if !near(pixel(t, up, 540, 900), red, 20) {
		t.Fatal("(540,900) moet achtergrond zijn na Y=-30")
	}
	if near(pixel(t, up, 540, 300), red, 30) {
		t.Fatal("(540,300) moet bronbeeld zijn na Y=-30")
	}
}

func TestLandscapeSourceContain(t *testing.T) {
	need(t)
	src := filepath.Join(t.TempDir(), "land.mp4")
	if out, err := exec.Command("ffmpeg", "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc2=size=1280x720:rate=30:duration=1",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", src).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	out, _ := run(t, src, 100, 0, 0)
	// 16:9 in 4:5 -> 1080x608 gecentreerd (y 371..979): boven en onder achtergrond.
	if !near(pixel(t, out, 540, 100), red, 20) || !near(pixel(t, out, 540, 1250), red, 20) {
		t.Fatal("boven/onder moet achtergrond zijn")
	}
	if near(pixel(t, out, 540, 675), red, 30) {
		t.Fatal("midden moet bronbeeld zijn")
	}
}

func TestUserClip(t *testing.T) {
	need(t)
	if _, err := os.Stat(userClip); err != nil {
		t.Skip("testclip ontbreekt")
	}
	out, _ := run(t, userClip, 60, -20, 10)
	p := probe(t, out)
	if p.W != 1080 || p.H != 1350 || p.Video != "h264" || p.Audio != "aac" || math.Abs(p.Duration-3) > 0.2 {
		t.Fatalf("onverwacht: %+v", p)
	}
}

func TestValidation(t *testing.T) {
	base := Options{Input: "a.mp4", Output: "b.mp4", Background: "#a9c9b6", Scale: 80}
	bad := []func(*Options){
		func(o *Options) { o.Format = "3:2" },
		func(o *Options) { o.Background = "red" },
		func(o *Options) { o.Scale = 10 },
		func(o *Options) { o.Scale = 131 },
		func(o *Options) { o.X = 101 },
		func(o *Options) { o.Y = -101 },
		func(o *Options) { o.Scale = math.NaN() },
		func(o *Options) { o.Output = o.Input },
		func(o *Options) { o.Input = "" },
	}
	for i, m := range bad {
		o := base
		m(&o)
		if _, err := FilterGraph(o, 30); err == nil {
			t.Errorf("case %d: verwachtte fout", i)
		}
	}
	if _, err := FilterGraph(base, 30); err != nil {
		t.Fatal(err)
	}
}

func TestMissingInputLeavesNoOutput(t *testing.T) {
	need(t)
	out := filepath.Join(t.TempDir(), "o.mp4")
	_, err := Export(context.Background(), Options{Input: "/nope.mp4", Output: out, Background: "#000000", Scale: 80})
	if err == nil {
		t.Fatal("verwachtte fout")
	}
	if _, e := os.Stat(out); e == nil {
		t.Fatal("uitvoer bestaat")
	}
}

func TestTrimKeepsOnlySelectedRange(t *testing.T) {
	need(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "source.mp4")
	out := filepath.Join(dir, "trimmed.mp4")
	source := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=red:s=128x160:r=30:d=1",
		"-f", "lavfi", "-i", "color=blue:s=128x160:r=30:d=1",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:d=2",
		"-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0[v]",
		"-map", "[v]", "-map", "2:a", "-c:v", "libx264", "-c:a", "aac", src)
	if output, err := source.CombinedOutput(); err != nil {
		t.Fatalf("testbron maken: %v %s", err, output)
	}

	end := 1.7
	settings := Options{Input: src, Output: out, Format: "4:5", Background: "#a9c9b6",
		Scale: 100, TrimStart: 1.1, TrimEnd: &end}
	if _, err := Export(context.Background(), settings); err != nil {
		t.Fatal(err)
	}
	p := probe(t, out)
	if p.Video != "h264" || p.Audio != "aac" || math.Abs(p.Duration-0.6) > 0.05 {
		t.Fatalf("onverwachte trim-uitvoer: %+v", p)
	}
	first := pixel(t, out, 540, 675)
	if first[0] > 40 || first[2] < 180 {
		t.Fatalf("eerste frame moet uit het blauwe deel na 1 s komen: %v", first)
	}

	end = 3
	settings.Output = filepath.Join(dir, "out-of-range.mp4")
	if _, err := Export(context.Background(), settings); err == nil {
		t.Fatal("eindpunt voorbij de bron werd geaccepteerd")
	}
	if _, err := os.Stat(settings.Output); !os.IsNotExist(err) {
		t.Fatalf("ongeldige trim mag geen uitvoer achterlaten: %v", err)
	}
}

func TestTimelineSkipsMiddleAndChangesSpeed(t *testing.T) {
	need(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "source.mp4")
	out := filepath.Join(dir, "timeline.mp4")
	create := exec.Command("ffmpeg", "-v", "error",
		"-f", "lavfi", "-i", "color=red:s=128x160:r=30:d=1",
		"-f", "lavfi", "-i", "color=green:s=128x160:r=30:d=1",
		"-f", "lavfi", "-i", "color=blue:s=128x160:r=30:d=1",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:d=3",
		"-filter_complex", "[0:v][1:v][2:v]concat=n=3:v=1:a=0[v]",
		"-map", "[v]", "-map", "3:a", "-c:v", "libx264", "-c:a", "aac", src)
	if output, err := create.CombinedOutput(); err != nil {
		t.Fatalf("testbron maken: %v %s", err, output)
	}
	o := Options{Input: src, Output: out, Format: "4:5", Background: "#a9c9b6", Scale: 100,
		Clips: []Clip{{Start: 0, End: 1}, {Start: 2, End: 3}}, Speed: 2}
	if _, err := Export(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	p := probe(t, out)
	if p.Video != "h264" || p.Audio != "aac" || math.Abs(p.Duration-1) > 0.06 {
		t.Fatalf("onverwachte montage-uitvoer: %+v", p)
	}
	assertColor := func(at string, channel int) {
		t.Helper()
		b, err := exec.Command("ffmpeg", "-v", "error", "-ss", at, "-i", out,
			"-vf", "format=rgb24,crop=1:1:540:675", "-frames:v", "1",
			"-f", "rawvideo", "-pix_fmt", "rgb24", "-").Output()
		if err != nil || len(b) != 3 || b[channel] < 185 || b[1] > 40 {
			t.Fatalf("verkeerd frame op %s: %v, %v", at, b, err)
		}
	}
	assertColor("0.2", 0)
	assertColor("0.8", 2)
}
