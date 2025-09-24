package repository

import (
	"database/sql"
	"fmt"
	"time"

	"ticket-booking-system/internal/models"

	"github.com/google/uuid"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(booking *models.Booking) error {
	query := `
		INSERT INTO bookings (user_id, event_id, quantity, status, total_amount, payment_deadline)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		booking.UserID,
		booking.EventID,
		booking.Quantity,
		booking.Status,
		booking.TotalAmount,
		booking.PaymentDeadline,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

	return err
}

func (r *BookingRepository) GetByID(id uuid.UUID) (*models.Booking, error) {
	query := `
		SELECT id, user_id, event_id, quantity, status, total_amount, payment_deadline, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	booking := &models.Booking{}
	err := r.db.QueryRow(query, id).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.EventID,
		&booking.Quantity,
		&booking.Status,
		&booking.TotalAmount,
		&booking.PaymentDeadline,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	return booking, nil
}

func (r *BookingRepository) GetByUserID(userID uuid.UUID) ([]*models.Booking, error) {
	query := `
		SELECT id, user_id, event_id, quantity, status, total_amount, payment_deadline, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		booking := &models.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.EventID,
			&booking.Quantity,
			&booking.Status,
			&booking.TotalAmount,
			&booking.PaymentDeadline,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

func (r *BookingRepository) UpdateStatus(id uuid.UUID, status models.BookingStatus) error {
	query := `
		UPDATE bookings
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id, status)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}

	return nil
}

func (r *BookingRepository) GetPendingBookings() ([]*models.Booking, error) {
	query := `
		SELECT id, user_id, event_id, quantity, status, total_amount, payment_deadline, created_at, updated_at
		FROM bookings
		WHERE status = 'PENDING' AND payment_deadline IS NOT NULL
		ORDER BY payment_deadline ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		booking := &models.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.EventID,
			&booking.Quantity,
			&booking.Status,
			&booking.TotalAmount,
			&booking.PaymentDeadline,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

func (r *BookingRepository) GetExpiredBookings() ([]*models.Booking, error) {
	query := `
		SELECT id, user_id, event_id, quantity, status, total_amount, payment_deadline, created_at, updated_at
		FROM bookings
		WHERE status = 'PENDING' AND payment_deadline < $1
		ORDER BY payment_deadline ASC
	`

	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		booking := &models.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.EventID,
			&booking.Quantity,
			&booking.Status,
			&booking.TotalAmount,
			&booking.PaymentDeadline,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

func (r *BookingRepository) CreateWithTransaction(booking *models.Booking) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert booking
	query := `
		INSERT INTO bookings (user_id, event_id, quantity, status, total_amount, payment_deadline)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		query,
		booking.UserID,
		booking.EventID,
		booking.Quantity,
		booking.Status,
		booking.TotalAmount,
		booking.PaymentDeadline,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}
