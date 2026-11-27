// Command videolab-export renders one clip to MP4 without Wails.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Aeroscoop/videolab/internal/export"
)

func main() {
	var o export.Options
	flag.StringVar(&o.Input, "in", "", "invoerbestand")
	flag.StringVar(&o.Output, "out", "", "uitvoerbestand (.mp4)")
	flag.StringVar(&o.Format, "format", "4:5", "4:5, 9:16, 16:9 of 1:1")
	flag.StringVar(&o.Background, "bg", "#a9c9b6", "achtergrondkleur #rrggbb")
	flag.Float64Var(&o.Scale, "scale", 80, "schaal in procent (20-130)")
	flag.Float64Var(&o.X, "x", 0, "horizontale positie in procent (-100..100)")
	flag.Float64Var(&o.Y, "y", 0, "verticale positie in procent (-100..100)")
	flag.Parse()
	info, err := export.Export(context.Background(), o)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fout:", err)
		os.Exit(1)
	}
	fmt.Printf("klaar: %s (audio: %v, %.0f fps)\n", o.Output, info.HasAudio, info.FPS)
}
