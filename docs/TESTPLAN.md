# Testplan — Linux Mint

## Automated
- `npm test`
- `go test -tags webkit2_41 ./...`

## Manual on dandylaptop
1. Run `wails dev -tags webkit2_41`.
2. Import the 2026-11-18 Nestar clip; confirm playback and duration.
3. Scrub to middle and jump ±5 seconds.
4. Test 4:5, 9:16, 16:9 and 1:1 and check the canvas is not cropped.
5. Move and scale the video; toggle the safe guide.
6. Save JSON, reopen it, and re-import the source media.
7. Verify WebKitGTK video codec support and file downloads.
8. Verify launcher skips Git update with local edits or on feature branches.

## Not yet supported
Trimming, clip rearrangement, audio editing, export, capture, actual platform UI overlays, MCP.
