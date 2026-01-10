package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionPlan tipos de plan
type SubscriptionPlan string

const (
	PlanFree    SubscriptionPlan = "free"
	PlanPremium SubscriptionPlan = "premium"
)

// SubscriptionStatus estados de suscripción
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusPaused   SubscriptionStatus = "paused"
)

// Subscription representa la suscripción de un usuario
type Subscription struct {
	ID                   uuid.UUID          `json:"id"`
	UserID               uuid.UUID          `json:"user_id"`
	StripeCustomerID     *string            `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string            `json:"stripe_subscription_id,omitempty"`
	Plan                 SubscriptionPlan   `json:"plan"`
	Status               SubscriptionStatus `json:"status"`
	MessagesLimit        int                `json:"messages_limit"`
	MessagesUsed         int                `json:"messages_used"`
	PeriodStart          *time.Time         `json:"period_start,omitempty"`
	PeriodEnd            *time.Time         `json:"period_end,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
}

// SubscriptionStatusResponse respuesta del endpoint /subscription/status
type SubscriptionStatusResponse struct {
	Plan              SubscriptionPlan `json:"plan"`
	Status            string           `json:"status"`
	MessagesUsed      int              `json:"messages_used"`
	MessagesLimit     int              `json:"messages_limit"`
	MessagesRemaining int              `json:"messages_remaining"`
	IsPremium         bool             `json:"is_premium"`
	PeriodEnd         *time.Time       `json:"period_end,omitempty"`
	CanSendMessage    bool             `json:"can_send_message"`
}

// IsLimitReached verifica si se alcanzó el límite de mensajes
func (s *Subscription) IsLimitReached() bool {
	if s.MessagesLimit == -1 {
		return false // Premium, sin límite
	}
	return s.MessagesUsed >= s.MessagesLimit
}

// MessagesRemaining calcula mensajes restantes
func (s *Subscription) MessagesRemaining() int {
	if s.MessagesLimit == -1 {
		return -1 // Ilimitado
	}
	remaining := s.MessagesLimit - s.MessagesUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsPremium verifica si es premium
func (s *Subscription) IsPremium() bool {
	return s.Plan == PlanPremium && s.Status == StatusActive
}

// MessageCountResult resultado de incrementar contador
type MessageCountResult struct {
	Success           bool `json:"success"`
	MessagesUsed      int  `json:"messages_used"`
	MessagesLimit     int  `json:"messages_limit"`
	MessagesRemaining int  `json:"messages_remaining"`
	IsPremium         bool `json:"is_premium"`
}

// PaymentHistory historial de pagos
type PaymentHistory struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"user_id"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id,omitempty"`
	StripeInvoiceID       *string   `json:"stripe_invoice_id,omitempty"`
	Amount                int       `json:"amount"` // Centavos
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	Description           *string   `json:"description,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// CheckoutSessionRequest request para crear sesión de checkout
type CheckoutSessionRequest struct {
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

// CheckoutSessionResponse respuesta con URL de checkout
type CheckoutSessionResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

// PortalSessionResponse respuesta con URL del portal de cliente
type PortalSessionResponse struct {
	PortalURL string `json:"portal_url"`
}
