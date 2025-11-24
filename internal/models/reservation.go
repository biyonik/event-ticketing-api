// -----------------------------------------------------------------------------
// Reservation & Payment Models
// -----------------------------------------------------------------------------
// Rezervasyon ve ödeme işlemlerini temsil eder
// -----------------------------------------------------------------------------

package models

import (
	"time"
)

// PaymentStatus, ödeme durumunu temsil eder
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Payment, bir ödeme işlemini temsil eder
type Payment struct {
	BaseModel
	TicketID      int64         `json:"ticket_id" db:"ticket_id"`
	UserID        int64         `json:"user_id" db:"user_id"`
	Amount        float64       `json:"amount" db:"amount"`
	Currency      string        `json:"currency" db:"currency"` // "TRY", "USD", "EUR"
	Status        PaymentStatus `json:"status" db:"status"`
	PaymentMethod string        `json:"payment_method" db:"payment_method"` // "credit_card", "debit_card", "paypal"
	TransactionID string        `json:"transaction_id,omitempty" db:"transaction_id"`
	PaidAt        *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
	RefundedAt    *time.Time    `json:"refunded_at,omitempty" db:"refunded_at"`

	// İlişkili veriler
	Ticket *Ticket `json:"ticket,omitempty" db:"-"`
	User   *User   `json:"user,omitempty" db:"-"`
}

// IsCompleted, ödemenin tamamlanıp tamamlanmadığını kontrol eder
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

// CanRefund, ödemenin iade edilip edilemeyeceğini kontrol eder
func (p *Payment) CanRefund() bool {
	return p.Status == PaymentStatusCompleted
}

// WaitingList, bekleme listesi kayıtlarını temsil eder
type WaitingList struct {
	BaseModel
	EventID     int64      `json:"event_id" db:"event_id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	SectionID   *int64     `json:"section_id,omitempty" db:"section_id"` // İsteğe bağlı bölüm tercihi
	IsNotified  bool       `json:"is_notified" db:"is_notified"`
	NotifiedAt  *time.Time `json:"notified_at,omitempty" db:"notified_at"`
	Position    int        `json:"position" db:"position"` // Sıradaki pozisyon

	// İlişkili veriler
	Event *Event `json:"event,omitempty" db:"-"`
	User  *User  `json:"user,omitempty" db:"-"`
}
