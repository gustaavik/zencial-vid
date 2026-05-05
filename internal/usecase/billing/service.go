package billing

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	portalsession "github.com/stripe/stripe-go/v83/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v83/checkout/session"
	customerapi "github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/invoice"
	subscriptionapi "github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/webhook"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

type Config struct {
	SecretKey     string
	WebhookSecret string
	Currency      string
}

type Service struct {
	userRepo repository.UserRepository
	planRepo repository.PlanRepository
	subRepo  repository.SubscriptionRepository
	cfg      Config
	log      *slog.Logger
}

type CheckoutInput struct {
	UserID     uuid.UUID
	PlanID     uuid.UUID
	SuccessURL string
	CancelURL  string
}

type HostedSession struct {
	URL string
}

type PortalInput struct {
	UserID    uuid.UUID
	ReturnURL string
}

type Invoice struct {
	ID               string
	Number           string
	Status           string
	AmountPaid       int64
	AmountDue        int64
	Currency         string
	HostedInvoiceURL string
	InvoicePDF       string
	CreatedAt        time.Time
}

func NewService(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	subRepo repository.SubscriptionRepository,
	cfg Config,
	log *slog.Logger,
) *Service {
	if cfg.SecretKey != "" {
		stripe.Key = cfg.SecretKey
	}
	return &Service{
		userRepo: userRepo,
		planRepo: planRepo,
		subRepo:  subRepo,
		cfg:      cfg,
		log:      log,
	}
}

func (s *Service) CreateCheckoutSession(ctx context.Context, input CheckoutInput) (*HostedSession, *apperror.AppError) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.SuccessURL) == "" || strings.TrimSpace(input.CancelURL) == "" {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "success_url and cancel_url are required", nil)
	}

	user, appErr := s.getUser(ctx, input.UserID)
	if appErr != nil {
		return nil, appErr
	}

	plan, err := s.planRepo.GetByID(ctx, input.PlanID)
	if err != nil {
		s.log.Error("getting plan for checkout", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}
	if plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}
	if !plan.IsActive {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "plan is not active", nil)
	}

	active, err := s.subRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		s.log.Error("checking active subscription before checkout", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check subscription", err)
	}
	if active != nil {
		return nil, apperror.Conflict(apperror.CodeActiveSubscriptionExists, "user already has an active subscription", domain.ErrActiveSubscriptionExists)
	}

	customerID, appErr := s.ensureStripeCustomer(ctx, user)
	if appErr != nil {
		return nil, appErr
	}

	metadata := map[string]string{
		"user_id": user.ID.String(),
		"plan_id": plan.ID.String(),
	}
	params := &stripe.CheckoutSessionParams{
		Params:              stripe.Params{Context: ctx},
		AllowPromotionCodes: new(true),
		CancelURL:           stripe.String(input.CancelURL),
		ClientReferenceID:   stripe.String(user.ID.String()),
		Customer:            stripe.String(customerID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			s.checkoutLineItem(plan),
		},
		Metadata: metadata,
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
		SuccessURL: stripe.String(input.SuccessURL),
	}

	session, err := checkoutsession.New(params)
	if err != nil {
		s.log.Error("creating stripe checkout session", "error", err)
		return nil, apperror.Internal(apperror.CodeBillingFailed, "failed to start checkout", err)
	}
	if session.URL == "" {
		return nil, apperror.Internal(apperror.CodeBillingFailed, "stripe checkout session did not include a URL", nil)
	}

	return &HostedSession{URL: session.URL}, nil
}

func (s *Service) CreatePortalSession(ctx context.Context, input PortalInput) (*HostedSession, *apperror.AppError) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.ReturnURL) == "" {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "return_url is required", nil)
	}

	user, appErr := s.getUser(ctx, input.UserID)
	if appErr != nil {
		return nil, appErr
	}
	if user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		return nil, apperror.BadRequest(apperror.CodeBillingFailed, "no Stripe billing account exists for this user", nil)
	}

	session, err := portalsession.New(&stripe.BillingPortalSessionParams{
		Params:    stripe.Params{Context: ctx},
		Customer:  stripe.String(*user.StripeCustomerID),
		ReturnURL: stripe.String(input.ReturnURL),
	})
	if err != nil {
		s.log.Error("creating stripe portal session", "error", err)
		return nil, apperror.Internal(apperror.CodeBillingFailed, "failed to open billing portal", err)
	}

	return &HostedSession{URL: session.URL}, nil
}

func (s *Service) ListInvoices(ctx context.Context, userID uuid.UUID, limit int64) ([]*Invoice, *apperror.AppError) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	user, appErr := s.getUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}
	if user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		return []*Invoice{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 12
	}

	iter := invoice.List(&stripe.InvoiceListParams{
		ListParams: stripe.ListParams{Context: ctx, Limit: new(limit)},
		Customer:   stripe.String(*user.StripeCustomerID),
	})

	invoices := make([]*Invoice, 0, limit)
	for iter.Next() {
		stripeInvoice := iter.Invoice()
		invoices = append(invoices, &Invoice{
			ID:               stripeInvoice.ID,
			Number:           stripeInvoice.Number,
			Status:           string(stripeInvoice.Status),
			AmountPaid:       stripeInvoice.AmountPaid,
			AmountDue:        stripeInvoice.AmountDue,
			Currency:         string(stripeInvoice.Currency),
			HostedInvoiceURL: stripeInvoice.HostedInvoiceURL,
			InvoicePDF:       stripeInvoice.InvoicePDF,
			CreatedAt:        time.Unix(stripeInvoice.Created, 0).UTC(),
		})
	}
	if err := iter.Err(); err != nil {
		s.log.Error("listing stripe invoices", "error", err)
		return nil, apperror.Internal(apperror.CodeBillingFailed, "failed to list invoices", err)
	}

	return invoices, nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) *apperror.AppError {
	if s.cfg.WebhookSecret == "" {
		return apperror.BadRequest(apperror.CodeBillingNotConfigured, "Stripe webhook secret is not configured", nil)
	}

	event, err := webhook.ConstructEvent(payload, signature, s.cfg.WebhookSecret)
	if err != nil {
		return apperror.BadRequest(apperror.CodeBadRequest, "invalid Stripe webhook signature", err)
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return apperror.BadRequest(apperror.CodeBadRequest, "invalid checkout session event", err)
		}
		return s.handleCheckoutCompleted(ctx, &session)
	case stripe.EventTypeCustomerSubscriptionCreated,
		stripe.EventTypeCustomerSubscriptionUpdated,
		stripe.EventTypeCustomerSubscriptionDeleted,
		stripe.EventTypeCustomerSubscriptionPaused,
		stripe.EventTypeCustomerSubscriptionResumed:
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return apperror.BadRequest(apperror.CodeBadRequest, "invalid subscription event", err)
		}
		return s.syncSubscription(ctx, &sub)
	default:
		return nil
	}
}

func (s *Service) requireConfigured() *apperror.AppError {
	if s.cfg.SecretKey == "" {
		return apperror.BadRequest(apperror.CodeBillingNotConfigured, "Stripe billing is not configured", nil)
	}
	return nil
}

func (s *Service) getUser(ctx context.Context, userID uuid.UUID) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user for billing", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	return user, nil
}

func (s *Service) ensureStripeCustomer(ctx context.Context, user *entity.User) (string, *apperror.AppError) {
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		return *user.StripeCustomerID, nil
	}

	params := &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
		Email:  stripe.String(user.Email.String()),
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
	}
	if displayName := strings.TrimSpace(user.Profile.DisplayName); displayName != "" {
		params.Name = stripe.String(displayName)
	}

	customer, err := customerapi.New(params)
	if err != nil {
		s.log.Error("creating stripe customer", "error", err)
		return "", apperror.Internal(apperror.CodeBillingFailed, "failed to create billing account", err)
	}

	user.StripeCustomerID = &customer.ID
	user.UpdatedAt = time.Now().UTC()
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("saving stripe customer id", "error", err)
		return "", apperror.Internal(apperror.CodeInternalError, "failed to save billing account", err)
	}

	return customer.ID, nil
}

func (s *Service) checkoutLineItem(plan *entity.Plan) *stripe.CheckoutSessionLineItemParams {
	lineItem := &stripe.CheckoutSessionLineItemParams{
		Quantity: stripe.Int64(1),
	}
	if plan.StripePriceID != "" {
		lineItem.Price = stripe.String(plan.StripePriceID)
		return lineItem
	}

	lineItem.PriceData = &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String(s.currency()),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name:        stripe.String(plan.Name),
			Description: stripe.String(plan.Description),
			Metadata: map[string]string{
				"plan_id": plan.ID.String(),
			},
		},
		Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		UnitAmount: new(int64(math.Round(plan.Price * 100))),
	}
	return lineItem
}

func (s *Service) currency() string {
	currency := strings.ToLower(strings.TrimSpace(s.cfg.Currency))
	if currency == "" {
		return "usd"
	}
	return currency
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, session *stripe.CheckoutSession) *apperror.AppError {
	if session.Subscription == nil || session.Subscription.ID == "" {
		s.log.Warn("checkout session completed without subscription", "session_id", session.ID)
		return nil
	}

	sub, err := subscriptionapi.Get(session.Subscription.ID, &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
	})
	if err != nil {
		s.log.Error("retrieving stripe subscription from checkout", "error", err, "session_id", session.ID)
		return apperror.Internal(apperror.CodeBillingFailed, "failed to retrieve subscription", err)
	}

	if sub.Metadata == nil {
		sub.Metadata = map[string]string{}
	}
	for key, value := range session.Metadata {
		if _, exists := sub.Metadata[key]; !exists {
			sub.Metadata[key] = value
		}
	}

	return s.syncSubscription(ctx, sub)
}

func (s *Service) syncSubscription(ctx context.Context, stripeSub *stripe.Subscription) *apperror.AppError {
	userID, planID, ok := stripeSubscriptionIDs(stripeSub)
	if !ok {
		s.log.Warn("stripe subscription missing local metadata", "subscription_id", stripeSub.ID)
		return nil
	}

	customerID := ""
	if stripeSub.Customer != nil {
		customerID = stripeSub.Customer.ID
	}

	user, appErr := s.getUser(ctx, userID)
	if appErr != nil {
		return appErr
	}
	if customerID != "" && (user.StripeCustomerID == nil || *user.StripeCustomerID != customerID) {
		user.StripeCustomerID = &customerID
		user.UpdatedAt = time.Now().UTC()
		if err := s.userRepo.Update(ctx, user); err != nil {
			s.log.Error("updating user stripe customer id from webhook", "error", err)
			return apperror.Internal(apperror.CodeInternalError, "failed to update billing account", err)
		}
	}

	status := localSubscriptionStatus(stripeSub)
	expiresAt := stripeSubscriptionPeriodEnd(stripeSub)

	localSub, err := s.subRepo.GetByStripeSubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		s.log.Error("getting subscription by stripe id for sync", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to sync subscription", err)
	}
	if localSub == nil {
		localSub, err = s.subRepo.GetActiveByUserID(ctx, userID)
		if err != nil {
			s.log.Error("getting active subscription for stripe sync", "error", err)
			return apperror.Internal(apperror.CodeInternalError, "failed to sync subscription", err)
		}
	}

	now := time.Now().UTC()
	if localSub == nil {
		if status != entity.SubscriptionStatusActive {
			return nil
		}
		localSub = entity.NewSubscription(userID, planID, expiresAt)
		localSub.StripeSubscriptionID = stripeSub.ID
		localSub.StripeCustomerID = customerID
		if err := s.subRepo.Create(ctx, localSub); err != nil {
			s.log.Error("creating local subscription from stripe", "error", err)
			return apperror.Internal(apperror.CodeInternalError, "failed to sync subscription", err)
		}
		return nil
	}

	localSub.PlanID = planID
	localSub.Status = status
	localSub.StripeSubscriptionID = stripeSub.ID
	localSub.StripeCustomerID = customerID
	localSub.ExpiresAt = expiresAt
	localSub.UpdatedAt = now
	if err := s.subRepo.Update(ctx, localSub); err != nil {
		s.log.Error("updating local subscription from stripe", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to sync subscription", err)
	}

	return nil
}

func stripeSubscriptionIDs(stripeSub *stripe.Subscription) (userID, planID uuid.UUID, success bool) {
	if stripeSub == nil || stripeSub.Metadata == nil {
		return uuid.Nil, uuid.Nil, false
	}

	userID, err := uuid.Parse(stripeSub.Metadata["user_id"])
	if err != nil {
		return uuid.Nil, uuid.Nil, false
	}
	planID, err = uuid.Parse(stripeSub.Metadata["plan_id"])
	if err != nil {
		return uuid.Nil, uuid.Nil, false
	}

	return userID, planID, true
}

func localSubscriptionStatus(stripeSub *stripe.Subscription) entity.SubscriptionStatus {
	switch string(stripeSub.Status) {
	case "active", "trialing":
		return entity.SubscriptionStatusActive
	case "canceled", "incomplete_expired", "unpaid":
		return entity.SubscriptionStatusCancelled
	default:
		return entity.SubscriptionStatusExpired
	}
}

func stripeSubscriptionPeriodEnd(stripeSub *stripe.Subscription) *time.Time {
	if stripeSub == nil || stripeSub.Items == nil || len(stripeSub.Items.Data) == 0 {
		return nil
	}
	periodEnd := stripeSub.Items.Data[0].CurrentPeriodEnd
	if periodEnd == 0 {
		return nil
	}
	expiresAt := time.Unix(periodEnd, 0).UTC()
	return &expiresAt
}
