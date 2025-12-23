# ============================================
# Test API con Burp Suite Proxy
# ============================================
# Configura Burp en 127.0.0.1:8080 (default)
# La API corre en localhost:9090
# ============================================

param(
    [string]$ApiUrl = "http://localhost:9090",
    [string]$BurpProxy = "http://127.0.0.1:8080"
)

Write-Host "`n============================================" -ForegroundColor Magenta
Write-Host "   TEST CON BURP SUITE PROXY" -ForegroundColor Magenta
Write-Host "============================================" -ForegroundColor Magenta
Write-Host "API: $ApiUrl"
Write-Host "Proxy: $BurpProxy"
Write-Host "============================================`n"

# Health Check
Write-Host "[1] Health Check..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/health" -Proxy $BurpProxy
    Write-Host "    Status: $($response.status)" -ForegroundColor Green
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Email seguro
Write-Host "`n[2] Analizar email seguro (Gmail)..." -ForegroundColor Cyan
$body = @{ email = "usuario@gmail.com" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/email" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    Email: $($response.email)" -ForegroundColor Green
    Write-Host "    Threat Level: $($response.analysis.threat_level)"
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Email malicioso
Write-Host "`n[3] Analizar email sospechoso..." -ForegroundColor Cyan
$body = @{
    email = "security@banco-falso.tk"
    context = "URGENTE: Tu cuenta sera bloqueada. Verifica tu password."
} | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/email" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    Email: $($response.email)" -ForegroundColor Yellow
    Write-Host "    Threat Level: $($response.analysis.threat_level)" -ForegroundColor Red
    Write-Host "    Reasons: $($response.analysis.reasons -join ', ')"
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# URL phishing
Write-Host "`n[4] Analizar URL de phishing..." -ForegroundColor Cyan
$body = @{ url = "http://paypal-login-secure.tk/verify?redirect=http://evil.com" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/url" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    URL: $($response.url)" -ForegroundColor Yellow
    Write-Host "    Malicious: $($response.analysis.is_malicious)" -ForegroundColor Red
    Write-Host "    Threat Level: $($response.analysis.threat_level)"
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# URL con IP
Write-Host "`n[5] Analizar URL con IP..." -ForegroundColor Cyan
$body = @{ url = "http://192.168.1.100/admin/login.php" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/url" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    URL: $($response.url)" -ForegroundColor Yellow
    Write-Host "    Threat Level: $($response.analysis.threat_level)"
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Telefono premium
Write-Host "`n[6] Analizar telefono premium..." -ForegroundColor Cyan
$body = @{ phone = "+34806123456"; country_code = "ES" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/phone" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    Phone: $($response.phone)" -ForegroundColor Yellow
    Write-Host "    Premium: $($response.phone_info.is_premium_rate)" -ForegroundColor Red
    Write-Host "    Threat Level: $($response.analysis.threat_level)"
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Batch analysis
Write-Host "`n[7] Analisis en lote (batch)..." -ForegroundColor Cyan
$body = @{
    emails = @("good@gmail.com", "bad@tempmail.com")
    urls = @("https://google.com", "http://malware.tk/download")
    phones = @("+34612345678", "+34900000000")
} | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/analyze/batch" -Method POST -Body $body -ContentType "application/json" -Proxy $BurpProxy
    Write-Host "    Total: $($response.summary.total_analyzed)" -ForegroundColor Green
    Write-Host "    Maliciosos: $($response.summary.malicious_count)" -ForegroundColor Red
    Write-Host "    Sospechosos: $($response.summary.suspicious_count)" -ForegroundColor Yellow
    Write-Host "    Seguros: $($response.summary.safe_count)" -ForegroundColor Green
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n============================================" -ForegroundColor Magenta
Write-Host "   Revisa Burp Suite para ver el trafico" -ForegroundColor Magenta
Write-Host "============================================`n" -ForegroundColor Magenta
