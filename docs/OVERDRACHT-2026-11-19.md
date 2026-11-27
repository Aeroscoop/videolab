# VideoLab-overdracht · 19 november 2026

De productie-app toont **VideoLab · fase 5**. De code bouwt voort op `936ebc3`, de originele exportcommit van Claude Lab, en is samengebracht op `feat/videolab-phase-5`.

## Wat werkt

- **Montage en export.** Importeer één bronbestand. Trim of splits het, verwijder afzonderlijke clips en stel de snelheid van de montage in. **Ongedaan maken** herstelt de vorige bewerking. Een opgeslagen JSON-project bewaart de instellingen en clipgrenzen. Importeer dezelfde bron bij het heropenen.
- **Platformpreview.** De richtzones voor YouTube Ads, TikTok Ads en LinkedIn helpen bij het kaderen. Mobiele en tabletrichtzones zijn op het virtuele scherm vergeleken. Controle op fysieke doelapparaten en in de actuele platform-apps blijft open.
- **Browseropname.** Open vanuit VideoLab de lokale opnamepagina. Geef in Chrome of Chromium toestemming voor het gekozen tabblad. Sla de opname als WebM op en importeer deze terug. VideoLab maakt voor WebKit een tijdelijke H.264/AAC-kopie met doorzoekbare tijdstempels.
- **MCP.** De stdio-server biedt `validate_project`, `probe_video` en `export_mp4`. Start met `-root` naar de gekozen projectmap. Gebruik relatieve paden daarbinnen. De export bewaart bestaande doelbestanden.

## Gecontroleerd

- Go-productiebuild met WebKitGTK 4.1, `go vet` en acht JavaScript-tests.
- Acht Go-exporttests op Mint. De extra testfixture op een vaste Claude Lab-locatie is op Mint overgeslagen.
- Native Wails-import en -export. Annuleren van een dialoog en weigeren van overschrijven zijn getest. De preview en de export van een echte bron zijn op hetzelfde tijdstip vergeleken met uit JSON herstelde kaderinstellingen.
- Montage van twee behouden clips. Verwijderen, ongedaan maken en export op 2× zijn getest. De uitvoer duurde 1,992 seconde en bevatte H.264 en AAC.
- Chrome-tabbladopname via de normale browserdialoog. De opname is terug geïmporteerd, gescrubd, geëxporteerd en volledig gedecodeerd.
- MCP-initialisatie, toolinventaris, projectvalidatie, videoprobe, export en weigering bij een bestaand doelbestand.

## Lokale installatie

Op Mint openen de menu- en bureaubladsnelkoppeling de gecompileerde app. De startopdrachten staan onder `~/.local/bin/`. FFmpeg en ffprobe zijn voor de gebruikersinstallatie meegeleverd. Via de geïnstalleerde app is een 1080 × 1350 MP4 met AAC gemaakt en gedecodeerd. De bijbehorende MCP-opdracht is met een proefclip getest. Er is nog geen MCP-client ingesteld voor een specifieke projectmap.

De oorspronkelijke ontwikkelsnelkoppelingen staan onder `~/.local/lib/videolab/launcher-backup/`. De [Linux-instructies](STARTEN-LINUX.md) beschrijven bouwen en bijwerken. De [eerste Cowork-overdracht](OVERDRACHT-COWORK-2026-11-18.md) documenteert fase 1.
