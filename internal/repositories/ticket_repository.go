package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ticket *models.Ticket) (int64, error) {
	query := `
		INSERT INTO tickets (event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		ticket.EventID, ticket.UserID, ticket.SeatID, ticket.SectionID,
		ticket.TicketNumber, ticket.TicketType, ticket.Status, ticket.Price,
		ticket.QRCodeData, ticket.QRCodeImage, ticket.VerificationCode,
		ticket.ReservationExpiry, ticket.CreatedAt, ticket.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create ticket: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *TicketRepository) FindByID(id int64) (*models.Ticket, error) {
	query := `
		SELECT id, event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, purchased_at, used_at, cancelled_at, created_at, updated_at
		FROM tickets
		WHERE id = ?
	`

	ticket := &models.Ticket{}
	err := r.db.QueryRow(query, id).Scan(
		&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.SeatID, &ticket.SectionID,
		&ticket.TicketNumber, &ticket.TicketType, &ticket.Status, &ticket.Price,
		&ticket.QRCodeData, &ticket.QRCodeImage, &ticket.VerificationCode,
		&ticket.ReservationExpiry, &ticket.PurchasedAt, &ticket.UsedAt,
		&ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ticket not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}

	return ticket, nil
}

func (r *TicketRepository) FindByTicketNumber(ticketNumber string) (*models.Ticket, error) {
	query := `
		SELECT id, event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, purchased_at, used_at, cancelled_at, created_at, updated_at
		FROM tickets
		WHERE ticket_number = ?
	`

	ticket := &models.Ticket{}
	err := r.db.QueryRow(query, ticketNumber).Scan(
		&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.SeatID, &ticket.SectionID,
		&ticket.TicketNumber, &ticket.TicketType, &ticket.Status, &ticket.Price,
		&ticket.QRCodeData, &ticket.QRCodeImage, &ticket.VerificationCode,
		&ticket.ReservationExpiry, &ticket.PurchasedAt, &ticket.UsedAt,
		&ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ticket not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}

	return ticket, nil
}

func (r *TicketRepository) FindByUserID(userID int64) ([]*models.Ticket, error) {
	query := `
		SELECT id, event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, purchased_at, used_at, cancelled_at, created_at, updated_at
		FROM tickets
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []*models.Ticket{}
	for rows.Next() {
		ticket := &models.Ticket{}
		err := rows.Scan(
			&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.SeatID, &ticket.SectionID,
			&ticket.TicketNumber, &ticket.TicketType, &ticket.Status, &ticket.Price,
			&ticket.QRCodeData, &ticket.QRCodeImage, &ticket.VerificationCode,
			&ticket.ReservationExpiry, &ticket.PurchasedAt, &ticket.UsedAt,
			&ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (r *TicketRepository) FindByEventID(eventID int64) ([]*models.Ticket, error) {
	query := `
		SELECT id, event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, purchased_at, used_at, cancelled_at, created_at, updated_at
		FROM tickets
		WHERE event_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []*models.Ticket{}
	for rows.Next() {
		ticket := &models.Ticket{}
		err := rows.Scan(
			&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.SeatID, &ticket.SectionID,
			&ticket.TicketNumber, &ticket.TicketType, &ticket.Status, &ticket.Price,
			&ticket.QRCodeData, &ticket.QRCodeImage, &ticket.VerificationCode,
			&ticket.ReservationExpiry, &ticket.PurchasedAt, &ticket.UsedAt,
			&ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (r *TicketRepository) Update(ticket *models.Ticket) error {
	query := `
		UPDATE tickets
		SET status = ?, reservation_expiry = ?, purchased_at = ?, used_at = ?,
			cancelled_at = ?, updated_at = ?
		WHERE id = ?
	`

	ticket.UpdatedAt = time.Now()

	result, err := r.db.Exec(query,
		ticket.Status, ticket.ReservationExpiry, ticket.PurchasedAt,
		ticket.UsedAt, ticket.CancelledAt, ticket.UpdatedAt, ticket.ID,
	)

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

func (r *TicketRepository) UpdateStatus(id int64, status models.TicketStatus) error {
	query := `
		UPDATE tickets
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
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

func (r *TicketRepository) MarkAsSold(id int64) error {
	query := `
		UPDATE tickets
		SET status = ?, purchased_at = ?, reservation_expiry = NULL, updated_at = ?
		WHERE id = ? AND status = ?
	`

	now := time.Now()
	result, err := r.db.Exec(query, models.TicketStatusSold, now, now, id, models.TicketStatusReserved)
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

func (r *TicketRepository) MarkAsUsed(id int64) error {
	query := `
		UPDATE tickets
		SET status = ?, used_at = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`

	now := time.Now()
	result, err := r.db.Exec(query, models.TicketStatusUsed, now, now, id, models.TicketStatusSold)
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

func (r *TicketRepository) MarkAsCancelled(id int64) error {
	query := `
		UPDATE tickets
		SET status = ?, cancelled_at = ?, updated_at = ?
		WHERE id = ? AND status IN (?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, models.TicketStatusCancelled, now, now, id,
		models.TicketStatusReserved, models.TicketStatusSold)
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

func (r *TicketRepository) FindExpiredReservations() ([]*models.Ticket, error) {
	query := `
		SELECT id, event_id, user_id, seat_id, section_id, ticket_number,
			ticket_type, status, price, qr_code_data, qr_code_image, verification_code,
			reservation_expiry, purchased_at, used_at, cancelled_at, created_at, updated_at
		FROM tickets
		WHERE status = ? AND reservation_expiry < ?
	`

	rows, err := r.db.Query(query, models.TicketStatusReserved, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query expired reservations: %w", err)
	}
	defer rows.Close()

	tickets := []*models.Ticket{}
	for rows.Next() {
		ticket := &models.Ticket{}
		err := rows.Scan(
			&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.SeatID, &ticket.SectionID,
			&ticket.TicketNumber, &ticket.TicketType, &ticket.Status, &ticket.Price,
			&ticket.QRCodeData, &ticket.QRCodeImage, &ticket.VerificationCode,
			&ticket.ReservationExpiry, &ticket.PurchasedAt, &ticket.UsedAt,
			&ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (r *TicketRepository) ExpireReservation(id int64) error {
	query := `
		UPDATE tickets
		SET status = ?, updated_at = ?
		WHERE id = ? AND status = ? AND reservation_expiry < ?
	`

	result, err := r.db.Exec(query, models.TicketStatusExpired, time.Now(), id,
		models.TicketStatusReserved, time.Now())
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
