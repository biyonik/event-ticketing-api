package services

import (
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/internal/patterns/observer"
	"github.com/biyonik/event-ticketing-api/internal/patterns/strategy"
	"github.com/biyonik/event-ticketing-api/internal/repositories"
)

type EventService struct {
	eventRepo         *repositories.EventRepository
	venueRepo         *repositories.VenueRepository
	pricingFactory    *strategy.PricingStrategyFactory
	eventPublisher    *observer.EventPublisher
}

func NewEventService(
	eventRepo *repositories.EventRepository,
	venueRepo *repositories.VenueRepository,
	eventPublisher *observer.EventPublisher,
) *EventService {
	return &EventService{
		eventRepo:      eventRepo,
		venueRepo:      venueRepo,
		pricingFactory: strategy.NewPricingStrategyFactory(),
		eventPublisher: eventPublisher,
	}
}

func (s *EventService) CreateEvent(
	name, description string,
	eventType models.EventType,
	venueID int64,
	startTime, endTime time.Time,
	basePrice float64,
	imageURL string,
	featured bool,
	metadata string,
) (*models.Event, error) {
	// 1. Validation
	if name == "" {
		return nil, fmt.Errorf("etkinlik adı boş olamaz")
	}

	if startTime.Before(time.Now()) {
		return nil, fmt.Errorf("etkinlik başlangıç zamanı geçmişte olamaz")
	}

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("bitiş zamanı başlangıç zamanından önce olamaz")
	}

	if basePrice <= 0 {
		return nil, fmt.Errorf("temel fiyat sıfırdan büyük olmalıdır")
	}

	// 2. Verify venue exists
	venue, err := s.venueRepo.FindByID(venueID)
	if err != nil {
		return nil, fmt.Errorf("mekan bulunamadı: %w", err)
	}

	// 3. Create event
	event := &models.Event{
		Name:           name,
		Description:    description,
		Type:           eventType,
		Status:         models.EventStatusDraft,
		VenueID:        venueID,
		StartTime:      startTime,
		EndTime:        endTime,
		BasePrice:      basePrice,
		TotalCapacity:  venue.Capacity,
		AvailableSeats: venue.Capacity,
		ImageURL:       imageURL,
		Featured:       featured,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	eventID, err := s.eventRepo.Create(event)
	if err != nil {
		return nil, fmt.Errorf("etkinlik oluşturulamadı: %w", err)
	}

	event.ID = eventID
	return event, nil
}

func (s *EventService) GetEventByID(id int64) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	return event, nil
}

func (s *EventService) ListEvents(filters map[string]interface{}, page, pageSize int) ([]*models.Event, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	events, err := s.eventRepo.FindAll(filters, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("etkinlikler listelenemedi: %w", err)
	}

	return events, nil
}

func (s *EventService) UpdateEvent(id int64, updates map[string]interface{}) (*models.Event, error) {
	// 1. Get existing event
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		event.Name = name
	}

	if description, ok := updates["description"].(string); ok {
		event.Description = description
	}

	if eventType, ok := updates["type"].(models.EventType); ok {
		event.Type = eventType
	}

	if startTime, ok := updates["start_time"].(time.Time); ok {
		if startTime.Before(time.Now()) {
			return nil, fmt.Errorf("etkinlik başlangıç zamanı geçmişte olamaz")
		}
		event.StartTime = startTime
	}

	if endTime, ok := updates["end_time"].(time.Time); ok {
		if endTime.Before(event.StartTime) {
			return nil, fmt.Errorf("bitiş zamanı başlangıç zamanından önce olamaz")
		}
		event.EndTime = endTime
	}

	if basePrice, ok := updates["base_price"].(float64); ok && basePrice > 0 {
		event.BasePrice = basePrice
	}

	if imageURL, ok := updates["image_url"].(string); ok {
		event.ImageURL = imageURL
	}

	if featured, ok := updates["featured"].(bool); ok {
		event.Featured = featured
	}

	if metadata, ok := updates["metadata"].(string); ok {
		event.Metadata = metadata
	}

	// 3. Update in database
	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("etkinlik güncellenemedi: %w", err)
	}

	return event, nil
}

func (s *EventService) DeleteEvent(id int64) error {
	// 1. Check if event exists
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Business rule: Can only delete draft events
	if event.Status != models.EventStatusDraft {
		return fmt.Errorf("sadece taslak etkinlikler silinebilir")
	}

	// 3. Delete
	if err := s.eventRepo.Delete(id); err != nil {
		return fmt.Errorf("etkinlik silinemedi: %w", err)
	}

	return nil
}

func (s *EventService) PublishEvent(id int64) error {
	// 1. Get event
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Business rules
	if event.Status != models.EventStatusDraft {
		return fmt.Errorf("sadece taslak etkinlikler yayınlanabilir")
	}

	if event.StartTime.Before(time.Now()) {
		return fmt.Errorf("geçmiş tarihli etkinlik yayınlanamaz")
	}

	// 3. Update status
	if err := s.eventRepo.UpdateStatus(id, models.EventStatusPublished); err != nil {
		return fmt.Errorf("etkinlik yayınlanamadı: %w", err)
	}

	// 4. Publish event notification
	s.eventPublisher.Notify(&observer.EventData{
		Type:      observer.EventTypeEventStatusChanged,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"event_id":    id,
			"event_name":  event.Name,
			"new_status":  models.EventStatusPublished,
			"start_time":  event.StartTime,
		},
	})

	return nil
}

func (s *EventService) ActivateSale(id int64) error {
	// 1. Get event
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Business rules
	if event.Status != models.EventStatusPublished {
		return fmt.Errorf("sadece yayınlanmış etkinliklerin satışı aktif edilebilir")
	}

	// 3. Update status
	if err := s.eventRepo.UpdateStatus(id, models.EventStatusSaleActive); err != nil {
		return fmt.Errorf("satış aktif edilemedi: %w", err)
	}

	return nil
}

func (s *EventService) MarkAsSoldOut(id int64) error {
	// 1. Get event
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Verify actually sold out
	if event.AvailableSeats > 0 {
		return fmt.Errorf("etkinlikte hala müsait koltuk var")
	}

	// 3. Update status
	if err := s.eventRepo.UpdateStatus(id, models.EventStatusSoldOut); err != nil {
		return fmt.Errorf("tükendi işareti eklenemedi: %w", err)
	}

	return nil
}

func (s *EventService) CancelEvent(id int64) error {
	// 1. Get event
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Business rules
	if event.Status == models.EventStatusCompleted || event.Status == models.EventStatusCancelled {
		return fmt.Errorf("tamamlanmış veya iptal edilmiş etkinlik tekrar iptal edilemez")
	}

	// 3. Update status
	if err := s.eventRepo.UpdateStatus(id, models.EventStatusCancelled); err != nil {
		return fmt.Errorf("etkinlik iptal edilemedi: %w", err)
	}

	return nil
}

func (s *EventService) GetUpcomingEvents(limit int) ([]*models.Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	events, err := s.eventRepo.GetUpcomingEvents(limit)
	if err != nil {
		return nil, fmt.Errorf("yaklaşan etkinlikler getirilemedi: %w", err)
	}

	return events, nil
}

func (s *EventService) GetFeaturedEvents(limit int) ([]*models.Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	events, err := s.eventRepo.GetFeaturedEvents(limit)
	if err != nil {
		return nil, fmt.Errorf("öne çıkan etkinlikler getirilemedi: %w", err)
	}

	return events, nil
}

func (s *EventService) SearchEvents(keyword string, limit int) ([]*models.Event, error) {
	if keyword == "" {
		return nil, fmt.Errorf("arama kelimesi boş olamaz")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	events, err := s.eventRepo.SearchByName(keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("etkinlik araması yapılamadı: %w", err)
	}

	return events, nil
}

// CalculateTicketPrice calculates ticket price using pricing strategies
func (s *EventService) CalculateTicketPrice(eventID int64, sectionType string) (float64, error) {
	// 1. Get event
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return 0, fmt.Errorf("etkinlik bulunamadı: %w", err)
	}

	// 2. Build pricing context
	now := time.Now()
	occupancyRate := event.GetOccupancyRate()
	isWeekend := now.Weekday() == time.Saturday || now.Weekday() == time.Sunday

	context := &strategy.PricingContext{
		EventStartTime:    event.StartTime,
		CurrentTime:       now,
		OccupancyRate:     occupancyRate,
		SectionType:       sectionType,
		IsWeekend:         isWeekend,
		RemainingCapacity: event.AvailableSeats,
		TotalCapacity:     event.TotalCapacity,
	}

	// 3. Create composite pricing strategy based on event type
	strategies := []strategy.PricingStrategy{}

	// Always apply early bird pricing (30 days before, 20% discount)
	strategies = append(strategies, s.pricingFactory.CreateEarlyBirdStrategy(30, 20))

	// VIP pricing for premium sections
	if sectionType == "VIP" || sectionType == "Premium" {
		strategies = append(strategies, s.pricingFactory.CreateVIPStrategy(2.5))
	}

	// Dynamic pricing based on demand
	strategies = append(strategies, s.pricingFactory.CreateDynamicStrategy(2.0, 0.8))

	// Seasonal pricing for concerts in summer
	if event.Type == models.EventTypeConcert {
		strategies = append(strategies, s.pricingFactory.CreateSeasonalStrategy(
			[]time.Month{time.June, time.July, time.August},
			0.15, // 15% summer markup
		))
	}

	// 4. Apply all strategies
	compositeStrategy := s.pricingFactory.CreateCompositeStrategy(strategies...)
	finalPrice := compositeStrategy.CalculatePrice(event.BasePrice, context)

	return finalPrice, nil
}
