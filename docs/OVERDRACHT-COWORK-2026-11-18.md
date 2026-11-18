# Overdracht volgende ChatGPT-desktopsessie / Cowork — 18 november 2026

## Doel
Richt op **Linux Mint** een vaste desktop-/menukoppeling voor VideoLab in en test de fase-1 Wails-app. De snelkoppeling moet opstarten via het vaste `start.sh`-script. Dat script haalt op een **schone main-branch** eerst de nieuwste GitHub-commits op. Dit is een handmatige update bij het starten, geen achtergrondtaak.

## Projectstand
- Repository: https://github.com/Aeroscoop/videolab (public).
- Branch met fase 1: `feat/phase-1-wails-studio`.
- Pull request: https://github.com/Aeroscoop/videolab/pull/1.
- PR is **niet gemerged**. De gebruiker test eerst lokaal; merge alleen op expliciete opdracht.
- Bron voor de workflow: `Aeroscoop/orchestrator`, vooral `start.sh`, `scripts/start-orchestrator.sh`, `scripts/orchestrator.desktop.example`.
- VideoLab is een los project, later ook inzetbaar buiten Nestar.

## Wat fase 1 bevat
Wails v2 + Go; HTML/CSS/JS-editor; één lokaal videobestand importeren en afspelen; eenvoudige tijdlijn met seek en ±5s; canvaspresets 4:5, 9:16, 16:9, 1:1; achtergrondkleur, schaal, X/Y-positie; indicatieve richtzone; JSON-projectinstellingen opslaan/openen. **Nog geen** splitsen/knippen, versnellen, FFmpeg-export, browseropname, werkelijke social-overlays of MCP.

## Cowork-opdracht op Linux Mint
1. Controleer apparaat, toegang tot lokale shell en beschikbare Go/Wails/Node/GTK/WebKitGTK. Wijzig geen andere repository.
2. Controleer of VideoLab al lokaal bestaat. Zo ja: toon pad en git status voordat je wijzigingen maakt. Zo nee: clone naar een in overleg gekozen map; geen aannames over paden.
3. Haal voor **eerste test** de PR-branch `feat/phase-1-wails-studio` op, zonder de hoofdbranch te wijzigen. Voer `go mod tidy` uit indien nodig; check wijzigingen in `go.mod`/`go.sum` en voorkom onbedoelde commits.
4. Draai `npm test`, `go test -tags webkit2_41 ./...` en start met `wails dev -tags webkit2_41`. Test met de Nestar-video. Noteer codecfouten en afwijkingen afzonderlijk.
5. Maak de `.desktop`-snelkoppeling op basis van `scripts/videolab.desktop.example`. Vervang `/ABSOLUTE/PATH/TO/videolab` door het **werkelijk gecontroleerde lokale pad**. Plaats kopie in `~/.local/share/applications/videolab.desktop`; maak indien nodig uitvoerbaar en zichtbaar in het menu. De `Exec` moet naar het vaste `start.sh` wijzen.
6. **Let op:** zolang de werkbranch niet `main` is, slaat het startscript automatische GitHub-updates bewust over. Dit is correct. Test de automatische update pas **na expliciete merge van PR #1 en een schone lokale `main`**, anders lijkt het updategedrag defect terwijl het juist beschermt tegen verlies van wijzigingen.
7. Bekijk de eerste start in Wails én de GUI, inclusief bestandkiezer, opslaan van JSON en video-afspelen. Geen succes claimen zonder de echte tests.
8. Leg gevonden problemen vast in PR #1 of een afzonderlijk issue. Geen merge, geen system-wide installatie en geen stille overschrijving van lokale data.

## Linux-commando's (uitvoeren vanuit de lokale repo)
```bash
git status --short --branch
git fetch origin
git switch feat/phase-1-wails-studio   # alleen als die branch lokaal bestaat
# anders: git switch --track origin/feat/phase-1-wails-studio
go mod tidy
npm test
go test -tags webkit2_41 ./...
wails dev -tags webkit2_41
```

## Snelkoppeling
De twee startpunten:
- `start.sh` = vast pad voor de snelkoppeling.
- `scripts/start-videolab.sh` = update-check + `wails dev -tags webkit2_41`.

Op **schone main** doet het script `git fetch origin main` en `git merge --ff-only origin/main`; bij offline modus blijft de lokale checkout bruikbaar. Op andere branches of met wijzigingen wordt niet geüpdatet. Het script start daarna ontwikkelmodus. Dit is **geen geïnstalleerde productie-build**; Go/Wails blijft als toolchain nodig.

Voorbeeld desktopbestand: `scripts/videolab.desktop.example`. Gebruik bij voorkeur `Terminal=true` zodat fouten bij starten zichtbaar zijn.

## Testpunten en bekende risico's
- Wails v2 Linux vereist GTK3 en WebKit2GTK 4.1; build tag `webkit2_41`.
- Beschikbare codecs, browser file-input, video-afspelen en downloaden binnen WebKitGTK zijn nog niet getest.
- JSON slaat alleen de layout en bestandsmetadata op; videodata wordt niet ingesloten en moet na openen opnieuw worden geïmporteerd.
- De richtzone is illustratief, geen nauwkeurige Instagram/LinkedIn-simulatie.
- Het startscript werkt met automatische update bij starten; op een PR-branch is die uitgeschakeld.
- Controleer Go-modules, Wails CLI en `go.sum` op het apparaat; GitHub-code is nog niet end-to-end getest.
- De eerste versie is een video-**layout-preview**, nog geen geëxporteerde montage.

## Waarom dit zo is ontworpen
De Nestar-video van 18 november was op Instagram en LinkedIn onduidelijk door overlappende social-UI, afgesneden delen en het telefoon-in-telefoon-effect. Doel: één rustige video geschikt voor beide kanalen, met schaal, marge en previews. Eerste voorkeur was 4:5 als gedeeld testformaat, zonder al te vroeg definitieve platformregels vast te leggen.

## Afsluiten volgende sessie
Rapporteer: lokaal pad, versies, startcommando, desktopkoppeling, testresultaten, openstaande defecten en eventueel PR/issue-links. Voer de merge alleen na expliciete gebruikersgoedkeuring uit.
