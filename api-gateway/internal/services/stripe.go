package services

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/trackfy/api-gateway/internal/config"
)

// StripeService maneja operaciones con Stripe
type StripeService struct {
	config *config.StripeConfig
}

// NewStripeService crea un nuevo servicio de Stripe
func NewStripeService(cfg *config.StripeConfig) *StripeService {
	stripe.Key = cfg.SecretKey
	return &StripeService{
		config: cfg,
	}
}

// CreateCustomer crea un cliente en Stripe
func (s *StripeService) CreateCustomer(phone, name, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Phone: stripe.String(phone),
		Name:  stripe.String(name),
	}
	if email != "" {
		params.Email = stripe.String(email)
	}

	// Metadata para identificar al usuario
	params.Metadata = map[string]string{
		"phone": phone,
	}

	c, err := customer.New(params)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Error creating Stripe customer")
		return nil, fmt.Errorf("error creating customer: %w", err)
	}

	log.Info().Str("customer_id", c.ID).Str("phone", phone).Msg("Stripe customer created")
	return c, nil
}

// GetOrCreateCustomer obtiene o crea un cliente
func (s *StripeService) GetOrCreateCustomer(stripeCustomerID *string, phone, name, email string) (string, error) {
	if stripeCustomerID != nil && *stripeCustomerID != "" {
		return *stripeCustomerID, nil
	}

	c, err := s.CreateCustomer(phone, name, email)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

// CreateCheckoutSession crea una sesión de checkout para suscripción
func (s *StripeService) CreateCheckoutSession(customerID, userID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.config.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.config.SuccessURL),
		CancelURL:  stripe.String(s.config.CancelURL),
		Metadata: map[string]string{
			"user_id": userID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": userID,
			},
		},
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		log.Error().Err(err).Str("customer_id", customerID).Msg("Error creating checkout session")
		return nil, fmt.Errorf("error creating checkout session: %w", err)
	}

	log.Info().
		Str("session_id", sess.ID).
		Str("customer_id", customerID).
		Msg("Checkout session created")

	return sess, nil
}

// CreatePortalSession crea una sesión del portal de cliente para gestionar suscripción
func (s *StripeService) CreatePortalSession(customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := session.New(params)
	if err != nil {
		log.Error().Err(err).Str("customer_id", customerID).Msg("Error creating portal session")
		return nil, fmt.Errorf("error creating portal session: %w", err)
	}

	log.Info().
		Str("customer_id", customerID).
		Msg("Portal session created")

	return sess, nil
}

// VerifyWebhookSignature verifica la firma del webhook de Stripe
func (s *StripeService) VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.config.WebhookSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying webhook signature")
		return nil, fmt.Errorf("invalid webhook signature: %w", err)
	}
	return &event, nil
}

// WebhookEventType tipos de eventos de webhook
type WebhookEventType string

const (
	EventCheckoutCompleted      WebhookEventType = "checkout.session.completed"
	EventSubscriptionUpdated    WebhookEventType = "customer.subscription.updated"
	EventSubscriptionDeleted    WebhookEventType = "customer.subscription.deleted"
	EventInvoicePaid            WebhookEventType = "invoice.paid"
	EventInvoicePaymentFailed   WebhookEventType = "invoice.payment_failed"
)

// IsConfigured verifica si Stripe está configurado
func (s *StripeService) IsConfigured() bool {
	return s.config.SecretKey != "" && s.config.PriceID != ""
}
