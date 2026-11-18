# VideoLab

A small, reusable desktop video-layout studio built with Go and Wails v2. Phase 1 supports importing a local video, previewing and scrubbing it, selecting 4:5 / 9:16 / 16:9 / 1:1, moving/scaling the recording, visual safe-area guides, and saving/opening a JSON project.

## Linux Mint requirements

- Go compatible with Wails v2, Wails CLI v2
- `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`
- Node.js for tests

After cloning and installing Wails, start in development mode:

```sh
wails dev -tags webkit2_41
```

Or run `bash start.sh`. The launcher follows Orchestrator's workflow: on a clean `main` it fetches fast-forward updates before starting Wails development mode. On other branches or with uncommitted work it leaves your checkout unchanged.

Tests:

```sh
npm test
go test -tags webkit2_41 ./...
```

## Phase 1 limitations

- One local video clip on a simple scrub timeline; no split, trim or multi-track editing yet.
- Browser file input, playback and downloads must be tested in Linux WebKitGTK. Codec support varies.
- Projects are JSON layout metadata; imported media is **not embedded**. Reimport the original video when reopening a project.
- Safe-area markings are approximate design guides, not verified Instagram/LinkedIn UI overlays.
- No MP4 export, recording, FFmpeg integration or MCP server yet. Those are later phases.
- A browser preview is not a reliable final-render guarantee.

## Project roadmap

1. **Phase 1:** Wails shell, local video import, preview, scrub timeline, presets, positioning, project JSON.
2. Phase 2: clips, trim/split, playback speed, undo and native FFmpeg export.
3. Phase 3: platform previews with tested device-specific overlays.
4. Phase 4: recording browser tabs via a supported capture workflow.
5. Phase 5: Go MCP service with shared project model and validation.

Repo: https://github.com/Aeroscoop/videolab
