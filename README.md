# Event Ticketing API

**Conduit-Go** framework kullanılarak geliştirilmiş, enterprise-level bir etkinlik ve biletleme sistemi. Bu proje, **Clean Architecture** prensipleri ve **Design Patterns** ile geliştirilmiş, ölçeklenebilir ve bakımı kolay bir API örneğidir.

## 🎯 Proje Özellikleri

- ✅ **Clean Architecture** (Controller → Service → Repository)
- ✅ **Design Patterns** (Strategy, Factory, Observer, State)
- ✅ **RESTful API** tasarımı
- ✅ **Ultra-thin Controllers** (şişkin controller yok!)
- ✅ **Domain-Driven Design** (DDD) yaklaşımı
- ✅ **SOLID Principles**
- ✅ **Docker** desteği
- ✅ **MySQL** database
- ✅ **Redis** caching
- ✅ **QR Code** oluşturma
- ✅ **Dynamic Pricing** (Dinamik fiyatlandırma)
- ✅ **Seat Mapping** (Koltuk haritası)
- ✅ **Waiting List** (Bekleme listesi)
- ✅ **Real-time Notifications** (Bildirimler)

## 📐 Mimari Tasarım

### Clean Architecture Katmanları

```
┌─────────────────────────────────────────────────────────────┐
│                      CONTROLLER LAYER                        │
│  • HTTP Request/Response handling                            │
│  • Input validation                                          │
│  • Ultra-thin (5-15 lines per method)                        │
│  • NO business logic!                                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                       SERVICE LAYER                          │
│  • ALL business logic here                                   │
│  • Design patterns integration                               │
│  • Orchestrates repositories                                 │
│  • Domain rules enforcement                                  │
│  • Independent from HTTP                                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                     REPOSITORY LAYER                         │
│  • Database interactions                                     │
│  • CRUD operations                                           │
│  • Query optimization                                        │
│  • Data access abstraction                                   │
└─────────────────────────────────────────────────────────────┘
```

### Proje Yapısı

```
event-ticketing-api/
├── cmd/
│   └── api/
│       └── main.go                 # Uygulama giriş noktası
├── internal/
│   ├── models/                     # Domain modelleri
│   │   ├── base.go
│   │   ├── event.go                # Event entity (business methods)
│   │   ├── ticket.go               # Ticket entity (State Pattern)
│   │   ├── venue.go                # Venue, Section, Seat
│   │   └── reservation.go          # Payment, WaitingList
│   ├── patterns/                   # Design Patterns
│   │   ├── strategy/              # Strategy Pattern (Pricing)
│   │   │   └── pricing_strategy.go
│   │   ├── factory/               # Factory Pattern (Tickets)
│   │   │   └── ticket_factory.go
│   │   ├── observer/              # Observer Pattern (Notifications)
│   │   │   └── notification_observer.go
│   │   └── state/                 # State Pattern (Ticket lifecycle)
│   ├── repositories/              # Data Access Layer
│   │   ├── event_repository.go
│   │   ├── ticket_repository.go
│   │   ├── venue_repository.go
│   │   └── reservation_repository.go
│   ├── services/                  # Business Logic Layer
│   │   ├── event_service.go       # Event business logic
│   │   ├── ticket_service.go      # Ticket business logic
│   │   └── reservation_service.go # Payment & WaitingList logic
│   └── controllers/               # HTTP Handlers (Ultra-thin!)
│       ├── event_controller.go
│       ├── ticket_controller.go
│       └── helpers.go
├── migrations/                    # Database migrations
│   ├── 001_create_venues_table.sql
│   ├── 002_create_events_table.sql
│   ├── 003_create_tickets_table.sql
│   └── 004_create_payments_and_waiting_lists.sql
├── docker-compose.yml            # Docker setup
├── Dockerfile
├── Makefile
├── .env.example
├── go.mod
└── README.md
```

## 🎨 Design Patterns (Tasarım Desenleri)

Bu projede 4 ana design pattern kullanılmıştır. Her pattern'in neden kullanıldığı ve nasıl çalıştığı aşağıda detaylı açıklanmıştır.

### 1. Strategy Pattern (Strateji Deseni) 💰

**Kullanım Alanı:** Dinamik fiyatlandırma sistemi

**Problem:**
Bilet fiyatları birçok faktöre bağlı olarak değişir:
- Erken rezervasyon indirimi (Early Bird)
- VIP koltuklar için premium fiyat
- Talebe göre dinamik fiyat (doluluk oranına göre)
- Mevsimsel fiyatlandırma (yaz konserleri vs.)
- Hafta sonu primleri

Bu kadar çok farklı fiyatlandırma kuralını if-else yığını ile yapmak kod karmaşasına yol açar.

**Çözüm:**
Strategy Pattern ile her fiyatlandırma kuralı ayrı bir strateji olarak implement edilir. İstediğiniz stratejileri birleştirebilirsiniz.

**Kod Örneği:**

```go
// internal/patterns/strategy/pricing_strategy.go

// Strategy interface
type PricingStrategy interface {
    CalculatePrice(basePrice float64, context *PricingContext) float64
    GetName() string
}

// Concrete Strategy 1: Early Bird
type EarlyBirdPricingStrategy struct {
    DaysBeforeEvent int
    DiscountPercent float64
}

func (s *EarlyBirdPricingStrategy) CalculatePrice(basePrice float64, ctx *PricingContext) float64 {
    daysUntilEvent := ctx.EventStartTime.Sub(ctx.CurrentTime).Hours() / 24

    if daysUntilEvent >= float64(s.DaysBeforeEvent) {
        discount := basePrice * (s.DiscountPercent / 100)
        return basePrice - discount
    }

    return basePrice
}

// Concrete Strategy 2: Dynamic Pricing (Demand-based)
type DynamicPricingStrategy struct {
    MaxPriceMultiplier float64
    MinPriceMultiplier float64
}

func (s *DynamicPricingStrategy) CalculatePrice(basePrice float64, ctx *PricingContext) float64 {
    // Doluluk oranına göre fiyat artışı
    priceMultiplier := s.MinPriceMultiplier +
        (ctx.OccupancyRate * (s.MaxPriceMultiplier - s.MinPriceMultiplier))

    price := basePrice * priceMultiplier

    // Hafta sonu premium
    if ctx.IsWeekend {
        price *= 1.15
    }

    return price
}

// Composite Strategy: Birden fazla stratejiyi birleştir
type CompositePricingStrategy struct {
    Strategies []PricingStrategy
}

func (s *CompositePricingStrategy) CalculatePrice(basePrice float64, ctx *PricingContext) float64 {
    finalPrice := basePrice

    for _, strategy := range s.Strategies {
        finalPrice = strategy.CalculatePrice(finalPrice, ctx)
    }

    return finalPrice
}
```

**Kullanım (Service Layer):**

```go
// internal/services/event_service.go

func (s *EventService) CalculateTicketPrice(eventID int64, sectionType string) (float64, error) {
    event, _ := s.eventRepo.FindByID(eventID)

    // Pricing context oluştur
    context := &strategy.PricingContext{
        EventStartTime:    event.StartTime,
        CurrentTime:       time.Now(),
        OccupancyRate:     event.GetOccupancyRate(),
        SectionType:       sectionType,
        IsWeekend:         time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday,
    }

    // Stratejileri oluştur
    strategies := []strategy.PricingStrategy{
        s.pricingFactory.CreateEarlyBirdStrategy(30, 20),  // 30 gün önce %20 indirim
        s.pricingFactory.CreateVIPStrategy(2.5),            // VIP için 2.5x fiyat
        s.pricingFactory.CreateDynamicStrategy(2.0, 0.8),   // Dinamik fiyat
    }

    // Composite strategy ile tüm kuralları uygula
    compositeStrategy := s.pricingFactory.CreateCompositeStrategy(strategies...)
    finalPrice := compositeStrategy.CalculatePrice(event.BasePrice, context)

    return finalPrice, nil
}
```

**Avantajlar:**
- ✅ Yeni fiyatlandırma kuralları eklemek kolay (Open/Closed Principle)
- ✅ Her strateji bağımsız test edilebilir
- ✅ Kod tekrarı yok
- ✅ If-else cehenneminden kurtulma

### 2. Factory Pattern (Fabrika Deseni) 🏭

**Kullanım Alanı:** Bilet ve QR kod oluşturma

**Problem:**
Bilet oluşturma karmaşık bir işlemdir:
- Unique bilet numarası oluştur
- QR kod oluştur
- Doğrulama kodu oluştur
- Bilet tipine göre farklı işlemler (VIP, Standard, Season Pass)
- Grup biletleri için toplu oluşturma

Bu mantığı her yerde tekrarlamak DRY prensibine aykırıdır.

**Çözüm:**
Factory Pattern ile bilet oluşturma işlemi tek bir yerde toplanır. Farklı bilet tipleri için factory methods kullanılır.

**Kod Örneği:**

```go
// internal/patterns/factory/ticket_factory.go

type TicketFactory struct {
    qrGenerator QRCodeGenerator
}

// Factory method: Standard bilet oluştur
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

    // Unique ticket number oluştur
    ticketNumber, _ := f.generateTicketNumber()
    ticket.TicketNumber = ticketNumber

    // 6-digit verification code oluştur
    verificationCode, _ := f.generateVerificationCode()
    ticket.VerificationCode = verificationCode

    // QR code data oluştur
    qrData := f.buildQRCodeData(ticket, req)
    ticket.QRCodeData = qrData

    // QR code image oluştur
    qrCodeImage, _ := f.qrGenerator.Generate(qrData)
    ticket.QRCodeImage = qrCodeImage

    // Rezervasyon süresi (15 dakika)
    expiryTime := time.Now().Add(15 * time.Minute)
    ticket.ReservationExpiry = &expiryTime

    return ticket, nil
}

// Factory method: VIP bilet oluştur
func (f *TicketFactory) CreateVIPTicket(req *TicketCreationRequest) (*models.Ticket, error) {
    req.TicketType = models.TicketTypeVIP

    ticket, _ := f.CreateTicket(req)

    // VIP biletler için yüksek kaliteli QR kod
    qrCodeImage, _ := f.qrGenerator.GenerateWithOptions(ticket.QRCodeData, 512, qrcode.High)
    ticket.QRCodeImage = qrCodeImage

    return ticket, nil
}

// Factory method: Grup bileti oluştur
func (f *TicketFactory) CreateGroupTickets(req *TicketCreationRequest, count int, seatIDs []int64) ([]*models.Ticket, error) {
    tickets := make([]*models.Ticket, 0, count)

    for i := 0; i < count; i++ {
        ticketReq := *req
        seatID := seatIDs[i]
        ticketReq.SeatID = &seatID

        ticket, _ := f.CreateTicket(&ticketReq)
        tickets = append(tickets, ticket)
    }

    return tickets, nil
}
```

**Kullanım (Service Layer):**

```go
// internal/services/ticket_service.go

func (s *TicketService) ReserveTicket(userID, eventID, sectionID int64, seatID *int64, price float64) (*models.Ticket, error) {
    // Business rules validation...

    // Factory kullanarak bilet oluştur
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

    ticket, _ := s.ticketFactory.CreateTicket(ticketReq)

    // Database'e kaydet
    ticketID, _ := s.ticketRepo.Create(ticket)
    ticket.ID = ticketID

    return ticket, nil
}
```

**Avantajlar:**
- ✅ Bilet oluşturma mantığı tek yerde (Single Responsibility)
- ✅ Yeni bilet tipleri eklemek kolay
- ✅ QR kod oluşturma test edilebilir (mock QR generator)
- ✅ Kod tekrarı yok

### 3. Observer Pattern (Gözlemci Deseni) 📢

**Kullanım Alanı:** Bildirim sistemi (Email, SMS, Analytics)

**Problem:**
Sistemde bir olay gerçekleştiğinde (bilet satın alındı, ödeme tamamlandı, rezervasyon süresi doldu) birden fazla işlem yapılması gerekir:
- Email gönder
- SMS gönder
- Analytics'e kaydet
- Bekleme listesine bildir

Bu işlemleri service içinde sırayla yazmak tight coupling yaratır ve test edilmesi zorlaşır.

**Çözüm:**
Observer Pattern ile event-driven bir yapı kurulur. Bir olay gerçekleştiğinde, tüm gözlemciler otomatik olarak bilgilendirilir.

**Kod Örneği:**

```go
// internal/patterns/observer/notification_observer.go

// Observer interface
type Observer interface {
    Update(event *EventData) error
    GetName() string
}

// Subject (Publisher)
type EventPublisher struct {
    observers []Observer
}

func (p *EventPublisher) Attach(observer Observer) {
    p.observers = append(p.observers, observer)
}

func (p *EventPublisher) Notify(event *EventData) {
    for _, observer := range p.observers {
        go func(obs Observer) {
            obs.Update(event)  // Asenkron olarak notify
        }(observer)
    }
}

// Concrete Observer 1: Email Notifications
type EmailNotificationObserver struct {
    EmailService EmailService
}

func (o *EmailNotificationObserver) Update(event *EventData) error {
    switch event.Type {
    case EventTypeTicketPurchased:
        return o.handleTicketPurchased(event)
    case EventTypeTicketCancelled:
        return o.handleTicketCancelled(event)
    case EventTypeWaitingListNotify:
        return o.handleWaitingListNotify(event)
    }
    return nil
}

func (o *EmailNotificationObserver) handleTicketPurchased(event *EventData) error {
    data := event.Data.(*TicketPurchaseData)

    subject := fmt.Sprintf("Biletiniz Başarıyla Satın Alındı - %s", data.EventName)
    body := fmt.Sprintf(`
Sayın %s,

%s etkinliği için biletiniz başarıyla satın alınmıştır.

Bilet No: %s
Doğrulama Kodu: %s
Koltuk: %s
Fiyat: %.2f TL

İyi eğlenceler!
`, data.UserEmail, data.EventName, data.TicketNumber, data.VerificationCode, data.SeatInfo, data.Price)

    return o.EmailService.SendEmail(data.UserEmail, subject, body)
}

// Concrete Observer 2: SMS Notifications
type SMSNotificationObserver struct {
    SMSService SMSService
}

func (o *SMSNotificationObserver) Update(event *EventData) error {
    switch event.Type {
    case EventTypeTicketPurchased:
        data := event.Data.(*TicketPurchaseData)
        message := fmt.Sprintf("Biletiniz alindi. Bilet No: %s, Dogrulama: %s",
            data.TicketNumber, data.VerificationCode)
        return o.SMSService.SendSMS(data.UserPhone, message)
    }
    return nil
}

// Concrete Observer 3: Analytics
type AnalyticsObserver struct {
    AnalyticsService AnalyticsService
}

func (o *AnalyticsObserver) Update(event *EventData) error {
    properties := map[string]interface{}{
        "timestamp":  event.Timestamp,
        "event_type": event.Type,
    }

    return o.AnalyticsService.TrackEvent(string(event.Type), userID, properties)
}
```

**Kullanım (Service Layer):**

```go
// internal/services/ticket_service.go

func (s *TicketService) PurchaseTicket(ticketID int64, userEmail, userPhone string) error {
    ticket, _ := s.ticketRepo.FindByID(ticketID)

    // Business logic...
    ticket.MarkAsSold()
    s.ticketRepo.Update(ticket)

    // Observer pattern: Tüm gözlemcileri bilgilendir
    s.eventPublisher.Notify(&observer.EventData{
        Type:      observer.EventTypeTicketPurchased,
        Timestamp: time.Now(),
        Data: &observer.TicketPurchaseData{
            UserID:           ticket.UserID,
            UserEmail:        userEmail,
            UserPhone:        userPhone,
            EventName:        event.Name,
            TicketNumber:     ticket.TicketNumber,
            VerificationCode: ticket.VerificationCode,
            Price:            ticket.Price,
        },
    })

    // Email, SMS, Analytics otomatik olarak tetiklenir!
    return nil
}
```

**Setup (Main):**

```go
// cmd/api/main.go

func main() {
    // Publisher oluştur
    eventPublisher := observer.NewEventPublisher()

    // Observers ekle
    emailObserver := observer.NewEmailNotificationObserver(emailService)
    smsObserver := observer.NewSMSNotificationObserver(smsService)
    analyticsObserver := observer.NewAnalyticsObserver(analyticsService)

    eventPublisher.Attach(emailObserver)
    eventPublisher.Attach(smsObserver)
    eventPublisher.Attach(analyticsObserver)

    // Service'leri oluştur
    ticketService := services.NewTicketService(ticketRepo, eventRepo, venueRepo, eventPublisher, db)
}
```

**Avantajlar:**
- ✅ Loose coupling (Service katmanı bildirim detaylarını bilmez)
- ✅ Yeni observer eklemek kolay (örn: Push notification)
- ✅ Asenkron çalışma (goroutine ile)
- ✅ Test edilebilir (mock observer)

### 4. State Pattern (Durum Deseni) 🎭

**Kullanım Alanı:** Bilet durumu yönetimi (lifecycle)

**Problem:**
Bir bilet birçok durumdan geçer:
- Reserved (Rezerve edildi)
- Sold (Satın alındı)
- Used (Kullanıldı)
- Cancelled (İptal edildi)
- Expired (Süresi doldu)

Her durumda farklı işlemler yapılabilir veya yapılamaz. Örneğin:
- Reserved bir bilet satın alınabilir, ama Used olamaz
- Sold bir bilet kullanılabilir, ama tekrar satın alınamaz
- Cancelled bir bilet kullanılamaz

Bu kontrolleri if-else ile yapmak karmaşık ve hata yapmaya açıktır.

**Çözüm:**
State Pattern ile her durum için geçerli işlemler açıkça tanımlanır. Geçersiz durum geçişleri otomatik olarak engellenir.

**Kod Örneği:**

```go
// internal/models/ticket.go

type TicketStatus string

const (
    TicketStatusReserved  TicketStatus = "reserved"
    TicketStatusSold      TicketStatus = "sold"
    TicketStatusUsed      TicketStatus = "used"
    TicketStatusCancelled TicketStatus = "cancelled"
    TicketStatusExpired   TicketStatus = "expired"
)

type Ticket struct {
    ID                int64
    Status            TicketStatus
    ReservationExpiry *time.Time
    PurchasedAt       *time.Time
    UsedAt            *time.Time
    CancelledAt       *time.Time
    // ... other fields
}

// State Pattern Methods

// CanPurchase checks if ticket can be purchased
func (t *Ticket) CanPurchase() bool {
    if t.Status != TicketStatusReserved {
        return false
    }

    if t.ReservationExpiry == nil {
        return false
    }

    return time.Now().Before(*t.ReservationExpiry)
}

// MarkAsSold transitions ticket to sold state
func (t *Ticket) MarkAsSold() error {
    if !t.CanPurchase() {
        return fmt.Errorf("bilet satın alınamaz: durum=%s", t.Status)
    }

    t.Status = TicketStatusSold
    now := time.Now()
    t.PurchasedAt = &now
    t.ReservationExpiry = nil

    return nil
}

// CanUse checks if ticket can be used
func (t *Ticket) CanUse() bool {
    return t.Status == TicketStatusSold
}

// MarkAsUsed transitions ticket to used state
func (t *Ticket) MarkAsUsed() error {
    if !t.CanUse() {
        return fmt.Errorf("bilet kullanılamaz: durum=%s", t.Status)
    }

    t.Status = TicketStatusUsed
    now := time.Now()
    t.UsedAt = &now

    return nil
}

// CanCancel checks if ticket can be cancelled
func (t *Ticket) CanCancel() bool {
    return t.Status == TicketStatusReserved || t.Status == TicketStatusSold
}

// MarkAsCancelled transitions ticket to cancelled state
func (t *Ticket) MarkAsCancelled() error {
    if !t.CanCancel() {
        return fmt.Errorf("bilet iptal edilemez: durum=%s", t.Status)
    }

    t.Status = TicketStatusCancelled
    now := time.Now()
    t.CancelledAt = &now

    return nil
}

// IsExpired checks if reservation has expired
func (t *Ticket) IsExpired() bool {
    if t.Status != TicketStatusReserved {
        return false
    }

    if t.ReservationExpiry == nil {
        return false
    }

    return time.Now().After(*t.ReservationExpiry)
}
```

**State Transition Diagram:**

```
                    ┌─────────────┐
                    │  RESERVED   │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
   ┌──────────┐     ┌───────────┐     ┌──────────┐
   │ EXPIRED  │     │   SOLD    │     │CANCELLED │
   └──────────┘     └─────┬─────┘     └──────────┘
                          │
                          ▼
                    ┌──────────┐
                    │  USED    │
                    └──────────┘
```

**Kullanım (Service Layer):**

```go
// internal/services/ticket_service.go

func (s *TicketService) PurchaseTicket(ticketID int64, userEmail, userPhone string) error {
    ticket, _ := s.ticketRepo.FindByID(ticketID)

    // State Pattern ile durum kontrolü
    if !ticket.CanPurchase() {
        return fmt.Errorf("bilet satın alınamaz durumda")
    }

    // Durum geçişi
    if err := ticket.MarkAsSold(); err != nil {
        return err
    }

    // Database'e kaydet
    s.ticketRepo.Update(ticket)

    return nil
}

func (s *TicketService) UseTicket(ticketNumber string) error {
    ticket, _ := s.ticketRepo.FindByTicketNumber(ticketNumber)

    // State Pattern ile durum kontrolü
    if !ticket.CanUse() {
        return fmt.Errorf("bilet kullanılamaz: %s", ticket.Status)
    }

    // Durum geçişi
    if err := ticket.MarkAsUsed(); err != nil {
        return err
    }

    s.ticketRepo.Update(ticket)

    return nil
}
```

**Avantajlar:**
- ✅ Geçersiz durum geçişleri engellenir
- ✅ Business rules açıkça tanımlı
- ✅ Her durum için izin verilen işlemler belirli
- ✅ Test edilebilir

## 🏛️ Clean Architecture Detayları

### Ultra-Thin Controllers

Controller'lar sadece 3 şey yapar:
1. Request parse et
2. Service'i çağır
3. Response dön

**Örnek:**

```go
// internal/controllers/ticket_controller.go

func (c *TicketController) Purchase(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    id, _ := parseIDFromPath(r.URL.Path, "/tickets/")

    var req struct {
        UserEmail string `json:"user_email"`
        UserPhone string `json:"user_phone"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Call service (ALL LOGIC HERE!)
    err := c.ticketService.PurchaseTicket(id, req.UserEmail, req.UserPhone)
    if err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // 3. Return response
    respondJSON(w, http.StatusOK, map[string]string{"message": "bilet satın alındı"})
}
```

**Neden Ultra-Thin?**
- ✅ Business logic service'te → HTTP'den bağımsız test
- ✅ Kod tekrarı yok
- ✅ Single Responsibility Principle
- ✅ Başka yerden (CLI, gRPC) aynı service kullanılabilir

### Service Layer: Business Logic Hub

Tüm iş mantığı burada:

```go
// internal/services/ticket_service.go

func (s *TicketService) ReserveTicket(userID, eventID, sectionID int64, seatID *int64, price float64) (*models.Ticket, error) {
    // 1. Business rule: Event validation
    event, _ := s.eventRepo.FindByID(eventID)
    if !event.IsSaleActive() {
        return nil, fmt.Errorf("bilet satışı aktif değil")
    }

    // 2. Business rule: Check availability
    if seatID != nil {
        isTaken, _ := s.ticketRepo.IsSeatTaken(eventID, *seatID)
        if isTaken {
            return nil, fmt.Errorf("koltuk dolu")
        }
    }

    // 3. Transaction: Prevent double booking
    tx, _ := s.db.Begin()
    defer tx.Rollback()

    s.eventRepo.DecrementAvailableSeats(eventID, 1)

    // 4. Factory Pattern: Create ticket
    ticket, _ := s.ticketFactory.CreateTicket(req)

    // 5. Save to database
    ticketID, _ := s.ticketRepo.Create(ticket)
    ticket.ID = ticketID

    tx.Commit()

    return ticket, nil
}
```

### Repository Layer: Data Access

Sadece database işlemleri:

```go
// internal/repositories/ticket_repository.go

func (r *TicketRepository) FindByID(id int64) (*models.Ticket, error) {
    query := `SELECT ... FROM tickets WHERE id = ?`

    ticket := &models.Ticket{}
    err := r.db.QueryRow(query, id).Scan(...)

    return ticket, err
}

func (r *TicketRepository) MarkAsSold(id int64) error {
    query := `UPDATE tickets SET status = ?, purchased_at = ? WHERE id = ?`

    _, err := r.db.Exec(query, models.TicketStatusSold, time.Now(), id)

    return err
}
```

## 🚀 Kurulum ve Çalıştırma

### Gereksinimler

- Go 1.22+
- MySQL 8.0+
- Redis 7+
- Docker & Docker Compose (opsiyonel)

### .env Dosyası

```bash
cp .env.example .env
```

```.env
# Application
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=3306
DB_NAME=event_ticketing
DB_USER=root
DB_PASSWORD=secret

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-super-secret-key-change-this-in-production
```

### Docker ile Çalıştırma (Önerilen)

```bash
# Build ve run
make docker-up

# Logs
make docker-logs

# Stop
make docker-down
```

### Manuel Kurulum

```bash
# Dependencies
go mod download

# Build
make build

# Run migrations
make migrate-up

# Run
make run
```

## 📚 API Endpoints

### Events

```bash
# Etkinlik oluştur
POST /events
{
  "name": "Tarkan Konseri",
  "description": "Unutulmaz bir gece",
  "type": "concert",
  "venue_id": 1,
  "start_time": "2024-06-15T20:00:00Z",
  "end_time": "2024-06-15T23:00:00Z",
  "base_price": 250.00,
  "featured": true
}

# Etkinlik listele
GET /events?status=sale_active&type=concert&page=1&page_size=20

# Etkinlik detay
GET /events/:id

# Etkinlik yayınla
POST /events/:id/publish

# Satışı aktif et
POST /events/:id/activate-sale

# Fiyat hesapla (Strategy Pattern kullanılır)
GET /events/:id/calculate-price?section_type=VIP
```

### Tickets

```bash
# Bilet rezerve et
POST /tickets/reserve
{
  "event_id": 1,
  "section_id": 2,
  "seat_id": 42,
  "price": 350.00
}

# Bilet satın al
POST /tickets/:id/purchase
{
  "user_email": "user@example.com",
  "user_phone": "+905551234567"
}

# Bilet iptal et
POST /tickets/:id/cancel
{
  "user_email": "user@example.com"
}

# Bilet doğrula (QR kod)
POST /tickets/validate
{
  "ticket_number": "TKT-20240520-abc123",
  "verification_code": "123456"
}

# Bilet kullan
POST /tickets/:id/use
{
  "ticket_number": "TKT-20240520-abc123"
}

# Kullanıcının biletleri
GET /tickets/my-tickets

# Etkinlik satış istatistikleri
GET /tickets/events/:id/stats
```

### Payments & Waiting List

```bash
# Ödeme oluştur
POST /payments
{
  "event_id": 1,
  "amount": 350.00,
  "payment_method": "credit_card"
}

# Bekleme listesine ekle
POST /waiting-lists
{
  "event_id": 1,
  "priority": 5
}

# Bekleme listesi bildir (sold-out'tan sonra iptal geldiğinde)
POST /waiting-lists/events/:id/notify
```

## 🧪 Testing

```bash
# Run all tests
make test

# Test with coverage
make test-coverage
```

**Test Örneği (Strategy Pattern):**

```go
func TestEarlyBirdPricing(t *testing.T) {
    strategy := &EarlyBirdPricingStrategy{
        DaysBeforeEvent: 30,
        DiscountPercent: 20,
    }

    context := &PricingContext{
        EventStartTime: time.Now().Add(45 * 24 * time.Hour),
        CurrentTime:    time.Now(),
    }

    price := strategy.CalculatePrice(100.0, context)

    expected := 80.0 // %20 indirim
    assert.Equal(t, expected, price)
}
```

## 📊 Database Schema

### Core Tables

- **venues**: Mekan bilgileri
- **sections**: Mekan bölümleri (VIP, Tribune, etc.)
- **seats**: Koltuklar
- **events**: Etkinlikler
- **tickets**: Biletler (State Pattern)
- **payments**: Ödemeler
- **waiting_lists**: Bekleme listeleri

### Key Relationships

```
venues (1) → (N) sections (1) → (N) seats
venues (1) → (N) events
events (1) → (N) tickets
seats (1) → (N) tickets
events (1) → (N) payments
events (1) → (N) waiting_lists
```

## 🔐 Güvenlik

- ✅ SQL Injection koruması (prepared statements)
- ✅ Input validation
- ✅ JWT authentication (placeholder)
- ✅ Rate limiting (Redis ile)
- ✅ Transaction management (double booking prevention)
- ✅ CORS configuration

## 🎯 Business Rules

### Bilet Rezervasyonu
- Rezervasyon 15 dakika geçerli
- Süresi dolan rezervasyonlar otomatik iptal
- Aynı koltuk için çift rezervasyon engelleniyor (transaction)

### Fiyatlandırma
- 30 gün öncesi: %20 erken rezervasyon indirimi
- VIP koltuklar: 2.5x premium
- Dinamik fiyat: Doluluk oranına göre 0.8x - 2.0x aralığında
- Hafta sonu: %15 ek fiyat

### İptal Politikası
- Etkinlikten 24 saat öncesine kadar iptal edilebilir
- İade süresi: 3-5 iş günü
- İptal edilen biletler bekleme listesine bildirilir

## 🌟 Öne Çıkan Özellikler

### 1. Double Booking Prevention
Transaction kullanarak aynı koltuğun iki kişiye satılması engelleniyor.

### 2. Dynamic Pricing
Strategy Pattern ile esnek fiyatlandırma. Yeni kurallar eklemek sadece yeni bir Strategy class yazmak kadar kolay.

### 3. QR Code Generation
Her bilet unique QR kod ile geliyor. Venue girişinde validate edilebilir.

### 4. Real-time Notifications
Observer Pattern ile bilet satın alındığında otomatik email, SMS ve analytics tracking.

### 5. Waiting List
Tükenen etkinlikler için akıllı bekleme listesi. İptal geldiğinde öncelik sırasına göre bilgilendirme.

## 🎓 Öğrenilecekler

Bu proje şu konuları öğrenmek için ideal:

1. **Clean Architecture**: Katmanlar arası bağımlılık yönetimi
2. **Design Patterns**: Strategy, Factory, Observer, State
3. **SOLID Principles**: Gerçek dünya uygulaması
4. **Domain-Driven Design**: Zengin domain modelleri
5. **Transaction Management**: Race condition önleme
6. **Test-Driven Development**: Pattern'ların test edilebilirliği

## 🤝 Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit atın (`git commit -m 'feat: Add amazing feature'`)
4. Push yapın (`git push origin feature/amazing-feature`)
5. Pull Request açın

## 📝 Lisans

MIT License

## 👨‍💻 Geliştirici

Bu proje [Conduit-Go](https://github.com/biyonik/conduit-go) framework kullanılarak geliştirilmiştir.

---

**Not:** Bu proje clean architecture ve design patterns öğrenmek için geliştirilmiş bir örnek projedir. Production'da kullanmadan önce güvenlik, ölçeklenebilirlik ve performans testlerini mutlaka yapın.

## 🔗 İlgili Projeler

- [Blog API](https://github.com/biyonik/blog-api-go) - Blog sistemi örneği
- [Task Management API](https://github.com/biyonik/task-management-api) - Clean Architecture örneği
- [Conduit-Go Framework](https://github.com/biyonik/conduit-go) - Laravel-inspired Go framework
