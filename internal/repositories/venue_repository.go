package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
)

type VenueRepository struct {
	db *sql.DB
}

func NewVenueRepository(db *sql.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

func (r *VenueRepository) Create(venue *models.Venue) (int64, error) {
	query := `
		INSERT INTO venues (name, address, city, country, capacity, latitude, longitude, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		venue.Name, venue.Address, venue.City, venue.Country, venue.Capacity,
		venue.Latitude, venue.Longitude, venue.CreatedAt, venue.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create venue: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *VenueRepository) FindByID(id int64) (*models.Venue, error) {
	query := `
		SELECT id, name, address, city, country, capacity, latitude, longitude, created_at, updated_at, deleted_at
		FROM venues
		WHERE id = ? AND deleted_at IS NULL
	`

	venue := &models.Venue{}
	err := r.db.QueryRow(query, id).Scan(
		&venue.ID, &venue.Name, &venue.Address, &venue.City, &venue.Country,
		&venue.Capacity, &venue.Latitude, &venue.Longitude, &venue.CreatedAt,
		&venue.UpdatedAt, &venue.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("venue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find venue: %w", err)
	}

	return venue, nil
}

func (r *VenueRepository) FindAll(limit, offset int) ([]*models.Venue, error) {
	query := `
		SELECT id, name, address, city, country, capacity, latitude, longitude, created_at, updated_at
		FROM venues
		WHERE deleted_at IS NULL
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query venues: %w", err)
	}
	defer rows.Close()

	venues := []*models.Venue{}
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Address, &venue.City, &venue.Country,
			&venue.Capacity, &venue.Latitude, &venue.Longitude, &venue.CreatedAt,
			&venue.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan venue: %w", err)
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

func (r *VenueRepository) Update(venue *models.Venue) error {
	query := `
		UPDATE venues
		SET name = ?, address = ?, city = ?, country = ?, capacity = ?,
			latitude = ?, longitude = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	venue.UpdatedAt = time.Now()

	result, err := r.db.Exec(query,
		venue.Name, venue.Address, venue.City, venue.Country, venue.Capacity,
		venue.Latitude, venue.Longitude, venue.UpdatedAt, venue.ID,
	)

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

func (r *VenueRepository) Delete(id int64) error {
	query := `
		UPDATE venues
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, time.Now(), id)
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

func (r *VenueRepository) SearchByName(keyword string) ([]*models.Venue, error) {
	query := `
		SELECT id, name, address, city, country, capacity, latitude, longitude, created_at, updated_at
		FROM venues
		WHERE deleted_at IS NULL AND name LIKE ?
		ORDER BY name ASC
	`

	searchTerm := "%" + keyword + "%"
	rows, err := r.db.Query(query, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues: %w", err)
	}
	defer rows.Close()

	venues := []*models.Venue{}
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Address, &venue.City, &venue.Country,
			&venue.Capacity, &venue.Latitude, &venue.Longitude, &venue.CreatedAt,
			&venue.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan venue: %w", err)
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

func (r *VenueRepository) FindByCity(city string) ([]*models.Venue, error) {
	query := `
		SELECT id, name, address, city, country, capacity, latitude, longitude, created_at, updated_at
		FROM venues
		WHERE deleted_at IS NULL AND city = ?
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, city)
	if err != nil {
		return nil, fmt.Errorf("failed to query venues by city: %w", err)
	}
	defer rows.Close()

	venues := []*models.Venue{}
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Address, &venue.City, &venue.Country,
			&venue.Capacity, &venue.Latitude, &venue.Longitude, &venue.CreatedAt,
			&venue.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan venue: %w", err)
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// Section Repository Methods
func (r *VenueRepository) CreateSection(section *models.Section) (int64, error) {
	query := `
		INSERT INTO sections (venue_id, name, row_count, seats_per_row, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		section.VenueID, section.Name, section.RowCount, section.SeatsPerRow,
		section.CreatedAt, section.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create section: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *VenueRepository) FindSectionsByVenueID(venueID int64) ([]*models.Section, error) {
	query := `
		SELECT id, venue_id, name, row_count, seats_per_row, created_at, updated_at
		FROM sections
		WHERE venue_id = ? AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, venueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sections: %w", err)
	}
	defer rows.Close()

	sections := []*models.Section{}
	for rows.Next() {
		section := &models.Section{}
		err := rows.Scan(
			&section.ID, &section.VenueID, &section.Name, &section.RowCount,
			&section.SeatsPerRow, &section.CreatedAt, &section.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func (r *VenueRepository) FindSectionByID(id int64) (*models.Section, error) {
	query := `
		SELECT id, venue_id, name, row_count, seats_per_row, created_at, updated_at
		FROM sections
		WHERE id = ? AND deleted_at IS NULL
	`

	section := &models.Section{}
	err := r.db.QueryRow(query, id).Scan(
		&section.ID, &section.VenueID, &section.Name, &section.RowCount,
		&section.SeatsPerRow, &section.CreatedAt, &section.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("section not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find section: %w", err)
	}

	return section, nil
}

// Seat Repository Methods
func (r *VenueRepository) CreateSeat(seat *models.Seat) (int64, error) {
	query := `
		INSERT INTO seats (section_id, row, number, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		seat.SectionID, seat.Row, seat.Number, seat.IsActive,
		seat.CreatedAt, seat.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create seat: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *VenueRepository) FindSeatsBySectionID(sectionID int64) ([]*models.Seat, error) {
	query := `
		SELECT id, section_id, row, number, is_active, created_at, updated_at
		FROM seats
		WHERE section_id = ? AND deleted_at IS NULL
		ORDER BY row, number
	`

	rows, err := r.db.Query(query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query seats: %w", err)
	}
	defer rows.Close()

	seats := []*models.Seat{}
	for rows.Next() {
		seat := &models.Seat{}
		err := rows.Scan(
			&seat.ID, &seat.SectionID, &seat.Row, &seat.Number, &seat.IsActive,
			&seat.CreatedAt, &seat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

func (r *VenueRepository) FindSeatByID(id int64) (*models.Seat, error) {
	query := `
		SELECT id, section_id, row, number, is_active, created_at, updated_at
		FROM seats
		WHERE id = ? AND deleted_at IS NULL
	`

	seat := &models.Seat{}
	err := r.db.QueryRow(query, id).Scan(
		&seat.ID, &seat.SectionID, &seat.Row, &seat.Number, &seat.IsActive,
		&seat.CreatedAt, &seat.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("seat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find seat: %w", err)
	}

	return seat, nil
}

func (r *VenueRepository) UpdateSeatStatus(id int64, isActive bool) error {
	query := `
		UPDATE seats
		SET is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, isActive, time.Now(), id)
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
