package models

import (
	"time"

	"github.com/google/uuid"
)

// User representa un usuario de la aplicación
type User struct {
	ID                   uuid.UUID `json:"id"`
	Phone                string    `json:"phone"`
	CountryCode          string    `json:"country_code"`
	Nombre               string    `json:"nombre"`
	Apellidos            string    `json:"apellidos"`
	IsActive             bool      `json:"is_active"`
	IsVerified           bool      `json:"is_verified"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	LastLogin            *time.Time `json:"last_login,omitempty"`
	Language             string    `json:"language"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
}

// UserPublic es la versión pública del usuario (sin datos sensibles)
type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Nombre    string    `json:"nombre"`
	Apellidos string    `json:"apellidos"`
	Phone     string    `json:"phone"` // Parcialmente oculto: +34***678
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

// ToPublic convierte User a UserPublic
func (u *User) ToPublic() UserPublic {
	maskedPhone := u.Phone
	if len(maskedPhone) > 6 {
		maskedPhone = maskedPhone[:4] + "***" + maskedPhone[len(maskedPhone)-3:]
	}
	return UserPublic{
		ID:        u.ID,
		Nombre:    u.Nombre,
		Apellidos: u.Apellidos,
		Phone:     maskedPhone,
		Language:  u.Language,
		CreatedAt: u.CreatedAt,
	}
}

// Session representa una sesión de usuario
type Session struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	DeviceID         string     `json:"device_id,omitempty"`
	DeviceName       string     `json:"device_name,omitempty"`
	DeviceType       string     `json:"device_type,omitempty"`
	AppVersion       string     `json:"app_version,omitempty"`
	IPAddress        string     `json:"ip_address,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	LastActivity     time.Time  `json:"last_activity"`
	IsActive         bool       `json:"is_active"`
}

// Conversation representa una conversación
type Conversation struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Title        string    `json:"title,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
	MessageCount int       `json:"message_count"`
	LastMessage  string    `json:"last_message,omitempty"`
	LastIntent   string    `json:"last_intent,omitempty"`
	HasThreats   bool      `json:"has_threats"`
}

// Message representa un mensaje en una conversación
type Message struct {
	ID                uuid.UUID              `json:"id"`
	ConversationID    uuid.UUID              `json:"conversation_id"`
	Role              string                 `json:"role"` // "user" o "assistant"
	Content           string                 `json:"content"`
	Intent            string                 `json:"intent,omitempty"`
	Mood              string                 `json:"mood,omitempty"`
	AnalysisPerformed bool                   `json:"analysis_performed"`
	EntitiesFound     map[string]interface{} `json:"entities_found,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// UserStats estadísticas del usuario
type UserStats struct {
	UserID          uuid.UUID `json:"user_id"`
	TotalMessages   int       `json:"total_messages"`
	TotalAnalyses   int       `json:"total_analyses"`
	ThreatsDetected int       `json:"threats_detected"`
	SafeVerified    int       `json:"safe_verified"`
	URLsAnalyzed    int       `json:"urls_analyzed"`
	EmailsAnalyzed  int       `json:"emails_analyzed"`
	PhonesAnalyzed  int       `json:"phones_analyzed"`
	UpdatedAt       time.Time `json:"updated_at"`
}
