package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	"github.com/biyonik/event-ticketing-api/internal/services"
)

// EventController handles HTTP requests for events (ultra-thin - no business logic!)
type EventController struct {
	eventService *services.EventService
}

func NewEventController(eventService *services.EventService) *EventController {
	return &EventController{
		eventService: eventService,
	}
}

// Create handles POST /events
func (c *EventController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Type        models.EventType  `json:"type"`
		VenueID     int64             `json:"venue_id"`
		StartTime   string            `json:"start_time"`
		EndTime     string            `json:"end_time"`
		BasePrice   float64           `json:"base_price"`
		ImageURL    string            `json:"image_url"`
		Featured    bool              `json:"featured"`
		Metadata    string            `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	startTime, _ := time.Parse(time.RFC3339, req.StartTime)
	endTime, _ := time.Parse(time.RFC3339, req.EndTime)

	// 2. Call service (ALL LOGIC HERE!)
	event, err := c.eventService.CreateEvent(
		req.Name, req.Description, req.Type, req.VenueID,
		startTime, endTime, req.BasePrice, req.ImageURL, req.Featured, req.Metadata,
	)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusCreated, event)
}

// GetByID handles GET /events/:id
func (c *EventController) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID from URL
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	event, err := c.eventService.GetEventByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, event)
}

// List handles GET /events
func (c *EventController) List(w http.ResponseWriter, r *http.Request) {
	// 1. Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))

	filters := make(map[string]interface{})
	if status := query.Get("status"); status != "" {
		filters["status"] = models.EventStatus(status)
	}
	if eventType := query.Get("type"); eventType != "" {
		filters["type"] = models.EventType(eventType)
	}

	// 2. Call service
	events, err := c.eventService.ListEvents(filters, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, events)
}

// Update handles PUT /events/:id
func (c *EventController) Update(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID and request
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	// 2. Call service
	event, err := c.eventService.UpdateEvent(id, updates)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, event)
}

// Delete handles DELETE /events/:id
func (c *EventController) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	if err := c.eventService.DeleteEvent(id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "etkinlik silindi"})
}

// Publish handles POST /events/:id/publish
func (c *EventController) Publish(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	if err := c.eventService.PublishEvent(id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "etkinlik yayınlandı"})
}

// ActivateSale handles POST /events/:id/activate-sale
func (c *EventController) ActivateSale(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	if err := c.eventService.ActivateSale(id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "satış aktif edildi"})
}

// Cancel handles POST /events/:id/cancel
func (c *EventController) Cancel(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	if err := c.eventService.CancelEvent(id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "etkinlik iptal edildi"})
}

// GetUpcoming handles GET /events/upcoming
func (c *EventController) GetUpcoming(w http.ResponseWriter, r *http.Request) {
	// 1. Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// 2. Call service
	events, err := c.eventService.GetUpcomingEvents(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, events)
}

// GetFeatured handles GET /events/featured
func (c *EventController) GetFeatured(w http.ResponseWriter, r *http.Request) {
	// 1. Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// 2. Call service
	events, err := c.eventService.GetFeaturedEvents(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, events)
}

// Search handles GET /events/search
func (c *EventController) Search(w http.ResponseWriter, r *http.Request) {
	// 1. Parse query parameters
	keyword := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// 2. Call service
	events, err := c.eventService.SearchEvents(keyword, limit)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, events)
}

// CalculatePrice handles GET /events/:id/calculate-price
func (c *EventController) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	// 1. Parse parameters
	id, err := parseIDFromPath(r.URL.Path, "/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	sectionType := r.URL.Query().Get("section_type")

	// 2. Call service
	price, err := c.eventService.CalculateTicketPrice(id, sectionType)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]float64{"price": price})
}
