package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	DateTime     time.Time `json:"date_time" db:"date_time"`
	TotalTickets int       `json:"total_tickets" db:"total_tickets"`
	TicketPrice  float64   `json:"ticket_price" db:"ticket_price"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	UserID          uuid.UUID     `json:"user_id" db:"user_id"`
	EventID         uuid.UUID     `json:"event_id" db:"event_id"`
	Quantity        int           `json:"quantity" db:"quantity"`
	Status          BookingStatus `json:"status" db:"status"`
	TotalAmount     float64       `json:"total_amount" db:"total_amount"`
	PaymentDeadline *time.Time    `json:"payment_deadline" db:"payment_deadline"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type EventStatistics struct {
	EventID          uuid.UUID `json:"event_id"`
	TotalSold        int       `json:"total_sold"`
	EstimatedRevenue float64   `json:"estimated_revenue"`
	AvailableTickets int       `json:"available_tickets"`
}

type CreateEventRequest struct {
	Name         string    `json:"name" binding:"required"`
	Description  string    `json:"description"`
	DateTime     time.Time `json:"date_time" binding:"required"`
	TotalTickets int       `json:"total_tickets" binding:"required,min=1"`
	TicketPrice  float64   `json:"ticket_price" binding:"required,min=0"`
}

type UpdateEventRequest struct {
	Name         *string    `json:"name"`
	Description  *string    `json:"description"`
	DateTime     *time.Time `json:"date_time"`
	TotalTickets *int       `json:"total_tickets" binding:"omitempty,min=1"`
	TicketPrice  *float64   `json:"ticket_price" binding:"omitempty,min=0"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email" binding:"omitempty,email"`
}

type CreateBookingRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	EventID  string `json:"event_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}
