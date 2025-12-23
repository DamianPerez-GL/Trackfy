# ============================================
# Demo - Demostración Interactiva de Fy-Analysis
# ============================================
# Este script muestra ejemplos de uso de la API
# Uso: .\demo.ps1
# ============================================

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Continue"

function Show-Analysis {
    param(
        [string]$Title,
        [string]$Endpoint,
        [object]$Body
    )

    Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║ $Title" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl$Endpoint" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($Body | ConvertTo-Json)

        # Mostrar resultado visual
        $emoji = if ($response.analysis.is_malicious) { "🚨" }
                 elseif ($response.analysis.threat_level -eq "safe") { "✅" }
                 else { "⚠️" }

        $color = switch ($response.analysis.threat_level) {
            "safe"     { "Green" }
            "low"      { "Yellow" }
            "medium"   { "DarkYellow" }
            "high"     { "Red" }
            "critical" { "DarkRed" }
            default    { "Gray" }
        }

        Write-Host "`n$emoji Resultado: " -NoNewline
        Write-Host $response.analysis.threat_level.ToUpper() -ForegroundColor $color

        if ($response.analysis.reasons) {
            foreach ($reason in $response.analysis.reasons) {
                Write-Host "   → $reason" -ForegroundColor Gray
            }
        }
    }
    catch {
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    }

    Start-Sleep -Milliseconds 500
}

Clear-Host

Write-Host @"

  ███████╗██╗   ██╗       █████╗ ███╗   ██╗ █████╗ ██╗  ██╗   ██╗███████╗██╗███████╗
  ██╔════╝╚██╗ ██╔╝      ██╔══██╗████╗  ██║██╔══██╗██║  ╚██╗ ██╔╝██╔════╝██║██╔════╝
  █████╗   ╚████╔╝ █████╗███████║██╔██╗ ██║███████║██║   ╚████╔╝ ███████╗██║███████╗
  ██╔══╝    ╚██╔╝  ╚════╝██╔══██║██║╚██╗██║██╔══██║██║    ╚██╔╝  ╚════██║██║╚════██║
  ██║        ██║         ██║  ██║██║ ╚████║██║  ██║███████╗██║   ███████║██║███████║
  ╚═╝        ╚═╝         ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝╚═╝   ╚══════╝╚═╝╚══════╝

                    🔒 Servicio de Análisis de Amenazas 🔒

"@ -ForegroundColor Cyan

Write-Host "Conectando a: $BaseUrl" -ForegroundColor Gray
Write-Host "Presiona cualquier tecla para iniciar la demo...`n"
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

# Demo 1: Email seguro
Show-Analysis -Title "EMAIL SEGURO - Gmail personal" -Endpoint "/api/v1/analyze/email" -Body @{
    email = "usuario.normal@gmail.com"
    context = ""
}

# Demo 2: Email desechable
Show-Analysis -Title "EMAIL DESECHABLE - Tempmail" -Endpoint "/api/v1/analyze/email" -Body @{
    email = "anonymous123@tempmail.com"
    context = ""
}

# Demo 3: Email phishing
Show-Analysis -Title "EMAIL PHISHING - Contexto sospechoso" -Endpoint "/api/v1/analyze/email" -Body @{
    email = "security-alert@unknown-service.com"
    context = "URGENTE: Su cuenta será suspendida. Verifique su contraseña ahora."
}

# Demo 4: URL segura
Show-Analysis -Title "URL SEGURA - Google" -Endpoint "/api/v1/analyze/url" -Body @{
    url = "https://www.google.com/search?q=test"
    context = ""
}

# Demo 5: URL acortada
Show-Analysis -Title "URL ACORTADA - bit.ly" -Endpoint "/api/v1/analyze/url" -Body @{
    url = "https://bit.ly/3xyz123"
    context = ""
}

# Demo 6: URL phishing
Show-Analysis -Title "URL PHISHING - Suplantación de PayPal" -Endpoint "/api/v1/analyze/url" -Body @{
    url = "http://paypal-secure-login.tk/verify-account?redirect=true"
    context = ""
}

# Demo 7: URL con IP
Show-Analysis -Title "URL SOSPECHOSA - Dirección IP" -Endpoint "/api/v1/analyze/url" -Body @{
    url = "http://192.168.1.100/admin/login.php"
    context = ""
}

# Demo 8: Teléfono normal
Show-Analysis -Title "TELÉFONO SEGURO - Móvil español" -Endpoint "/api/v1/analyze/phone" -Body @{
    phone = "+34612345678"
    country_code = "ES"
    context = ""
}

# Demo 9: Teléfono premium
Show-Analysis -Title "TELÉFONO PREMIUM - Tarificación especial" -Endpoint "/api/v1/analyze/phone" -Body @{
    phone = "+34806123456"
    country_code = "ES"
    context = ""
}

# Demo 10: Teléfono scam
Show-Analysis -Title "TELÉFONO SCAM - Contexto de estafa" -Endpoint "/api/v1/analyze/phone" -Body @{
    phone = "+34900555123"
    country_code = "ES"
    context = "¡Felicidades! Has ganado 10,000€ en nuestra lotería. Llama para reclamar tu premio."
}

Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                    DEMO COMPLETADA                           ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Green

Write-Host @"

📋 Comandos útiles:
   .\quick-test.ps1 -Type email -Value "test@example.com"
   .\quick-test.ps1 -Type url -Value "https://suspicious-site.tk"
   .\quick-test.ps1 -Type phone -Value "+34600000000"
   .\test-api.ps1  (ejecutar suite completa de tests)

"@ -ForegroundColor Gray
