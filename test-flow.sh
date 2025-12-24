#!/bin/bash
# ===========================================
# Test del flujo completo Trackfy
# ===========================================

echo "=========================================="
echo "  TRACKFY - Test del Flujo Completo"
echo "=========================================="
echo ""

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# URLs base
ANALYSIS_URL="http://localhost:9090"
ENGINE_URL="http://localhost:8082"

# Funcion para imprimir resultados
print_result() {
    local test_name=$1
    local status=$2
    local response=$3

    if [ "$status" == "OK" ]; then
        echo -e "${GREEN}[OK]${NC} $test_name"
    else
        echo -e "${RED}[FAIL]${NC} $test_name"
    fi
    echo "    Response: $response"
    echo ""
}

# ===========================================
# 1. Test Health Checks
# ===========================================
echo -e "${BLUE}[1] Testing Health Checks...${NC}"
echo "-------------------------------------------"

# Fy-Analysis
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" $ANALYSIS_URL/health)
if [ "$HEALTH" == "200" ]; then
    print_result "Fy-Analysis Health" "OK" "HTTP $HEALTH"
else
    print_result "Fy-Analysis Health" "FAIL" "HTTP $HEALTH"
fi

# Fy-Engine
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" $ENGINE_URL/health)
if [ "$HEALTH" == "200" ]; then
    print_result "Fy-Engine Health" "OK" "HTTP $HEALTH"
else
    print_result "Fy-Engine Health" "FAIL" "HTTP $HEALTH"
fi

# ===========================================
# 2. Test Fy-Analysis Endpoints (Direct)
# ===========================================
echo -e "${BLUE}[2] Testing Fy-Analysis Direct Endpoints...${NC}"
echo "-------------------------------------------"

# Test URL maliciosa (bbva phishing)
echo -e "${YELLOW}Test: URL Phishing (bbva-verificacion.com)${NC}"
RESPONSE=$(curl -s $ANALYSIS_URL/analyze/url \
    -H "Content-Type: application/json" \
    -d '{"url": "https://bbva-verificacion.com/login"}')
echo "Response: $RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test URL segura
echo -e "${YELLOW}Test: URL Segura (bbva.es)${NC}"
RESPONSE=$(curl -s $ANALYSIS_URL/analyze/url \
    -H "Content-Type: application/json" \
    -d '{"url": "https://www.bbva.es"}')
echo "Response: $RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test Email phishing
echo -e "${YELLOW}Test: Email Phishing${NC}"
RESPONSE=$(curl -s $ANALYSIS_URL/analyze/email \
    -H "Content-Type: application/json" \
    -d '{"email": "soporte@bbva-verificacion.com"}')
echo "Response: $RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test Phone premium
echo -e "${YELLOW}Test: Telefono Premium (806)${NC}"
RESPONSE=$(curl -s $ANALYSIS_URL/analyze/phone \
    -H "Content-Type: application/json" \
    -d '{"phone": "+34 806 123 456"}')
echo "Response: $RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# ===========================================
# 3. Test Fy-Engine Chat (Flujo Completo)
# ===========================================
echo -e "${BLUE}[3] Testing Fy-Engine Chat (Flujo Completo)...${NC}"
echo "-------------------------------------------"

# Test mensaje con URL maliciosa
echo -e "${YELLOW}Test: Chat con URL maliciosa${NC}"
RESPONSE=$(curl -s $ENGINE_URL/chat \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test-user-1",
        "message": "Oye, me ha llegado este enlace https://bbva-verificacion.com/login es seguro?",
        "context": []
    }')
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test mensaje con URL segura
echo -e "${YELLOW}Test: Chat con URL segura${NC}"
RESPONSE=$(curl -s $ENGINE_URL/chat \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test-user-1",
        "message": "Es seguro entrar a https://www.bbva.es?",
        "context": []
    }')
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test saludo
echo -e "${YELLOW}Test: Chat saludo (smalltalk)${NC}"
RESPONSE=$(curl -s $ENGINE_URL/chat \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test-user-1",
        "message": "Hola Fy!",
        "context": []
    }')
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Test pregunta
echo -e "${YELLOW}Test: Chat pregunta sobre phishing${NC}"
RESPONSE=$(curl -s $ENGINE_URL/chat \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test-user-1",
        "message": "Que es el phishing?",
        "context": []
    }')
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# ===========================================
# 4. Test API v1 (Endpoint unificado)
# ===========================================
echo -e "${BLUE}[4] Testing API v1 Endpoint Unificado...${NC}"
echo "-------------------------------------------"

echo -e "${YELLOW}Test: POST /api/v1/analyze${NC}"
RESPONSE=$(curl -s $ANALYSIS_URL/api/v1/analyze \
    -H "Content-Type: application/json" \
    -d '{"input": "https://santander-seguridad.net/verify", "type": "url"}')
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

echo "=========================================="
echo "  Tests completados!"
echo "=========================================="