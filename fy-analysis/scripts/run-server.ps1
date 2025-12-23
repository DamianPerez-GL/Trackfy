# ============================================
# Fy-Analysis Server Runner Script (Docker)
# ============================================
# Este script ejecuta el servidor usando Docker
# Uso: .\run-server.ps1
#      .\run-server.ps1 -Stop     (detener)
#      .\run-server.ps1 -Rebuild  (reconstruir imagen)
# ============================================

param(
    [string]$Port = "8080",
    [switch]$Stop,
    [switch]$Rebuild,
    [switch]$Logs
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $PSScriptRoot

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "   FY-ANALYSIS SERVER (Docker)" -ForegroundColor Cyan
Write-Host "============================================`n" -ForegroundColor Cyan

Set-Location $ScriptDir

# Verificar Docker
try {
    $dockerVersion = docker --version
    Write-Host "Docker: $dockerVersion" -ForegroundColor Green
}
catch {
    Write-Host "Error: Docker no está instalado o no está corriendo" -ForegroundColor Red
    Write-Host "Instala Docker Desktop desde: https://www.docker.com/products/docker-desktop/" -ForegroundColor Yellow
    exit 1
}

if ($Stop) {
    Write-Host "Deteniendo contenedor..." -ForegroundColor Yellow
    docker-compose down
    Write-Host "Contenedor detenido" -ForegroundColor Green
    exit 0
}

if ($Logs) {
    docker-compose logs -f
    exit 0
}

if ($Rebuild) {
    Write-Host "Reconstruyendo imagen..." -ForegroundColor Yellow
    docker-compose down
    docker-compose build --no-cache
}

Write-Host "Iniciando servidor en http://localhost:$Port" -ForegroundColor Green
Write-Host "Presiona Ctrl+C para ver logs, luego ejecuta: .\run-server.ps1 -Stop`n" -ForegroundColor Gray

docker-compose up --build
