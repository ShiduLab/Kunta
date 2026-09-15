# Kunta Desktop — Windows x64

Da root del repository:

```bat
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-H windowsgui -s -w" -o release\Kunta.exe .\desktop
```

Il contenuto di `desktop/web/` viene incorporato dentro `Kunta.exe` in fase di build.
A runtime non servono file HTML/JS esterni.
