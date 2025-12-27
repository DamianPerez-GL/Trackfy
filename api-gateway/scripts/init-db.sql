-- ============================================
-- API Gateway Database Schema
-- Usuarios, Sesiones, Conversaciones
-- ============================================

-- Extensiones
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- TABLA: users
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL UNIQUE,  -- Número normalizado: +34612345678
    country_code VARCHAR(4) DEFAULT '34',

    -- Datos personales
    nombre VARCHAR(50) NOT NULL,
    apellidos VARCHAR(100) NOT NULL,

    -- Estado
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,

    -- Metadatos
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP,

    -- Configuración
    language VARCHAR(5) DEFAULT 'es',
    notifications_enabled BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(id) WHERE is_active = true;

-- ============================================
-- TABLA: verification_codes
-- Para autenticación por SMS
-- ============================================
CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL,
    code VARCHAR(6) NOT NULL,

    -- Control
    attempts SMALLINT DEFAULT 0,
    max_attempts SMALLINT DEFAULT 3,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,

    -- Estado
    is_used BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_verification_phone ON verification_codes(phone, is_used);
CREATE INDEX IF NOT EXISTS idx_verification_expires ON verification_codes(expires_at);

-- ============================================
-- TABLA: sessions
-- Sesiones de usuario (backup de Redis)
-- ============================================
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Token
    token_hash BYTEA NOT NULL,  -- Hash del JWT
    refresh_token_hash BYTEA,

    -- Metadatos del dispositivo
    device_id VARCHAR(64),
    device_name VARCHAR(100),
    device_type VARCHAR(20),  -- ios, android, web
    app_version VARCHAR(20),

    -- IP y ubicación
    ip_address INET,
    user_agent TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Estado
    is_active BOOLEAN DEFAULT true,
    revoked_at TIMESTAMP,
    revoke_reason VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at) WHERE is_active = true;

-- ============================================
-- TABLA: conversations
-- Historial de conversaciones
-- ============================================
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Metadatos
    title VARCHAR(100),  -- Generado automáticamente o por usuario

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Estado
    is_active BOOLEAN DEFAULT true,
    message_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_conversations_user ON conversations(user_id, created_at DESC);

-- ============================================
-- TABLA: messages
-- Mensajes dentro de conversaciones
-- ============================================
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    -- Contenido
    role VARCHAR(10) NOT NULL,  -- 'user' o 'assistant'
    content TEXT NOT NULL,

    -- Metadatos del análisis (si aplica)
    intent VARCHAR(20),
    mood VARCHAR(20),
    analysis_performed BOOLEAN DEFAULT false,
    entities_found JSONB,  -- URLs, emails, phones detectados

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at);

-- ============================================
-- TABLA: user_stats
-- Estadísticas del usuario
-- ============================================
CREATE TABLE IF NOT EXISTS user_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- Contadores
    total_messages INTEGER DEFAULT 0,
    total_analyses INTEGER DEFAULT 0,
    threats_detected INTEGER DEFAULT 0,
    safe_verified INTEGER DEFAULT 0,

    -- Por tipo
    urls_analyzed INTEGER DEFAULT 0,
    emails_analyzed INTEGER DEFAULT 0,
    phones_analyzed INTEGER DEFAULT 0,

    -- Timestamps
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================
-- FUNCIONES
-- ============================================

-- Función para actualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para updated_at
DROP TRIGGER IF EXISTS trigger_users_updated ON users;
CREATE TRIGGER trigger_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_conversations_updated ON conversations;
CREATE TRIGGER trigger_conversations_updated
    BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Función para generar código de verificación
CREATE OR REPLACE FUNCTION generate_verification_code(p_phone TEXT, p_ttl_minutes INTEGER DEFAULT 5)
RETURNS TEXT AS $$
DECLARE
    v_code TEXT;
BEGIN
    -- Generar código de 6 dígitos
    v_code := LPAD(FLOOR(RANDOM() * 1000000)::TEXT, 6, '0');

    -- Invalidar códigos anteriores
    UPDATE verification_codes
    SET is_used = true
    WHERE phone = p_phone AND is_used = false;

    -- Insertar nuevo código
    INSERT INTO verification_codes (phone, code, expires_at)
    VALUES (p_phone, v_code, NOW() + (p_ttl_minutes || ' minutes')::INTERVAL);

    RETURN v_code;
END;
$$ LANGUAGE plpgsql;

-- Función para verificar código
CREATE OR REPLACE FUNCTION verify_code(p_phone TEXT, p_code TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    v_valid BOOLEAN := false;
    v_id UUID;
BEGIN
    -- Buscar código válido
    SELECT id INTO v_id
    FROM verification_codes
    WHERE phone = p_phone
      AND code = p_code
      AND is_used = false
      AND expires_at > NOW()
      AND attempts < max_attempts
    LIMIT 1;

    IF v_id IS NOT NULL THEN
        -- Marcar como usado
        UPDATE verification_codes
        SET is_used = true, used_at = NOW()
        WHERE id = v_id;

        v_valid := true;
    ELSE
        -- Incrementar intentos
        UPDATE verification_codes
        SET attempts = attempts + 1
        WHERE phone = p_phone AND is_used = false;
    END IF;

    RETURN v_valid;
END;
$$ LANGUAGE plpgsql;

-- Función para limpiar datos expirados
CREATE OR REPLACE FUNCTION cleanup_expired_data()
RETURNS TABLE(codes_deleted INTEGER, sessions_deleted INTEGER) AS $$
DECLARE
    v_codes INTEGER;
    v_sessions INTEGER;
BEGIN
    -- Limpiar códigos expirados
    DELETE FROM verification_codes WHERE expires_at < NOW();
    GET DIAGNOSTICS v_codes = ROW_COUNT;

    -- Limpiar sesiones expiradas
    UPDATE sessions SET is_active = false, revoke_reason = 'expired'
    WHERE expires_at < NOW() AND is_active = true;
    GET DIAGNOSTICS v_sessions = ROW_COUNT;

    RETURN QUERY SELECT v_codes, v_sessions;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- MENSAJE DE CONFIRMACIÓN
-- ============================================
DO $$
BEGIN
    RAISE NOTICE '==========================================';
    RAISE NOTICE 'API Gateway Database Schema - Instalado';
    RAISE NOTICE '==========================================';
    RAISE NOTICE 'Tablas: users, verification_codes, sessions, conversations, messages, user_stats';
    RAISE NOTICE '==========================================';
END $$;
