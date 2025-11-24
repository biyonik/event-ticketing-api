// -----------------------------------------------------------------------------
// Ticket Model
// -----------------------------------------------------------------------------
// Biletleri temsil eder. State Pattern kullanır.
// States: Reserved, Sold, Used, Cancelled, Expired
// -----------------------------------------------------------------------------

package models

import (
	"time"
)

// TicketStatus, bilet durumunu temsil eder (State Pattern)
type TicketStatus string

const (
	TicketStatusReserved  TicketStatus = "reserved"  // Rezerve edilmiş (15 dk)
	TicketStatusSold      TicketStatus = "sold"      // Satılmış
	TicketStatusUsed      TicketStatus = "used"      // Kullanılmış (check-in yapıldı)
	TicketStatusCancelled TicketStatus = "cancelled" // İptal edilmiş
	TicketStatusExpired   TicketStatus = "expired"   // Süresi dolmuş rezervasyon
)

// Ticket, bir bileti temsil eder
type Ticket struct {
	BaseModel
	TicketCode     string       `json:"ticket_code" db:"ticket_code"` // Unique ticket code
	EventID        int64        `json:"event_id" db:"event_id"`
	SeatID         int64        `json:"seat_id" db:"seat_id"`
	UserID         int64        `json:"user_id" db:"user_id"`
	Status         TicketStatus `json:"status" db:"status"`
	Price          float64      `json:"price" db:"price"`
	PricingType    string       `json:"pricing_type" db:"pricing_type"` // "early_bird", "vip", "dynamic"
	QRCodeURL      string       `json:"qr_code_url,omitempty" db:"qr_code_url"`
	ReservedAt     *time.Time   `json:"reserved_at,omitempty" db:"reserved_at"`
	PurchasedAt    *time.Time   `json:"purchased_at,omitempty" db:"purchased_at"`
	CancelledAt    *time.Time   `json:"cancelled_at,omitempty" db:"cancelled_at"`
	UsedAt         *time.Time   `json:"used_at,omitempty" db:"used_at"`
	ReservationExpiry *time.Time `json:"reservation_expiry,omitempty" db:"reservation_expiry"`
	DeletedAt      *time.Time   `json:"-" db:"deleted_at"`

	// İlişkili veriler
	Event *Event `json:"event,omitempty" db:"-"`
	Seat  *Seat  `json:"seat,omitempty" db:"-"`
	User  *User  `json:"user,omitempty" db:"-"`
}

// State Pattern Methods

// CanPurchase, biletin satın alınıp alınamayacağını kontrol eder
func (t *Ticket) CanPurchase() bool {
	return t.Status == TicketStatusReserved && time.Now().Before(*t.ReservationExpiry)
}

// CanCancel, biletin iptal edilip edilemeyeceğini kontrol eder
func (t *Ticket) CanCancel() bool {
	return t.Status == TicketStatusReserved || t.Status == TicketStatusSold
}

// CanUse, biletin kullanılıp kullanılamayacağını kontrol eder (check-in)
func (t *Ticket) CanUse() bool {
	return t.Status == TicketStatusSold
}

// IsExpired, rezervasyonun süresinin dolup dolmadığını kontrol eder
func (t *Ticket) IsExpired() bool {
	if t.Status != TicketStatusReserved {
		return false
	}
	if t.ReservationExpiry == nil {
		return false
	}
	return time.Now().After(*t.ReservationExpiry)
}

// MarkAsSold, bileti satılmış olarak işaretler (State transition)
func (t *Ticket) MarkAsSold() error {
	if !t.CanPurchase() {
		return ErrInvalidStateTransition
	}
	t.Status = TicketStatusSold
	now := time.Now()
	t.PurchasedAt = &now
	return nil
}

// MarkAsUsed, bileti kullanılmış olarak işaretler (Check-in)
func (t *Ticket) MarkAsUsed() error {
	if !t.CanUse() {
		return ErrInvalidStateTransition
	}
	t.Status = TicketStatusUsed
	now := time.Now()
	t.UsedAt = &now
	return nil
}

// MarkAsCancelled, bileti iptal edilmiş olarak işaretler
func (t *Ticket) MarkAsCancelled() error {
	if !t.CanCancel() {
		return ErrInvalidStateTransition
	}
	t.Status = TicketStatusCancelled
	now := time.Now()
	t.CancelledAt = &now
	return nil
}

// MarkAsExpired, rezervasyonu süresi dolmuş olarak işaretler
func (t *Ticket) MarkAsExpired() error {
	if t.Status != TicketStatusReserved {
		return ErrInvalidStateTransition
	}
	t.Status = TicketStatusExpired
	return nil
}

// GetReservationTimeLeft, rezervasyonun kalan süresini döndürür
func (t *Ticket) GetReservationTimeLeft() time.Duration {
	if t.Status != TicketStatusReserved || t.ReservationExpiry == nil {
		return 0
	}
	remaining := time.Until(*t.ReservationExpiry)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Errors
var (
	ErrInvalidStateTransition = &AppError{Code: "INVALID_STATE_TRANSITION", Message: "Geçersiz durum geçişi"}
)

// AppError, uygulama hata yapısı
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
