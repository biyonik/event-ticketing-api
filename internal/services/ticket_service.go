package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/internal/patterns/factory"
	"github.com/biyonik/event-ticketing-api/internal/patterns/observer"
	"github.com/biyonik/event-ticketing-api/internal/repositories"
	v "github.com/biyonik/event-ticketing-api/pkg/validation"
	"github.com/biyonik/event-ticketing-api/pkg/validation/types"
)

type TicketService struct {
	ticketRepo     *repositories.TicketRepository
	eventRepo      *repositories.EventRepository
	venueRepo      *repositories.VenueRepository
	ticketFactory  *factory.TicketFactory
	ticketValidator *factory.TicketValidator
	eventPublisher *observer.EventPublisher
	db             *sql.DB
}

func NewTicketService(
	ticketRepo *repositories.TicketRepository,
	eventRepo *repositories.EventRepository,
	venueRepo *repositories.VenueRepository,
	eventPublisher *observer.EventPublisher,
	db *sql.DB,
) *TicketService {
	return &TicketService{
		ticketRepo:      ticketRepo,
		eventRepo:       eventRepo,
		venueRepo:       venueRepo,
		ticketFactory:   factory.NewTicketFactory(),
		ticketValidator: factory.NewTicketValidator(),
		eventPublisher:  eventPublisher,
		db:              db,
	}
}

// ReserveTicket reserves a ticket for a limited time
func (s *TicketService) ReserveTicket(userID, eventID, sectionID int64, seatID *int64, price float64) (*models.Ticket, error) {
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
		"section_id": types.Number().
			Required().
			Min(1).
			Label("Bölüm ID"),
		"price": types.Number().
			Required().
			Min(0.01).
			Label("Fiyat"),
	})

	rawData := map[string]any{
		"user_id":    float64(userID),
		"event_id":   float64(eventID),
		"section_id": float64(sectionID),
		"price":      price,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return nil, fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Validate event
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 3. Business rules
	if !event.IsSaleActive() {
		return nil, fmt.Errorf("bilet satışı aktif değil")
	}

	if event.IsSoldOut() {
		return nil, fmt.Errorf("etkinlik tükendi")
	}

	// 4. Check seat availability if specific seat requested
	if seatID != nil {
		isTaken, err := s.ticketRepo.IsSeatTaken(eventID, *seatID)
		if err != nil {
			return nil, fmt.Errorf("koltuk kontrolü yapılamadı: %w", err)
		}
		if isTaken {
			return nil, fmt.Errorf("koltuk dolu")
		}
	}

	// 5. Start transaction to prevent double booking
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// 6. Decrement available seats
	if err := s.eventRepo.DecrementAvailableSeats(eventID, 1); err != nil {
		return nil, fmt.Errorf("koltuk rezervasyonu yapılamadı: %w", err)
	}

	// 7. Get venue and section info for ticket
	venue, err := s.venueRepo.FindByID(event.VenueID)
	if err != nil {
		return nil, fmt.Errorf("mekan bulunamadı: %w", err)
	}

	section, err := s.venueRepo.FindSectionByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("bölüm bulunamadı: %w", err)
	}

	seatInfo := section.Name
	if seatID != nil {
		seat, err := s.venueRepo.FindSeatByID(*seatID)
		if err != nil {
			return nil, fmt.Errorf("koltuk bulunamadı: %w", err)
		}
		seatInfo = fmt.Sprintf("%s - Sıra: %s, Koltuk: %s", section.Name, seat.Row, seat.Number)
	}

	// 8. Create ticket using Factory pattern
	ticketReq := &factory.TicketCreationRequest{
		EventID:    eventID,
		UserID:     userID,
		SeatID:     seatID,
		SectionID:  sectionID,
		Price:      price,
		TicketType: models.TicketTypeStandard,
		EventName:  event.Name,
		VenueName:  venue.Name,
		SeatInfo:   seatInfo,
	}

	ticket, err := s.ticketFactory.CreateTicket(ticketReq)
	if err != nil {
		return nil, fmt.Errorf("bilet oluşturulamadı: %w", err)
	}

	// 9. Save ticket to database
	ticketID, err := s.ticketRepo.Create(ticket)
	if err != nil {
		return nil, fmt.Errorf("bilet kaydedilemedi: %w", err)
	}
	ticket.ID = ticketID

	// 10. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit edilemedi: %w", err)
	}

	return ticket, nil
}

// PurchaseTicket completes a ticket purchase
func (s *TicketService) PurchaseTicket(ticketID int64, userEmail, userPhone string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"ticket_id": types.Number().
			Required().
			Min(1).
			Label("Bilet ID"),
		"user_email": types.String().
			Required().
			Email().
			Label("E-posta"),
		"user_phone": types.String().
			Required().
			Min(10).
			Max(15).
			Label("Telefon"),
	})

	rawData := map[string]any{
		"ticket_id":  float64(ticketID),
		"user_email": userEmail,
		"user_phone": userPhone,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Get ticket
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return fmt.Errorf("bilet bulunamadı: %w", err)
	}

	// 3. Business rules - use State Pattern methods
	if !ticket.CanPurchase() {
		return fmt.Errorf("bilet satın alınamaz durumda")
	}

	// 4. Mark as sold
	if err := ticket.MarkAsSold(); err != nil {
		return fmt.Errorf("bilet satış durumuna geçirilemedi: %w", err)
	}

	// 5. Update in database
	if err := s.ticketRepo.Update(ticket); err != nil {
		return fmt.Errorf("bilet güncellenemedi: %w", err)
	}

	// 6. Get event details for notification
	event, err := s.eventRepo.FindByID(ticket.EventID)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	venue, err := s.venueRepo.FindByID(event.VenueID)
	if err != nil {
		return fmt.Errorf("mekan bulunamadı: %w", err)
	}

	seatInfo := "Genel"
	if ticket.SeatID != nil {
		seat, _ := s.venueRepo.FindSeatByID(*ticket.SeatID)
		if seat != nil {
			section, _ := s.venueRepo.FindSectionByID(seat.SectionID)
			if section != nil {
				seatInfo = fmt.Sprintf("%s - Sıra: %s, Koltuk: %s", section.Name, seat.Row, seat.Number)
			}
		}
	}

	// 7. Notify observers using Observer pattern
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypeTicketPurchased,
		Timestamp: time.Now(),
		Data: &observer.TicketPurchaseData{
			UserID:           ticket.UserID,
			UserEmail:        userEmail,
			UserPhone:        userPhone,
			EventID:          event.ID,
			EventName:        event.Name,
			VenueName:        venue.Name,
			EventDateTime:    event.StartTime.Format("02.01.2006 15:04"),
			TicketNumber:     ticket.TicketNumber,
			VerificationCode: ticket.VerificationCode,
			SeatInfo:         seatInfo,
			Price:            ticket.Price,
		},
	})

	// 8. Check if event sold out
	if event.AvailableSeats == 0 {
		s.eventRepo.UpdateStatus(event.ID, models.EventStatusSoldOut)
	}

	return nil
}

// CancelTicket cancels a ticket and refunds
func (s *TicketService) CancelTicket(ticketID int64, userEmail string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"ticket_id": types.Number().
			Required().
			Min(1).
			Label("Bilet ID"),
		"user_email": types.String().
			Required().
			Email().
			Label("E-posta"),
	})

	rawData := map[string]any{
		"ticket_id":  float64(ticketID),
		"user_email": userEmail,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Get ticket
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return fmt.Errorf("bilet bulunamadı: %w", err)
	}

	// 3. Business rules - use State Pattern methods
	if !ticket.CanCancel() {
		return fmt.Errorf("bilet iptal edilemez durumda")
	}

	// 4. Get event to check cancellation policy
	event, err := s.eventRepo.FindByID(ticket.EventID)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// Business rule: Can't cancel within 24 hours of event
	if time.Until(event.StartTime) < 24*time.Hour {
		return fmt.Errorf("etkinlikten 24 saat kala iptal yapılamaz")
	}

	// 5. Mark as cancelled
	if err := ticket.MarkAsCancelled(); err != nil {
		return fmt.Errorf("bilet iptal durumuna geçirilemedi: %w", err)
	}

	// 6. Update in database
	if err := s.ticketRepo.Update(ticket); err != nil {
		return fmt.Errorf("bilet güncellenemedi: %w", err)
	}

	// 7. Increment available seats
	if err := s.eventRepo.IncrementAvailableSeats(ticket.EventID, 1); err != nil {
		return fmt.Errorf("koltuk sayısı artırılamadı: %w", err)
	}

	// 8. Notify observers
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypeTicketCancelled,
		Timestamp: time.Now(),
		Data: &observer.TicketCancellationData{
			UserEmail:    userEmail,
			TicketNumber: ticket.TicketNumber,
			RefundAmount: ticket.Price,
		},
	})

	return nil
}

// ValidateTicket validates a ticket at venue entrance
func (s *TicketService) ValidateTicket(ticketNumber, verificationCode string) (*models.Ticket, error) {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"ticket_number": types.String().
			Required().
			Min(10).
			Max(50).
			Label("Bilet Numarası"),
		"verification_code": types.String().
			Required().
			Min(6).
			Max(10).
			Label("Doğrulama Kodu"),
	})

	rawData := map[string]any{
		"ticket_number":     ticketNumber,
		"verification_code": verificationCode,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return nil, fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Find ticket
	ticket, err := s.ticketRepo.FindByTicketNumber(ticketNumber)
	if err != nil {
		return nil, fmt.Errorf("bilet bulunamadı: %w", err)
	}

	// 3. Validate verification code
	if !s.ticketValidator.ValidateVerificationCode(verificationCode, ticket) {
		return nil, fmt.Errorf("doğrulama kodu hatalı")
	}

	// 4. Business rules - use State Pattern methods
	if !ticket.CanUse() {
		return nil, fmt.Errorf("bilet kullanılamaz durumda: %s", ticket.Status)
	}

	// 5. Check if ticket already used
	if ticket.Status == models.TicketStatusUsed {
		return nil, fmt.Errorf("bilet daha önce kullanılmış")
	}

	return ticket, nil
}

// UseTicket marks a ticket as used
func (s *TicketService) UseTicket(ticketNumber string) error {
	// 1. Validate input using Conduit-Go Validation
	schema := v.Make().Shape(map[string]v.Type{
		"ticket_number": types.String().
			Required().
			Min(10).
			Max(50).
			Label("Bilet Numarası"),
	})

	rawData := map[string]any{
		"ticket_number": ticketNumber,
	}

	result := schema.Validate(rawData)
	if result.HasErrors() {
		for field, errs := range result.Errors() {
			return fmt.Errorf("%s: %s", field, errs[0])
		}
	}

	// 2. Find ticket
	ticket, err := s.ticketRepo.FindByTicketNumber(ticketNumber)
	if err != nil {
		return fmt.Errorf("bilet bulunamadı: %w", err)
	}

	// 3. Business rules
	if !ticket.CanUse() {
		return fmt.Errorf("bilet kullanılamaz durumda")
	}

	// 4. Mark as used
	if err := ticket.MarkAsUsed(); err != nil {
		return fmt.Errorf("bilet kullanıldı olarak işaretlenemedi: %w", err)
	}

	// 5. Update in database
	if err := s.ticketRepo.Update(ticket); err != nil {
		return fmt.Errorf("bilet güncellenemedi: %w", err)
	}

	// 6. Notify observers
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypeTicketUsed,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"ticket_id":     ticket.ID,
			"ticket_number": ticket.TicketNumber,
			"event_id":      ticket.EventID,
		},
	})

	return nil
}

// GetUserTickets retrieves all tickets for a user
func (s *TicketService) GetUserTickets(userID int64) ([]*models.Ticket, error) {
	tickets, err := s.ticketRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("biletler getirilemedi: %w", err)
	}

	return tickets, nil
}

// GetTicketByID retrieves a ticket by ID
func (s *TicketService) GetTicketByID(ticketID int64) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("bilet bulunamadı: %w", err)
	}

	return ticket, nil
}

// ExpireReservations expires all reservations that have passed their expiry time
func (s *TicketService) ExpireReservations() error {
	// 1. Find expired reservations
	expiredTickets, err := s.ticketRepo.FindExpiredReservations()
	if err != nil {
		return fmt.Errorf("süresi geçmiş rezervasyonlar bulunamadı: %w", err)
	}

	// 2. Process each expired ticket
	for _, ticket := range expiredTickets {
		// Mark as expired
		if err := s.ticketRepo.ExpireReservation(ticket.ID); err != nil {
			continue // Log error but continue processing others
		}

		// Increment available seats
		s.eventRepo.IncrementAvailableSeats(ticket.EventID, 1)

		// Notify observers
		s.eventPublisher.Notify(&observer.EventData{
			Type:      observer.EventTypeReservationExpired,
			Timestamp: time.Now(),
			Data: &observer.ReservationExpiredData{
				UserEmail:     fmt.Sprintf("user_%d@email.com", ticket.UserID), // In real app, fetch from user service
				EventName:     "Event", // In real app, fetch event name
				ReservationID: ticket.TicketNumber,
			},
		})
	}

	return nil
}

// GetEventRevenue calculates total revenue for an event
func (s *TicketService) GetEventRevenue(eventID int64) (float64, error) {
	revenue, err := s.ticketRepo.GetRevenueByEvent(eventID)
	if err != nil {
		return 0, fmt.Errorf("gelir hesaplanamadı: %w", err)
	}

	return revenue, nil
}

// GetEventSalesStats returns sales statistics for an event
func (s *TicketService) GetEventSalesStats(eventID int64) (map[string]interface{}, error) {
	// 1. Get event
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Get ticket count
	soldCount, err := s.ticketRepo.GetSoldTicketCountByEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf("satış sayısı alınamadı: %w", err)
	}

	// 3. Get revenue
	revenue, err := s.ticketRepo.GetRevenueByEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf("gelir alınamadı: %w", err)
	}

	// 4. Calculate stats
	occupancyRate := event.GetOccupancyRate()

	stats := map[string]interface{}{
		"total_capacity":  event.TotalCapacity,
		"available_seats": event.AvailableSeats,
		"sold_tickets":    soldCount,
		"occupancy_rate":  fmt.Sprintf("%.2f%%", occupancyRate*100),
		"total_revenue":   revenue,
		"average_price":   0.0,
	}

	if soldCount > 0 {
		stats["average_price"] = revenue / float64(soldCount)
	}

	return stats, nil
}
