package observer

import (
	"fmt"
	"log"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
)

// Event types for the observer pattern
type EventType string

const (
	EventTypeTicketPurchased    EventType = "ticket_purchased"
	EventTypeTicketCancelled    EventType = "ticket_cancelled"
	EventTypeTicketUsed         EventType = "ticket_used"
	EventTypeWaitingListAdded   EventType = "waiting_list_added"
	EventTypeWaitingListNotify  EventType = "waiting_list_notify"
	EventTypeEventStatusChanged EventType = "event_status_changed"
	EventTypePaymentCompleted   EventType = "payment_completed"
	EventTypePaymentFailed      EventType = "payment_failed"
	EventTypeReservationExpired EventType = "reservation_expired"
)

// EventData holds data for an event
type EventData struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// Observer interface for subscribers
type Observer interface {
	Update(event *EventData) error
	GetName() string
}

// Subject (Publisher) interface
type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify(event *EventData)
}

// EventPublisher manages observers and publishes events
type EventPublisher struct {
	observers []Observer
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		observers: make([]Observer, 0),
	}
}

func (p *EventPublisher) Attach(observer Observer) {
	p.observers = append(p.observers, observer)
	log.Printf("[EventPublisher] Attached observer: %s", observer.GetName())
}

func (p *EventPublisher) Detach(observer Observer) {
	for i, obs := range p.observers {
		if obs.GetName() == observer.GetName() {
			p.observers = append(p.observers[:i], p.observers[i+1:]...)
			log.Printf("[EventPublisher] Detached observer: %s", observer.GetName())
			return
		}
	}
}

func (p *EventPublisher) Notify(event *EventData) {
	log.Printf("[EventPublisher] Notifying %d observers about event: %s", len(p.observers), event.Type)

	for _, observer := range p.observers {
		go func(obs Observer) {
			if err := obs.Update(event); err != nil {
				log.Printf("[EventPublisher] Error notifying %s: %v", obs.GetName(), err)
			}
		}(observer)
	}
}

// EmailNotificationObserver sends email notifications
type EmailNotificationObserver struct {
	EmailService EmailService
}

type EmailService interface {
	SendEmail(to, subject, body string) error
}

func NewEmailNotificationObserver(emailService EmailService) *EmailNotificationObserver {
	return &EmailNotificationObserver{
		EmailService: emailService,
	}
}

func (o *EmailNotificationObserver) Update(event *EventData) error {
	switch event.Type {
	case EventTypeTicketPurchased:
		return o.handleTicketPurchased(event)
	case EventTypeTicketCancelled:
		return o.handleTicketCancelled(event)
	case EventTypeWaitingListNotify:
		return o.handleWaitingListNotify(event)
	case EventTypePaymentCompleted:
		return o.handlePaymentCompleted(event)
	case EventTypePaymentFailed:
		return o.handlePaymentFailed(event)
	case EventTypeReservationExpired:
		return o.handleReservationExpired(event)
	}

	return nil
}

func (o *EmailNotificationObserver) GetName() string {
	return "EmailNotificationObserver"
}

func (o *EmailNotificationObserver) handleTicketPurchased(event *EventData) error {
	data, ok := event.Data.(*TicketPurchaseData)
	if !ok {
		return fmt.Errorf("invalid data type for ticket purchase event")
	}

	subject := fmt.Sprintf("Biletiniz Başarıyla Satın Alındı - %s", data.EventName)
	body := fmt.Sprintf(`
Sayın %s,

%s etkinliği için biletiniz başarıyla satın alınmıştır.

Bilet Detayları:
- Bilet No: %s
- Etkinlik: %s
- Mekan: %s
- Tarih: %s
- Koltuk: %s
- Fiyat: %.2f TL
- Doğrulama Kodu: %s

Biletinizi göstermek için QR kodunuzu kullanabilirsiniz.

İyi eğlenceler!
`, data.UserEmail, data.EventName, data.TicketNumber, data.EventName, data.VenueName, data.EventDateTime, data.SeatInfo, data.Price, data.VerificationCode)

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

func (o *EmailNotificationObserver) handleTicketCancelled(event *EventData) error {
	data, ok := event.Data.(*TicketCancellationData)
	if !ok {
		return fmt.Errorf("invalid data type for ticket cancellation event")
	}

	subject := "Biletiniz İptal Edildi"
	body := fmt.Sprintf(`
Sayın %s,

%s numaralı biletiniz iptal edilmiştir.

İade tutarı: %.2f TL
İade süresi: 3-5 iş günü

İyi günler dileriz.
`, data.UserEmail, data.TicketNumber, data.RefundAmount)

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

func (o *EmailNotificationObserver) handleWaitingListNotify(event *EventData) error {
	data, ok := event.Data.(*WaitingListNotifyData)
	if !ok {
		return fmt.Errorf("invalid data type for waiting list notify event")
	}

	subject := fmt.Sprintf("Bilet Müsaitliği - %s", data.EventName)
	body := fmt.Sprintf(`
Sayın %s,

Bekleme listesinde olduğunuz %s etkinliği için bilet açıldı!

Lütfen en kısa sürede satın alma işleminizi tamamlayın.
Bu fırsat sınırlı sayıdadır.

Etkinlik: %s
Mekan: %s
Tarih: %s

İyi eğlenceler!
`, data.UserEmail, data.EventName, data.EventName, data.VenueName, data.EventDateTime)

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

func (o *EmailNotificationObserver) handlePaymentCompleted(event *EventData) error {
	data, ok := event.Data.(*PaymentData)
	if !ok {
		return fmt.Errorf("invalid data type for payment completed event")
	}

	subject := "Ödeme Onayı"
	body := fmt.Sprintf(`
Sayın %s,

%.2f TL tutarındaki ödemeniz başarıyla alınmıştır.

İşlem No: %s
Tarih: %s

Teşekkür ederiz.
`, data.UserEmail, data.Amount, data.TransactionID, data.Timestamp.Format("02.01.2006 15:04"))

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

func (o *EmailNotificationObserver) handlePaymentFailed(event *EventData) error {
	data, ok := event.Data.(*PaymentData)
	if !ok {
		return fmt.Errorf("invalid data type for payment failed event")
	}

	subject := "Ödeme Başarısız"
	body := fmt.Sprintf(`
Sayın %s,

%.2f TL tutarındaki ödemeniz alınamadı.

Hata: %s

Lütfen bilgilerinizi kontrol edip tekrar deneyin.
`, data.UserEmail, data.Amount, data.ErrorMessage)

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

func (o *EmailNotificationObserver) handleReservationExpired(event *EventData) error {
	data, ok := event.Data.(*ReservationExpiredData)
	if !ok {
		return fmt.Errorf("invalid data type for reservation expired event")
	}

	subject := "Rezervasyon Süresi Doldu"
	body := fmt.Sprintf(`
Sayın %s,

%s etkinliği için yaptığınız rezervasyonun süresi dolmuştur.

Rezervasyon No: %s

Tekrar rezervasyon yapmak için lütfen web sitemizi ziyaret edin.
`, data.UserEmail, data.EventName, data.ReservationID)

	return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

// SMSNotificationObserver sends SMS notifications
type SMSNotificationObserver struct {
	SMSService SMSService
}

type SMSService interface {
	SendSMS(to, message string) error
}

func NewSMSNotificationObserver(smsService SMSService) *SMSNotificationObserver {
	return &SMSNotificationObserver{
		SMSService: smsService,
	}
}

func (o *SMSNotificationObserver) Update(event *EventData) error {
	switch event.Type {
	case EventTypeTicketPurchased:
		return o.handleTicketPurchased(event)
	case EventTypeWaitingListNotify:
		return o.handleWaitingListNotify(event)
	case EventTypeReservationExpired:
		return o.handleReservationExpired(event)
	}

	return nil
}

func (o *SMSNotificationObserver) GetName() string {
	return "SMSNotificationObserver"
}

func (o *SMSNotificationObserver) handleTicketPurchased(event *EventData) error {
	data, ok := event.Data.(*TicketPurchaseData)
	if !ok {
		return fmt.Errorf("invalid data type for ticket purchase event")
	}

	message := fmt.Sprintf("Biletiniz alindi. Bilet No: %s, Dogrulama: %s. %s - %s",
		data.TicketNumber, data.VerificationCode, data.EventName, data.EventDateTime)

	return o.SMSService.SendSMS(data.UserPhone, message)
}

func (o *SMSNotificationObserver) handleWaitingListNotify(event *EventData) error {
	data, ok := event.Data.(*WaitingListNotifyData)
	if !ok {
		return fmt.Errorf("invalid data type for waiting list notify event")
	}

	message := fmt.Sprintf("%s etkinligi icin bilet acildi! Hemen satin al.", data.EventName)

	return o.SMSService.SendSMS(data.UserPhone, message)
}

func (o *SMSNotificationObserver) handleReservationExpired(event *EventData) error {
	data, ok := event.Data.(*ReservationExpiredData)
	if !ok {
		return fmt.Errorf("invalid data type for reservation expired event")
	}

	message := fmt.Sprintf("Rezervasyonunuz iptal edildi: %s", data.EventName)

	return o.SMSService.SendSMS(data.UserPhone, message)
}

// AnalyticsObserver tracks events for analytics
type AnalyticsObserver struct {
	AnalyticsService AnalyticsService
}

type AnalyticsService interface {
	TrackEvent(eventType, userID string, properties map[string]interface{}) error
}

func NewAnalyticsObserver(analyticsService AnalyticsService) *AnalyticsObserver {
	return &AnalyticsObserver{
		AnalyticsService: analyticsService,
	}
}

func (o *AnalyticsObserver) Update(event *EventData) error {
	properties := map[string]interface{}{
		"timestamp":  event.Timestamp,
		"event_type": event.Type,
	}

	// Extract user ID based on event type
	var userID string
	switch event.Type {
	case EventTypeTicketPurchased:
		if data, ok := event.Data.(*TicketPurchaseData); ok {
			userID = fmt.Sprintf("%d", data.UserID)
			properties["event_id"] = data.EventID
			properties["ticket_number"] = data.TicketNumber
			properties["price"] = data.Price
		}
	case EventTypePaymentCompleted:
		if data, ok := event.Data.(*PaymentData); ok {
			userID = fmt.Sprintf("%d", data.UserID)
			properties["amount"] = data.Amount
			properties["transaction_id"] = data.TransactionID
		}
	}

	return o.AnalyticsService.TrackEvent(string(event.Type), userID, properties)
}

func (o *AnalyticsObserver) GetName() string {
	return "AnalyticsObserver"
}

// Event Data Structures
type TicketPurchaseData struct {
	UserID           int64
	UserEmail        string
	UserPhone        string
	EventID          int64
	EventName        string
	VenueName        string
	EventDateTime    string
	TicketNumber     string
	VerificationCode string
	SeatInfo         string
	Price            float64
}

type TicketCancellationData struct {
	UserEmail    string
	TicketNumber string
	RefundAmount float64
}

type WaitingListNotifyData struct {
	UserEmail     string
	UserPhone     string
	EventName     string
	VenueName     string
	EventDateTime string
}

type PaymentData struct {
	UserID        int64
	UserEmail     string
	Amount        float64
	TransactionID string
	Timestamp     time.Time
	ErrorMessage  string
}

type ReservationExpiredData struct {
	UserEmail     string
	UserPhone     string
	EventName     string
	ReservationID string
}
