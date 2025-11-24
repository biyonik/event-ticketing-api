package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/pkg/database"
)

type ReservationRepository struct {
	db      *sql.DB
	grammar database.Grammar
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{
		db:      db,
		grammar: database.NewMySQLGrammar(),
	}
}

// Payment Repository Methods

// CreatePayment - Conduit-Go Builder ile payment oluşturma
func (r *ReservationRepository) CreatePayment(payment *models.Payment) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("payments").
		ExecInsert(map[string]interface{}{
			"user_id":           payment.UserID,
			"event_id":          payment.EventID,
			"amount":            payment.Amount,
			"currency":          payment.Currency,
			"status":            payment.Status,
			"payment_method":    payment.PaymentMethod,
			"transaction_id":    payment.TransactionID,
			"provider_response": payment.ProviderResponse,
			"created_at":        payment.CreatedAt,
			"updated_at":        payment.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create payment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindPaymentByID - Conduit-Go Builder ile single payment
func (r *ReservationRepository) FindPaymentByID(id int64) (*models.Payment, error) {
	var payment models.Payment

	err := database.NewBuilder(r.db, r.grammar).
		Table("payments").
		Where("id", "=", id).
		First(&payment)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindPaymentByTransactionID - Builder ile transaction ID search
func (r *ReservationRepository) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment

	err := database.NewBuilder(r.db, r.grammar).
		Table("payments").
		Where("transaction_id", "=", transactionID).
		First(&payment)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindPaymentsByUserID - Builder ile user payments
func (r *ReservationRepository) FindPaymentsByUserID(userID int64) ([]*models.Payment, error) {
	var payments []*models.Payment

	err := database.NewBuilder(r.db, r.grammar).
		Table("payments").
		Where("user_id", "=", userID).
		OrderBy("created_at", "DESC").
		Get(&payments)

	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}

	return payments, nil
}

// UpdatePaymentStatus - Builder ile payment status güncelleme
func (r *ReservationRepository) UpdatePaymentStatus(id int64, status models.PaymentStatus, providerResponse string) error {
	now := time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("payments").
		Where("id", "=", id).
		ExecUpdate(map[string]interface{}{
			"status":            status,
			"provider_response": providerResponse,
			"processed_at":      now,
			"updated_at":        now,
		})

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

// GetTotalRevenueByEvent - SUM query (raw SQL for aggregate functions)
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

// AddToWaitingList - Conduit-Go Builder ile waiting list ekleme
func (r *ReservationRepository) AddToWaitingList(waitingList *models.WaitingList) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		ExecInsert(map[string]interface{}{
			"event_id":   waitingList.EventID,
			"user_id":    waitingList.UserID,
			"status":     waitingList.Status,
			"priority":   waitingList.Priority,
			"created_at": waitingList.CreatedAt,
			"updated_at": waitingList.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to add to waiting list: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindWaitingListByID - Builder ile single waiting list entry
func (r *ReservationRepository) FindWaitingListByID(id int64) (*models.WaitingList, error) {
	var waitingList models.WaitingList

	err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("id", "=", id).
		First(&waitingList)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("waiting list entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find waiting list entry: %w", err)
	}

	return &waitingList, nil
}

// FindWaitingListByEvent - Builder ile event waiting list
func (r *ReservationRepository) FindWaitingListByEvent(eventID int64, limit int) ([]*models.WaitingList, error) {
	var waitingLists []*models.WaitingList

	err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("event_id", "=", eventID).
		Where("status", "=", models.WaitingListStatusWaiting).
		OrderBy("priority", "DESC").
		OrderBy("created_at", "ASC").
		Limit(limit).
		Get(&waitingLists)

	if err != nil {
		return nil, fmt.Errorf("failed to query waiting list: %w", err)
	}

	return waitingLists, nil
}

// FindWaitingListByUser - Builder ile user waiting list
func (r *ReservationRepository) FindWaitingListByUser(userID int64) ([]*models.WaitingList, error) {
	var waitingLists []*models.WaitingList

	err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("user_id", "=", userID).
		OrderBy("created_at", "DESC").
		Get(&waitingLists)

	if err != nil {
		return nil, fmt.Errorf("failed to query waiting list: %w", err)
	}

	return waitingLists, nil
}

// UpdateWaitingListStatus - Builder ile waiting list status güncelleme
func (r *ReservationRepository) UpdateWaitingListStatus(id int64, status models.WaitingListStatus) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("id", "=", id).
		ExecUpdate(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

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

// MarkAsNotified - Builder ile notified işareti
func (r *ReservationRepository) MarkAsNotified(id int64) error {
	now := time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("id", "=", id).
		Where("status", "=", models.WaitingListStatusWaiting).
		ExecUpdate(map[string]interface{}{
			"status":      models.WaitingListStatusNotified,
			"notified_at": now,
			"updated_at":  now,
		})

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

// RemoveFromWaitingList - Builder ile waiting list silme
func (r *ReservationRepository) RemoveFromWaitingList(id int64) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("waiting_lists").
		Where("id", "=", id).
		ExecDelete()

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

// GetWaitingListCount - COUNT query (raw SQL for aggregate functions)
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

// IsUserInWaitingList - COUNT query (raw SQL for aggregate functions)
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
