// -----------------------------------------------------------------------------
// Venue Model
// -----------------------------------------------------------------------------
// Etkinlik mekanlarını temsil eder (stadyum, konser salonu, tiyatro vb.)
// Section'lar (bölümler) ve seat'ler (koltuklar) içerir
// -----------------------------------------------------------------------------

package models

import (
	"time"
)

// Venue, bir etkinlik mekanını temsil eder
type Venue struct {
	BaseModel
	Name        string     `json:"name" db:"name"`
	Address     string     `json:"address" db:"address"`
	City        string     `json:"city" db:"city"`
	Country     string     `json:"country" db:"country"`
	Capacity    int        `json:"capacity" db:"capacity"`
	Description string     `json:"description,omitempty" db:"description"`
	ImageURL    string     `json:"image_url,omitempty" db:"image_url"`
	Latitude    float64    `json:"latitude" db:"latitude"`
	Longitude   float64    `json:"longitude" db:"longitude"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`

	// İlişkili veriler
	Sections []Section `json:"sections,omitempty" db:"-"`
}

// Section, mekan içindeki bir bölümü temsil eder (VIP, Normal, Balkon vb.)
type Section struct {
	BaseModel
	VenueID     int64      `json:"venue_id" db:"venue_id"`
	Name        string     `json:"name" db:"name"` // "VIP", "Tribune A", "Balkon 1"
	Description string     `json:"description,omitempty" db:"description"`
	Capacity    int        `json:"capacity" db:"capacity"`
	RowCount    int        `json:"row_count" db:"row_count"`       // Kaç sıra var
	SeatsPerRow int        `json:"seats_per_row" db:"seats_per_row"` // Sıra başına koltuk
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`

	// İlişkili veriler
	Venue *Venue `json:"venue,omitempty" db:"-"`
	Seats []Seat `json:"seats,omitempty" db:"-"`
}

// Seat, bir koltuğu temsil eder
type Seat struct {
	BaseModel
	SectionID int64  `json:"section_id" db:"section_id"`
	Row       string `json:"row" db:"row"`        // "A", "B", "C" veya "1", "2", "3"
	Number    string `json:"number" db:"number"`  // "1", "2", "3" veya "A1", "B2"
	IsActive  bool   `json:"is_active" db:"is_active"` // Koltuk kullanılabilir mi?

	// İlişkili veriler
	Section *Section `json:"section,omitempty" db:"-"`
}

// GetFullName, koltuğun tam adını döndürür (örn: "VIP - A12")
func (s *Seat) GetFullName() string {
	return s.Row + s.Number
}

// IsAvailable, koltuğun müsait olup olmadığını kontrol eder
func (s *Seat) IsAvailable() bool {
	return s.IsActive
}
