$ErrorActionPreference = 'Stop'
$RepoUrl = 'https://github.com/ShiduLab/Kunta.git'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ''
Write-Host 'KUNTA - pubblicazione atomica del repository' -ForegroundColor Cyan
Write-Host 'Il repository remoto non viene svuotato prima del push.' -ForegroundColor DarkGray
Write-Host ''

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw 'Git non e installato o non e nel PATH.'
}

Set-Location $Root

if (-not (Test-Path '.git')) {
    git init
    git branch -M main
}

$origin = git remote 2>$null | Where-Object { $_ -eq 'origin' }
if ($origin) {
    git remote set-url origin $RepoUrl
} else {
    git remote add origin $RepoUrl
}

# Identita locale soltanto se Git non ne ha gia una configurata.
if (-not (git config user.name)) { git config user.name 'ShiduLab' }
if (-not (git config user.email)) { git config user.email 'ShiduLab@users.noreply.github.com' }

# Conosciamo lo stato remoto prima del force-with-lease: nessuna finestra in cui il repo e vuoto.
git fetch origin main --prune

git add -A
$pending = git status --porcelain
if (-not $pending) {
    Write-Host 'Nessuna modifica da pubblicare.' -ForegroundColor Yellow
    exit 0
}

git commit -m 'Kunta: ricostruzione completa pulita Desktop PWA Android'

git push --force-with-lease -u origin main

Write-Host ''
Write-Host 'PUBBLICAZIONE COMPLETATA.' -ForegroundColor Green
Write-Host 'Ora GitHub Actions puo eseguire Forge Kunta.' -ForegroundColor Green
