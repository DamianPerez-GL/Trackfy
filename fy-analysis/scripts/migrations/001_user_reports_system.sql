-- ============================================
-- MIGRACIÓN: Sistema de Reportes de Usuarios v1.0
-- Con protección anti-spam y sistema de confianza
-- ============================================

-- ============================================
-- 1. TABLA: user_trust_scores
-- Sistema de confianza por usuario
-- ============================================
CREATE TABLE IF NOT EXISTS user_trust_scores (
    user_id VARCHAR(64) PRIMARY KEY,

    -- Puntuación de confianza (0-100)
    -- Nuevos usuarios empiezan en 50
    trust_score SMALLINT NOT NULL DEFAULT 50,

    -- Contadores de reportes
    total_reports INTEGER DEFAULT 0,
    confirmed_reports INTEGER DEFAULT 0,    -- Reportes que fueron confirmados como válidos
    rejected_reports INTEGER DEFAULT 0,     -- Reportes rechazados (spam/falsos)

    -- Metadata
    first_report_at TIMESTAMP,
    last_report_at TIMESTAMP,

    -- Flags (bit 0: is_banned, bit 1: is_trusted_reviewer)
    flags SMALLINT DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_trust_scores IS 'Sistema de confianza por usuario para ponderar reportes';
COMMENT ON COLUMN user_trust_scores.trust_score IS '0-100: <20=bajo, 20-50=normal, 50-80=confiable, >80=muy confiable';

-- ============================================
-- 2. TABLA: reported_urls (URLs reportadas por usuarios)
-- Tabla agregada para evitar duplicados
-- ============================================
CREATE TABLE IF NOT EXISTS reported_urls (
    -- Hash de la URL normalizada
    url_hash BYTEA PRIMARY KEY,

    -- URL normalizada (sin tracking params, lowercase)
    url VARCHAR(2048) NOT NULL,

    -- Dominio extraído
    domain VARCHAR(253) NOT NULL,

    -- Tipo de amenaza reportada más común
    primary_threat_type threat_type_enum,

    -- Puntuación agregada de reportes (0-100)
    -- Calculada en base a cantidad y calidad de reportes
    aggregated_score SMALLINT NOT NULL DEFAULT 0,

    -- Contadores
    total_reports INTEGER NOT NULL DEFAULT 0,
    unique_reporters INTEGER NOT NULL DEFAULT 0,  -- Usuarios únicos que reportaron

    -- Suma ponderada de confianza de reportadores
    weighted_trust_sum NUMERIC(10,2) DEFAULT 0,

    -- Estado de revisión
    status report_status_enum DEFAULT 'pending',
    reviewed_by VARCHAR(64),
    reviewed_at TIMESTAMP,

    -- Si fue promovido a threat_domains
    promoted_to_threats BOOLEAN DEFAULT FALSE,
    promoted_at TIMESTAMP,

    -- Timestamps
    first_reported_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_reported_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Flags
    flags SMALLINT DEFAULT 1  -- bit 0: is_active
);

CREATE INDEX IF NOT EXISTS idx_reported_urls_domain ON reported_urls USING HASH(domain);
CREATE INDEX IF NOT EXISTS idx_reported_urls_score ON reported_urls(aggregated_score DESC) WHERE (flags & 1) = 1;
CREATE INDEX IF NOT EXISTS idx_reported_urls_pending ON reported_urls(first_reported_at) WHERE status = 'pending';

COMMENT ON TABLE reported_urls IS 'URLs reportadas por usuarios con puntuación agregada anti-spam';

-- ============================================
-- 3. TABLA: user_url_reports (Reportes individuales)
-- Cada reporte de usuario
-- ============================================
CREATE TABLE IF NOT EXISTS user_url_reports (
    id BIGSERIAL PRIMARY KEY,

    -- Referencia a la URL reportada
    url_hash BYTEA NOT NULL REFERENCES reported_urls(url_hash) ON DELETE CASCADE,

    -- Usuario que reporta
    user_id VARCHAR(64) NOT NULL,

    -- Confianza del usuario al momento del reporte
    user_trust_at_report SMALLINT NOT NULL,

    -- Tipo de amenaza reportada
    threat_type threat_type_enum NOT NULL,

    -- Descripción opcional
    description VARCHAR(500),

    -- Contexto del reporte
    report_context VARCHAR(50),  -- 'chat', 'manual', 'browser_extension'

    -- Metadata
    user_ip INET,
    user_agent VARCHAR(255),

    -- Estado individual
    status report_status_enum DEFAULT 'pending',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Evitar reportes duplicados del mismo usuario para la misma URL
    UNIQUE(url_hash, user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_url_reports_user ON user_url_reports(user_id);
CREATE INDEX IF NOT EXISTS idx_user_url_reports_created ON user_url_reports(created_at DESC);

COMMENT ON TABLE user_url_reports IS 'Reportes individuales de usuarios con deduplicación';

-- ============================================
-- 4. FUNCIONES PARA SISTEMA DE CONFIANZA
-- ============================================

-- Función para obtener o crear trust score de usuario
CREATE OR REPLACE FUNCTION get_or_create_user_trust(p_user_id VARCHAR(64))
RETURNS SMALLINT AS $$
DECLARE
    v_trust SMALLINT;
BEGIN
    SELECT trust_score INTO v_trust
    FROM user_trust_scores
    WHERE user_id = p_user_id;

    IF NOT FOUND THEN
        INSERT INTO user_trust_scores (user_id, trust_score)
        VALUES (p_user_id, 50)
        ON CONFLICT (user_id) DO NOTHING
        RETURNING trust_score INTO v_trust;

        IF v_trust IS NULL THEN
            v_trust := 50;
        END IF;
    END IF;

    RETURN v_trust;
END;
$$ LANGUAGE plpgsql;

-- Función para calcular score agregado de una URL
-- ANTI-SPAM: Requiere múltiples reportadores confiables
CREATE OR REPLACE FUNCTION calculate_url_aggregated_score(p_url_hash BYTEA)
RETURNS SMALLINT AS $$
DECLARE
    v_unique_reporters INTEGER;
    v_weighted_trust NUMERIC;
    v_avg_trust NUMERIC;
    v_base_score NUMERIC;
    v_final_score SMALLINT;
BEGIN
    -- Obtener métricas de reportes
    SELECT
        unique_reporters,
        weighted_trust_sum,
        CASE WHEN unique_reporters > 0
             THEN weighted_trust_sum / unique_reporters
             ELSE 0 END
    INTO v_unique_reporters, v_weighted_trust, v_avg_trust
    FROM reported_urls
    WHERE url_hash = p_url_hash;

    IF v_unique_reporters IS NULL OR v_unique_reporters = 0 THEN
        RETURN 0;
    END IF;

    -- ALGORITMO ANTI-SPAM:
    -- 1. Base: Confianza promedio de reportadores (0-100)
    -- 2. Multiplicador por cantidad de reportadores únicos
    -- 3. Cap máximo según cantidad de reportadores

    v_base_score := v_avg_trust;

    -- Multiplicador logarítmico por cantidad de reportadores
    -- 1 reportador: x0.3, 2: x0.5, 5: x0.7, 10: x0.85, 20+: x1.0
    v_base_score := v_base_score * (
        CASE
            WHEN v_unique_reporters = 1 THEN 0.3
            WHEN v_unique_reporters = 2 THEN 0.5
            WHEN v_unique_reporters BETWEEN 3 AND 4 THEN 0.6
            WHEN v_unique_reporters BETWEEN 5 AND 9 THEN 0.7
            WHEN v_unique_reporters BETWEEN 10 AND 19 THEN 0.85
            ELSE 1.0
        END
    );

    -- Cap máximo según reportadores (previene 4 personas inflando score)
    -- 1 reportador: max 30, 2: max 45, 3-4: max 60, 5-9: max 75, 10+: max 100
    v_final_score := LEAST(
        v_base_score::SMALLINT,
        CASE
            WHEN v_unique_reporters = 1 THEN 30
            WHEN v_unique_reporters = 2 THEN 45
            WHEN v_unique_reporters BETWEEN 3 AND 4 THEN 60
            WHEN v_unique_reporters BETWEEN 5 AND 9 THEN 75
            ELSE 100
        END
    );

    RETURN v_final_score;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_url_aggregated_score IS
'Calcula score anti-spam: requiere múltiples reportadores confiables para scores altos';

-- ============================================
-- 5. FUNCIÓN PRINCIPAL: Reportar URL
-- ============================================
CREATE OR REPLACE FUNCTION report_url(
    p_url VARCHAR(2048),
    p_domain VARCHAR(253),
    p_user_id VARCHAR(64),
    p_threat_type threat_type_enum,
    p_description VARCHAR(500) DEFAULT NULL,
    p_context VARCHAR(50) DEFAULT 'manual',
    p_user_ip INET DEFAULT NULL,
    p_user_agent VARCHAR(255) DEFAULT NULL
) RETURNS TABLE (
    success BOOLEAN,
    message VARCHAR(200),
    url_score SMALLINT,
    is_new_report BOOLEAN
) AS $$
DECLARE
    v_url_hash BYTEA;
    v_user_trust SMALLINT;
    v_is_new BOOLEAN := FALSE;
    v_existing_report_id BIGINT;
    v_new_score SMALLINT;
    v_is_banned BOOLEAN;
BEGIN
    -- Verificar si el usuario está baneado
    SELECT (flags & 1) = 1 INTO v_is_banned
    FROM user_trust_scores
    WHERE user_id = p_user_id;

    IF v_is_banned THEN
        RETURN QUERY SELECT FALSE, 'Usuario no puede reportar'::VARCHAR(200), 0::SMALLINT, FALSE;
        RETURN;
    END IF;

    -- Calcular hash de la URL
    v_url_hash := sha256_bytea(LOWER(p_url));

    -- Obtener trust score del usuario
    v_user_trust := get_or_create_user_trust(p_user_id);

    -- Verificar si ya existe un reporte de este usuario para esta URL
    SELECT id INTO v_existing_report_id
    FROM user_url_reports
    WHERE url_hash = v_url_hash AND user_id = p_user_id;

    IF v_existing_report_id IS NOT NULL THEN
        -- Ya existe un reporte, no permitir duplicados
        SELECT aggregated_score INTO v_new_score FROM reported_urls WHERE url_hash = v_url_hash;
        RETURN QUERY SELECT FALSE, 'Ya reportaste esta URL anteriormente'::VARCHAR(200), COALESCE(v_new_score, 0::SMALLINT), FALSE;
        RETURN;
    END IF;

    -- Crear o actualizar la URL reportada
    INSERT INTO reported_urls (
        url_hash, url, domain, primary_threat_type,
        total_reports, unique_reporters, weighted_trust_sum,
        first_reported_at, last_reported_at
    ) VALUES (
        v_url_hash, LOWER(p_url), LOWER(p_domain), p_threat_type,
        1, 1, v_user_trust,
        NOW(), NOW()
    )
    ON CONFLICT (url_hash) DO UPDATE SET
        total_reports = reported_urls.total_reports + 1,
        unique_reporters = reported_urls.unique_reporters + 1,
        weighted_trust_sum = reported_urls.weighted_trust_sum + v_user_trust,
        last_reported_at = NOW(),
        -- Actualizar threat_type al más reportado (simplificado: último)
        primary_threat_type = EXCLUDED.primary_threat_type;

    -- Verificar si es nuevo (para saber si fue INSERT o UPDATE)
    GET DIAGNOSTICS v_is_new = ROW_COUNT;
    v_is_new := (v_is_new = 1);

    -- Crear el reporte individual
    INSERT INTO user_url_reports (
        url_hash, user_id, user_trust_at_report, threat_type,
        description, report_context, user_ip, user_agent
    ) VALUES (
        v_url_hash, p_user_id, v_user_trust, p_threat_type,
        p_description, p_context, p_user_ip, p_user_agent
    );

    -- Actualizar stats del usuario
    UPDATE user_trust_scores SET
        total_reports = total_reports + 1,
        last_report_at = NOW(),
        first_report_at = COALESCE(first_report_at, NOW()),
        updated_at = NOW()
    WHERE user_id = p_user_id;

    -- Recalcular score agregado
    v_new_score := calculate_url_aggregated_score(v_url_hash);

    UPDATE reported_urls SET aggregated_score = v_new_score WHERE url_hash = v_url_hash;

    RETURN QUERY SELECT TRUE, 'Reporte registrado correctamente'::VARCHAR(200), v_new_score, v_is_new;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION report_url IS
'Registra un reporte de URL con protección anti-spam y deduplicación';

-- ============================================
-- 6. FUNCIÓN: Buscar URL en reportes
-- ============================================
CREATE OR REPLACE FUNCTION find_reported_url(p_url VARCHAR(2048))
RETURNS TABLE (
    url VARCHAR,
    domain VARCHAR,
    threat_type threat_type_enum,
    aggregated_score SMALLINT,
    total_reports INTEGER,
    unique_reporters INTEGER,
    status report_status_enum,
    first_reported_at TIMESTAMP,
    last_reported_at TIMESTAMP
) AS $$
DECLARE
    v_url_hash BYTEA;
BEGIN
    v_url_hash := sha256_bytea(LOWER(p_url));

    RETURN QUERY
    SELECT
        r.url,
        r.domain,
        r.primary_threat_type,
        r.aggregated_score,
        r.total_reports,
        r.unique_reporters,
        r.status,
        r.first_reported_at,
        r.last_reported_at
    FROM reported_urls r
    WHERE r.url_hash = v_url_hash
      AND (r.flags & 1) = 1;  -- Solo activas
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 7. FUNCIÓN: Actualizar trust score tras revisión
-- ============================================
CREATE OR REPLACE FUNCTION update_user_trust_after_review(
    p_user_id VARCHAR(64),
    p_was_correct BOOLEAN
) RETURNS VOID AS $$
DECLARE
    v_adjustment SMALLINT;
BEGIN
    IF p_was_correct THEN
        -- Reporte correcto: +3 puntos (max 100)
        v_adjustment := 3;
        UPDATE user_trust_scores SET
            confirmed_reports = confirmed_reports + 1,
            trust_score = LEAST(trust_score + v_adjustment, 100),
            updated_at = NOW()
        WHERE user_id = p_user_id;
    ELSE
        -- Reporte incorrecto: -10 puntos (min 0)
        v_adjustment := 10;
        UPDATE user_trust_scores SET
            rejected_reports = rejected_reports + 1,
            trust_score = GREATEST(trust_score - v_adjustment, 0),
            updated_at = NOW()
        WHERE user_id = p_user_id;

        -- Si el trust baja de 10, banear automáticamente
        UPDATE user_trust_scores SET
            flags = flags | 1  -- Set bit 0 (is_banned)
        WHERE user_id = p_user_id AND trust_score < 10;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 8. VISTA: Estadísticas de reportes
-- ============================================
CREATE OR REPLACE VIEW reported_urls_stats AS
SELECT
    COUNT(*) as total_reported_urls,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_review,
    COUNT(*) FILTER (WHERE status = 'confirmed') as confirmed,
    COUNT(*) FILTER (WHERE status = 'rejected') as rejected,
    COUNT(*) FILTER (WHERE aggregated_score >= 70) as high_score,
    COUNT(*) FILTER (WHERE aggregated_score BETWEEN 40 AND 69) as medium_score,
    COUNT(*) FILTER (WHERE aggregated_score < 40) as low_score,
    AVG(unique_reporters)::NUMERIC(5,2) as avg_reporters_per_url,
    SUM(total_reports) as total_individual_reports
FROM reported_urls
WHERE (flags & 1) = 1;

-- ============================================
-- 9. DATOS DE EJEMPLO
-- ============================================

-- Insertar algunos usuarios de ejemplo con trust scores
INSERT INTO user_trust_scores (user_id, trust_score, total_reports, confirmed_reports)
VALUES
    ('user-demo-1', 75, 10, 8),
    ('user-demo-2', 50, 5, 3),
    ('user-demo-3', 30, 3, 0)
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- MENSAJE DE CONFIRMACIÓN
-- ============================================
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Sistema de Reportes de Usuarios v1.0';
    RAISE NOTICE 'Con protección anti-spam implementada';
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Tablas creadas:';
    RAISE NOTICE '  - user_trust_scores: Confianza por usuario';
    RAISE NOTICE '  - reported_urls: URLs agregadas';
    RAISE NOTICE '  - user_url_reports: Reportes individuales';
    RAISE NOTICE '';
    RAISE NOTICE 'Funciones anti-spam:';
    RAISE NOTICE '  - get_or_create_user_trust()';
    RAISE NOTICE '  - calculate_url_aggregated_score()';
    RAISE NOTICE '  - report_url()';
    RAISE NOTICE '  - find_reported_url()';
    RAISE NOTICE '  - update_user_trust_after_review()';
    RAISE NOTICE '';
    RAISE NOTICE 'REGLAS ANTI-SPAM:';
    RAISE NOTICE '  - 1 reportador: max score 30';
    RAISE NOTICE '  - 2 reportadores: max score 45';
    RAISE NOTICE '  - 3-4 reportadores: max score 60';
    RAISE NOTICE '  - 5-9 reportadores: max score 75';
    RAISE NOTICE '  - 10+ reportadores: max score 100';
    RAISE NOTICE '===========================================';
END $$;
