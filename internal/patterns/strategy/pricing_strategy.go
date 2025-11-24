package strategy

import (
	"time"
)

// PricingStrategy defines the interface for different pricing strategies
type PricingStrategy interface {
	CalculatePrice(basePrice float64, context *PricingContext) float64
	GetName() string
}

// PricingContext holds context information for pricing calculations
type PricingContext struct {
	EventStartTime    time.Time
	CurrentTime       time.Time
	OccupancyRate     float64 // 0.0 to 1.0
	SectionType       string  // "VIP", "Regular", "Balcony", etc.
	IsWeekend         bool
	RemainingCapacity int
	TotalCapacity     int
}

// EarlyBirdPricingStrategy - Discount for early purchases
type EarlyBirdPricingStrategy struct {
	DaysBeforeEvent int
	DiscountPercent float64
}

func (s *EarlyBirdPricingStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	daysUntilEvent := context.EventStartTime.Sub(context.CurrentTime).Hours() / 24

	if daysUntilEvent >= float64(s.DaysBeforeEvent) {
		discount := basePrice * (s.DiscountPercent / 100)
		return basePrice - discount
	}

	return basePrice
}

func (s *EarlyBirdPricingStrategy) GetName() string {
	return "Early Bird Pricing"
}

// VIPPricingStrategy - Premium pricing for VIP sections
type VIPPricingStrategy struct {
	PremiumMultiplier float64 // e.g., 2.5 for 150% markup
}

func (s *VIPPricingStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	if context.SectionType == "VIP" || context.SectionType == "Premium" {
		return basePrice * s.PremiumMultiplier
	}

	return basePrice
}

func (s *VIPPricingStrategy) GetName() string {
	return "VIP Pricing"
}

// DynamicPricingStrategy - Demand-based dynamic pricing
type DynamicPricingStrategy struct {
	MaxPriceMultiplier float64 // Maximum price increase (e.g., 2.0 for 100% increase)
	MinPriceMultiplier float64 // Minimum price (e.g., 0.7 for 30% discount)
}

func (s *DynamicPricingStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	// Price increases as occupancy increases
	priceMultiplier := s.MinPriceMultiplier + (context.OccupancyRate * (s.MaxPriceMultiplier - s.MinPriceMultiplier))

	price := basePrice * priceMultiplier

	// Weekend premium
	if context.IsWeekend {
		price *= 1.15 // 15% weekend premium
	}

	// Last-minute premium (less than 3 days before event)
	daysUntilEvent := context.EventStartTime.Sub(context.CurrentTime).Hours() / 24
	if daysUntilEvent < 3 && daysUntilEvent > 0 {
		price *= 1.10 // 10% last-minute premium
	}

	return price
}

func (s *DynamicPricingStrategy) GetName() string {
	return "Dynamic Pricing"
}

// SeasonalPricingStrategy - Seasonal adjustments
type SeasonalPricingStrategy struct {
	HighSeasonMonths []time.Month
	HighSeasonMarkup float64
}

func (s *SeasonalPricingStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	currentMonth := context.CurrentTime.Month()

	for _, month := range s.HighSeasonMonths {
		if currentMonth == month {
			return basePrice * (1 + s.HighSeasonMarkup)
		}
	}

	return basePrice
}

func (s *SeasonalPricingStrategy) GetName() string {
	return "Seasonal Pricing"
}

// GroupDiscountStrategy - Discounts for bulk purchases
type GroupDiscountStrategy struct {
	MinTickets      int
	DiscountPercent float64
}

func (s *GroupDiscountStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	// This would need ticket count in context, simplified for now
	return basePrice * (1 - s.DiscountPercent/100)
}

func (s *GroupDiscountStrategy) GetName() string {
	return "Group Discount"
}

// CompositePricingStrategy - Combines multiple strategies
type CompositePricingStrategy struct {
	Strategies []PricingStrategy
}

func (s *CompositePricingStrategy) CalculatePrice(basePrice float64, context *PricingContext) float64 {
	finalPrice := basePrice

	for _, strategy := range s.Strategies {
		finalPrice = strategy.CalculatePrice(finalPrice, context)
	}

	return finalPrice
}

func (s *CompositePricingStrategy) GetName() string {
	return "Composite Pricing"
}

// PricingStrategyFactory - Factory for creating pricing strategies
type PricingStrategyFactory struct{}

func NewPricingStrategyFactory() *PricingStrategyFactory {
	return &PricingStrategyFactory{}
}

func (f *PricingStrategyFactory) CreateEarlyBirdStrategy(daysBeforeEvent int, discountPercent float64) PricingStrategy {
	return &EarlyBirdPricingStrategy{
		DaysBeforeEvent: daysBeforeEvent,
		DiscountPercent: discountPercent,
	}
}

func (f *PricingStrategyFactory) CreateVIPStrategy(multiplier float64) PricingStrategy {
	return &VIPPricingStrategy{
		PremiumMultiplier: multiplier,
	}
}

func (f *PricingStrategyFactory) CreateDynamicStrategy(maxMultiplier, minMultiplier float64) PricingStrategy {
	return &DynamicPricingStrategy{
		MaxPriceMultiplier: maxMultiplier,
		MinPriceMultiplier: minMultiplier,
	}
}

func (f *PricingStrategyFactory) CreateSeasonalStrategy(highSeasonMonths []time.Month, markup float64) PricingStrategy {
	return &SeasonalPricingStrategy{
		HighSeasonMonths: highSeasonMonths,
		HighSeasonMarkup: markup,
	}
}

func (f *PricingStrategyFactory) CreateCompositeStrategy(strategies ...PricingStrategy) PricingStrategy {
	return &CompositePricingStrategy{
		Strategies: strategies,
	}
}
