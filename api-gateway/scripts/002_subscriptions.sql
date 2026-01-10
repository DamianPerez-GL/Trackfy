-- ============================================
-- MIGRACIÓN: Sistema de Suscripciones con Stripe
-- Trackfy Premium: 4,99€/mes - Mensajes ilimitados
-- ============================================

-- ============================================
-- 1. TABLA: subscriptions
-- Estado de suscripción por usuario
-- ============================================
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Stripe IDs
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),

    -- Plan
    plan VARCHAR(20) NOT NULL DEFAULT 'free',  -- free, premium
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, canceled, past_due, paused

    -- Límites de mensajes
    messages_limit INTEGER NOT NULL DEFAULT 5,  -- -1 = ilimitado
    messages_used INTEGER NOT NULL DEFAULT 0,

    -- Período actual (para reseteo mensual)
    period_start TIMESTAMP,
    period_end TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    canceled_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_customer ON subscriptions(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_sub ON subscriptions(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_period_end ON subscriptions(period_end);

COMMENT ON TABLE subscriptions IS 'Estado de suscripción de usuarios - Free: 5 msg/mes, Premium: ilimitado';
COMMENT ON COLUMN subscriptions.messages_limit IS '5 para free, -1 para premium (ilimitado)';

-- ============================================
-- 2. TABLA: payment_history
-- Historial de pagos
-- ============================================
CREATE TABLE IF NOT EXISTS payment_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Stripe
    stripe_payment_intent_id VARCHAR(255),
    stripe_invoice_id VARCHAR(255),

    -- Detalles
    amount INTEGER NOT NULL,  -- Cantidad en centavos (499 = 4,99€)
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    status VARCHAR(20) NOT NULL,  -- succeeded, failed, pending, refunded

    -- Descripción
    description VARCHAR(255),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_history_user ON payment_history(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_history_stripe ON payment_history(stripe_payment_intent_id);

COMMENT ON TABLE payment_history IS 'Historial de pagos de suscripciones';

-- ============================================
-- 3. TRIGGER: updated_at para subscriptions
-- ============================================
DROP TRIGGER IF EXISTS trigger_subscriptions_updated ON subscriptions;
CREATE TRIGGER trigger_subscriptions_updated
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 4. FUNCIÓN: Crear suscripción para nuevo usuario
-- Se llama automáticamente al registrar usuario
-- ============================================
CREATE OR REPLACE FUNCTION create_subscription_for_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO subscriptions (user_id, plan, messages_limit, period_start, period_end)
    VALUES (
        NEW.id,
        'free',
        5,
        DATE_TRUNC('month', NOW()),
        DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_create_subscription ON users;
CREATE TRIGGER trigger_create_subscription
    AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION create_subscription_for_user();

-- ============================================
-- 5. FUNCIÓN: Incrementar contador de mensajes
-- Retorna: remaining_messages, is_limited
-- ============================================
CREATE OR REPLACE FUNCTION increment_message_count(p_user_id UUID)
RETURNS TABLE(
    success BOOLEAN,
    messages_used INTEGER,
    messages_limit INTEGER,
    messages_remaining INTEGER,
    is_premium BOOLEAN
) AS $$
DECLARE
    v_sub subscriptions%ROWTYPE;
    v_remaining INTEGER;
BEGIN
    -- Obtener suscripción con lock
    SELECT * INTO v_sub
    FROM subscriptions
    WHERE user_id = p_user_id
    FOR UPDATE;

    -- Si no existe suscripción, crear una
    IF NOT FOUND THEN
        INSERT INTO subscriptions (user_id, plan, messages_limit, period_start, period_end)
        VALUES (
            p_user_id,
            'free',
            5,
            DATE_TRUNC('month', NOW()),
            DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
        )
        RETURNING * INTO v_sub;
    END IF;

    -- Verificar si hay que resetear el contador (nuevo mes)
    IF v_sub.period_end < NOW() THEN
        UPDATE subscriptions
        SET messages_used = 0,
            period_start = DATE_TRUNC('month', NOW()),
            period_end = DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
        WHERE user_id = p_user_id
        RETURNING * INTO v_sub;
    END IF;

    -- Si es premium (-1), no hay límite
    IF v_sub.messages_limit = -1 THEN
        UPDATE subscriptions
        SET messages_used = messages_used + 1
        WHERE user_id = p_user_id;

        RETURN QUERY SELECT
            TRUE,
            v_sub.messages_used + 1,
            -1,
            -1,  -- Ilimitado
            TRUE;
        RETURN;
    END IF;

    -- Verificar límite
    IF v_sub.messages_used >= v_sub.messages_limit THEN
        -- Límite alcanzado
        RETURN QUERY SELECT
            FALSE,
            v_sub.messages_used,
            v_sub.messages_limit,
            0,
            FALSE;
        RETURN;
    END IF;

    -- Incrementar contador
    UPDATE subscriptions
    SET messages_used = messages_used + 1
    WHERE user_id = p_user_id
    RETURNING messages_used INTO v_sub.messages_used;

    v_remaining := v_sub.messages_limit - v_sub.messages_used;

    RETURN QUERY SELECT
        TRUE,
        v_sub.messages_used,
        v_sub.messages_limit,
        v_remaining,
        FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION increment_message_count IS 'Incrementa el contador de mensajes y verifica límites. Retorna FALSE si se alcanzó el límite.';

-- ============================================
-- 6. FUNCIÓN: Obtener estado de suscripción
-- ============================================
CREATE OR REPLACE FUNCTION get_subscription_status(p_user_id UUID)
RETURNS TABLE(
    plan VARCHAR,
    status VARCHAR,
    messages_used INTEGER,
    messages_limit INTEGER,
    messages_remaining INTEGER,
    is_premium BOOLEAN,
    period_end TIMESTAMP,
    stripe_customer_id VARCHAR
) AS $$
DECLARE
    v_sub subscriptions%ROWTYPE;
    v_remaining INTEGER;
BEGIN
    SELECT * INTO v_sub
    FROM subscriptions s
    WHERE s.user_id = p_user_id;

    IF NOT FOUND THEN
        -- Crear suscripción por defecto
        INSERT INTO subscriptions (user_id, plan, messages_limit, period_start, period_end)
        VALUES (
            p_user_id,
            'free',
            5,
            DATE_TRUNC('month', NOW()),
            DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
        )
        RETURNING * INTO v_sub;
    END IF;

    -- Verificar si hay que resetear (nuevo mes)
    IF v_sub.period_end < NOW() THEN
        UPDATE subscriptions
        SET messages_used = 0,
            period_start = DATE_TRUNC('month', NOW()),
            period_end = DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
        WHERE user_id = p_user_id
        RETURNING * INTO v_sub;
    END IF;

    -- Calcular remaining
    IF v_sub.messages_limit = -1 THEN
        v_remaining := -1;
    ELSE
        v_remaining := GREATEST(0, v_sub.messages_limit - v_sub.messages_used);
    END IF;

    RETURN QUERY SELECT
        v_sub.plan,
        v_sub.status,
        v_sub.messages_used,
        v_sub.messages_limit,
        v_remaining,
        v_sub.plan = 'premium',
        v_sub.period_end,
        v_sub.stripe_customer_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 7. FUNCIÓN: Actualizar a Premium
-- ============================================
CREATE OR REPLACE FUNCTION upgrade_to_premium(
    p_user_id UUID,
    p_stripe_customer_id VARCHAR,
    p_stripe_subscription_id VARCHAR
) RETURNS VOID AS $$
BEGIN
    UPDATE subscriptions
    SET plan = 'premium',
        status = 'active',
        messages_limit = -1,
        stripe_customer_id = p_stripe_customer_id,
        stripe_subscription_id = p_stripe_subscription_id,
        period_start = NOW(),
        period_end = NOW() + INTERVAL '1 month'
    WHERE user_id = p_user_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 8. FUNCIÓN: Cancelar Premium (volver a Free)
-- ============================================
CREATE OR REPLACE FUNCTION cancel_premium(p_user_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE subscriptions
    SET plan = 'free',
        status = 'canceled',
        messages_limit = 5,
        canceled_at = NOW(),
        -- Mantener stripe IDs para referencia
        period_start = DATE_TRUNC('month', NOW()),
        period_end = DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
    WHERE user_id = p_user_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 9. CREAR SUSCRIPCIONES PARA USUARIOS EXISTENTES
-- ============================================
INSERT INTO subscriptions (user_id, plan, messages_limit, period_start, period_end)
SELECT
    id,
    'free',
    5,
    DATE_TRUNC('month', NOW()),
    DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
FROM users
WHERE id NOT IN (SELECT user_id FROM subscriptions)
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- MENSAJE DE CONFIRMACIÓN
-- ============================================
DO $$
BEGIN
    RAISE NOTICE '==========================================';
    RAISE NOTICE 'Sistema de Suscripciones - Instalado';
    RAISE NOTICE '==========================================';
    RAISE NOTICE 'Tablas: subscriptions, payment_history';
    RAISE NOTICE 'Funciones: increment_message_count, get_subscription_status';
    RAISE NOTICE 'Funciones: upgrade_to_premium, cancel_premium';
    RAISE NOTICE '';
    RAISE NOTICE 'Plan Free: 5 mensajes/mes';
    RAISE NOTICE 'Plan Premium: Mensajes ilimitados (4,99€/mes)';
    RAISE NOTICE '==========================================';
END $$;
