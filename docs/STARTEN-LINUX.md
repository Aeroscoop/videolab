# VideoLab lokaal starten en automatisch bijwerken (Linux Mint)

## Vereisten
- `git`, Go, Wails CLI v2, Node.js
- GTK3/WebKitGTK 4.1 ontwikkelbibliotheken (`libgtk-3-dev`, `libwebkit2gtk-4.1-dev`)
- Voor fase 1: `go mod tidy` en daarna `wails dev -tags webkit2_41`

## Het Orchestrator-patroon
Net als bij Orchestrator blijft het bestand `start.sh` in de repo-root staan, zodat een menu- of bureaubladsnelkoppeling nooit naar een bewegend script hoeft te verwijzen. Dit roept `scripts/start-videolab.sh` aan.

De launcher controleert `git branch --show-current` en `git status --porcelain`. **Alleen** op een schone `main` wordt `origin/main` opgehaald en een fast-forward merge geprobeerd. Bij netwerkproblemen, lokale wijzigingen of een andere branch blijft de werkboom ongewijzigd. Vervolgens start `wails dev -tags webkit2_41`.

Dit is een ontwikkelworkflow: er is geen afzonderlijk geïnstalleerde, gecompileerde productie-app. Wails kan Go-code bij starten compileren en wijzigingen herladen. Voor een snelle dagelijkse start zonder ontwikkeltoolchain kan later de Orchestrator-productiebuild-variant worden toegevoegd, na testen.

## PR #1 eerst lokaal testen
Zolang PR #1 niet in `main` zit, kan de normale updatecheck de fase-1-code niet op een schone `main` ophalen. Check dus voor de eerste test de PR-branch uit en start deze handmatig. Na goedgekeurde merge: wissel naar een schone `main` en gebruik de vaste snelkoppeling.

## Desktopkoppeling maken
1. Bepaal de exacte lokale map met `pwd`. Gebruik geen aangenomen pad.
2. Kopieer `scripts/videolab.desktop.example` naar `~/.local/share/applications/videolab.desktop`.
3. Vervang beide `/ABSOLUTE/PATH/TO/videolab`-waarden door het gecontroleerde absolute pad.
4. Controleer dat `start.sh` uitvoerbaar is en werkt via `bash start.sh`.
5. Test de koppeling vanuit het Linux Mint-menu. Met `Terminal=true` zijn opstartfouten zichtbaar.

## Handmatige checks
```bash
git status --short --branch
bash start.sh
npm test
go test -tags webkit2_41 ./...
```

Zie `docs/TESTPLAN.md` en `docs/OVERDRACHT-COWORK-2026-11-18.md`.
