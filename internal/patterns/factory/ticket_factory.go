package factory

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/biyonik/event-ticketing-api/internal/models"
	qrcode "github.com/skip2/go-qrcode"
)

// TicketFactory handles the creation of different ticket types
type TicketFactory struct {
	qrGenerator QRCodeGenerator
}

// QRCodeGenerator interface for generating QR codes
type QRCodeGenerator interface {
	Generate(data string) ([]byte, error)
	GenerateWithOptions(data string, size int, level qrcode.RecoveryLevel) ([]byte, error)
}

// DefaultQRCodeGenerator implements QRCodeGenerator
type DefaultQRCodeGenerator struct{}

func (g *DefaultQRCodeGenerator) Generate(data string) ([]byte, error) {
	return qrcode.Encode(data, qrcode.Medium, 256)
}

func (g *DefaultQRCodeGenerator) GenerateWithOptions(data string, size int, level qrcode.RecoveryLevel) ([]byte, error) {
	return qrcode.Encode(data, level, size)
}

// NewTicketFactory creates a new ticket factory
func NewTicketFactory() *TicketFactory {
	return &TicketFactory{
		qrGenerator: &DefaultQRCodeGenerator{},
	}
}

// NewTicketFactoryWithQRGenerator creates a factory with custom QR generator
func NewTicketFactoryWithQRGenerator(qrGenerator QRCodeGenerator) *TicketFactory {
	return &TicketFactory{
		qrGenerator: qrGenerator,
	}
}

// TicketCreationRequest holds parameters for ticket creation
type TicketCreationRequest struct {
	EventID       int64
	UserID        int64
	SeatID        *int64
	SectionID     int64
	Price         float64
	TicketType    models.TicketType
	ReservationID *int64
	EventName     string
	VenueName     string
	SeatInfo      string
}

// CreateTicket creates a standard ticket with QR code
func (f *TicketFactory) CreateTicket(req *TicketCreationRequest) (*models.Ticket, error) {
	ticket := &models.Ticket{
		EventID:    req.EventID,
		UserID:     req.UserID,
		SeatID:     req.SeatID,
		SectionID:  req.SectionID,
		Price:      req.Price,
		TicketType: req.TicketType,
		Status:     models.TicketStatusReserved,
		CreatedAt:  time.Now(),
	}

	// Generate unique ticket number
	ticketNumber, err := f.generateTicketNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ticket number: %w", err)
	}
	ticket.TicketNumber = ticketNumber

	// Generate verification code (6-digit)
	verificationCode, err := f.generateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}
	ticket.VerificationCode = verificationCode

	// Generate QR code data
	qrData := f.buildQRCodeData(ticket, req)
	ticket.QRCodeData = qrData

	// Generate QR code image
	qrCodeImage, err := f.qrGenerator.Generate(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	ticket.QRCodeImage = qrCodeImage

	// Set reservation expiry (15 minutes for reserved tickets)
	expiryTime := time.Now().Add(15 * time.Minute)
	ticket.ReservationExpiry = &expiryTime

	return ticket, nil
}

// CreateVIPTicket creates a VIP ticket with enhanced QR code
func (f *TicketFactory) CreateVIPTicket(req *TicketCreationRequest) (*models.Ticket, error) {
	req.TicketType = models.TicketTypeVIP

	ticket, err := f.CreateTicket(req)
	if err != nil {
		return nil, err
	}

	// VIP tickets get high-quality QR codes
	qrCodeImage, err := f.qrGenerator.GenerateWithOptions(ticket.QRCodeData, 512, qrcode.High)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VIP QR code: %w", err)
	}
	ticket.QRCodeImage = qrCodeImage

	return ticket, nil
}

// CreateGroupTickets creates multiple tickets for a group purchase
func (f *TicketFactory) CreateGroupTickets(req *TicketCreationRequest, count int, seatIDs []int64) ([]*models.Ticket, error) {
	if count != len(seatIDs) {
		return nil, fmt.Errorf("seat count mismatch: requested %d tickets but provided %d seats", count, len(seatIDs))
	}

	tickets := make([]*models.Ticket, 0, count)

	for i := 0; i < count; i++ {
		ticketReq := *req
		seatID := seatIDs[i]
		ticketReq.SeatID = &seatID

		ticket, err := f.CreateTicket(&ticketReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create ticket %d: %w", i+1, err)
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// CreateSeasonTicket creates a season pass ticket
func (f *TicketFactory) CreateSeasonTicket(req *TicketCreationRequest) (*models.Ticket, error) {
	req.TicketType = models.TicketTypeSeason

	ticket, err := f.CreateTicket(req)
	if err != nil {
		return nil, err
	}

	// Season tickets don't expire during reservation
	ticket.ReservationExpiry = nil
	ticket.Status = models.TicketStatusSold

	return ticket, nil
}

// RegenerateQRCode regenerates the QR code for a ticket
func (f *TicketFactory) RegenerateQRCode(ticket *models.Ticket) error {
	qrCodeImage, err := f.qrGenerator.Generate(ticket.QRCodeData)
	if err != nil {
		return fmt.Errorf("failed to regenerate QR code: %w", err)
	}

	ticket.QRCodeImage = qrCodeImage
	return nil
}

// generateTicketNumber generates a unique ticket number
// Format: TKT-YYYYMMDD-XXXXXXXX (20 characters)
func (f *TicketFactory) generateTicketNumber() (string, error) {
	dateStr := time.Now().Format("20060102")
	randomBytes := make([]byte, 4)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("TKT-%s-%s", dateStr, randomStr), nil
}

// generateVerificationCode generates a 6-digit verification code
func (f *TicketFactory) generateVerificationCode() (string, error) {
	randomBytes := make([]byte, 3)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Convert to 6-digit number
	code := int(randomBytes[0])<<16 | int(randomBytes[1])<<8 | int(randomBytes[2])
	code = code % 1000000 // Ensure 6 digits

	return fmt.Sprintf("%06d", code), nil
}

// buildQRCodeData builds the QR code data string
func (f *TicketFactory) buildQRCodeData(ticket *models.Ticket, req *TicketCreationRequest) string {
	seatInfo := "Genel"
	if req.SeatInfo != "" {
		seatInfo = req.SeatInfo
	}

	return fmt.Sprintf(
		"TICKET:%s|EVENT:%d|USER:%d|SEAT:%s|CODE:%s|TYPE:%s",
		ticket.TicketNumber,
		ticket.EventID,
		ticket.UserID,
		seatInfo,
		ticket.VerificationCode,
		ticket.TicketType,
	)
}

// TicketValidator validates QR codes and verification codes
type TicketValidator struct{}

func NewTicketValidator() *TicketValidator {
	return &TicketValidator{}
}

// ValidateQRCode validates a QR code data string
func (v *TicketValidator) ValidateQRCode(qrData string, ticket *models.Ticket) bool {
	expectedData := fmt.Sprintf(
		"TICKET:%s|EVENT:%d|USER:%d|",
		ticket.TicketNumber,
		ticket.EventID,
		ticket.UserID,
	)

	// Check if QR data starts with expected format
	return len(qrData) > len(expectedData) && qrData[:len(expectedData)] == expectedData
}

// ValidateVerificationCode validates a verification code
func (v *TicketValidator) ValidateVerificationCode(code string, ticket *models.Ticket) bool {
	return code == ticket.VerificationCode
}

// TicketPrinter generates printable ticket data
type TicketPrinter struct{}

func NewTicketPrinter() *TicketPrinter {
	return &TicketPrinter{}
}

// PrintableTicket represents printable ticket data
type PrintableTicket struct {
	TicketNumber     string
	EventName        string
	VenueName        string
	DateTime         string
	SeatInfo         string
	Price            string
	VerificationCode string
	QRCodeImage      []byte
	TicketType       string
}

// GeneratePrintableTicket converts a ticket to printable format
func (p *TicketPrinter) GeneratePrintableTicket(ticket *models.Ticket, eventName, venueName string, eventTime time.Time, seatInfo string) *PrintableTicket {
	return &PrintableTicket{
		TicketNumber:     ticket.TicketNumber,
		EventName:        eventName,
		VenueName:        venueName,
		DateTime:         eventTime.Format("02.01.2006 15:04"),
		SeatInfo:         seatInfo,
		Price:            fmt.Sprintf("%.2f TL", ticket.Price),
		VerificationCode: ticket.VerificationCode,
		QRCodeImage:      ticket.QRCodeImage,
		TicketType:       string(ticket.TicketType),
	}
}
