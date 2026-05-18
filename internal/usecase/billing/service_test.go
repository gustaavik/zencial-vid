package billing

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

type mockUserRepo struct {
	createFn        func(context.Context, *entity.User) error
	getByIDFn       func(context.Context, uuid.UUID) (*entity.User, error)
	getByEmailFn    func(context.Context, valueobject.Email) (*entity.User, error)
	updateFn        func(context.Context, *entity.User) error
	deleteFn        func(context.Context, uuid.UUID) error
	existsByEmailFn func(context.Context, valueobject.Email) (bool, error)
	listFn          func(context.Context, *filter.FilterSet) ([]entity.User, int64, error)
	updateStatusFn  func(context.Context, uuid.UUID, entity.UserStatus) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *entity.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(ctx, email)
	}
	return false, nil
}

func (m *mockUserRepo) List(ctx context.Context, fs *filter.FilterSet) ([]entity.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.UserStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

type mockPlanRepo struct {
	createFn       func(context.Context, *entity.Plan) error
	getByIDFn      func(context.Context, uuid.UUID) (*entity.Plan, error)
	getBySlugFn    func(context.Context, valueobject.Slug) (*entity.Plan, error)
	updateFn       func(context.Context, *entity.Plan) error
	deleteFn       func(context.Context, uuid.UUID) error
	listFn         func(context.Context, *filter.FilterSet) ([]entity.Plan, int64, error)
	listActiveFn   func(context.Context) ([]entity.Plan, error)
	existsBySlugFn func(context.Context, valueobject.Slug) (bool, error)
}

func (m *mockPlanRepo) Create(ctx context.Context, plan *entity.Plan) error {
	if m.createFn != nil {
		return m.createFn(ctx, plan)
	}
	return nil
}

func (m *mockPlanRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPlanRepo) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Plan, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	return nil, nil
}

func (m *mockPlanRepo) Update(ctx context.Context, plan *entity.Plan) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, plan)
	}
	return nil
}

func (m *mockPlanRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockPlanRepo) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Plan, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockPlanRepo) ListActive(ctx context.Context) ([]entity.Plan, error) {
	if m.listActiveFn != nil {
		return m.listActiveFn(ctx)
	}
	return nil, nil
}

func (m *mockPlanRepo) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	if m.existsBySlugFn != nil {
		return m.existsBySlugFn(ctx, slug)
	}
	return false, nil
}

type mockSubRepo struct {
	createFn                    func(context.Context, *entity.Subscription) error
	getByIDFn                   func(context.Context, uuid.UUID) (*entity.Subscription, error)
	getByStripeSubscriptionIDFn func(context.Context, string) (*entity.Subscription, error)
	getActiveByUserIDFn         func(context.Context, uuid.UUID) (*entity.Subscription, error)
	updateFn                    func(context.Context, *entity.Subscription) error
	cancelFn                    func(context.Context, uuid.UUID) error
	listByUserIDFn              func(context.Context, uuid.UUID) ([]entity.Subscription, error)
}

func (m *mockSubRepo) Create(ctx context.Context, sub *entity.Subscription) error {
	if m.createFn != nil {
		return m.createFn(ctx, sub)
	}
	return nil
}

func (m *mockSubRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSubRepo) GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entity.Subscription, error) {
	if m.getByStripeSubscriptionIDFn != nil {
		return m.getByStripeSubscriptionIDFn(ctx, stripeSubscriptionID)
	}
	return nil, nil
}

func (m *mockSubRepo) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	if m.getActiveByUserIDFn != nil {
		return m.getActiveByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockSubRepo) Update(ctx context.Context, sub *entity.Subscription) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, sub)
	}
	return nil
}

func (m *mockSubRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, id)
	}
	return nil
}

func (m *mockSubRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testUser(id uuid.UUID, stripeCustomerID *string) *entity.User {
	now := time.Now().UTC()
	return &entity.User{
		ID:               id,
		Email:            valueobject.EmailFromTrusted("viewer@example.com"),
		PasswordHash:     valueobject.NewHashedPassword("hash"),
		Roles:            []entity.UserRole{entity.RoleUser},
		Status:           entity.UserStatusActive,
		StripeCustomerID: stripeCustomerID,
		Profile: entity.UserProfile{
			UserID:      id,
			DisplayName: "Viewer",
			Language:    "en",
			UpdatedAt:   now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestService_CreateCheckoutSession_MissingConfig(t *testing.T) {
	svc := NewService(&mockUserRepo{}, &mockPlanRepo{}, &mockSubRepo{}, Config{}, testLogger())

	_, appErr := svc.CreateCheckoutSession(context.Background(), CheckoutInput{})

	require.NotNil(t, appErr)
	assert.Equal(t, apperror.CodeBillingNotConfigured, appErr.Code)
}

func TestService_CreateCheckoutSession_RequiresReturnURLs(t *testing.T) {
	svc := NewService(&mockUserRepo{}, &mockPlanRepo{}, &mockSubRepo{}, Config{SecretKey: "sk_test_123"}, testLogger())

	_, appErr := svc.CreateCheckoutSession(context.Background(), CheckoutInput{
		UserID: uuid.New(),
		PlanID: uuid.New(),
	})

	require.NotNil(t, appErr)
	assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
}

func TestService_ListInvoices_NoStripeCustomerReturnsEmpty(t *testing.T) {
	userID := uuid.New()
	svc := NewService(&mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
			assert.Equal(t, userID, id)
			return testUser(userID, nil), nil
		},
	}, &mockPlanRepo{}, &mockSubRepo{}, Config{SecretKey: "sk_test_123"}, testLogger())

	invoices, appErr := svc.ListInvoices(context.Background(), userID, 12)

	require.Nil(t, appErr)
	assert.Empty(t, invoices)
}

func TestService_SyncSubscription_CreatesActiveLocalSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	planID := uuid.New()
	periodEnd := time.Now().UTC().Add(30 * 24 * time.Hour)
	var updatedUser *entity.User
	var createdSub *entity.Subscription

	svc := NewService(
		&mockUserRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
				assert.Equal(t, userID, id)
				return testUser(userID, nil), nil
			},
			updateFn: func(_ context.Context, user *entity.User) error {
				updatedUser = user
				return nil
			},
		},
		&mockPlanRepo{},
		&mockSubRepo{
			createFn: func(_ context.Context, sub *entity.Subscription) error {
				createdSub = sub
				return nil
			},
		},
		Config{SecretKey: "sk_test_123"},
		testLogger(),
	)

	appErr := svc.syncSubscription(ctx, &stripe.Subscription{
		ID:       "sub_123",
		Customer: &stripe.Customer{ID: "cus_123"},
		Status:   stripe.SubscriptionStatusActive,
		Metadata: map[string]string{
			"user_id": userID.String(),
			"plan_id": planID.String(),
		},
		Items: &stripe.SubscriptionItemList{
			Data: []*stripe.SubscriptionItem{
				{CurrentPeriodEnd: periodEnd.Unix()},
			},
		},
	})

	require.Nil(t, appErr)
	require.NotNil(t, updatedUser)
	require.NotNil(t, updatedUser.StripeCustomerID)
	assert.Equal(t, "cus_123", *updatedUser.StripeCustomerID)
	require.NotNil(t, createdSub)
	assert.Equal(t, userID, createdSub.UserID)
	assert.Equal(t, planID, createdSub.PlanID)
	assert.Equal(t, entity.SubscriptionStatusActive, createdSub.Status)
	assert.Equal(t, "sub_123", createdSub.StripeSubscriptionID)
	assert.Equal(t, "cus_123", createdSub.StripeCustomerID)
	require.NotNil(t, createdSub.ExpiresAt)
	assert.Equal(t, periodEnd.Unix(), createdSub.ExpiresAt.Unix())
}
