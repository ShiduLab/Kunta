$ErrorActionPreference='Stop'
$root=Split-Path -Parent $MyInvocation.MyCommand.Path
$must=@(
'.github/workflows/forge-kunta.yml',
'.github/workflows/pages.yml',
'desktop/main.go',
'desktop/web/index.html',
'desktop/web/app.js',
'pwa/index.html',
'pwa/app.js',
'android/settings.gradle',
'android/build.gradle',
'android/app/build.gradle',
'android/app/src/main/AndroidManifest.xml',
'android/app/src/main/java/io/github/shidulab/kunta/MainActivity.java',
'android/app/src/main/assets/index.html',
'android/app/src/main/assets/app.js',
'Kunta.exe',
'Kunta-PWA.zip'
)
$missing=@()
foreach($p in $must){ if(-not (Test-Path (Join-Path $root $p))){ $missing += $p } }
if($missing.Count){ Write-Host 'MANCANO:' -ForegroundColor Red; $missing | ForEach-Object {Write-Host $_}; exit 1 }
Write-Host 'Struttura Kunta completa.' -ForegroundColor Green
