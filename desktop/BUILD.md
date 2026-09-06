# Build di Kunta

Il programma usa Go e le API Win32 native.

## Build base

Da una shell con Go installato:

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui -s -w" -o Kunta_base.exe main.go
```

Il sorgente incorpora `assets/Kunta.ico` e `assets/Botolo.ico`; la cartella `assets/` deve quindi restare accanto a `main.go`.

## Icona PE per pinning sulla taskbar

Perché Windows mantenga l'icona di Kunta anche quando l'eseguibile viene fissato alla barra delle applicazioni, la build distribuita contiene anche la risorsa icona nel PE.

Lo script incluso `embed_icon.py` aggiunge la risorsa a una build base e richiede Python 3 e GNU `objcopy`:

```bash
python embed_icon.py Kunta_base.exe assets/Kunta.ico Kunta.exe
```

La build finale usa inoltre l'AppUserModelID:

```text
ShiduLab.Kunta
```

Gli utenti del pacchetto Portable non devono installare nulla di tutto questo: serve solo per ricompilare il sorgente.
