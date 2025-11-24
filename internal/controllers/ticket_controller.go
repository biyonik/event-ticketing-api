package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/biyonik/event-ticketing-api/internal/services"
)

// TicketController handles HTTP requests for tickets (ultra-thin!)
type TicketController struct {
	ticketService *services.TicketService
}

func NewTicketController(ticketService *services.TicketService) *TicketController {
	return &TicketController{
		ticketService: ticketService,
	}
}

// Reserve handles POST /tickets/reserve
func (c *TicketController) Reserve(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	var req struct {
		EventID   int64  `json:"event_id"`
		SectionID int64  `json:"section_id"`
		SeatID    *int64 `json:"seat_id"`
		Price     float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	userID := getUserIDFromContext(r)

	// 2. Call service
	ticket, err := c.ticketService.ReserveTicket(userID, req.EventID, req.SectionID, req.SeatID, req.Price)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusCreated, ticket)
}

// Purchase handles POST /tickets/:id/purchase
func (c *TicketController) Purchase(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	id, err := parseIDFromPath(r.URL.Path, "/tickets/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	var req struct {
		UserEmail string `json:"user_email"`
		UserPhone string `json:"user_phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	// 2. Call service
	if err := c.ticketService.PurchaseTicket(id, req.UserEmail, req.UserPhone); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "bilet satın alındı"})
}

// Cancel handles POST /tickets/:id/cancel
func (c *TicketController) Cancel(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	id, err := parseIDFromPath(r.URL.Path, "/tickets/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	var req struct {
		UserEmail string `json:"user_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	// 2. Call service
	if err := c.ticketService.CancelTicket(id, req.UserEmail); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "bilet iptal edildi"})
}

// Validate handles POST /tickets/validate
func (c *TicketController) Validate(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	var req struct {
		TicketNumber     string `json:"ticket_number"`
		VerificationCode string `json:"verification_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	// 2. Call service
	ticket, err := c.ticketService.ValidateTicket(req.TicketNumber, req.VerificationCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"ticket": ticket,
	})
}

// Use handles POST /tickets/:id/use
func (c *TicketController) Use(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	var req struct {
		TicketNumber string `json:"ticket_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz istek")
		return
	}

	// 2. Call service
	if err := c.ticketService.UseTicket(req.TicketNumber); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, map[string]string{"message": "bilet kullanıldı"})
}

// GetUserTickets handles GET /tickets/my-tickets
func (c *TicketController) GetUserTickets(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID
	userID := getUserIDFromContext(r)

	// 2. Call service
	tickets, err := c.ticketService.GetUserTickets(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, tickets)
}

// GetByID handles GET /tickets/:id
func (c *TicketController) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/tickets/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	ticket, err := c.ticketService.GetTicketByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, ticket)
}

// GetEventStats handles GET /tickets/events/:id/stats
func (c *TicketController) GetEventStats(w http.ResponseWriter, r *http.Request) {
	// 1. Parse ID
	id, err := parseIDFromPath(r.URL.Path, "/tickets/events/")
	if err != nil {
		respondError(w, http.StatusBadRequest, "geçersiz ID")
		return
	}

	// 2. Call service
	stats, err := c.ticketService.GetEventSalesStats(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return response
	respondJSON(w, http.StatusOK, stats)
}
