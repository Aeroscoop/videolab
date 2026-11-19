# VideoLab op Linux Mint

## Vereisten

- Go 1.25 of hoger en GTK3/WebKitGTK 4.1 voor de build. Installeer de ontwikkelbibliotheken `libgtk-3-dev` en `libwebkit2gtk-4.1-dev`.
- Zorg voor `ffmpeg` en `ffprobe` in het zoekpad van VideoLab. FFmpeg heeft de encoders `libx264` en `aac` nodig.
- Gebruik Chrome of Chromium voor tabbladopname. Node.js is nodig voor de JavaScript-tests.

## Productiebuild

Bouw vanuit de repo-root. De uitvoer komt in de genegeerde map `build/bin/`.

```bash
mkdir -p build/bin
go build -mod=readonly -tags production,webkit2_41 -o build/bin/videolab .
go build -mod=readonly -o build/bin/videolab-mcp ./cmd/videolab-mcp
./build/bin/videolab
```

Maak voor dagelijks gebruik een menu-item dat naar een vaste launcher of een vast binair pad wijst. Stel `Terminal=false` in en geef `Exec` het absolute pad. De productie-app opent direct. Voor een nieuwe versie bouw je de binaries opnieuw en activeer je die via het vaste pad.

De geteste installatie gebruikt `~/.local/bin/videolab` en `~/.local/bin/videolab-mcp`. Beide launchers gebruiken `~/.local/lib/videolab/current`. Die symlink wijst naar een release onder `~/.local/lib/videolab/releases`. Wanneer FFmpeg via het gebruikersaccount is geïnstalleerd, staan de benodigde binaries bij de release onder `tools/`. De launchers voegen die map toe aan hun zoekpad.

Bewaar bij een update een werkende release. Zet de nieuwe desktop- en MCP-binaries samen in een nieuwe release, met eventueel dezelfde tools. Richt `current` vervolgens op de nieuwe release en herstart VideoLab. Menu- en bureaubladsnelkoppelingen blijven naar het vaste launcherpad wijzen.

## Ontwikkelen met automatische Git-updates

Gebruik `bash start.sh` voor de ontwikkelmodus. Dit start `wails dev -tags webkit2_41` via `scripts/start-videolab.sh`. Alleen op een schone lokale `main` probeert het script eerst een fast-forward naar `origin/main`. Het voorbeeld in [scripts/videolab.desktop.example](../scripts/videolab.desktop.example) hoort bij deze ontwikkelmodus.

```bash
git status --short --branch
npm test
go test -mod=readonly -tags production,webkit2_41 ./...
```

Zie het [testplan](TESTPLAN.md) en de [overdracht van 19 november](OVERDRACHT-2026-11-19.md).
