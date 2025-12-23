# ============================================
# Quick Test - Pruebas Rápidas Individuales
# ============================================
# Uso: .\quick-test.ps1 -Type email -Value "test@gmail.com"
#      .\quick-test.ps1 -Type url -Value "https://suspicious-site.tk"
#      .\quick-test.ps1 -Type phone -Value "+34612345678"
# ============================================

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("email", "url", "phone")]
    [string]$Type,

    [Parameter(Mandatory=$true)]
    [string]$Value,

    [string]$Context = "",
    [string]$CountryCode = "",
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

function Format-ThreatLevel {
    param([string]$Level)
    switch ($Level) {
        "safe"     { Write-Host $Level -ForegroundColor Green -NoNewline }
        "low"      { Write-Host $Level -ForegroundColor Yellow -NoNewline }
        "medium"   { Write-Host $Level -ForegroundColor DarkYellow -NoNewline }
        "high"     { Write-Host $Level -ForegroundColor Red -NoNewline }
        "critical" { Write-Host $Level -ForegroundColor DarkRed -NoNewline }
        default    { Write-Host $Level -ForegroundColor Gray -NoNewline }
    }
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "   ANÁLISIS DE $($Type.ToUpper())" -ForegroundColor Cyan
Write-Host "============================================`n" -ForegroundColor Cyan

$endpoint = "/api/v1/analyze/$Type"
$headers = @{ "Content-Type" = "application/json" }

switch ($Type) {
    "email" {
        $body = @{ email = $Value; context = $Context } | ConvertTo-Json
    }
    "url" {
        $body = @{ url = $Value; context = $Context } | ConvertTo-Json
    }
    "phone" {
        $body = @{ phone = $Value; country_code = $CountryCode; context = $Context } | ConvertTo-Json
    }
}

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl$endpoint" -Method POST -Headers $headers -Body $body

    Write-Host "Valor analizado: " -NoNewline
    Write-Host $Value -ForegroundColor White

    Write-Host "`n--- RESULTADO ---" -ForegroundColor Yellow

    Write-Host "Es malicioso: " -NoNewline
    if ($response.analysis.is_malicious) {
        Write-Host "SÍ" -ForegroundColor Red
    } else {
        Write-Host "NO" -ForegroundColor Green
    }

    Write-Host "Nivel de amenaza: " -NoNewline
    Format-ThreatLevel $response.analysis.threat_level
    Write-Host ""

    Write-Host "Confianza: " -NoNewline
    $confidence = [math]::Round($response.analysis.confidence * 100, 1)
    Write-Host "$confidence%" -ForegroundColor Cyan

    if ($response.analysis.threat_types -and $response.analysis.threat_types.Count -gt 0) {
        Write-Host "`nTipos de amenaza:"
        foreach ($threat in $response.analysis.threat_types) {
            Write-Host "  - $threat" -ForegroundColor DarkYellow
        }
    }

    if ($response.analysis.reasons -and $response.analysis.reasons.Count -gt 0) {
        Write-Host "`nRazones:"
        foreach ($reason in $response.analysis.reasons) {
            Write-Host "  - $reason" -ForegroundColor Gray
        }
    }

    if ($response.recommendations -and $response.recommendations.Count -gt 0) {
        Write-Host "`nRecomendaciones:" -ForegroundColor Magenta
        foreach ($rec in $response.recommendations) {
            Write-Host "  → $rec" -ForegroundColor Magenta
        }
    }

    # Información adicional según tipo
    Write-Host "`n--- INFORMACIÓN ADICIONAL ---" -ForegroundColor Yellow

    switch ($Type) {
        "email" {
            if ($response.domain_info) {
                Write-Host "Dominio: $($response.domain_info.domain)"
                Write-Host "Es desechable: $($response.domain_info.is_disposable)"
                Write-Host "Es freemail: $($response.domain_info.is_freemail)"
                Write-Host "Tiene registros MX: $($response.domain_info.has_mx_records)"
            }
        }
        "url" {
            if ($response.url_info) {
                Write-Host "Dominio: $($response.url_info.domain)"
                Write-Host "Esquema: $($response.url_info.scheme)"
                Write-Host "Es URL acortada: $($response.url_info.is_shortened)"
                Write-Host "SSL válido: $($response.url_info.ssl_valid)"
                Write-Host "Params sospechosos: $($response.url_info.has_suspicious_params)"
            }
        }
        "phone" {
            if ($response.phone_info) {
                Write-Host "País: $($response.phone_info.country) ($($response.phone_info.country_code))"
                Write-Host "Tipo: $($response.phone_info.type)"
                Write-Host "Es válido: $($response.phone_info.is_valid)"
                Write-Host "Es premium: $($response.phone_info.is_premium_rate)"
            }
        }
    }

    Write-Host "`n"
}
catch {
    Write-Host "Error al analizar: $($_.Exception.Message)" -ForegroundColor Red
}
