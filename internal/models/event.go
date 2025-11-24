// -----------------------------------------------------------------------------
// Event Model
// -----------------------------------------------------------------------------
// Etkinlikleri temsil eder (konser, tiyatro, spor maçı vb.)
// -----------------------------------------------------------------------------

package models

import (
	"time"
)

// EventType, etkinlik tipini temsil eder
type EventType string

const (
	EventTypeConcert   EventType = "concert"
	EventTypeTheater   EventType = "theater"
	EventTypeSports    EventType = "sports"
	EventTypeConference EventType = "conference"
	EventTypeFestival  EventType = "festival"
)

// EventStatus, etkinlik durumunu temsil eder
type EventStatus string

const (
	EventStatusDraft      EventStatus = "draft"
	EventStatusPublished  EventStatus = "published"
	EventStatusSaleActive EventStatus = "sale_active"
	EventStatusSoldOut    EventStatus = "sold_out"
	EventStatusCancelled  EventStatus = "cancelled"
	EventStatusCompleted  EventStatus = "completed"
)

// Event, bir etkinliği temsil eder
type Event struct {
	BaseModel
	Name            string      `json:"name" db:"name"`
	Description     string      `json:"description" db:"description"`
	Type            EventType   `json:"type" db:"type"`
	Status          EventStatus `json:"status" db:"status"`
	VenueID         int64       `json:"venue_id" db:"venue_id"`
	StartTime       time.Time   `json:"start_time" db:"start_time"`
	EndTime         time.Time   `json:"end_time" db:"end_time"`
	ImageURL        string      `json:"image_url,omitempty" db:"image_url"`
	TotalCapacity   int         `json:"total_capacity" db:"total_capacity"`
	AvailableSeats  int         `json:"available_seats" db:"available_seats"`
	BasePr ice       float64     `json:"base_price" db:"base_price"`
	OrganizerId     int64       `json:"organizer_id" db:"organizer_id"`
	IsFeatured      bool        `json:"is_featured" db:"is_featured"`
	SaleStartTime   *time.Time  `json:"sale_start_time,omitempty" db:"sale_start_time"`
	SaleEndTime     *time.Time  `json:"sale_end_time,omitempty" db:"sale_end_time"`
	DeletedAt       *time.Time  `json:"-" db:"deleted_at"`

	// İlişkili veriler
	Venue     *Venue `json:"venue,omitempty" db:"-"`
	Organizer *User  `json:"organizer,omitempty" db:"-"`
}

// IsSaleActive, bilet satışının aktif olup olmadığını kontrol eder
func (e *Event) IsSaleActive() bool {
	now := time.Now()

	if e.Status != EventStatusSaleActive {
		return false
	}

	if e.SaleStartTime != nil && now.Before(*e.SaleStartTime) {
		return false
	}

	if e.SaleEndTime != nil && now.After(*e.SaleEndTime) {
		return false
	}

	return e.AvailableSeats > 0
}

// IsSoldOut, etkinliğin tükenip tükenmediğini kontrol eder
func (e *Event) IsSoldOut() bool {
	return e.AvailableSeats <= 0
}

// CanCancel, etkinliğin iptal edilip edilemeyeceğini kontrol eder
func (e *Event) CanCancel() bool {
	// Etkinlik henüz başlamadıysa iptal edilebilir
	return time.Now().Before(e.StartTime)
}

// IsUpcoming, etkinliğin yaklaşan olup olmadığını kontrol eder
func (e *Event) IsUpcoming() bool {
	return time.Now().Before(e.StartTime)
}

// IsCompleted, etkinliğin tamamlanıp tamamlanmadığını kontrol eder
func (e *Event) IsCompleted() bool {
	return time.Now().After(e.EndTime)
}

// GetOccupancyRate, doluluk oranını hesaplar (%)
func (e *Event) GetOccupancyRate() float64 {
	if e.TotalCapacity == 0 {
		return 0
	}
	soldSeats := e.TotalCapacity - e.AvailableSeats
	return (float64(soldSeats) / float64(e.TotalCapacity)) * 100
}
