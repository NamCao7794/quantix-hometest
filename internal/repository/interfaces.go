package repository

import (
	"ticket-booking-system/internal/models"

	"github.com/google/uuid"
)

type EventRepositoryInterface interface {
	Create(event *models.Event) error
	GetByID(id uuid.UUID) (*models.Event, error)
	GetAll() ([]*models.Event, error)
	Update(event *models.Event) error
	Delete(id uuid.UUID) error
	GetStatistics(eventID uuid.UUID) (*models.EventStatistics, error)
	GetAvailableTickets(eventID uuid.UUID) (int, error)
	ReserveTickets(eventID uuid.UUID, quantity int) error
}

type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
}

type BookingRepositoryInterface interface {
	Create(booking *models.Booking) error
	GetByID(id uuid.UUID) (*models.Booking, error)
	GetByUserID(userID uuid.UUID) ([]*models.Booking, error)
	UpdateStatus(id uuid.UUID, status models.BookingStatus) error
	GetPendingBookings() ([]*models.Booking, error)
	GetExpiredBookings() ([]*models.Booking, error)
	CreateWithTransaction(booking *models.Booking) error
}
