package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/pkg/database"
)

type EventRepository struct {
	db      *sql.DB
	grammar database.Grammar
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db:      db,
		grammar: database.NewMySQLGrammar(),
	}
}

// Create - Conduit-Go Database Builder ile event oluşturma
func (r *EventRepository) Create(event *models.Event) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		ExecInsert(map[string]interface{}{
			"name":            event.Name,
			"description":     event.Description,
			"type":            event.Type,
			"status":          event.Status,
			"venue_id":        event.VenueID,
			"start_time":      event.StartTime,
			"end_time":        event.EndTime,
			"base_price":      event.BasePrice,
			"total_capacity":  event.TotalCapacity,
			"available_seats": event.AvailableSeats,
			"image_url":       event.ImageURL,
			"featured":        event.Featured,
			"metadata":        event.Metadata,
			"created_at":      event.CreatedAt,
			"updated_at":      event.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindByID - Conduit-Go Query Builder ile single event
func (r *EventRepository) FindByID(id int64) (*models.Event, error) {
	var event models.Event

	err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		Where("id", "=", id).
		WhereNull("deleted_at").
		First(&event)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	return &event, nil
}

// FindAll - Conduit-Go Query Builder ile filtreleme
func (r *EventRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]*models.Event, error) {
	builder := database.NewBuilder(r.db, r.grammar).
		Table("events").
		WhereNull("deleted_at")

	// Apply filters using Conduit-Go query builder
	if status, ok := filters["status"].(models.EventStatus); ok {
		builder.Where("status", "=", status)
	}

	if eventType, ok := filters["type"].(models.EventType); ok {
		builder.Where("type", "=", eventType)
	}

	if featured, ok := filters["featured"].(bool); ok {
		builder.Where("featured", "=", featured)
	}

	if startDate, ok := filters["start_date"].(time.Time); ok {
		builder.Where("start_time", ">=", startDate)
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		builder.Where("start_time", "<=", endDate)
	}

	var events []*models.Event
	err := builder.
		OrderBy("start_time", "DESC").
		Limit(limit).
		Offset(offset).
		Get(&events)

	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	return events, nil
}

// Update - Conduit-Go Builder ile güncelleme
func (r *EventRepository) Update(event *models.Event) error {
	event.UpdatedAt = time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		Where("id", "=", event.ID).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"name":            event.Name,
			"description":     event.Description,
			"type":            event.Type,
			"status":          event.Status,
			"start_time":      event.StartTime,
			"end_time":        event.EndTime,
			"base_price":      event.BasePrice,
			"available_seats": event.AvailableSeats,
			"image_url":       event.ImageURL,
			"featured":        event.Featured,
			"metadata":        event.Metadata,
			"updated_at":      event.UpdatedAt,
		})

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

// Delete - Soft delete using Conduit-Go Builder
func (r *EventRepository) Delete(id int64) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		Where("id", "=", id).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"deleted_at": time.Now(),
		})

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

// UpdateStatus - Builder ile status güncelleme
func (r *EventRepository) UpdateStatus(id int64, status models.EventStatus) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		Where("id", "=", id).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

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

// DecrementAvailableSeats - Atomic decrement (Conduit-Go'da increment/decrement yok, raw SQL kullanıyoruz ama prepared statement ile)
func (r *EventRepository) DecrementAvailableSeats(id int64, count int) error {
	// Conduit-Go'da henüz Decrement/Increment yok, raw SQL ama güvenli
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

// IncrementAvailableSeats - Atomic increment
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

// GetUpcomingEvents - Builder ile WhereIn kullanımı
func (r *EventRepository) GetUpcomingEvents(limit int) ([]*models.Event, error) {
	var events []*models.Event

	err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		WhereNull("deleted_at").
		WhereIn("status", []interface{}{models.EventStatusPublished, models.EventStatusSaleActive}).
		Where("start_time", ">", time.Now()).
		OrderBy("start_time", "ASC").
		Limit(limit).
		Get(&events)

	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming events: %w", err)
	}

	return events, nil
}

// GetFeaturedEvents - Builder ile featured events
func (r *EventRepository) GetFeaturedEvents(limit int) ([]*models.Event, error] {
	var events []*models.Event

	err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		WhereNull("deleted_at").
		Where("featured", "=", true).
		WhereIn("status", []interface{}{models.EventStatusPublished, models.EventStatusSaleActive}).
		Where("start_time", ">", time.Now()).
		OrderBy("start_time", "ASC").
		Limit(limit).
		Get(&events)

	if err != nil {
		return nil, fmt.Errorf("failed to get featured events: %w", err)
	}

	return events, nil
}

// SearchByName - Builder ile LIKE search (OrWhere ile)
func (r *EventRepository) SearchByName(keyword string, limit int) ([]*models.Event, error) {
	var events []*models.Event

	searchTerm := "%" + keyword + "%"

	err := database.NewBuilder(r.db, r.grammar).
		Table("events").
		WhereNull("deleted_at").
		Where("name", "LIKE", searchTerm).
		OrWhere("description", "LIKE", searchTerm).
		WhereIn("status", []interface{}{models.EventStatusPublished, models.EventStatusSaleActive}).
		OrderBy("start_time", "ASC").
		Limit(limit).
		Get(&events)

	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	return events, nil
}
