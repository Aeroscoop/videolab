package export

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const MinClipDuration = 0.05
const MaxClips = 64

type Clip struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func (o Options) validateTimeline() error {
	if !finite(o.TrimStart) || o.TrimStart < 0 || (o.TrimEnd != nil &&
		(!finite(*o.TrimEnd) || *o.TrimEnd-o.TrimStart+0.000001 < MinClipDuration)) {
		return errors.New("ongeldige trimgrenzen")
	}
	if o.Speed != 0 && (!finite(o.Speed) || o.Speed < 0.5 || o.Speed > 2) {
		return errors.New("afspeelsnelheid moet tussen 0,5× en 2× liggen")
	}
	if len(o.Clips) > MaxClips {
		return fmt.Errorf("maximaal %d clips", MaxClips)
	}
	previousEnd := 0.0
	for _, clip := range o.Clips {
		if !finite(clip.Start) || !finite(clip.End) || clip.Start < previousEnd ||
			clip.End-clip.Start+0.000001 < MinClipDuration {
			return errors.New("ongeldige of overlappende clips")
		}
		previousEnd = clip.End
	}
	return nil
}

func (o Options) timelineGraph(hasAudio bool) string {
	var graph []string
	var inputs strings.Builder
	for i, clip := range o.Clips {
		bounds := "start=" + f(clip.Start) + ":end=" + f(clip.End)
		graph = append(graph, fmt.Sprintf("[0:v:0]trim=%s,setpts=PTS-STARTPTS[v%d]", bounds, i))
		fmt.Fprintf(&inputs, "[v%d]", i)
		if hasAudio {
			graph = append(graph, fmt.Sprintf("[0:a:0]atrim=%s,asetpts=PTS-STARTPTS[a%d]", bounds, i))
			fmt.Fprintf(&inputs, "[a%d]", i)
		}
	}
	audioCount, output := 0, "[timeline]"
	if hasAudio {
		audioCount, output = 1, "[timeline][timelineaudio]"
	}
	graph = append(graph, fmt.Sprintf("%sconcat=n=%d:v=1:a=%d%s", inputs.String(), len(o.Clips), audioCount, output))
	return strings.Join(graph, ";")
}
