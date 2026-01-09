-- ============================================
-- PRE-MIGRACIÓN: Enums y funciones requeridas
-- Ejecutar ANTES de 001_user_reports_system.sql
-- ============================================

-- Extensión para funciones criptográficas (si no existe)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enum para tipos de amenaza
DO $$ BEGIN
    CREATE TYPE threat_type_enum AS ENUM (
        'phishing', 'malware', 'scam', 'spam',
        'vishing', 'smishing', 'other'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Enum para estado de reportes
DO $$ BEGIN
    CREATE TYPE report_status_enum AS ENUM (
        'pending', 'confirmed', 'rejected'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Función sha256_bytea (wrapper de pgcrypto)
CREATE OR REPLACE FUNCTION sha256_bytea(data TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN digest(data, 'sha256');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Verificación
DO $$
BEGIN
    RAISE NOTICE 'Pre-migración completada:';
    RAISE NOTICE '  - threat_type_enum creado';
    RAISE NOTICE '  - report_status_enum creado';
    RAISE NOTICE '  - sha256_bytea() disponible';
END $$;
