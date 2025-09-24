package services

import (
	"fmt"
	"time"

	"ticket-booking-system/internal/models"
	"ticket-booking-system/internal/repository"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo     repository.BookingRepositoryInterface
	eventRepo       repository.EventRepositoryInterface
	paymentDeadline int // in minutes
}

func NewBookingService(
	bookingRepo repository.BookingRepositoryInterface,
	eventRepo repository.EventRepositoryInterface,
	paymentDeadline int,
) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		eventRepo:       eventRepo,
		paymentDeadline: paymentDeadline,
	}
}

func (s *BookingService) CreateBooking(req *models.CreateBookingRequest) (*models.Booking, error) {
	// Parse UUIDs
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}

	// Get event details
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Check if event is in the future
	if event.DateTime.Before(time.Now()) {
		return nil, fmt.Errorf("cannot book tickets for past events")
	}

	// Reserve tickets with row locking to prevent race conditions
	err = s.eventRepo.ReserveTickets(eventID, req.Quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve tickets: %w", err)
	}

	// Calculate total amount
	totalAmount := float64(req.Quantity) * event.TicketPrice

	// Set payment deadline from config
	paymentDeadline := time.Now().Add(time.Duration(s.paymentDeadline) * time.Minute)

	// Create booking
	booking := &models.Booking{
		UserID:          userID,
		EventID:         eventID,
		Quantity:        req.Quantity,
		Status:          models.BookingStatusPending,
		TotalAmount:     totalAmount,
		PaymentDeadline: &paymentDeadline,
	}

	// Create booking in database
	err = s.bookingRepo.CreateWithTransaction(booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return booking, nil
}

func (s *BookingService) GetBooking(id uuid.UUID) (*models.Booking, error) {
	return s.bookingRepo.GetByID(id)
}

func (s *BookingService) GetUserBookings(userID uuid.UUID) ([]*models.Booking, error) {
	return s.bookingRepo.GetByUserID(userID)
}

func (s *BookingService) CancelBooking(id uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Only allow cancellation of pending bookings
	if booking.Status != models.BookingStatusPending {
		return fmt.Errorf("only pending bookings can be cancelled")
	}

	// Update booking status to cancelled
	return s.bookingRepo.UpdateStatus(id, models.BookingStatusCancelled)
}

func (s *BookingService) ConfirmBooking(id uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Only allow confirmation of pending bookings
	if booking.Status != models.BookingStatusPending {
		return fmt.Errorf("only pending bookings can be confirmed")
	}

	// Update booking status to confirmed
	return s.bookingRepo.UpdateStatus(id, models.BookingStatusConfirmed)
}
