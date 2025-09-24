package repository

import (
	"database/sql"
	"fmt"

	"ticket-booking-system/internal/models"

	"github.com/google/uuid"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	query := `
		INSERT INTO events (name, description, date_time, total_tickets, ticket_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		event.Name,
		event.Description,
		event.DateTime,
		event.TotalTickets,
		event.TicketPrice,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	return err
}

func (r *EventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	query := `
		SELECT id, name, description, date_time, total_tickets, ticket_price, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	event := &models.Event{}
	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Description,
		&event.DateTime,
		&event.TotalTickets,
		&event.TicketPrice,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) GetAll() ([]*models.Event, error) {
	query := `
		SELECT id, name, description, date_time, total_tickets, ticket_price, created_at, updated_at
		FROM events
		ORDER BY date_time ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Description,
			&event.DateTime,
			&event.TotalTickets,
			&event.TicketPrice,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) Update(event *models.Event) error {
	query := `
		UPDATE events
		SET name = $2, description = $3, date_time = $4, total_tickets = $5, ticket_price = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		event.ID,
		event.Name,
		event.Description,
		event.DateTime,
		event.TotalTickets,
		event.TicketPrice,
	).Scan(&event.UpdatedAt)

	return err
}

func (r *EventRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *EventRepository) GetStatistics(eventID uuid.UUID) (*models.EventStatistics, error) {
	query := `
		SELECT 
			e.id,
			e.total_tickets,
			COALESCE(SUM(CASE WHEN b.status = 'CONFIRMED' THEN b.quantity ELSE 0 END), 0) as total_sold,
			COALESCE(SUM(CASE WHEN b.status = 'CONFIRMED' THEN b.total_amount ELSE 0 END), 0) as estimated_revenue
		FROM events e
		LEFT JOIN bookings b ON e.id = b.event_id
		WHERE e.id = $1
		GROUP BY e.id, e.total_tickets
	`

	stats := &models.EventStatistics{}
	err := r.db.QueryRow(query, eventID).Scan(
		&stats.EventID,
		&stats.TotalSold,
		&stats.EstimatedRevenue,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, err
	}

	// Get total tickets from the event
	eventQuery := `SELECT total_tickets FROM events WHERE id = $1`
	var totalTickets int
	err = r.db.QueryRow(eventQuery, eventID).Scan(&totalTickets)
	if err != nil {
		return nil, err
	}

	stats.AvailableTickets = totalTickets - stats.TotalSold

	return stats, nil
}

func (r *EventRepository) GetAvailableTickets(eventID uuid.UUID) (int, error) {
	query := `
		SELECT 
			e.total_tickets - COALESCE(SUM(CASE WHEN b.status IN ('PENDING', 'CONFIRMED') THEN b.quantity ELSE 0 END), 0) as available_tickets
		FROM events e
		LEFT JOIN bookings b ON e.id = b.event_id
		WHERE e.id = $1
		GROUP BY e.total_tickets
	`

	var availableTickets int
	err := r.db.QueryRow(query, eventID).Scan(&availableTickets)
	if err != nil {
		return 0, err
	}

	return availableTickets, nil
}

func (r *EventRepository) ReserveTickets(eventID uuid.UUID, quantity int) error {
	// Use a transaction with row locking to prevent race conditions
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the event row for update
	query := `SELECT total_tickets FROM events WHERE id = $1 FOR UPDATE`
	var totalTickets int
	err = tx.QueryRow(query, eventID).Scan(&totalTickets)
	if err != nil {
		return err
	}

	// Check available tickets
	availableQuery := `
		SELECT 
			$1 - COALESCE(SUM(CASE WHEN status IN ('PENDING', 'CONFIRMED') THEN quantity ELSE 0 END), 0) as available_tickets
		FROM bookings
		WHERE event_id = $2
	`

	var availableTickets int
	err = tx.QueryRow(availableQuery, totalTickets, eventID).Scan(&availableTickets)
	if err != nil {
		return err
	}

	if availableTickets < quantity {
		return fmt.Errorf("insufficient tickets available. Available: %d, Requested: %d", availableTickets, quantity)
	}

	// Commit the transaction
	return tx.Commit()
}
