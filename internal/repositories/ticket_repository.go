package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/pkg/database"
)

type TicketRepository struct {
	db      *sql.DB
	grammar database.Grammar
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{
		db:      db,
		grammar: database.NewMySQLGrammar(),
	}
}

// Create - Conduit-Go Builder ile ticket oluşturma
func (r *TicketRepository) Create(ticket *models.Ticket) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		ExecInsert(map[string]interface{}{
			"event_id":           ticket.EventID,
			"user_id":            ticket.UserID,
			"seat_id":            ticket.SeatID,
			"section_id":         ticket.SectionID,
			"ticket_number":      ticket.TicketNumber,
			"ticket_type":        ticket.TicketType,
			"status":             ticket.Status,
			"price":              ticket.Price,
			"qr_code_data":       ticket.QRCodeData,
			"qr_code_image":      ticket.QRCodeImage,
			"verification_code":  ticket.VerificationCode,
			"reservation_expiry": ticket.ReservationExpiry,
			"created_at":         ticket.CreatedAt,
			"updated_at":         ticket.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create ticket: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindByID - Builder ile single ticket
func (r *TicketRepository) FindByID(id int64) (*models.Ticket, error) {
	var ticket models.Ticket

	err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		First(&ticket)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ticket not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}

	return &ticket, nil
}

// FindByTicketNumber - Builder ile ticket number search
func (r *TicketRepository) FindByTicketNumber(ticketNumber string) (*models.Ticket, error) {
	var ticket models.Ticket

	err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("ticket_number", "=", ticketNumber).
		First(&ticket)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ticket not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}

	return &ticket, nil
}

// FindByUserID - Builder ile user tickets
func (r *TicketRepository) FindByUserID(userID int64) ([]*models.Ticket, error) {
	var tickets []*models.Ticket

	err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("user_id", "=", userID).
		OrderBy("created_at", "DESC").
		Get(&tickets)

	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}

	return tickets, nil
}

// FindByEventID - Builder ile event tickets
func (r *TicketRepository) FindByEventID(eventID int64) ([]*models.Ticket, error) {
	var tickets []*models.Ticket

	err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("event_id", "=", eventID).
		OrderBy("created_at", "DESC").
		Get(&tickets)

	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}

	return tickets, nil
}

// Update - Builder ile ticket güncelleme
func (r *TicketRepository) Update(ticket *models.Ticket) error {
	ticket.UpdatedAt = time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", ticket.ID).
		ExecUpdate(map[string]interface{}{
			"status":             ticket.Status,
			"reservation_expiry": ticket.ReservationExpiry,
			"purchased_at":       ticket.PurchasedAt,
			"used_at":            ticket.UsedAt,
			"cancelled_at":       ticket.CancelledAt,
			"updated_at":         ticket.UpdatedAt,
		})

	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found")
	}

	return nil
}

// UpdateStatus - Builder ile status güncelleme
func (r *TicketRepository) UpdateStatus(id int64, status models.TicketStatus) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		ExecUpdate(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found")
	}

	return nil
}

// MarkAsSold - Builder ile sold işareti
func (r *TicketRepository) MarkAsSold(id int64) error {
	now := time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		Where("status", "=", models.TicketStatusReserved).
		ExecUpdate(map[string]interface{}{
			"status":             models.TicketStatusSold,
			"purchased_at":       now,
			"reservation_expiry": nil,
			"updated_at":         now,
		})

	if err != nil {
		return fmt.Errorf("failed to mark ticket as sold: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found or invalid status")
	}

	return nil
}

// MarkAsUsed - Builder ile used işareti
func (r *TicketRepository) MarkAsUsed(id int64) error {
	now := time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		Where("status", "=", models.TicketStatusSold).
		ExecUpdate(map[string]interface{}{
			"status":     models.TicketStatusUsed,
			"used_at":    now,
			"updated_at": now,
		})

	if err != nil {
		return fmt.Errorf("failed to mark ticket as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found or invalid status")
	}

	return nil
}

// MarkAsCancelled - Builder ile cancelled işareti (WhereIn kullanımı)
func (r *TicketRepository) MarkAsCancelled(id int64) error {
	now := time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		WhereIn("status", []interface{}{models.TicketStatusReserved, models.TicketStatusSold}).
		ExecUpdate(map[string]interface{}{
			"status":       models.TicketStatusCancelled,
			"cancelled_at": now,
			"updated_at":   now,
		})

	if err != nil {
		return fmt.Errorf("failed to mark ticket as cancelled: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found or invalid status")
	}

	return nil
}

// FindExpiredReservations - Builder ile expired tickets
func (r *TicketRepository) FindExpiredReservations() ([]*models.Ticket, error) {
	var tickets []*models.Ticket

	err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("status", "=", models.TicketStatusReserved).
		Where("reservation_expiry", "<", time.Now()).
		Get(&tickets)

	if err != nil {
		return nil, fmt.Errorf("failed to query expired reservations: %w", err)
	}

	return tickets, nil
}

// ExpireReservation - Builder ile reservation expire
func (r *TicketRepository) ExpireReservation(id int64) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("tickets").
		Where("id", "=", id).
		Where("status", "=", models.TicketStatusReserved).
		Where("reservation_expiry", "<", time.Now()).
		ExecUpdate(map[string]interface{}{
			"status":     models.TicketStatusExpired,
			"updated_at": time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to expire reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found or invalid status")
	}

	return nil
}

// GetSoldTicketCountByEvent - COUNT query (raw SQL needed for aggregate functions)
func (r *TicketRepository) GetSoldTicketCountByEvent(eventID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tickets
		WHERE event_id = ? AND status IN (?, ?)
	`

	var count int
	err := r.db.QueryRow(query, eventID, models.TicketStatusSold, models.TicketStatusUsed).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get sold ticket count: %w", err)
	}

	return count, nil
}

// IsSeatTaken - COUNT query (raw SQL for aggregate)
func (r *TicketRepository) IsSeatTaken(eventID, seatID int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM tickets
		WHERE event_id = ? AND seat_id = ? AND status IN (?, ?, ?)
	`

	var count int
	err := r.db.QueryRow(query, eventID, seatID,
		models.TicketStatusReserved, models.TicketStatusSold, models.TicketStatusUsed).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check seat availability: %w", err)
	}

	return count > 0, nil
}

// GetRevenueByEvent - SUM query (raw SQL for aggregate)
func (r *TicketRepository) GetRevenueByEvent(eventID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM tickets
		WHERE event_id = ? AND status IN (?, ?)
	`

	var revenue float64
	err := r.db.QueryRow(query, eventID, models.TicketStatusSold, models.TicketStatusUsed).Scan(&revenue)
	if err != nil {
		return 0, fmt.Errorf("failed to get revenue: %w", err)
	}

	return revenue, nil
}
