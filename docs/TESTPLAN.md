# Testplan — Linux Mint

Status 2026-11-18: code in PR #1; tests en Wails-GUI zijn **nog niet lokaal uitgevoerd**.

## Automated
- `go mod tidy` indien go.sum nog ontbreekt
- `npm test`
- `go test -tags webkit2_41 ./...`

## Manual op dandylaptop
1. Controleer Go/Wails/Node/WebKitGTK-versies en `git status`; checkout van de featurebranch zonder werk te overschrijven.
2. Start `wails dev -tags webkit2_41`. Bevestig dat het venster opent.
3. Importeer de Nestar-video van 18 november. Controleer decodering, geluid, duur en zichtbaarheid van boven- en onderkant.
4. Speel af/pauzeer, scrub naar midden en spring ±5 seconden.
5. Test 4:5, 9:16, 16:9, 1:1 op uitsnijden en overloop. Vergelijk met Instagram-/LinkedIn-screenshots.
6. Schaal, verplaats horizontaal/verticaal, verander achtergrond en richtzones.
7. Sla JSON op, laad het opnieuw en importeer de oorspronkelijke video opnieuw.
8. Controleer dat `bash start.sh` de repo op featurebranch niet update en wél de app start.
9. Maak/test desktopkoppeling volgens `docs/STARTEN-LINUX.md`.
10. Controleer updategedrag pas na merge op een schone `main` met toestemming.

## Bekende grenzen
- Slechts één clip; tijdlijn is scrub/seek, geen echte montage.
- Geen cliptrim, splitsen, herschikken, snelheidsregeling, audio-edit, export, browseropname of MCP.
- Geen embedded media in JSON.
- Codec, HTML file-input en bestandsdownload op WebKitGTK niet gegarandeerd.
- Richtzones zijn illustratief, niet een garantie voor elk social-platform.
