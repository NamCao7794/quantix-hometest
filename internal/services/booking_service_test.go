package services

import (
	"testing"
	"time"

	"ticket-booking-system/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) Create(booking *models.Booking) error {
	args := m.Called(booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetByID(id uuid.UUID) (*models.Booking, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetByUserID(userID uuid.UUID) ([]*models.Booking, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) UpdateStatus(id uuid.UUID, status models.BookingStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockBookingRepository) GetPendingBookings() ([]*models.Booking, error) {
	args := m.Called()
	return args.Get(0).([]*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetExpiredBookings() ([]*models.Booking, error) {
	args := m.Called()
	return args.Get(0).([]*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) CreateWithTransaction(booking *models.Booking) error {
	args := m.Called(booking)
	return args.Error(0)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventRepository) ReserveTickets(eventID uuid.UUID, quantity int) error {
	args := m.Called(eventID, quantity)
	return args.Error(0)
}

func (m *MockEventRepository) Create(event *models.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) GetAll() ([]*models.Event, error) {
	args := m.Called()
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *MockEventRepository) Update(event *models.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEventRepository) GetStatistics(eventID uuid.UUID) (*models.EventStatistics, error) {
	args := m.Called(eventID)
	return args.Get(0).(*models.EventStatistics), args.Error(1)
}

func (m *MockEventRepository) GetAvailableTickets(eventID uuid.UUID) (int, error) {
	args := m.Called(eventID)
	return args.Int(0), args.Error(1)
}

func TestBookingService_CreateBooking_Success(t *testing.T) {
	// Setup
	mockBookingRepo := &MockBookingRepository{}
	mockEventRepo := &MockEventRepository{}
	service := NewBookingService(mockBookingRepo, mockEventRepo, 15)

	userID := uuid.New()
	eventID := uuid.New()
	quantity := 2

	// Mock event
	event := &models.Event{
		ID:           eventID,
		Name:         "Test Event",
		DateTime:     time.Now().Add(24 * time.Hour),
		TotalTickets: 100,
		TicketPrice:  50.0,
	}

	// Mock expectations
	mockEventRepo.On("GetByID", eventID).Return(event, nil)
	mockEventRepo.On("ReserveTickets", eventID, quantity).Return(nil)
	mockBookingRepo.On("CreateWithTransaction", mock.AnythingOfType("*models.Booking")).Return(nil)

	// Test
	req := &models.CreateBookingRequest{
		UserID:   userID.String(),
		EventID:  eventID.String(),
		Quantity: quantity,
	}

	booking, err := service.CreateBooking(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, eventID, booking.EventID)
	assert.Equal(t, quantity, booking.Quantity)
	assert.Equal(t, models.BookingStatusPending, booking.Status)
	assert.Equal(t, 100.0, booking.TotalAmount) // 2 * 50.0
	assert.NotNil(t, booking.PaymentDeadline)

	// Verify all expectations were met
	mockEventRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_PastEvent(t *testing.T) {
	// Setup
	mockBookingRepo := &MockBookingRepository{}
	mockEventRepo := &MockEventRepository{}
	service := NewBookingService(mockBookingRepo, mockEventRepo, 15)

	userID := uuid.New()
	eventID := uuid.New()

	// Mock past event
	event := &models.Event{
		ID:           eventID,
		Name:         "Past Event",
		DateTime:     time.Now().Add(-24 * time.Hour), // Past event
		TotalTickets: 100,
		TicketPrice:  50.0,
	}

	// Mock expectations
	mockEventRepo.On("GetByID", eventID).Return(event, nil)

	// Test
	req := &models.CreateBookingRequest{
		UserID:   userID.String(),
		EventID:  eventID.String(),
		Quantity: 2,
	}

	booking, err := service.CreateBooking(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "cannot book tickets for past events")

	// Verify expectations
	mockEventRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_InsufficientTickets(t *testing.T) {
	// Setup
	mockBookingRepo := &MockBookingRepository{}
	mockEventRepo := &MockEventRepository{}
	service := NewBookingService(mockBookingRepo, mockEventRepo, 15)

	userID := uuid.New()
	eventID := uuid.New()

	// Mock event
	event := &models.Event{
		ID:           eventID,
		Name:         "Test Event",
		DateTime:     time.Now().Add(24 * time.Hour),
		TotalTickets: 100,
		TicketPrice:  50.0,
	}

	// Mock expectations
	mockEventRepo.On("GetByID", eventID).Return(event, nil)
	mockEventRepo.On("ReserveTickets", eventID, 150).Return(assert.AnError) // Simulate insufficient tickets

	// Test
	req := &models.CreateBookingRequest{
		UserID:   userID.String(),
		EventID:  eventID.String(),
		Quantity: 150, // More than available
	}

	booking, err := service.CreateBooking(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "failed to reserve tickets")

	// Verify expectations
	mockEventRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_Success(t *testing.T) {
	// Setup
	mockBookingRepo := &MockBookingRepository{}
	mockEventRepo := &MockEventRepository{}
	service := NewBookingService(mockBookingRepo, mockEventRepo, 15)

	bookingID := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()

	// Mock pending booking
	booking := &models.Booking{
		ID:      bookingID,
		UserID:  userID,
		EventID: eventID,
		Status:  models.BookingStatusPending,
	}

	// Mock expectations
	mockBookingRepo.On("GetByID", bookingID).Return(booking, nil)
	mockBookingRepo.On("UpdateStatus", bookingID, models.BookingStatusCancelled).Return(nil)

	// Test
	err := service.CancelBooking(bookingID)

	// Assertions
	assert.NoError(t, err)

	// Verify expectations
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_NotPending(t *testing.T) {
	// Setup
	mockBookingRepo := &MockBookingRepository{}
	mockEventRepo := &MockEventRepository{}
	service := NewBookingService(mockBookingRepo, mockEventRepo, 15)

	bookingID := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()

	// Mock confirmed booking
	booking := &models.Booking{
		ID:      bookingID,
		UserID:  userID,
		EventID: eventID,
		Status:  models.BookingStatusConfirmed,
	}

	// Mock expectations
	mockBookingRepo.On("GetByID", bookingID).Return(booking, nil)

	// Test
	err := service.CancelBooking(bookingID)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only pending bookings can be cancelled")

	// Verify expectations
	mockBookingRepo.AssertExpectations(t)
}
