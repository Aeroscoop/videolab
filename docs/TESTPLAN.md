# Testplan — Linux Mint

Gebruik een aparte checkout en een virtueel scherm wanneer de werkdesktop bezet is.

## Automated
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
9. Splits een bron in drie delen, verwijder het middelste deel en zet de snelheid op 2×. Bekijk de sprong in de preview, exporteer en controleer duur, kleuren en audio met FFmpeg. Maak de verwijdering ongedaan en controleer de tijdlijn.
10. Selecteer verschillende platformrichtzones. Controleer de overgang naar het juiste canvas en vergelijk mobiel, tablet en desktop. Exporteer terwijl de richtzone zichtbaar is.
11. Neem in een geïsoleerd Chrome-profiel een synthetisch tabblad op via de deelkeuze. Stop, download, importeer de WebM, scrub en exporteer.
12. Start `videolab-mcp -root <testmap>` via stdio. Test toolinventaris, projectvalidatie, videoprobe, export, een bestaand doelbestand en een pad buiten de testmap.

## Bekende grenzen

- De tijdlijn maakt meerdere clips uit één bronbestand.
- Het JSON-project bewaart het montageplan. Herimport van de video is nodig.
- De browser en WebKitGTK bepalen welke codecs kunnen worden afgespeeld.
- Platformrichtzones zijn ontwerphulpen. Controleer de echte plaatsing op het doelapparaat.
