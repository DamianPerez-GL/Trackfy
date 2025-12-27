package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/trackfy/api-gateway/internal/models"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(databaseURL string, maxConns int) (*PostgresDB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Info().Msg("[PostgresDB] Connected successfully")
	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// ==================== USERS ====================

func (p *PostgresDB) CreateUser(ctx context.Context, phone, nombre, apellidos string) (*models.User, error) {
	user := &models.User{}
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO users (phone, nombre, apellidos)
		VALUES ($1, $2, $3)
		RETURNING id, phone, country_code, nombre, apellidos, is_active, is_verified,
				  created_at, updated_at, language, notifications_enabled
	`, phone, nombre, apellidos).Scan(
		&user.ID, &user.Phone, &user.CountryCode, &user.Nombre, &user.Apellidos,
		&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		&user.Language, &user.NotificationsEnabled,
	)
	if err != nil {
		return nil, err
	}

	// Crear estadísticas iniciales
	_, _ = p.db.ExecContext(ctx, `INSERT INTO user_stats (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, user.ID)

	return user, nil
}

func (p *PostgresDB) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	user := &models.User{}
	err := p.db.QueryRowContext(ctx, `
		SELECT id, phone, country_code, nombre, apellidos, is_active, is_verified,
			   created_at, updated_at, last_login, language, notifications_enabled
		FROM users
		WHERE phone = $1 AND is_active = true
	`, phone).Scan(
		&user.ID, &user.Phone, &user.CountryCode, &user.Nombre, &user.Apellidos,
		&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLogin, &user.Language, &user.NotificationsEnabled,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *PostgresDB) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := p.db.QueryRowContext(ctx, `
		SELECT id, phone, country_code, nombre, apellidos, is_active, is_verified,
			   created_at, updated_at, last_login, language, notifications_enabled
		FROM users
		WHERE id = $1 AND is_active = true
	`, userID).Scan(
		&user.ID, &user.Phone, &user.CountryCode, &user.Nombre, &user.Apellidos,
		&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLogin, &user.Language, &user.NotificationsEnabled,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *PostgresDB) UpdateUserLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE users SET last_login = NOW(), is_verified = true WHERE id = $1
	`, userID)
	return err
}

func (p *PostgresDB) UserExists(ctx context.Context, phone string) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)
	`, phone).Scan(&exists)
	return exists, err
}

// ==================== VERIFICATION ====================

func (p *PostgresDB) GenerateVerificationCode(ctx context.Context, phone string) (string, error) {
	var code string
	err := p.db.QueryRowContext(ctx, `SELECT generate_verification_code($1)`, phone).Scan(&code)
	return code, err
}

func (p *PostgresDB) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
	var valid bool
	err := p.db.QueryRowContext(ctx, `SELECT verify_code($1, $2)`, phone, code).Scan(&valid)
	return valid, err
}

// ==================== SESSIONS ====================

func (p *PostgresDB) CreateSession(ctx context.Context, session *models.Session, tokenHash []byte) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, device_id, device_name, device_type,
							  app_version, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, session.ID, session.UserID, tokenHash, session.DeviceID, session.DeviceName,
		session.DeviceType, session.AppVersion, session.IPAddress, session.ExpiresAt)
	return err
}

func (p *PostgresDB) InvalidateSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE sessions
		SET is_active = false, revoked_at = NOW(), revoke_reason = $2
		WHERE id = $1
	`, sessionID, reason)
	return err
}

func (p *PostgresDB) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE sessions
		SET is_active = false, revoked_at = NOW(), revoke_reason = 'logout_all'
		WHERE user_id = $1 AND is_active = true
	`, userID)
	return err
}

func (p *PostgresDB) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, user_id, device_id, device_name, device_type, app_version,
			   ip_address::text, created_at, expires_at, last_activity, is_active
		FROM sessions
		WHERE user_id = $1 AND is_active = true
		ORDER BY last_activity DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		var ip sql.NullString
		if err := rows.Scan(&s.ID, &s.UserID, &s.DeviceID, &s.DeviceName, &s.DeviceType,
			&s.AppVersion, &ip, &s.CreatedAt, &s.ExpiresAt, &s.LastActivity, &s.IsActive); err != nil {
			continue
		}
		if ip.Valid {
			s.IPAddress = ip.String
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// ==================== CONVERSATIONS ====================

func (p *PostgresDB) CreateConversation(ctx context.Context, userID uuid.UUID, title string) (*models.Conversation, error) {
	conv := &models.Conversation{}
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO conversations (user_id, title)
		VALUES ($1, $2)
		RETURNING id, user_id, title, created_at, updated_at, is_active, message_count
	`, userID, title).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt,
		&conv.UpdatedAt, &conv.IsActive, &conv.MessageCount,
	)
	return conv, err
}

func (p *PostgresDB) GetConversation(ctx context.Context, conversationID, userID uuid.UUID) (*models.Conversation, error) {
	conv := &models.Conversation{}
	var title sql.NullString
	err := p.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, created_at, updated_at, is_active, message_count
		FROM conversations
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`, conversationID, userID).Scan(
		&conv.ID, &conv.UserID, &title, &conv.CreatedAt,
		&conv.UpdatedAt, &conv.IsActive, &conv.MessageCount,
	)
	if title.Valid {
		conv.Title = title.String
	}
	return conv, err
}

func (p *PostgresDB) GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Conversation, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, user_id, title, created_at, updated_at, is_active, message_count
		FROM conversations
		WHERE user_id = $1 AND is_active = true
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var c models.Conversation
		var title sql.NullString
		if err := rows.Scan(&c.ID, &c.UserID, &title, &c.CreatedAt, &c.UpdatedAt, &c.IsActive, &c.MessageCount); err != nil {
			continue
		}
		if title.Valid {
			c.Title = title.String
		}
		conversations = append(conversations, c)
	}
	return conversations, nil
}

// ==================== MESSAGES ====================

func (p *PostgresDB) AddMessage(ctx context.Context, msg *models.Message) error {
	entitiesJSON, _ := json.Marshal(msg.EntitiesFound)

	_, err := p.db.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, role, content, intent, mood, analysis_performed, entities_found)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.Intent, msg.Mood, msg.AnalysisPerformed, entitiesJSON)

	if err == nil {
		// Actualizar contador de mensajes
		_, _ = p.db.ExecContext(ctx, `
			UPDATE conversations SET message_count = message_count + 1, updated_at = NOW()
			WHERE id = $1
		`, msg.ConversationID)
	}

	return err
}

func (p *PostgresDB) GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.Message, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, conversation_id, role, content, intent, mood, analysis_performed, entities_found, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		var intent, mood sql.NullString
		var entitiesJSON []byte
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &intent, &mood,
			&m.AnalysisPerformed, &entitiesJSON, &m.CreatedAt); err != nil {
			continue
		}
		if intent.Valid {
			m.Intent = intent.String
		}
		if mood.Valid {
			m.Mood = mood.String
		}
		if len(entitiesJSON) > 0 {
			_ = json.Unmarshal(entitiesJSON, &m.EntitiesFound)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// ==================== STATS ====================

func (p *PostgresDB) UpdateUserStats(ctx context.Context, userID uuid.UUID, analysisPerformed bool, threatDetected bool) error {
	query := `
		UPDATE user_stats SET
			total_messages = total_messages + 1,
			total_analyses = total_analyses + CASE WHEN $2 THEN 1 ELSE 0 END,
			threats_detected = threats_detected + CASE WHEN $3 THEN 1 ELSE 0 END,
			safe_verified = safe_verified + CASE WHEN $2 AND NOT $3 THEN 1 ELSE 0 END,
			updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := p.db.ExecContext(ctx, query, userID, analysisPerformed, threatDetected)
	return err
}

func (p *PostgresDB) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	stats := &models.UserStats{}
	err := p.db.QueryRowContext(ctx, `
		SELECT user_id, total_messages, total_analyses, threats_detected, safe_verified,
			   urls_analyzed, emails_analyzed, phones_analyzed, updated_at
		FROM user_stats
		WHERE user_id = $1
	`, userID).Scan(
		&stats.UserID, &stats.TotalMessages, &stats.TotalAnalyses, &stats.ThreatsDetected,
		&stats.SafeVerified, &stats.URLsAnalyzed, &stats.EmailsAnalyzed, &stats.PhonesAnalyzed,
		&stats.UpdatedAt,
	)
	return stats, err
}
