package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v76"
	"github.com/trackfy/api-gateway/internal/db"
	"github.com/trackfy/api-gateway/internal/models"
	"github.com/trackfy/api-gateway/internal/services"
)

// SubscriptionHandler maneja endpoints de suscripción
type SubscriptionHandler struct {
	db     *db.PostgresDB
	stripe *services.StripeService
}

// NewSubscriptionHandler crea un nuevo handler de suscripción
func NewSubscriptionHandler(db *db.PostgresDB, stripe *services.StripeService) *SubscriptionHandler {
	return &SubscriptionHandler{
		db:     db,
		stripe: stripe,
	}
}

// GetStatus GET /api/v1/subscription/status
func (h *SubscriptionHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	status, err := h.db.GetSubscriptionStatus(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Error getting subscription status")
		respondError(w, http.StatusInternalServerError, "server_error", "Error al obtener estado de suscripción")
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// CreateCheckout POST /api/v1/subscription/checkout
func (h *SubscriptionHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Verificar si Stripe está configurado
	if !h.stripe.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "stripe_not_configured", "Pagos no configurados")
		return
	}

	// Obtener usuario para datos de Stripe
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Error getting user")
		respondError(w, http.StatusInternalServerError, "server_error", "Error interno")
		return
	}

	// Obtener o crear Stripe Customer
	existingCustomerID, _ := h.db.GetStripeCustomerID(r.Context(), userID)
	customerID, err := h.stripe.GetOrCreateCustomer(
		existingCustomerID,
		user.Phone,
		user.Nombre+" "+user.Apellidos,
		"", // Sin email por ahora
	)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Error creating Stripe customer")
		respondError(w, http.StatusInternalServerError, "stripe_error", "Error al crear cliente")
		return
	}

	// Guardar customer ID si es nuevo
	if existingCustomerID == nil || *existingCustomerID == "" {
		if err := h.db.UpdateStripeCustomerID(r.Context(), userID, customerID); err != nil {
			log.Error().Err(err).Msg("Error saving stripe customer ID")
		}
	}

	// Crear sesión de checkout
	session, err := h.stripe.CreateCheckoutSession(customerID, userID.String())
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Error creating checkout session")
		respondError(w, http.StatusInternalServerError, "checkout_error", "Error al crear sesión de pago")
		return
	}

	respondJSON(w, http.StatusOK, models.CheckoutSessionResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
	})
}

// CreatePortal POST /api/v1/subscription/portal
func (h *SubscriptionHandler) CreatePortal(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Verificar si Stripe está configurado
	if !h.stripe.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "stripe_not_configured", "Pagos no configurados")
		return
	}

	// Obtener Stripe Customer ID
	customerID, err := h.db.GetStripeCustomerID(r.Context(), userID)
	if err != nil || customerID == nil || *customerID == "" {
		respondError(w, http.StatusBadRequest, "no_subscription", "No tienes una suscripción activa")
		return
	}

	// Crear sesión del portal
	session, err := h.stripe.CreatePortalSession(*customerID, "trackfy://subscription")
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Error creating portal session")
		respondError(w, http.StatusInternalServerError, "portal_error", "Error al crear sesión del portal")
		return
	}

	respondJSON(w, http.StatusOK, models.PortalSessionResponse{
		PortalURL: session.URL,
	})
}

// HandleWebhook POST /webhook/stripe (público, sin auth)
func (h *SubscriptionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Leer body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading webhook body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verificar firma
	signature := r.Header.Get("Stripe-Signature")
	event, err := h.stripe.VerifyWebhookSignature(payload, signature)
	if err != nil {
		log.Error().Err(err).Msg("Invalid webhook signature")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Str("type", string(event.Type)).Msg("Webhook received")

	// Procesar evento
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(r.Context(), event)

	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(r.Context(), event)

	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(r.Context(), event)

	case "invoice.paid":
		h.handleInvoicePaid(r.Context(), event)

	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(r.Context(), event)

	default:
		log.Debug().Str("type", string(event.Type)).Msg("Unhandled webhook event")
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Error().Err(err).Msg("Error parsing checkout session")
		return
	}

	// Obtener user_id del metadata
	userIDStr, ok := session.Metadata["user_id"]
	if !ok {
		log.Error().Msg("No user_id in checkout session metadata")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Invalid user_id in metadata")
		return
	}

	// Activar premium
	if err := h.db.UpgradeToPremium(ctx, userID, session.Customer.ID, session.Subscription.ID); err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Error upgrading to premium")
		return
	}

	log.Info().Str("user_id", userIDStr).Msg("User upgraded to premium")
}

func (h *SubscriptionHandler) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Error().Err(err).Msg("Error parsing subscription")
		return
	}

	// Buscar usuario por customer ID
	userID, err := h.db.GetUserIDByStripeCustomer(ctx, subscription.Customer.ID)
	if err != nil || userID == nil {
		log.Error().Err(err).Str("customer_id", subscription.Customer.ID).Msg("User not found for customer")
		return
	}

	// Actualizar según estado
	switch subscription.Status {
	case stripe.SubscriptionStatusActive:
		if err := h.db.UpgradeToPremium(ctx, *userID, subscription.Customer.ID, subscription.ID); err != nil {
			log.Error().Err(err).Msg("Error upgrading subscription")
		}
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusUnpaid:
		if err := h.db.CancelPremium(ctx, *userID); err != nil {
			log.Error().Err(err).Msg("Error canceling subscription")
		}
	}

	log.Info().Str("user_id", userID.String()).Str("status", string(subscription.Status)).Msg("Subscription updated")
}

func (h *SubscriptionHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Error().Err(err).Msg("Error parsing subscription")
		return
	}

	userID, err := h.db.GetUserIDByStripeCustomer(ctx, subscription.Customer.ID)
	if err != nil || userID == nil {
		log.Error().Str("customer_id", subscription.Customer.ID).Msg("User not found for deleted subscription")
		return
	}

	if err := h.db.CancelPremium(ctx, *userID); err != nil {
		log.Error().Err(err).Msg("Error canceling premium after subscription deleted")
	}

	log.Info().Str("user_id", userID.String()).Msg("Subscription deleted, premium canceled")
}

func (h *SubscriptionHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Error().Err(err).Msg("Error parsing invoice")
		return
	}

	userID, err := h.db.GetUserIDByStripeCustomer(ctx, invoice.Customer.ID)
	if err != nil || userID == nil {
		log.Error().Str("customer_id", invoice.Customer.ID).Msg("User not found for invoice")
		return
	}

	// Registrar pago
	paymentIntentID := ""
	if invoice.PaymentIntent != nil {
		paymentIntentID = invoice.PaymentIntent.ID
	}

	if err := h.db.AddPaymentHistory(
		ctx,
		*userID,
		paymentIntentID,
		invoice.ID,
		int(invoice.AmountPaid),
		string(invoice.Currency),
		"succeeded",
		"Trackfy Premium - Suscripción mensual",
	); err != nil {
		log.Error().Err(err).Msg("Error recording payment")
	}

	log.Info().Str("user_id", userID.String()).Int64("amount", invoice.AmountPaid).Msg("Invoice paid")
}

func (h *SubscriptionHandler) handleInvoicePaymentFailed(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Error().Err(err).Msg("Error parsing invoice")
		return
	}

	userID, err := h.db.GetUserIDByStripeCustomer(ctx, invoice.Customer.ID)
	if err != nil || userID == nil {
		log.Error().Str("customer_id", invoice.Customer.ID).Msg("User not found for failed invoice")
		return
	}

	log.Warn().Str("user_id", userID.String()).Msg("Invoice payment failed")
}
