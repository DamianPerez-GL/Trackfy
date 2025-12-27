-- ============================================
-- Fy Threats Database Schema v2.0 - OPTIMIZADO
-- Diseñado para millones de registros
-- ============================================

-- Extensiones
CREATE EXTENSION IF NOT EXISTS "pg_trgm";   -- Búsquedas fuzzy
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- SHA256 hashing

-- ============================================
-- TIPOS ENUMERADOS (más eficientes que VARCHAR)
-- ============================================
DO $$ BEGIN
    CREATE TYPE threat_type_enum AS ENUM (
        'phishing', 'malware', 'scam', 'spam',
        'vishing', 'smishing', 'premium_fraud',
        'ransomware', 'cryptojacking', 'other'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE severity_enum AS ENUM ('low', 'medium', 'high', 'critical');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE source_enum AS ENUM (
        'urlhaus', 'phishtank', 'virustotal', 'webrisk',
        'urlscan', 'manual', 'user_report', 'honeypot',
        'partner', 'osint'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE input_type_enum AS ENUM ('url', 'email', 'phone', 'domain');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE report_status_enum AS ENUM ('pending', 'reviewed', 'confirmed', 'rejected');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================
-- TABLA: threat_domains (PRINCIPAL)
-- Solo dominios, NO URLs completas
-- ============================================
CREATE TABLE IF NOT EXISTS threat_domains (
    -- PK: hash binario del dominio (32 bytes vs 64 chars)
    domain_hash BYTEA PRIMARY KEY,

    -- Dominio normalizado (lowercase, sin www)
    domain VARCHAR(253) NOT NULL,  -- Max DNS length

    -- Clasificación (ENUMs = 4 bytes vs VARCHAR)
    threat_type threat_type_enum NOT NULL,
    severity severity_enum NOT NULL DEFAULT 'medium',
    confidence SMALLINT NOT NULL DEFAULT 80,  -- 0-100, SMALLINT = 2 bytes

    -- Fuente
    source source_enum NOT NULL,
    source_id VARCHAR(64),  -- ID en fuente original

    -- Análisis
    impersonates_hash BYTEA,  -- FK a whitelist si es typosquatting
    tld VARCHAR(20),

    -- Contadores (para priorización)
    hit_count INTEGER DEFAULT 0,
    report_count SMALLINT DEFAULT 1,

    -- Timestamps (sin timezone = 8 bytes vs 12)
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,  -- NULL = no expira

    -- Flags compactos (1 byte total con bit fields)
    flags SMALLINT DEFAULT 0
    -- bit 0: is_active
    -- bit 1: verified
    -- bit 2: is_typosquatting
    -- bit 3: has_ssl
    -- bit 4: is_parked
);

-- Índices optimizados
CREATE INDEX IF NOT EXISTS idx_domains_domain ON threat_domains USING HASH(domain);
CREATE INDEX IF NOT EXISTS idx_domains_active ON threat_domains(domain_hash) WHERE (flags & 1) = 1;
CREATE INDEX IF NOT EXISTS idx_domains_type ON threat_domains(threat_type) WHERE (flags & 1) = 1;
CREATE INDEX IF NOT EXISTS idx_domains_expires ON threat_domains(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_domains_last_seen ON threat_domains(last_seen DESC);

-- ============================================
-- TABLA: threat_paths (URLs = dominio + path)
-- Referencia a threat_domains para evitar duplicación
-- ============================================
CREATE TABLE IF NOT EXISTS threat_paths (
    -- PK compuesto: hash del path
    path_hash BYTEA PRIMARY KEY,

    -- FK al dominio
    domain_hash BYTEA NOT NULL REFERENCES threat_domains(domain_hash) ON DELETE CASCADE,

    -- Solo el path (sin dominio)
    path VARCHAR(2048),  -- Max URL path

    -- Clasificación (puede diferir del dominio)
    threat_type threat_type_enum NOT NULL,
    severity severity_enum NOT NULL DEFAULT 'medium',
    confidence SMALLINT NOT NULL DEFAULT 80,

    -- Metadatos específicos del path
    payload_type VARCHAR(20),  -- exe, js, doc, apk
    target_brand VARCHAR(50),

    -- Source
    source source_enum NOT NULL,

    -- Timestamps
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Flags
    flags SMALLINT DEFAULT 1  -- bit 0: is_active
);

-- Índice para buscar paths de un dominio
CREATE INDEX IF NOT EXISTS idx_paths_domain ON threat_paths(domain_hash);
CREATE INDEX IF NOT EXISTS idx_paths_active ON threat_paths(path_hash) WHERE (flags & 1) = 1;

-- ============================================
-- TABLA: threat_emails
-- Optimizada para búsqueda por hash
-- ============================================
CREATE TABLE IF NOT EXISTS threat_emails (
    email_hash BYTEA PRIMARY KEY,

    -- Email normalizado (lowercase)
    email VARCHAR(254) NOT NULL,  -- RFC 5321 max

    -- FK al dominio del email (reutiliza threat_domains si existe)
    domain_hash BYTEA,

    -- Clasificación
    threat_type threat_type_enum NOT NULL,
    severity severity_enum NOT NULL DEFAULT 'medium',
    confidence SMALLINT NOT NULL DEFAULT 80,

    -- Suplantación
    impersonates VARCHAR(50),

    -- Source
    source source_enum NOT NULL,

    -- Contadores
    report_count SMALLINT DEFAULT 1,

    -- Timestamps
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Flags
    flags SMALLINT DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_emails_domain ON threat_emails(domain_hash) WHERE domain_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_emails_active ON threat_emails(email_hash) WHERE (flags & 1) = 1;

-- ============================================
-- TABLA: threat_phones
-- Optimizada para teléfonos españoles
-- ============================================
CREATE TABLE IF NOT EXISTS threat_phones (
    -- Número nacional sin código país (9 dígitos para España)
    phone_national VARCHAR(15) PRIMARY KEY,

    -- Código país
    country_code VARCHAR(4) NOT NULL DEFAULT '34',

    -- Clasificación
    threat_type threat_type_enum NOT NULL,
    severity severity_enum NOT NULL DEFAULT 'medium',
    confidence SMALLINT NOT NULL DEFAULT 80,

    -- Descripción corta del fraude
    description VARCHAR(200),

    -- Source
    source source_enum NOT NULL,

    -- Contadores
    report_count SMALLINT DEFAULT 1,

    -- Timestamps
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Flags (bit 0: active, bit 1: is_premium, bit 2: verified)
    flags SMALLINT DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_phones_premium ON threat_phones(phone_national) WHERE (flags & 2) = 2;
CREATE INDEX IF NOT EXISTS idx_phones_active ON threat_phones(phone_national) WHERE (flags & 1) = 1;
CREATE INDEX IF NOT EXISTS idx_phones_prefix ON threat_phones(LEFT(phone_national, 3));

-- ============================================
-- TABLA: whitelist_domains
-- Dominios legítimos conocidos
-- ============================================
CREATE TABLE IF NOT EXISTS whitelist_domains (
    domain_hash BYTEA PRIMARY KEY,
    domain VARCHAR(253) NOT NULL UNIQUE,

    -- Categoría
    category VARCHAR(20),  -- banking, telecom, government, ecommerce
    brand VARCHAR(50),
    country CHAR(2) DEFAULT 'ES',

    -- Metadatos
    official_name VARCHAR(100),

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_whitelist_domain ON whitelist_domains USING HASH(domain);
CREATE INDEX IF NOT EXISTS idx_whitelist_brand ON whitelist_domains(brand) WHERE brand IS NOT NULL;

-- ============================================
-- TABLA: domain_tags (many-to-many normalizado)
-- Evita arrays TEXT[] que son ineficientes
-- ============================================
CREATE TABLE IF NOT EXISTS tags (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS domain_tags (
    domain_hash BYTEA NOT NULL REFERENCES threat_domains(domain_hash) ON DELETE CASCADE,
    tag_id SMALLINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (domain_hash, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_domain_tags_tag ON domain_tags(tag_id);

-- ============================================
-- TABLA: analysis_cache (con TTL automático)
-- ============================================
CREATE TABLE IF NOT EXISTS analysis_cache (
    input_hash BYTEA PRIMARY KEY,
    input_type input_type_enum NOT NULL,

    -- Resultado compacto
    risk_score SMALLINT NOT NULL,
    risk_level VARCHAR(10) NOT NULL,

    -- Response completo comprimido
    response BYTEA NOT NULL,  -- JSONB comprimido con pg_lz

    -- TTL
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    hit_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_cache_expires ON analysis_cache(expires_at);

-- ============================================
-- TABLA: user_reports (particionada por mes)
-- ============================================
CREATE TABLE IF NOT EXISTS user_reports (
    id BIGSERIAL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    input_type input_type_enum NOT NULL,
    input_hash BYTEA NOT NULL,

    report_type VARCHAR(20) NOT NULL,
    description VARCHAR(500),

    -- Usuario
    user_id VARCHAR(64),
    user_ip INET,

    -- Estado
    status report_status_enum DEFAULT 'pending',
    reviewed_at TIMESTAMP,

    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

-- Crear particiones para los próximos 12 meses
DO $$
DECLARE
    start_date DATE := DATE_TRUNC('month', CURRENT_DATE);
    partition_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..11 LOOP
        partition_date := start_date + (i || ' months')::INTERVAL;
        partition_name := 'user_reports_' || TO_CHAR(partition_date, 'YYYY_MM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF user_reports
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            partition_date,
            partition_date + INTERVAL '1 month'
        );
    END LOOP;
END $$;

-- ============================================
-- TABLA: sync_status (para sincronización)
-- ============================================
CREATE TABLE IF NOT EXISTS sync_status (
    source source_enum PRIMARY KEY,
    last_sync TIMESTAMP NOT NULL DEFAULT NOW(),
    last_count INTEGER DEFAULT 0,
    last_error TEXT,
    next_sync TIMESTAMP
);

-- ============================================
-- FUNCIONES OPTIMIZADAS
-- ============================================

-- Hash SHA256 a BYTEA (más eficiente que VARCHAR)
CREATE OR REPLACE FUNCTION sha256_bytea(input TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN decode(encode(digest(input, 'sha256'), 'hex'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Buscar dominio por texto
CREATE OR REPLACE FUNCTION find_threat_domain(p_domain TEXT)
RETURNS TABLE (
    domain_hash BYTEA,
    domain VARCHAR,
    threat_type threat_type_enum,
    severity severity_enum,
    confidence SMALLINT,
    impersonates VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        td.domain_hash,
        td.domain,
        td.threat_type,
        td.severity,
        td.confidence,
        wd.domain as impersonates
    FROM threat_domains td
    LEFT JOIN whitelist_domains wd ON td.impersonates_hash = wd.domain_hash
    WHERE td.domain = LOWER(p_domain)
      AND (td.flags & 1) = 1
      AND (td.expires_at IS NULL OR td.expires_at > NOW());
END;
$$ LANGUAGE plpgsql;

-- Upsert optimizado para dominios
CREATE OR REPLACE FUNCTION upsert_threat_domain(
    p_domain TEXT,
    p_threat_type threat_type_enum,
    p_severity severity_enum,
    p_confidence SMALLINT,
    p_source source_enum,
    p_impersonates TEXT DEFAULT NULL
) RETURNS VOID AS $$
DECLARE
    v_domain_hash BYTEA;
    v_impersonates_hash BYTEA;
BEGIN
    v_domain_hash := sha256_bytea(LOWER(p_domain));

    IF p_impersonates IS NOT NULL THEN
        SELECT domain_hash INTO v_impersonates_hash
        FROM whitelist_domains
        WHERE domain = LOWER(p_impersonates);
    END IF;

    INSERT INTO threat_domains (
        domain_hash, domain, threat_type, severity, confidence,
        source, impersonates_hash, tld, first_seen, last_seen, flags
    ) VALUES (
        v_domain_hash,
        LOWER(p_domain),
        p_threat_type,
        p_severity,
        p_confidence,
        p_source,
        v_impersonates_hash,
        SUBSTRING(p_domain FROM '\.([^.]+)$'),
        NOW(),
        NOW(),
        1  -- is_active
    )
    ON CONFLICT (domain_hash) DO UPDATE SET
        last_seen = NOW(),
        hit_count = threat_domains.hit_count + 1,
        confidence = GREATEST(threat_domains.confidence, EXCLUDED.confidence);
END;
$$ LANGUAGE plpgsql;

-- Limpiar cache expirado
CREATE OR REPLACE FUNCTION cleanup_expired_cache() RETURNS INTEGER AS $$
DECLARE
    deleted INTEGER;
BEGIN
    DELETE FROM analysis_cache WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted = ROW_COUNT;
    RETURN deleted;
END;
$$ LANGUAGE plpgsql;

-- Limpiar dominios expirados
CREATE OR REPLACE FUNCTION cleanup_expired_threats() RETURNS INTEGER AS $$
DECLARE
    deleted INTEGER;
BEGIN
    DELETE FROM threat_domains WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted = ROW_COUNT;
    RETURN deleted;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- VISTA: Estadísticas
-- ============================================
CREATE OR REPLACE VIEW threat_stats AS
SELECT
    'domains' as type,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE (flags & 1) = 1) as active,
    COUNT(*) FILTER (WHERE threat_type = 'phishing') as phishing,
    COUNT(*) FILTER (WHERE threat_type = 'malware') as malware,
    COUNT(*) FILTER (WHERE threat_type = 'scam') as scam,
    pg_size_pretty(pg_total_relation_size('threat_domains')) as size
FROM threat_domains
UNION ALL
SELECT
    'paths' as type,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE (flags & 1) = 1) as active,
    COUNT(*) FILTER (WHERE threat_type = 'phishing') as phishing,
    COUNT(*) FILTER (WHERE threat_type = 'malware') as malware,
    COUNT(*) FILTER (WHERE threat_type = 'scam') as scam,
    pg_size_pretty(pg_total_relation_size('threat_paths')) as size
FROM threat_paths
UNION ALL
SELECT
    'emails' as type,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE (flags & 1) = 1) as active,
    COUNT(*) FILTER (WHERE threat_type = 'phishing') as phishing,
    0 as malware,
    COUNT(*) FILTER (WHERE threat_type = 'scam') as scam,
    pg_size_pretty(pg_total_relation_size('threat_emails')) as size
FROM threat_emails
UNION ALL
SELECT
    'phones' as type,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE (flags & 1) = 1) as active,
    0 as phishing,
    0 as malware,
    COUNT(*) FILTER (WHERE threat_type = 'scam') as scam,
    pg_size_pretty(pg_total_relation_size('threat_phones')) as size
FROM threat_phones;

-- ============================================
-- DATOS DE EJEMPLO
-- ============================================

-- Tags comunes
INSERT INTO tags (name) VALUES
    ('banking'), ('telecom'), ('government'), ('ecommerce'),
    ('spain'), ('typosquatting'), ('premium'), ('sms'),
    ('email'), ('social'), ('crypto'), ('romance')
ON CONFLICT (name) DO NOTHING;

-- Whitelist de dominios españoles
INSERT INTO whitelist_domains (domain_hash, domain, category, brand, country, official_name) VALUES
    -- BANCOS
    (sha256_bytea('bbva.es'), 'bbva.es', 'banking', 'BBVA', 'ES', 'Banco Bilbao Vizcaya Argentaria'),
    (sha256_bytea('bbva.com'), 'bbva.com', 'banking', 'BBVA', 'ES', 'Banco Bilbao Vizcaya Argentaria'),
    (sha256_bytea('santander.es'), 'santander.es', 'banking', 'Santander', 'ES', 'Banco Santander'),
    (sha256_bytea('bancosantander.es'), 'bancosantander.es', 'banking', 'Santander', 'ES', 'Banco Santander'),
    (sha256_bytea('caixabank.es'), 'caixabank.es', 'banking', 'CaixaBank', 'ES', 'CaixaBank'),
    (sha256_bytea('lacaixa.es'), 'lacaixa.es', 'banking', 'CaixaBank', 'ES', 'La Caixa'),
    (sha256_bytea('ing.es'), 'ing.es', 'banking', 'ING', 'ES', 'ING Direct'),
    (sha256_bytea('bankinter.com'), 'bankinter.com', 'banking', 'Bankinter', 'ES', 'Bankinter'),
    (sha256_bytea('openbank.es'), 'openbank.es', 'banking', 'Openbank', 'ES', 'Openbank'),
    (sha256_bytea('unicajabanco.es'), 'unicajabanco.es', 'banking', 'Unicaja', 'ES', 'Unicaja Banco'),
    (sha256_bytea('abanca.com'), 'abanca.com', 'banking', 'Abanca', 'ES', 'Abanca'),
    (sha256_bytea('kutxabank.es'), 'kutxabank.es', 'banking', 'Kutxabank', 'ES', 'Kutxabank'),
    (sha256_bytea('sabadell.com'), 'sabadell.com', 'banking', 'Sabadell', 'ES', 'Banco Sabadell'),
    (sha256_bytea('bancsabadell.com'), 'bancsabadell.com', 'banking', 'Sabadell', 'ES', 'Banco Sabadell'),
    -- TELECOMUNICACIONES
    (sha256_bytea('movistar.es'), 'movistar.es', 'telecom', 'Movistar', 'ES', 'Movistar'),
    (sha256_bytea('vodafone.es'), 'vodafone.es', 'telecom', 'Vodafone', 'ES', 'Vodafone'),
    (sha256_bytea('orange.es'), 'orange.es', 'telecom', 'Orange', 'ES', 'Orange'),
    (sha256_bytea('masmovil.es'), 'masmovil.es', 'telecom', 'MasMovil', 'ES', 'MasMovil'),
    (sha256_bytea('yoigo.com'), 'yoigo.com', 'telecom', 'Yoigo', 'ES', 'Yoigo'),
    (sha256_bytea('jazztel.com'), 'jazztel.com', 'telecom', 'Jazztel', 'ES', 'Jazztel'),
    -- ORGANISMOS OFICIALES DEL ESTADO
    (sha256_bytea('dgt.es'), 'dgt.es', 'government', 'DGT', 'ES', 'Dirección General de Tráfico'),
    (sha256_bytea('dgt.gob.es'), 'dgt.gob.es', 'government', 'DGT', 'ES', 'Dirección General de Tráfico'),
    (sha256_bytea('correos.es'), 'correos.es', 'government', 'Correos', 'ES', 'Correos'),
    (sha256_bytea('agenciatributaria.es'), 'agenciatributaria.es', 'government', 'AEAT', 'ES', 'Agencia Tributaria'),
    (sha256_bytea('agenciatributaria.gob.es'), 'agenciatributaria.gob.es', 'government', 'AEAT', 'ES', 'Agencia Tributaria'),
    (sha256_bytea('hacienda.gob.es'), 'hacienda.gob.es', 'government', 'Hacienda', 'ES', 'Ministerio de Hacienda'),
    (sha256_bytea('seg-social.es'), 'seg-social.es', 'government', 'Seguridad Social', 'ES', 'Seguridad Social'),
    (sha256_bytea('seg-social.gob.es'), 'seg-social.gob.es', 'government', 'Seguridad Social', 'ES', 'Seguridad Social'),
    (sha256_bytea('sepe.es'), 'sepe.es', 'government', 'SEPE', 'ES', 'Servicio Público de Empleo Estatal'),
    (sha256_bytea('sepe.gob.es'), 'sepe.gob.es', 'government', 'SEPE', 'ES', 'Servicio Público de Empleo Estatal'),
    (sha256_bytea('policia.es'), 'policia.es', 'government', 'Policia', 'ES', 'Policía Nacional'),
    (sha256_bytea('policia.gob.es'), 'policia.gob.es', 'government', 'Policia', 'ES', 'Policía Nacional'),
    (sha256_bytea('guardiacivil.es'), 'guardiacivil.es', 'government', 'Guardia Civil', 'ES', 'Guardia Civil'),
    (sha256_bytea('guardiacivil.gob.es'), 'guardiacivil.gob.es', 'government', 'Guardia Civil', 'ES', 'Guardia Civil'),
    (sha256_bytea('mjusticia.gob.es'), 'mjusticia.gob.es', 'government', 'Justicia', 'ES', 'Ministerio de Justicia'),
    (sha256_bytea('interior.gob.es'), 'interior.gob.es', 'government', 'Interior', 'ES', 'Ministerio del Interior'),
    (sha256_bytea('mscbs.gob.es'), 'mscbs.gob.es', 'government', 'Sanidad', 'ES', 'Ministerio de Sanidad'),
    (sha256_bytea('sanidad.gob.es'), 'sanidad.gob.es', 'government', 'Sanidad', 'ES', 'Ministerio de Sanidad'),
    (sha256_bytea('educacion.gob.es'), 'educacion.gob.es', 'government', 'Educacion', 'ES', 'Ministerio de Educación'),
    (sha256_bytea('boe.es'), 'boe.es', 'government', 'BOE', 'ES', 'Boletín Oficial del Estado'),
    (sha256_bytea('congreso.es'), 'congreso.es', 'government', 'Congreso', 'ES', 'Congreso de los Diputados'),
    (sha256_bytea('senado.es'), 'senado.es', 'government', 'Senado', 'ES', 'Senado de España'),
    (sha256_bytea('lamoncloa.gob.es'), 'lamoncloa.gob.es', 'government', 'Moncloa', 'ES', 'Presidencia del Gobierno'),
    (sha256_bytea('exteriores.gob.es'), 'exteriores.gob.es', 'government', 'Exteriores', 'ES', 'Ministerio de Asuntos Exteriores'),
    (sha256_bytea('defensa.gob.es'), 'defensa.gob.es', 'government', 'Defensa', 'ES', 'Ministerio de Defensa'),
    (sha256_bytea('ine.es'), 'ine.es', 'government', 'INE', 'ES', 'Instituto Nacional de Estadística'),
    (sha256_bytea('cnmv.es'), 'cnmv.es', 'government', 'CNMV', 'ES', 'Comisión Nacional del Mercado de Valores'),
    (sha256_bytea('bde.es'), 'bde.es', 'government', 'Banco de España', 'ES', 'Banco de España'),
    -- SERVICIOS PÚBLICOS Y UTILIDADES
    (sha256_bytea('iberdrola.es'), 'iberdrola.es', 'utilities', 'Iberdrola', 'ES', 'Iberdrola'),
    (sha256_bytea('endesa.com'), 'endesa.com', 'utilities', 'Endesa', 'ES', 'Endesa'),
    (sha256_bytea('naturgy.es'), 'naturgy.es', 'utilities', 'Naturgy', 'ES', 'Naturgy'),
    (sha256_bytea('repsol.es'), 'repsol.es', 'utilities', 'Repsol', 'ES', 'Repsol'),
    (sha256_bytea('renfe.com'), 'renfe.com', 'transport', 'Renfe', 'ES', 'Renfe'),
    (sha256_bytea('aena.es'), 'aena.es', 'transport', 'Aena', 'ES', 'Aena Aeropuertos'),
    -- ECOMMERCE Y SERVICIOS
    (sha256_bytea('amazon.es'), 'amazon.es', 'ecommerce', 'Amazon', 'ES', 'Amazon España'),
    (sha256_bytea('elcorteingles.es'), 'elcorteingles.es', 'ecommerce', 'El Corte Ingles', 'ES', 'El Corte Inglés'),
    (sha256_bytea('mediamarkt.es'), 'mediamarkt.es', 'ecommerce', 'MediaMarkt', 'ES', 'MediaMarkt'),
    (sha256_bytea('pccomponentes.com'), 'pccomponentes.com', 'ecommerce', 'PcComponentes', 'ES', 'PcComponentes'),
    (sha256_bytea('wallapop.com'), 'wallapop.com', 'ecommerce', 'Wallapop', 'ES', 'Wallapop'),
    -- SEGUROS
    (sha256_bytea('mapfre.es'), 'mapfre.es', 'insurance', 'Mapfre', 'ES', 'Mapfre'),
    (sha256_bytea('axa.es'), 'axa.es', 'insurance', 'AXA', 'ES', 'AXA Seguros'),
    (sha256_bytea('mutua.es'), 'mutua.es', 'insurance', 'Mutua', 'ES', 'Mutua Madrileña'),
    (sha256_bytea('generali.es'), 'generali.es', 'insurance', 'Generali', 'ES', 'Generali Seguros')
ON CONFLICT (domain_hash) DO NOTHING;

-- Dominios de ejemplo (amenazas)
SELECT upsert_threat_domain('bbva-verificacion.com', 'phishing'::threat_type_enum, 'high'::severity_enum, 95::SMALLINT, 'manual'::source_enum, 'bbva.es');
SELECT upsert_threat_domain('santander-seguridad.net', 'phishing'::threat_type_enum, 'high'::severity_enum, 95::SMALLINT, 'manual'::source_enum, 'santander.es');
SELECT upsert_threat_domain('caixabank-login.xyz', 'phishing'::threat_type_enum, 'high'::severity_enum, 92::SMALLINT, 'manual'::source_enum, 'caixabank.es');
SELECT upsert_threat_domain('movistar-factura.tk', 'phishing'::threat_type_enum, 'medium'::severity_enum, 85::SMALLINT, 'manual'::source_enum, 'movistar.es');
SELECT upsert_threat_domain('correos-paquete.xyz', 'phishing'::threat_type_enum, 'high'::severity_enum, 90::SMALLINT, 'manual'::source_enum, 'correos.es');

-- Teléfonos de ejemplo
INSERT INTO threat_phones (phone_national, country_code, threat_type, severity, confidence, source, description, flags) VALUES
    ('806123456', '34', 'premium_fraud', 'high', 90, 'manual', 'Línea premium fraudulenta', 3),  -- active + premium
    ('807654321', '34', 'premium_fraud', 'high', 90, 'manual', 'SMS falso premio', 3),
    ('612345678', '34', 'vishing', 'medium', 75, 'user_report', 'Se hacen pasar por BBVA', 1),
    ('698765432', '34', 'scam', 'medium', 70, 'user_report', 'Estafa romántica', 1),
    ('687120072', '34', 'vishing', 'high', 85, 'user_report', 'Llamadas fraudulentas recurrentes - posible estafa telefónica', 1),
    ('666555444', '34', 'scam', 'high', 90, 'manual', 'Número usado para estafas de WhatsApp', 1),
    ('622334455', '34', 'smishing', 'high', 88, 'manual', 'SMS fraudulentos suplantando a Correos', 1)
ON CONFLICT (phone_national) DO NOTHING;

-- Emails de ejemplo
INSERT INTO threat_emails (email_hash, email, threat_type, severity, confidence, source, impersonates, flags) VALUES
    (sha256_bytea('soporte@bbva-verificacion.com'), 'soporte@bbva-verificacion.com', 'phishing', 'high', 95, 'manual', 'BBVA', 1),
    (sha256_bytea('no-reply@correos-tracking.xyz'), 'no-reply@correos-tracking.xyz', 'phishing', 'high', 90, 'manual', 'Correos', 1),
    (sha256_bytea('premio@amazon-winner.tk'), 'premio@amazon-winner.tk', 'scam', 'medium', 85, 'manual', 'Amazon', 1)
ON CONFLICT (email_hash) DO NOTHING;

-- Sync status inicial
INSERT INTO sync_status (source, last_sync, last_count) VALUES
    ('urlhaus', NOW(), 0),
    ('phishtank', NOW(), 0)
ON CONFLICT (source) DO NOTHING;

-- ============================================
-- MANTENIMIENTO PROGRAMADO
-- ============================================
-- Ejecutar diariamente con pg_cron o similar:
-- SELECT cleanup_expired_cache();
-- SELECT cleanup_expired_threats();
-- VACUUM ANALYZE threat_domains;
-- VACUUM ANALYZE threat_paths;

-- Mensaje de confirmación
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Fy Threats Database v2.0 - OPTIMIZADO';
    RAISE NOTICE 'Diseñado para millones de registros';
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Tablas: threat_domains, threat_paths, threat_emails, threat_phones';
    RAISE NOTICE 'Optimizaciones:';
    RAISE NOTICE '  - BYTEA hashes (32 bytes vs 64 chars)';
    RAISE NOTICE '  - ENUMs (4 bytes vs VARCHAR)';
    RAISE NOTICE '  - Flags compactos (bit fields)';
    RAISE NOTICE '  - Índices HASH para lookups exactos';
    RAISE NOTICE '  - Particionamiento por fecha en reports';
    RAISE NOTICE '  - Separación dominio/path para evitar duplicados';
    RAISE NOTICE '===========================================';
END $$;