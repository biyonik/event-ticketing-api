package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) (int64, error) {
	query := `
		INSERT INTO events (name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		event.Name, event.Description, event.Type, event.Status, event.VenueID,
		event.StartTime, event.EndTime, event.BasePrice, event.TotalCapacity,
		event.AvailableSeats, event.ImageURL, event.Featured, event.Metadata,
		event.CreatedAt, event.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *EventRepository) FindByID(id int64) (*models.Event, error) {
	query := `
		SELECT id, name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata,
			created_at, updated_at, deleted_at
		FROM events
		WHERE id = ? AND deleted_at IS NULL
	`

	event := &models.Event{}
	err := r.db.QueryRow(query, id).Scan(
		&event.ID, &event.Name, &event.Description, &event.Type, &event.Status,
		&event.VenueID, &event.StartTime, &event.EndTime, &event.BasePrice,
		&event.TotalCapacity, &event.AvailableSeats, &event.ImageURL,
		&event.Featured, &event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		&event.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	return event, nil
}

func (r *EventRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]*models.Event, error) {
	query := `
		SELECT id, name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata,
			created_at, updated_at
		FROM events
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}

	// Apply filters
	if status, ok := filters["status"].(models.EventStatus); ok {
		query += " AND status = ?"
		args = append(args, status)
	}

	if eventType, ok := filters["type"].(models.EventType); ok {
		query += " AND type = ?"
		args = append(args, eventType)
	}

	if featured, ok := filters["featured"].(bool); ok {
		query += " AND featured = ?"
		args = append(args, featured)
	}

	if startDate, ok := filters["start_date"].(time.Time); ok {
		query += " AND start_time >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		query += " AND start_time <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY start_time DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := []*models.Event{}
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID, &event.Name, &event.Description, &event.Type, &event.Status,
			&event.VenueID, &event.StartTime, &event.EndTime, &event.BasePrice,
			&event.TotalCapacity, &event.AvailableSeats, &event.ImageURL,
			&event.Featured, &event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) Update(event *models.Event) error {
	query := `
		UPDATE events
		SET name = ?, description = ?, type = ?, status = ?, start_time = ?, end_time = ?,
			base_price = ?, available_seats = ?, image_url = ?, featured = ?, metadata = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	event.UpdatedAt = time.Now()

	result, err := r.db.Exec(query,
		event.Name, event.Description, event.Type, event.Status, event.StartTime,
		event.EndTime, event.BasePrice, event.AvailableSeats, event.ImageURL,
		event.Featured, event.Metadata, event.UpdatedAt, event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *EventRepository) Delete(id int64) error {
	query := `
		UPDATE events
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *EventRepository) UpdateStatus(id int64, status models.EventStatus) error {
	query := `
		UPDATE events
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *EventRepository) DecrementAvailableSeats(id int64, count int) error {
	query := `
		UPDATE events
		SET available_seats = available_seats - ?, updated_at = ?
		WHERE id = ? AND available_seats >= ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, count, time.Now(), id, count)
	if err != nil {
		return fmt.Errorf("failed to decrement available seats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient seats or event not found")
	}

	return nil
}

func (r *EventRepository) IncrementAvailableSeats(id int64, count int) error {
	query := `
		UPDATE events
		SET available_seats = available_seats + ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, count, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment available seats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *EventRepository) GetUpcomingEvents(limit int) ([]*models.Event, error) {
	query := `
		SELECT id, name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata,
			created_at, updated_at
		FROM events
		WHERE deleted_at IS NULL
			AND status IN (?, ?)
			AND start_time > NOW()
		ORDER BY start_time ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, models.EventStatusPublished, models.EventStatusSaleActive, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming events: %w", err)
	}
	defer rows.Close()

	events := []*models.Event{}
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID, &event.Name, &event.Description, &event.Type, &event.Status,
			&event.VenueID, &event.StartTime, &event.EndTime, &event.BasePrice,
			&event.TotalCapacity, &event.AvailableSeats, &event.ImageURL,
			&event.Featured, &event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) GetFeaturedEvents(limit int) ([]*models.Event, error) {
	query := `
		SELECT id, name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata,
			created_at, updated_at
		FROM events
		WHERE deleted_at IS NULL
			AND featured = true
			AND status IN (?, ?)
			AND start_time > NOW()
		ORDER BY start_time ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, models.EventStatusPublished, models.EventStatusSaleActive, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured events: %w", err)
	}
	defer rows.Close()

	events := []*models.Event{}
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID, &event.Name, &event.Description, &event.Type, &event.Status,
			&event.VenueID, &event.StartTime, &event.EndTime, &event.BasePrice,
			&event.TotalCapacity, &event.AvailableSeats, &event.ImageURL,
			&event.Featured, &event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) SearchByName(keyword string, limit int) ([]*models.Event, error) {
	query := `
		SELECT id, name, description, type, status, venue_id, start_time, end_time,
			base_price, total_capacity, available_seats, image_url, featured, metadata,
			created_at, updated_at
		FROM events
		WHERE deleted_at IS NULL
			AND (name LIKE ? OR description LIKE ?)
			AND status IN (?, ?)
		ORDER BY start_time ASC
		LIMIT ?
	`

	searchTerm := "%" + keyword + "%"
	rows, err := r.db.Query(query, searchTerm, searchTerm, models.EventStatusPublished, models.EventStatusSaleActive, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer rows.Close()

	events := []*models.Event{}
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID, &event.Name, &event.Description, &event.Type, &event.Status,
			&event.VenueID, &event.StartTime, &event.EndTime, &event.BasePrice,
			&event.TotalCapacity, &event.AvailableSeats, &event.ImageURL,
			&event.Featured, &event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
