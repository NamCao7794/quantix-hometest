package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ticket-booking-system/internal/models"
	"ticket-booking-system/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PaymentService struct {
	rdb         *redis.Client
	bookingRepo repository.BookingRepositoryInterface
}

type PaymentJob struct {
	BookingID uuid.UUID `json:"booking_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

func NewPaymentService(rdb *redis.Client, bookingRepo repository.BookingRepositoryInterface) *PaymentService {
	return &PaymentService{
		rdb:         rdb,
		bookingRepo: bookingRepo,
	}
}

func (s *PaymentService) ProcessPayment(bookingID uuid.UUID) error {
	// Simulate payment processing
	// In a real application, this would integrate with a payment gateway

	// Simulate random payment success/failure (80% success rate)
	// In production, this would be replaced with actual payment processing
	time.Sleep(2 * time.Second) // Simulate processing time

	// For demo purposes, we'll simulate successful payment
	// In reality, this would depend on the payment gateway response
	err := s.bookingRepo.UpdateStatus(bookingID, models.BookingStatusConfirmed)
	if err != nil {
		log.Printf("Failed to confirm booking %s: %v", bookingID, err)
		return err
	}

	log.Printf("Payment processed successfully for booking %s", bookingID)
	return nil
}

func (s *PaymentService) QueuePayment(bookingID uuid.UUID, amount float64) error {
	job := PaymentJob{
		BookingID: bookingID,
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal payment job: %w", err)
	}

	// Add job to Redis queue
	err = s.rdb.LPush(context.Background(), "payment_queue", jobData).Err()
	if err != nil {
		return fmt.Errorf("failed to queue payment job: %w", err)
	}

	log.Printf("Payment job queued for booking %s", bookingID)
	return nil
}

func (s *PaymentService) StartProcessor() {
	log.Println("Starting payment processor...")

	for {
		// Block and wait for jobs from the queue
		result, err := s.rdb.BRPop(context.Background(), 0, "payment_queue").Result()
		if err != nil {
			log.Printf("Error waiting for payment jobs: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Parse job data
		var job PaymentJob
		err = json.Unmarshal([]byte(result[1]), &job)
		if err != nil {
			log.Printf("Failed to unmarshal payment job: %v", err)
			continue
		}

		// Process payment
		err = s.ProcessPayment(job.BookingID)
		if err != nil {
			log.Printf("Failed to process payment for booking %s: %v", job.BookingID, err)
			// In production, you might want to retry failed payments or move them to a dead letter queue
		}
	}
}

func (s *PaymentService) ProcessExpiredBookings() error {
	// Get all expired bookings
	expiredBookings, err := s.bookingRepo.GetExpiredBookings()
	if err != nil {
		return fmt.Errorf("failed to get expired bookings: %w", err)
	}

	// Cancel expired bookings
	for _, booking := range expiredBookings {
		err := s.bookingRepo.UpdateStatus(booking.ID, models.BookingStatusCancelled)
		if err != nil {
			log.Printf("Failed to cancel expired booking %s: %v", booking.ID, err)
			continue
		}
		log.Printf("Cancelled expired booking %s", booking.ID)
	}

	return nil
}

func (s *PaymentService) StartExpiredBookingProcessor() {
	log.Println("Starting expired booking processor...")

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		err := s.ProcessExpiredBookings()
		if err != nil {
			log.Printf("Error processing expired bookings: %v", err)
		}
	}
}
