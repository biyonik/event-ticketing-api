package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/pkg/database"
)

type VenueRepository struct {
	db      *sql.DB
	grammar database.Grammar
}

func NewVenueRepository(db *sql.DB) *VenueRepository {
	return &VenueRepository{
		db:      db,
		grammar: database.NewMySQLGrammar(),
	}
}

// Create - Conduit-Go Builder ile venue oluşturma
func (r *VenueRepository) Create(venue *models.Venue) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		ExecInsert(map[string]interface{}{
			"name":       venue.Name,
			"address":    venue.Address,
			"city":       venue.City,
			"country":    venue.Country,
			"capacity":   venue.Capacity,
			"latitude":   venue.Latitude,
			"longitude":  venue.Longitude,
			"created_at": venue.CreatedAt,
			"updated_at": venue.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create venue: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindByID - Conduit-Go Builder ile single venue
func (r *VenueRepository) FindByID(id int64) (*models.Venue, error) {
	var venue models.Venue

	err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		Where("id", "=", id).
		WhereNull("deleted_at").
		First(&venue)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("venue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find venue: %w", err)
	}

	return &venue, nil
}

// FindAll - Conduit-Go Builder ile tüm venues
func (r *VenueRepository) FindAll(limit, offset int) ([]*models.Venue, error) {
	var venues []*models.Venue

	err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		WhereNull("deleted_at").
		OrderBy("name", "ASC").
		Limit(limit).
		Offset(offset).
		Get(&venues)

	if err != nil {
		return nil, fmt.Errorf("failed to query venues: %w", err)
	}

	return venues, nil
}

// Update - Conduit-Go Builder ile venue güncelleme
func (r *VenueRepository) Update(venue *models.Venue) error {
	venue.UpdatedAt = time.Now()

	result, err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		Where("id", "=", venue.ID).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"name":       venue.Name,
			"address":    venue.Address,
			"city":       venue.City,
			"country":    venue.Country,
			"capacity":   venue.Capacity,
			"latitude":   venue.Latitude,
			"longitude":  venue.Longitude,
			"updated_at": venue.UpdatedAt,
		})

	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

// Delete - Soft delete using Conduit-Go Builder
func (r *VenueRepository) Delete(id int64) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		Where("id", "=", id).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"deleted_at": time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to delete venue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

// SearchByName - Builder ile LIKE search
func (r *VenueRepository) SearchByName(keyword string) ([]*models.Venue, error) {
	var venues []*models.Venue

	searchTerm := "%" + keyword + "%"

	err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		WhereNull("deleted_at").
		Where("name", "LIKE", searchTerm).
		OrderBy("name", "ASC").
		Get(&venues)

	if err != nil {
		return nil, fmt.Errorf("failed to search venues: %w", err)
	}

	return venues, nil
}

// FindByCity - Builder ile city filter
func (r *VenueRepository) FindByCity(city string) ([]*models.Venue, error) {
	var venues []*models.Venue

	err := database.NewBuilder(r.db, r.grammar).
		Table("venues").
		WhereNull("deleted_at").
		Where("city", "=", city).
		OrderBy("name", "ASC").
		Get(&venues)

	if err != nil {
		return nil, fmt.Errorf("failed to query venues by city: %w", err)
	}

	return venues, nil
}

// Section Repository Methods

// CreateSection - Conduit-Go Builder ile section oluşturma
func (r *VenueRepository) CreateSection(section *models.Section) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("sections").
		ExecInsert(map[string]interface{}{
			"venue_id":      section.VenueID,
			"name":          section.Name,
			"row_count":     section.RowCount,
			"seats_per_row": section.SeatsPerRow,
			"created_at":    section.CreatedAt,
			"updated_at":    section.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create section: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindSectionsByVenueID - Builder ile venue sections
func (r *VenueRepository) FindSectionsByVenueID(venueID int64) ([]*models.Section, error) {
	var sections []*models.Section

	err := database.NewBuilder(r.db, r.grammar).
		Table("sections").
		Where("venue_id", "=", venueID).
		WhereNull("deleted_at").
		OrderBy("name", "ASC").
		Get(&sections)

	if err != nil {
		return nil, fmt.Errorf("failed to query sections: %w", err)
	}

	return sections, nil
}

// FindSectionByID - Builder ile single section
func (r *VenueRepository) FindSectionByID(id int64) (*models.Section, error) {
	var section models.Section

	err := database.NewBuilder(r.db, r.grammar).
		Table("sections").
		Where("id", "=", id).
		WhereNull("deleted_at").
		First(&section)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("section not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find section: %w", err)
	}

	return &section, nil
}

// Seat Repository Methods

// CreateSeat - Conduit-Go Builder ile seat oluşturma
func (r *VenueRepository) CreateSeat(seat *models.Seat) (int64, error) {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("seats").
		ExecInsert(map[string]interface{}{
			"section_id": seat.SectionID,
			"row":        seat.Row,
			"number":     seat.Number,
			"is_active":  seat.IsActive,
			"created_at": seat.CreatedAt,
			"updated_at": seat.UpdatedAt,
		})

	if err != nil {
		return 0, fmt.Errorf("failed to create seat: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// FindSeatsBySectionID - Builder ile section seats
func (r *VenueRepository) FindSeatsBySectionID(sectionID int64) ([]*models.Seat, error) {
	var seats []*models.Seat

	err := database.NewBuilder(r.db, r.grammar).
		Table("seats").
		Where("section_id", "=", sectionID).
		WhereNull("deleted_at").
		OrderBy("row", "ASC").
		OrderBy("number", "ASC").
		Get(&seats)

	if err != nil {
		return nil, fmt.Errorf("failed to query seats: %w", err)
	}

	return seats, nil
}

// FindSeatByID - Builder ile single seat
func (r *VenueRepository) FindSeatByID(id int64) (*models.Seat, error) {
	var seat models.Seat

	err := database.NewBuilder(r.db, r.grammar).
		Table("seats").
		Where("id", "=", id).
		WhereNull("deleted_at").
		First(&seat)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("seat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find seat: %w", err)
	}

	return &seat, nil
}

// UpdateSeatStatus - Builder ile seat status güncelleme
func (r *VenueRepository) UpdateSeatStatus(id int64, isActive bool) error {
	result, err := database.NewBuilder(r.db, r.grammar).
		Table("seats").
		Where("id", "=", id).
		WhereNull("deleted_at").
		ExecUpdate(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to update seat status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("seat not found")
	}

	return nil
}

// GetVenueWithSections - Builder ile nested loading
func (r *VenueRepository) GetVenueWithSections(venueID int64) (*models.Venue, error) {
	venue, err := r.FindByID(venueID)
	if err != nil {
		return nil, err
	}

	sections, err := r.FindSectionsByVenueID(venueID)
	if err != nil {
		return nil, err
	}

	// Load seats for each section
	for _, section := range sections {
		seats, err := r.FindSeatsBySectionID(section.ID)
		if err != nil {
			return nil, err
		}
		section.Seats = seats
	}

	venue.Sections = sections
	return venue, nil
}
