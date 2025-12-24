package urlengine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// Normalizer maneja la normalización y expansión de URLs, emails y teléfonos
type Normalizer struct {
	httpClient       *http.Client
	shortenerDomains map[string]bool
	phoneRegex       *regexp.Regexp
	premiumPrefixes  []string // Prefijos premium españoles
}

// NewNormalizer crea un nuevo normalizador de URLs
func NewNormalizer() *Normalizer {
	return &Normalizer{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// No seguir redirects automáticamente, queremos capturarlos
				return http.ErrUseLastResponse
			},
		},
		shortenerDomains: map[string]bool{
			"bit.ly":       true,
			"tinyurl.com":  true,
			"t.co":         true,
			"goo.gl":       true,
			"ow.ly":        true,
			"is.gd":        true,
			"buff.ly":      true,
			"adf.ly":       true,
			"bl.ink":       true,
			"lnkd.in":      true,
			"rebrand.ly":   true,
			"short.io":     true,
			"cutt.ly":      true,
			"shorturl.at":  true,
			"rb.gy":        true,
			"v.gd":         true,
			"clck.ru":      true,
			"shorte.st":    true,
			"bc.vc":        true,
			"j.mp":         true,
		},
		// Regex para limpiar teléfonos: solo dígitos y +
		phoneRegex: regexp.MustCompile(`[^\d+]`),
		// Prefijos premium españoles (tarificación adicional)
		premiumPrefixes: []string{"803", "806", "807", "905", "907"},
	}
}

// NormalizeInput es el punto de entrada unificado para normalizar cualquier tipo de input
func (n *Normalizer) NormalizeInput(ctx context.Context, input string, inputType checkers.InputType) (*checkers.Indicators, error) {
	switch inputType {
	case checkers.InputTypeURL:
		return n.NormalizeURLToIndicators(ctx, input)
	case checkers.InputTypeEmail:
		return n.NormalizeEmail(ctx, input)
	case checkers.InputTypePhone:
		return n.NormalizePhone(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported input type: %s", inputType)
	}
}

// NormalizeURLToIndicators normaliza una URL y retorna Indicators
func (n *Normalizer) NormalizeURLToIndicators(ctx context.Context, rawURL string) (*checkers.Indicators, error) {
	result := n.Normalize(ctx, rawURL)
	if result.Error != nil {
		return nil, result.Error
	}

	// Determinar la URL final (expandida si es shortener)
	finalURL := result.NormalizedURL
	if result.ExpandedURL != "" {
		finalURL = result.ExpandedURL
	}

	// Extraer TLD
	tld := extractTLD(result.Domain)

	// Parsear para obtener path
	parsed, _ := url.Parse(finalURL)
	path := ""
	if parsed != nil {
		path = parsed.Path
	}

	indicators := &checkers.Indicators{
		Original:   rawURL,
		Normalized: finalURL,
		Hash:       hashSHA256(finalURL),
		InputType:  checkers.InputTypeURL,

		FullURL:    finalURL,
		Domain:     result.Domain,
		DomainHash: hashSHA256(result.Domain),
		IP:         result.IP,
		TLD:        tld,
		Scheme:     result.Scheme,
		Path:       path,
		URLHash:    hashSHA256(finalURL),
	}

	log.Debug().
		Str("original", rawURL).
		Str("normalized", finalURL).
		Str("domain", result.Domain).
		Str("tld", tld).
		Msg("[Normalizer] URL indicators extracted")

	return indicators, nil
}

// NormalizeEmail normaliza una dirección de email y extrae indicadores
func (n *Normalizer) NormalizeEmail(ctx context.Context, rawEmail string) (*checkers.Indicators, error) {
	rawEmail = strings.TrimSpace(rawEmail)
	rawEmail = strings.ToLower(rawEmail)

	// Validar formato de email
	addr, err := mail.ParseAddress(rawEmail)
	if err != nil {
		// Intentar parsear solo el email sin nombre
		if !strings.Contains(rawEmail, "@") {
			return nil, fmt.Errorf("invalid email format: missing @")
		}
		parts := strings.Split(rawEmail, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid email format")
		}
		addr = &mail.Address{Address: rawEmail}
	}

	email := strings.ToLower(addr.Address)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email format")
	}

	user := parts[0]
	domain := parts[1]

	indicators := &checkers.Indicators{
		Original:    rawEmail,
		Normalized:  email,
		Hash:        hashSHA256(email),
		InputType:   checkers.InputTypeEmail,
		Domain:      domain,
		DomainHash:  hashSHA256(domain),
		TLD:         extractTLD(domain),
		EmailUser:   user,
		EmailDomain: domain,
	}

	log.Debug().
		Str("email", email).
		Str("domain", domain).
		Msg("[Normalizer] Email indicators extracted")

	return indicators, nil
}

// NormalizePhone normaliza un número de teléfono y extrae indicadores
func (n *Normalizer) NormalizePhone(ctx context.Context, rawPhone string) (*checkers.Indicators, error) {
	rawPhone = strings.TrimSpace(rawPhone)

	// Limpiar: solo dígitos y +
	cleaned := n.phoneRegex.ReplaceAllString(rawPhone, "")

	// Si empieza con 00, reemplazar por +
	if strings.HasPrefix(cleaned, "00") {
		cleaned = "+" + cleaned[2:]
	}

	// Si no tiene código de país y parece español (9 dígitos), añadir +34
	if !strings.HasPrefix(cleaned, "+") {
		if len(cleaned) == 9 {
			// Teléfono español sin código
			cleaned = "+34" + cleaned
		} else if len(cleaned) > 9 {
			// Asumir que ya tiene código de país
			cleaned = "+" + cleaned
		}
	}

	// Extraer código de país y número nacional
	countryCode := ""
	nationalNum := cleaned

	if strings.HasPrefix(cleaned, "+") {
		// Códigos de país comunes (simplificado)
		if strings.HasPrefix(cleaned, "+34") {
			countryCode = "+34"
			nationalNum = cleaned[3:]
		} else if strings.HasPrefix(cleaned, "+1") {
			countryCode = "+1"
			nationalNum = cleaned[2:]
		} else if strings.HasPrefix(cleaned, "+44") {
			countryCode = "+44"
			nationalNum = cleaned[3:]
		} else {
			// Asumir código de 2-3 dígitos
			if len(cleaned) > 3 {
				countryCode = cleaned[:3]
				nationalNum = cleaned[3:]
			}
		}
	}

	// Detectar si es número premium español
	isPremium := false
	for _, prefix := range n.premiumPrefixes {
		if strings.HasPrefix(nationalNum, prefix) {
			isPremium = true
			break
		}
	}

	indicators := &checkers.Indicators{
		Original:    rawPhone,
		Normalized:  cleaned,
		Hash:        hashSHA256(cleaned),
		InputType:   checkers.InputTypePhone,
		PhoneNumber: cleaned,
		CountryCode: countryCode,
		NationalNum: nationalNum,
		IsPremium:   isPremium,
	}

	log.Debug().
		Str("original", rawPhone).
		Str("normalized", cleaned).
		Str("country", countryCode).
		Bool("premium", isPremium).
		Msg("[Normalizer] Phone indicators extracted")

	return indicators, nil
}

// extractTLD extrae el TLD de un dominio
func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	// Manejar TLDs compuestos como .co.uk, .com.es
	if len(parts) >= 3 {
		secondLast := parts[len(parts)-2]
		if secondLast == "co" || secondLast == "com" || secondLast == "org" || secondLast == "net" {
			return parts[len(parts)-2] + "." + parts[len(parts)-1]
		}
	}
	return parts[len(parts)-1]
}

// NormalizeResult resultado de la normalización
type NormalizeResult struct {
	OriginalURL   string
	NormalizedURL string
	ExpandedURL   string   // URL expandida si era shortener
	Domain        string
	IP            string   // IP resuelta o directa si es IP-based URL
	Scheme        string
	IsShortener   bool
	ExpandChain   []string // Cadena de redirects si es shortener
	Error         error
}

// Normalize normaliza una URL y expande shorteners si es necesario
func (n *Normalizer) Normalize(ctx context.Context, rawURL string) *NormalizeResult {
	result := &NormalizeResult{
		OriginalURL: rawURL,
		ExpandChain: []string{},
	}

	// Limpiar espacios
	rawURL = strings.TrimSpace(rawURL)

	// Añadir scheme si no tiene
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	// Parsear URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		result.Error = fmt.Errorf("invalid URL format: %w", err)
		log.Debug().Err(err).Str("url", rawURL).Msg("[Normalizer] Failed to parse URL")
		return result
	}

	// Normalizar
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Remover puerto default
	host := parsed.Hostname()
	port := parsed.Port()
	if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		parsed.Host = host
	}

	// Remover fragmento (#)
	parsed.Fragment = ""

	// Normalizar path
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	result.NormalizedURL = parsed.String()
	result.Domain = host
	result.Scheme = parsed.Scheme

	log.Debug().
		Str("original", result.OriginalURL).
		Str("normalized", result.NormalizedURL).
		Str("domain", result.Domain).
		Msg("[Normalizer] URL normalized")

	// Verificar si es IP directa
	if ip := net.ParseIP(host); ip != nil {
		result.IP = host
		log.Debug().Str("ip", host).Msg("[Normalizer] URL uses direct IP")
	} else {
		// Resolver DNS para obtener IP
		result.IP = n.resolveIP(ctx, host)
	}

	// Verificar si es shortener y expandir
	if n.isShortener(host) {
		result.IsShortener = true
		log.Debug().Str("domain", host).Msg("[Normalizer] Detected URL shortener, expanding...")
		expanded, chain := n.expandShortener(ctx, result.NormalizedURL)
		if expanded != "" && expanded != result.NormalizedURL {
			result.ExpandedURL = expanded
			result.ExpandChain = chain
			log.Debug().
				Str("expanded", expanded).
				Int("chain_length", len(chain)).
				Msg("[Normalizer] Shortener expanded")
		}
	}

	return result
}

// isShortener verifica si el dominio es un servicio de acortamiento
func (n *Normalizer) isShortener(domain string) bool {
	// Verificar dominio exacto
	if n.shortenerDomains[domain] {
		return true
	}

	// Verificar sin www
	if strings.HasPrefix(domain, "www.") {
		return n.shortenerDomains[domain[4:]]
	}

	return false
}

// expandShortener expande una URL acortada siguiendo redirects
func (n *Normalizer) expandShortener(ctx context.Context, shortURL string) (string, []string) {
	chain := []string{shortURL}
	currentURL := shortURL
	maxRedirects := 10

	for i := 0; i < maxRedirects; i++ {
		select {
		case <-ctx.Done():
			log.Debug().Msg("[Normalizer] Context cancelled during expansion")
			return currentURL, chain
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "HEAD", currentURL, nil)
		if err != nil {
			log.Debug().Err(err).Str("url", currentURL).Msg("[Normalizer] Failed to create request")
			break
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := n.httpClient.Do(req)
		if err != nil {
			log.Debug().Err(err).Str("url", currentURL).Msg("[Normalizer] Failed to follow redirect")
			break
		}
		resp.Body.Close()

		// Verificar si hay redirect
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location == "" {
				break
			}

			// Resolver URL relativa
			base, _ := url.Parse(currentURL)
			ref, err := url.Parse(location)
			if err != nil {
				break
			}
			nextURL := base.ResolveReference(ref).String()

			log.Debug().
				Int("status", resp.StatusCode).
				Str("from", currentURL).
				Str("to", nextURL).
				Msg("[Normalizer] Following redirect")

			chain = append(chain, nextURL)
			currentURL = nextURL

			// Verificar si el nuevo dominio ya no es shortener
			parsed, _ := url.Parse(nextURL)
			if parsed != nil && !n.isShortener(parsed.Hostname()) {
				// Llegamos al destino final
				break
			}
		} else {
			// No hay más redirects
			break
		}
	}

	finalURL := currentURL
	if len(chain) > 1 {
		finalURL = chain[len(chain)-1]
	}

	return finalURL, chain
}

// resolveIP resuelve el dominio a IP
func (n *Normalizer) resolveIP(ctx context.Context, domain string) string {
	resolver := net.Resolver{}
	ips, err := resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		log.Debug().Err(err).Str("domain", domain).Msg("[Normalizer] Failed to resolve IP")
		return ""
	}

	if len(ips) > 0 {
		ip := ips[0].String()
		log.Debug().Str("domain", domain).Str("ip", ip).Msg("[Normalizer] Resolved IP")
		return ip
	}

	return ""
}

// DecodeURL decodifica una URL con encoding (percent-encoding, punycode, etc)
func (n *Normalizer) DecodeURL(rawURL string) string {
	// Decodificar percent-encoding
	decoded, err := url.QueryUnescape(rawURL)
	if err != nil {
		return rawURL
	}

	return decoded
}