# ============================================
# Fy-Analysis API Test Script
# ============================================
# Este script prueba todos los endpoints de la API
# Uso: .\test-api.ps1 [-BaseUrl "http://localhost:8080"]
# ============================================

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

# Colores para output
function Write-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Error { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Test { param($msg) Write-Host "`n=== $msg ===" -ForegroundColor Yellow }

# Función para hacer requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null
    )

    $uri = "$BaseUrl$Endpoint"
    $headers = @{ "Content-Type" = "application/json" }

    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        return @{ Success = $true; Data = $response }
    }
    catch {
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# ============================================
# TESTS
# ============================================

Write-Host "`n============================================" -ForegroundColor Magenta
Write-Host "   FY-ANALYSIS API TEST SUITE" -ForegroundColor Magenta
Write-Host "============================================" -ForegroundColor Magenta
Write-Host "Base URL: $BaseUrl`n"

$totalTests = 0
$passedTests = 0

# --------------------------------------------
# Test 1: Health Check
# --------------------------------------------
Write-Test "Test 1: Health Check"
$totalTests++

$result = Invoke-ApiRequest -Method "GET" -Endpoint "/health"
if ($result.Success -and $result.Data.status -eq "healthy") {
    Write-Success "Health check passed"
    Write-Info "Version: $($result.Data.version)"
    $passedTests++
} else {
    Write-Error "Health check failed: $($result.Error)"
}

# --------------------------------------------
# Test 2: Analizar Email Seguro
# --------------------------------------------
Write-Test "Test 2: Analizar Email Seguro (Gmail)"
$totalTests++

$body = @{ email = "usuario@gmail.com"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/email" -Body $body

if ($result.Success) {
    Write-Success "Email analizado correctamente"
    Write-Info "Email: $($result.Data.email)"
    Write-Info "Es malicioso: $($result.Data.analysis.is_malicious)"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    Write-Info "Confianza: $($result.Data.analysis.confidence)"
    $passedTests++
} else {
    Write-Error "Fallo al analizar email: $($result.Error)"
}

# --------------------------------------------
# Test 3: Analizar Email Desechable
# --------------------------------------------
Write-Test "Test 3: Analizar Email Desechable"
$totalTests++

$body = @{ email = "test@tempmail.com"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/email" -Body $body

if ($result.Success -and $result.Data.analysis.threat_level -ne "safe") {
    Write-Success "Email desechable detectado correctamente"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    Write-Info "Razones: $($result.Data.analysis.reasons -join ', ')"
    $passedTests++
} else {
    Write-Error "No se detectó el email desechable"
}

# --------------------------------------------
# Test 4: Analizar Email con Contexto Sospechoso
# --------------------------------------------
Write-Test "Test 4: Analizar Email con Contexto de Phishing"
$totalTests++

$body = @{
    email = "security-alert@unknown-bank.com"
    context = "URGENTE: Tu cuenta ha sido suspendida. Verifica tu contraseña inmediatamente."
}
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/email" -Body $body

if ($result.Success) {
    Write-Success "Email con contexto analizado"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    Write-Info "Tipos de amenaza: $($result.Data.analysis.threat_types -join ', ')"
    Write-Info "Razones: $($result.Data.analysis.reasons -join ', ')"
    $passedTests++
} else {
    Write-Error "Fallo al analizar: $($result.Error)"
}

# --------------------------------------------
# Test 5: Analizar URL Segura
# --------------------------------------------
Write-Test "Test 5: Analizar URL Segura (Google)"
$totalTests++

$body = @{ url = "https://www.google.com"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/url" -Body $body

if ($result.Success -and -not $result.Data.analysis.is_malicious) {
    Write-Success "URL segura detectada correctamente"
    Write-Info "URL: $($result.Data.url)"
    Write-Info "Dominio: $($result.Data.url_info.domain)"
    Write-Info "SSL válido: $($result.Data.url_info.ssl_valid)"
    $passedTests++
} else {
    Write-Error "Fallo al analizar URL segura"
}

# --------------------------------------------
# Test 6: Analizar URL Acortada
# --------------------------------------------
Write-Test "Test 6: Analizar URL Acortada"
$totalTests++

$body = @{ url = "https://bit.ly/abc123"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/url" -Body $body

if ($result.Success -and $result.Data.url_info.is_shortened) {
    Write-Success "URL acortada detectada"
    Write-Info "Es acortada: $($result.Data.url_info.is_shortened)"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    $passedTests++
} else {
    Write-Error "No se detectó URL acortada"
}

# --------------------------------------------
# Test 7: Analizar URL Sospechosa
# --------------------------------------------
Write-Test "Test 7: Analizar URL Sospechosa (Phishing)"
$totalTests++

$body = @{ url = "http://paypal-secure-login.tk/verify-account"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/url" -Body $body

if ($result.Success -and $result.Data.analysis.threat_level -ne "safe") {
    Write-Success "URL sospechosa detectada"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    Write-Info "Tipos: $($result.Data.analysis.threat_types -join ', ')"
    Write-Info "Razones: $($result.Data.analysis.reasons -join ', ')"
    $passedTests++
} else {
    Write-Error "No se detectó URL sospechosa"
}

# --------------------------------------------
# Test 8: Analizar Teléfono Válido
# --------------------------------------------
Write-Test "Test 8: Analizar Teléfono Español Válido"
$totalTests++

$body = @{ phone = "+34612345678"; country_code = "ES"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/phone" -Body $body

if ($result.Success) {
    Write-Success "Teléfono analizado correctamente"
    Write-Info "Teléfono: $($result.Data.phone)"
    Write-Info "País: $($result.Data.phone_info.country)"
    Write-Info "Tipo: $($result.Data.phone_info.type)"
    Write-Info "Es válido: $($result.Data.phone_info.is_valid)"
    $passedTests++
} else {
    Write-Error "Fallo al analizar teléfono: $($result.Error)"
}

# --------------------------------------------
# Test 9: Analizar Teléfono Premium
# --------------------------------------------
Write-Test "Test 9: Analizar Teléfono Premium (Tarificación especial)"
$totalTests++

$body = @{ phone = "+34806123456"; country_code = "ES"; context = "" }
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/phone" -Body $body

if ($result.Success -and $result.Data.phone_info.is_premium_rate) {
    Write-Success "Teléfono premium detectado"
    Write-Info "Es premium: $($result.Data.phone_info.is_premium_rate)"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    $passedTests++
} else {
    Write-Error "No se detectó teléfono premium"
}

# --------------------------------------------
# Test 10: Analizar Teléfono con Contexto Scam
# --------------------------------------------
Write-Test "Test 10: Analizar Teléfono con Contexto de Estafa"
$totalTests++

$body = @{
    phone = "+34900123456"
    country_code = "ES"
    context = "Has ganado un premio de lotería! Llama para reclamar tu dinero."
}
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/phone" -Body $body

if ($result.Success -and $result.Data.analysis.threat_level -ne "safe") {
    Write-Success "Contexto de estafa detectado"
    Write-Info "Nivel de amenaza: $($result.Data.analysis.threat_level)"
    Write-Info "Razones: $($result.Data.analysis.reasons -join ', ')"
    $passedTests++
} else {
    Write-Error "No se detectó contexto de estafa"
}

# --------------------------------------------
# Test 11: Análisis en Lote
# --------------------------------------------
Write-Test "Test 11: Análisis en Lote (Batch)"
$totalTests++

$body = @{
    emails = @("user@gmail.com", "test@tempmail.com", "spam@phishing-site.net")
    urls = @("https://google.com", "http://malware-site.tk/download")
    phones = @("+34612345678", "+34806000000")
}
$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/v1/analyze/batch" -Body $body

if ($result.Success) {
    Write-Success "Análisis en lote completado"
    Write-Info "Total analizado: $($result.Data.summary.total_analyzed)"
    Write-Info "Maliciosos: $($result.Data.summary.malicious_count)"
    Write-Info "Sospechosos: $($result.Data.summary.suspicious_count)"
    Write-Info "Seguros: $($result.Data.summary.safe_count)"
    $passedTests++
} else {
    Write-Error "Fallo en análisis en lote: $($result.Error)"
}

# --------------------------------------------
# Test 12: Error - Campo Requerido Faltante
# --------------------------------------------
Write-Test "Test 12: Validación - Email Vacío"
$totalTests++

$body = @{ email = ""; context = "" }
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/analyze/email" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($body | ConvertTo-Json)
    Write-Error "Debería haber fallado con email vacío"
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        Write-Success "Validación correcta - rechazó email vacío"
        $passedTests++
    } else {
        Write-Error "Error inesperado: $($_.Exception.Message)"
    }
}

# ============================================
# RESUMEN
# ============================================
Write-Host "`n============================================" -ForegroundColor Magenta
Write-Host "   RESUMEN DE TESTS" -ForegroundColor Magenta
Write-Host "============================================" -ForegroundColor Magenta

$percentage = [math]::Round(($passedTests / $totalTests) * 100, 2)

if ($passedTests -eq $totalTests) {
    Write-Host "`nResultado: $passedTests/$totalTests tests pasados ($percentage%)" -ForegroundColor Green
    Write-Host "¡Todos los tests pasaron exitosamente!" -ForegroundColor Green
} else {
    Write-Host "`nResultado: $passedTests/$totalTests tests pasados ($percentage%)" -ForegroundColor Yellow
    Write-Host "Algunos tests fallaron. Revisa los errores arriba." -ForegroundColor Yellow
}

Write-Host "`n"
