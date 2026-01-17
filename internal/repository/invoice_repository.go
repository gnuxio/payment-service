package repository

import (
	"database/sql"
	"fmt"

	"github.com/naventro/payment-service/internal/models"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(invoice *models.Invoice) error {
	query := `
		INSERT INTO invoices (
			subscription_id, stripe_invoice_id, user_id, tenant,
			amount_paid, currency, status, invoice_pdf, hosted_invoice_url,
			period_start, period_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		invoice.SubscriptionID,
		invoice.StripeInvoiceID,
		invoice.UserID,
		invoice.Tenant,
		invoice.AmountPaid,
		invoice.Currency,
		invoice.Status,
		invoice.InvoicePDF,
		invoice.HostedInvoiceURL,
		invoice.PeriodStart,
		invoice.PeriodEnd,
	).Scan(&invoice.ID, &invoice.CreatedAt)

	if err != nil {
		return fmt.Errorf("error creating invoice: %w", err)
	}

	return nil
}

func (r *InvoiceRepository) GetByStripeInvoiceID(stripeInvoiceID string) (*models.Invoice, error) {
	query := `
		SELECT
			id, subscription_id, stripe_invoice_id, user_id, tenant,
			amount_paid, currency, status, invoice_pdf, hosted_invoice_url,
			period_start, period_end, created_at
		FROM invoices
		WHERE stripe_invoice_id = $1
	`

	invoice := &models.Invoice{}
	err := r.db.QueryRow(query, stripeInvoiceID).Scan(
		&invoice.ID,
		&invoice.SubscriptionID,
		&invoice.StripeInvoiceID,
		&invoice.UserID,
		&invoice.Tenant,
		&invoice.AmountPaid,
		&invoice.Currency,
		&invoice.Status,
		&invoice.InvoicePDF,
		&invoice.HostedInvoiceURL,
		&invoice.PeriodStart,
		&invoice.PeriodEnd,
		&invoice.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching invoice: %w", err)
	}

	return invoice, nil
}

func (r *InvoiceRepository) GetByUserID(userID, tenant string) ([]*models.Invoice, error) {
	query := `
		SELECT
			id, subscription_id, stripe_invoice_id, user_id, tenant,
			amount_paid, currency, status, invoice_pdf, hosted_invoice_url,
			period_start, period_end, created_at
		FROM invoices
		WHERE user_id = $1 AND tenant = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID, tenant)
	if err != nil {
		return nil, fmt.Errorf("error fetching invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		err := rows.Scan(
			&invoice.ID,
			&invoice.SubscriptionID,
			&invoice.StripeInvoiceID,
			&invoice.UserID,
			&invoice.Tenant,
			&invoice.AmountPaid,
			&invoice.Currency,
			&invoice.Status,
			&invoice.InvoicePDF,
			&invoice.HostedInvoiceURL,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoices: %w", err)
	}

	return invoices, nil
}

func (r *InvoiceRepository) Update(invoice *models.Invoice) error {
	query := `
		UPDATE invoices
		SET status = $1, amount_paid = $2, invoice_pdf = $3, hosted_invoice_url = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(
		query,
		invoice.Status,
		invoice.AmountPaid,
		invoice.InvoicePDF,
		invoice.HostedInvoiceURL,
		invoice.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}
