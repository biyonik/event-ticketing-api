package services

import (
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/internal/patterns/observer"
	"github.com/biyonik/event-ticketing-api/internal/repositories"
	v "github.com/biyonik/event-ticketing-api/pkg/validation"
	"github.com/biyonik/event-ticketing-api/pkg/validation/types"
)

type ReservationService struct {
	reservationRepo *repositories.ReservationRepository
	eventRepo       *repositories.EventRepository
	ticketRepo      *repositories.TicketRepository
	eventPublisher  *observer.EventPublisher
}

func NewReservationService(
	reservationRepo *repositories.ReservationRepository,
	eventRepo *repositories.EventRepository,
	ticketRepo *repositories.TicketRepository,
	eventPublisher *observer.EventPublisher,
) *ReservationService {
	return &ReservationService{
		reservationRepo: reservationRepo,
		eventRepo:       eventRepo,
		ticketRepo:      ticketRepo,
		eventPublisher:  eventPublisher,
	}
}

// CreatePayment creates a payment record
func (s *ReservationService) CreatePayment(
	userID, eventID int64,
	amount float64,
	currency string,
	paymentMethod models.PaymentMethod,
	transactionID string,
) (*models.Payment, error) {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"user_id": types.Number().
			Required().
			Min(1).
			Label("Kullanıcı ID"),
		"event_id": types.Number().
			Required().
			Min(1).
			Label("Etkinlik ID"),
		"amount": types.Number().
			Required().
			Min(0.01).
			Label("Ödeme Tutarı"),
		"currency": types.String().
			Min(3).
			Max(3).
			Label("Para Birimi"),
		"transaction_id": types.String().
			Required().
			Min(10).
			Max(100).
			Label("İşlem ID"),
	})

	// Default currency
	if currency == "" {
		currency = "TRY"
	}

	rawData := map[string]any{
		"user_id":        float64(userID),
		"event_id":       float64(eventID),
		"amount":         amount,
		"currency":       currency,
		"transaction_id": transactionID,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return nil, fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Create payment
	payment := &models.Payment{
		UserID:        userID,
		EventID:       eventID,
		Amount:        amount,
		Currency:      currency,
		Status:        models.PaymentStatusPending,
		PaymentMethod: paymentMethod,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	paymentID, err := s.reservationRepo.CreatePayment(payment)
	if err != nil {
		return nil, fmt.Errorf("ödeme oluşturulamadı: %w", err)
	}

	payment.ID = paymentID
	return payment, nil
}

// ProcessPayment processes a payment (simulated payment gateway)
func (s *ReservationService) ProcessPayment(paymentID int64, userEmail string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"payment_id": types.Number().
			Required().
			Min(1).
			Label("Ödeme ID"),
		"user_email": types.String().
			Required().
			Email().
			Label("E-posta"),
	})

	rawData := map[string]any{
		"payment_id": float64(paymentID),
		"user_email": userEmail,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Get payment
	payment, err := s.reservationRepo.FindPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("ödeme bulunamadı: %w", err)
	}

	// 3. Business rules
	if payment.Status != models.PaymentStatusPending {
		return fmt.Errorf("ödeme zaten işlendi")
	}

	// 4. Simulate payment processing (in real app, call payment gateway)
	// For demo, we'll just mark as completed
	providerResponse := fmt.Sprintf("Payment processed successfully. Amount: %.2f %s", payment.Amount, payment.Currency)

	if err := s.reservationRepo.UpdatePaymentStatus(paymentID, models.PaymentStatusCompleted, providerResponse); err != nil {
		return fmt.Errorf("ödeme durumu güncellenemedi: %w", err)
	}

	// 5. Notify observers
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypePaymentCompleted,
		Timestamp: time.Now(),
		Data: &observer.PaymentData{
			UserID:        payment.UserID,
			UserEmail:     userEmail,
			Amount:        payment.Amount,
			TransactionID: payment.TransactionID,
			Timestamp:     time.Now(),
		},
	})

	return nil
}

// FailPayment marks a payment as failed
func (s *ReservationService) FailPayment(paymentID int64, userEmail, errorMessage string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"payment_id": types.Number().
			Required().
			Min(1).
			Label("Ödeme ID"),
		"user_email": types.String().
			Required().
			Email().
			Label("E-posta"),
		"error_message": types.String().
			Required().
			Min(3).
			Max(500).
			Label("Hata Mesajı"),
	})

	rawData := map[string]any{
		"payment_id":    float64(paymentID),
		"user_email":    userEmail,
		"error_message": errorMessage,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Get payment
	payment, err := s.reservationRepo.FindPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("ödeme bulunamadı: %w", err)
	}

	// 3. Update status
	providerResponse := fmt.Sprintf("Payment failed: %s", errorMessage)
	if err := s.reservationRepo.UpdatePaymentStatus(paymentID, models.PaymentStatusFailed, providerResponse); err != nil {
		return fmt.Errorf("ödeme durumu güncellenemedi: %w", err)
	}

	// 4. Notify observers
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypePaymentFailed,
		Timestamp: time.Now(),
		Data: &observer.PaymentData{
			UserID:       payment.UserID,
			UserEmail:    userEmail,
			Amount:       payment.Amount,
			Timestamp:    time.Now(),
			ErrorMessage: errorMessage,
		},
	})

	return nil
}

// RefundPayment processes a refund
func (s *ReservationService) RefundPayment(paymentID int64, userEmail string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"payment_id": types.Number().
			Required().
			Min(1).
			Label("Ödeme ID"),
		"user_email": types.String().
			Required().
			Email().
			Label("E-posta"),
	})

	rawData := map[string]any{
		"payment_id": float64(paymentID),
		"user_email": userEmail,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Get payment
	payment, err := s.reservationRepo.FindPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("ödeme bulunamadı: %w", err)
	}

	// 3. Business rules
	if payment.Status != models.PaymentStatusCompleted {
		return fmt.Errorf("sadece tamamlanmış ödemeler iade edilebilir")
	}

	// 4. Process refund (in real app, call payment gateway)
	providerResponse := fmt.Sprintf("Refund processed. Amount: %.2f %s", payment.Amount, payment.Currency)

	if err := s.reservationRepo.UpdatePaymentStatus(paymentID, models.PaymentStatusRefunded, providerResponse); err != nil {
		return fmt.Errorf("iade işlemi yapılamadı: %w", err)
	}

	return nil
}

// GetUserPayments retrieves all payments for a user
func (s *ReservationService) GetUserPayments(userID int64) ([]*models.Payment, error) {
	payments, err := s.reservationRepo.FindPaymentsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("ödemeler getirilemedi: %w", err)
	}

	return payments, nil
}

// AddToWaitingList adds a user to the waiting list for a sold-out event
func (s *ReservationService) AddToWaitingList(userID, eventID int64, priority int) (*models.WaitingList, error) {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"user_id": types.Number().
			Required().
			Min(1).
			Label("Kullanıcı ID"),
		"event_id": types.Number().
			Required().
			Min(1).
			Label("Etkinlik ID"),
		"priority": types.Number().
			Required().
			Min(0).
			Max(10).
			Label("Öncelik"),
	})

	rawData := map[string]any{
		"user_id":  float64(userID),
		"event_id": float64(eventID),
		"priority": float64(priority),
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return nil, fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Verify event exists and is sold out
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	if event.Status != models.EventStatusSoldOut {
		return nil, fmt.Errorf("etkinlik tükenmedi, bekleme listesine eklenemez")
	}

	// 3. Check if user already in waiting list
	isInList, err := s.reservationRepo.IsUserInWaitingList(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("bekleme listesi kontrolü yapılamadı: %w", err)
	}

	if isInList {
		return nil, fmt.Errorf("kullanıcı zaten bekleme listesinde")
	}

	// 4. Add to waiting list
	waitingList := &models.WaitingList{
		EventID:   eventID,
		UserID:    userID,
		Status:    models.WaitingListStatusWaiting,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := s.reservationRepo.AddToWaitingList(waitingList)
	if err != nil {
		return nil, fmt.Errorf("bekleme listesine eklenemedi: %w", err)
	}

	waitingList.ID = id

	// 5. Notify observers
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypeWaitingListAdded,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id":  userID,
			"event_id": eventID,
		},
	})

	return waitingList, nil
}

// RemoveFromWaitingList removes a user from the waiting list
func (s *ReservationService) RemoveFromWaitingList(waitingListID int64) error {
	if err := s.reservationRepo.RemoveFromWaitingList(waitingListID); err != nil {
		return fmt.Errorf("bekleme listesinden kaldırılamadı: %w", err)
	}

	return nil
}

// NotifyWaitingList notifies users in the waiting list when tickets become available
func (s *ReservationService) NotifyWaitingList(eventID int64, availableSeats int) error {
	// 1. Get event
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Get waiting list (prioritized)
	waitingList, err := s.reservationRepo.FindWaitingListByEvent(eventID, availableSeats)
	if err != nil {
		return fmt.Errorf("bekleme listesi getirilemedi: %w", err)
	}

	// 3. Get venue for notification
	venue, err := s.eventRepo.FindByID(event.VenueID)
	if err != nil {
		venue = &models.Event{} // Fallback
	}

	// 4. Notify each user in the waiting list
	for _, entry := range waitingList {
		// Mark as notified
		if err := s.reservationRepo.MarkAsNotified(entry.ID); err != nil {
			continue // Log error but continue notifying others
		}

		// Notify observers
		s.eventPublisher.Notify(&observer.EventData{
			Type:      observer.EventTypeWaitingListNotify,
			Timestamp: time.Now(),
			Data: &observer.WaitingListNotifyData{
				UserEmail:     fmt.Sprintf("user_%d@email.com", entry.UserID), // In real app, fetch from user service
				UserPhone:     fmt.Sprintf("+90555%07d", entry.UserID),       // In real app, fetch from user service
				EventName:     event.Name,
				VenueName:     venue.Name,
				EventDateTime: event.StartTime.Format("02.01.2006 15:04"),
			},
		})
	}

	return nil
}

// GetWaitingListPosition gets a user's position in the waiting list
func (s *ReservationService) GetWaitingListPosition(eventID, userID int64) (int, error) {
	// Get all waiting list entries for the event
	waitingList, err := s.reservationRepo.FindWaitingListByEvent(eventID, 1000) // High limit
	if err != nil {
		return 0, fmt.Errorf("bekleme listesi getirilemedi: %w", err)
	}

	// Find user's position
	for i, entry := range waitingList {
		if entry.UserID == userID {
			return i + 1, nil // Position is 1-indexed
		}
	}

	return 0, fmt.Errorf("kullanıcı bekleme listesinde değil")
}

// GetUserWaitingLists retrieves all waiting list entries for a user
func (s *ReservationService) GetUserWaitingLists(userID int64) ([]*models.WaitingList, error) {
	waitingLists, err := s.reservationRepo.FindWaitingListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("bekleme listeleri getirilemedi: %w", err)
	}

	return waitingLists, nil
}

// GetEventRevenue calculates total revenue for an event
func (s *ReservationService) GetEventRevenue(eventID int64) (float64, error) {
	revenue, err := s.reservationRepo.GetTotalRevenueByEvent(eventID)
	if err != nil {
		return 0, fmt.Errorf("gelir hesaplanamadı: %w", err)
	}

	return revenue, nil
}

// GetWaitingListCount returns the number of users in the waiting list
func (s *ReservationService) GetWaitingListCount(eventID int64) (int, error) {
	count, err := s.reservationRepo.GetWaitingListCount(eventID)
	if err != nil {
		return 0, fmt.Errorf("bekleme listesi sayısı alınamadı: %w", err)
	}

	return count, nil
}
