package repository

import (
	"database/sql"
	"fmt"

	"github.com/naventro/payment-service/internal/models"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			user_id, tenant, stripe_customer_id, stripe_subscription_id,
			status, plan, current_period_start, current_period_end, cancel_at_period_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		sub.UserID,
		sub.Tenant,
		sub.StripeCustomerID,
		sub.StripeSubscriptionID,
		sub.Status,
		sub.Plan,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByUserID(userID, tenant string) (*models.Subscription, error) {
	query := `
		SELECT
			id, user_id, tenant, stripe_customer_id, stripe_subscription_id,
			status, plan, current_period_start, current_period_end,
			cancel_at_period_end, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1 AND tenant = $2
	`

	sub := &models.Subscription{}
	err := r.db.QueryRow(query, userID, tenant).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.Tenant,
		&sub.StripeCustomerID,
		&sub.StripeSubscriptionID,
		&sub.Status,
		&sub.Plan,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByStripeSubscriptionID(stripeSubID string) (*models.Subscription, error) {
	query := `
		SELECT
			id, user_id, tenant, stripe_customer_id, stripe_subscription_id,
			status, plan, current_period_start, current_period_end,
			cancel_at_period_end, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	sub := &models.Subscription{}
	err := r.db.QueryRow(query, stripeSubID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.Tenant,
		&sub.StripeCustomerID,
		&sub.StripeSubscriptionID,
		&sub.Status,
		&sub.Plan,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $1, plan = $2, current_period_start = $3,
		    current_period_end = $4, cancel_at_period_end = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`

	result, err := r.db.Exec(
		query,
		sub.Status,
		sub.Plan,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *SubscriptionRepository) Delete(id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}
