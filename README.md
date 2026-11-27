# VideoLab

VideoLab is a desktop editor for turning one source video into a social-format MP4. It runs locally with Go, Wails and FFmpeg.

## Linux Mint requirements

- Go compatible with Wails v2, Wails CLI v2
- `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`
- `ffmpeg` and `ffprobe` on the app’s PATH for export and preparing browser recordings
- Chrome or Chromium for tab recording
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

## Editing and export

- Import a video. Use the canvas preset, background, scale and position to frame it.
- Scrub to a source time and use **Begin hier**, **Einde hier** or **Splits hier**. Select a clip on the timeline to trim or remove it. Choose a playback speed for the whole edit. **Ongedaan maken** or Ctrl+Z restores the previous edit.
- **MP4 exporteren** renders the kept clips in source order. Output uses H.264 and AAC when the source contains audio.
- Saved projects are JSON. Reopen the project and reimport the same source video to restore the edit.

## Platform previews

Choose **Platformrichtzone** to set the associated canvas and see the guide. The YouTube Ads profiles use Google's mobile, tablet and desktop illustrations. The TikTok Ads profile uses a conservative rectangle inside the standard LTR In-Feed template. The LinkedIn feed guide uses editorial margins based on the platform's edge guidance. Guides only affect the preview. Check the target placement on the actual device before publishing.

Platform references are available from [Google Ads safe zones](https://support.google.com/google-ads/answer/13547298), [TikTok In-Feed ad templates](https://ads.tiktok.com/resources/help/article/tiktok-auction-in-feed-ads), [LinkedIn video guidance](https://www.linkedin.com/help/linkedin/answer/a7494039).

## Record a browser tab

Choose **Tabblad opnemen** in VideoLab. In Chrome or Chromium, select **Kies tabblad** and approve the specific tab in the browser dialog. Enable tab audio in that dialog when needed. Stop the recording, save the WebM and import it into VideoLab. On Linux, VideoLab prepares a temporary H.264/AAC copy for preview and editing. The recorder page runs on localhost and keeps the recording in the browser until downloaded.

## MCP

Build the local stdio server with Go.

```sh
go build -o /absolute/path/videolab-mcp ./cmd/videolab-mcp
```

Set the client's command to that executable and its arguments to `["-root", "/absolute/path/to/video-projects"]`. The server provides `validate_project`, `probe_video` and `export_mp4`. Tool paths are relative to the configured root. Export takes a saved project, a source and a new `.mp4` destination. It preserves existing files at the destination.

Repo: https://github.com/Aeroscoop/videolab

## Next desktop session / local installation

The [original Cowork handoff](docs/OVERDRACHT-COWORK-2026-11-18.md) records the phase 1 setup.

- [Linux production build, startup and updates](docs/STARTEN-LINUX.md)
- [Phase 5 handoff and verification](docs/OVERDRACHT-2026-11-19.md)
- [Desktop launcher template](scripts/videolab.desktop.example)
- [Manual testing checklist](docs/TESTPLAN.md)
- [Technical decisions](docs/BESLUITEN.md)
