package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Payment Repository Methods
func (r *ReservationRepository) CreatePayment(payment *models.Payment) (int64, error) {
	query := `
		INSERT INTO payments (user_id, event_id, amount, currency, status, payment_method,
			transaction_id, provider_response, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		payment.UserID, payment.EventID, payment.Amount, payment.Currency, payment.Status,
		payment.PaymentMethod, payment.TransactionID, payment.ProviderResponse,
		payment.CreatedAt, payment.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create payment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *ReservationRepository) FindPaymentByID(id int64) (*models.Payment, error) {
	query := `
		SELECT id, user_id, event_id, amount, currency, status, payment_method,
			transaction_id, provider_response, processed_at, created_at, updated_at
		FROM payments
		WHERE id = ?
	`

	payment := &models.Payment{}
	err := r.db.QueryRow(query, id).Scan(
		&payment.ID, &payment.UserID, &payment.EventID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.PaymentMethod, &payment.TransactionID,
		&payment.ProviderResponse, &payment.ProcessedAt, &payment.CreatedAt, &payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return payment, nil
}

func (r *ReservationRepository) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	query := `
		SELECT id, user_id, event_id, amount, currency, status, payment_method,
			transaction_id, provider_response, processed_at, created_at, updated_at
		FROM payments
		WHERE transaction_id = ?
	`

	payment := &models.Payment{}
	err := r.db.QueryRow(query, transactionID).Scan(
		&payment.ID, &payment.UserID, &payment.EventID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.PaymentMethod, &payment.TransactionID,
		&payment.ProviderResponse, &payment.ProcessedAt, &payment.CreatedAt, &payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return payment, nil
}

func (r *ReservationRepository) FindPaymentsByUserID(userID int64) ([]*models.Payment, error) {
	query := `
		SELECT id, user_id, event_id, amount, currency, status, payment_method,
			transaction_id, provider_response, processed_at, created_at, updated_at
		FROM payments
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	payments := []*models.Payment{}
	for rows.Next() {
		payment := &models.Payment{}
		err := rows.Scan(
			&payment.ID, &payment.UserID, &payment.EventID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.PaymentMethod, &payment.TransactionID,
			&payment.ProviderResponse, &payment.ProcessedAt, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *ReservationRepository) UpdatePaymentStatus(id int64, status models.PaymentStatus, providerResponse string) error {
	query := `
		UPDATE payments
		SET status = ?, provider_response = ?, processed_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(query, status, providerResponse, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

func (r *ReservationRepository) GetTotalRevenueByEvent(eventID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE event_id = ? AND status = ?
	`

	var revenue float64
	err := r.db.QueryRow(query, eventID, models.PaymentStatusCompleted).Scan(&revenue)
	if err != nil {
		return 0, fmt.Errorf("failed to get total revenue: %w", err)
	}

	return revenue, nil
}

// Waiting List Repository Methods
func (r *ReservationRepository) AddToWaitingList(waitingList *models.WaitingList) (int64, error) {
	query := `
		INSERT INTO waiting_lists (event_id, user_id, status, priority, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		waitingList.EventID, waitingList.UserID, waitingList.Status,
		waitingList.Priority, waitingList.CreatedAt, waitingList.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to add to waiting list: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *ReservationRepository) FindWaitingListByID(id int64) (*models.WaitingList, error) {
	query := `
		SELECT id, event_id, user_id, status, priority, notified_at, created_at, updated_at
		FROM waiting_lists
		WHERE id = ?
	`

	waitingList := &models.WaitingList{}
	err := r.db.QueryRow(query, id).Scan(
		&waitingList.ID, &waitingList.EventID, &waitingList.UserID, &waitingList.Status,
		&waitingList.Priority, &waitingList.NotifiedAt, &waitingList.CreatedAt, &waitingList.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("waiting list entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find waiting list entry: %w", err)
	}

	return waitingList, nil
}

func (r *ReservationRepository) FindWaitingListByEvent(eventID int64, limit int) ([]*models.WaitingList, error) {
	query := `
		SELECT id, event_id, user_id, status, priority, notified_at, created_at, updated_at
		FROM waiting_lists
		WHERE event_id = ? AND status = ?
		ORDER BY priority DESC, created_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, eventID, models.WaitingListStatusWaiting, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query waiting list: %w", err)
	}
	defer rows.Close()

	waitingLists := []*models.WaitingList{}
	for rows.Next() {
		waitingList := &models.WaitingList{}
		err := rows.Scan(
			&waitingList.ID, &waitingList.EventID, &waitingList.UserID, &waitingList.Status,
			&waitingList.Priority, &waitingList.NotifiedAt, &waitingList.CreatedAt, &waitingList.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan waiting list entry: %w", err)
		}
		waitingLists = append(waitingLists, waitingList)
	}

	return waitingLists, nil
}

func (r *ReservationRepository) FindWaitingListByUser(userID int64) ([]*models.WaitingList, error) {
	query := `
		SELECT id, event_id, user_id, status, priority, notified_at, created_at, updated_at
		FROM waiting_lists
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query waiting list: %w", err)
	}
	defer rows.Close()

	waitingLists := []*models.WaitingList{}
	for rows.Next() {
		waitingList := &models.WaitingList{}
		err := rows.Scan(
			&waitingList.ID, &waitingList.EventID, &waitingList.UserID, &waitingList.Status,
			&waitingList.Priority, &waitingList.NotifiedAt, &waitingList.CreatedAt, &waitingList.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan waiting list entry: %w", err)
		}
		waitingLists = append(waitingLists, waitingList)
	}

	return waitingLists, nil
}

func (r *ReservationRepository) UpdateWaitingListStatus(id int64, status models.WaitingListStatus) error {
	query := `
		UPDATE waiting_lists
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update waiting list status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("waiting list entry not found")
	}

	return nil
}

func (r *ReservationRepository) MarkAsNotified(id int64) error {
	query := `
		UPDATE waiting_lists
		SET status = ?, notified_at = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`

	now := time.Now()
	result, err := r.db.Exec(query, models.WaitingListStatusNotified, now, now, id, models.WaitingListStatusWaiting)
	if err != nil {
		return fmt.Errorf("failed to mark as notified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("waiting list entry not found or invalid status")
	}

	return nil
}

func (r *ReservationRepository) RemoveFromWaitingList(id int64) error {
	query := `
		DELETE FROM waiting_lists
		WHERE id = ?
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove from waiting list: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("waiting list entry not found")
	}

	return nil
}

func (r *ReservationRepository) GetWaitingListCount(eventID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM waiting_lists
		WHERE event_id = ? AND status = ?
	`

	var count int
	err := r.db.QueryRow(query, eventID, models.WaitingListStatusWaiting).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get waiting list count: %w", err)
	}

	return count, nil
}

func (r *ReservationRepository) IsUserInWaitingList(eventID, userID int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM waiting_lists
		WHERE event_id = ? AND user_id = ? AND status = ?
	`

	var count int
	err := r.db.QueryRow(query, eventID, userID, models.WaitingListStatusWaiting).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check waiting list: %w", err)
	}

	return count > 0, nil
}
