package services

import (
	"ticket-booking-system/internal/models"
	"ticket-booking-system/internal/repository"

	"github.com/google/uuid"
)

type EventService struct {
	eventRepo repository.EventRepositoryInterface
}

func NewEventService(eventRepo repository.EventRepositoryInterface) *EventService {
	return &EventService{
		eventRepo: eventRepo,
	}
}

func (s *EventService) CreateEvent(req *models.CreateEventRequest) (*models.Event, error) {
	event := &models.Event{
		Name:         req.Name,
		Description:  req.Description,
		DateTime:     req.DateTime,
		TotalTickets: req.TotalTickets,
		TicketPrice:  req.TicketPrice,
	}

	err := s.eventRepo.Create(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(id uuid.UUID) (*models.Event, error) {
	return s.eventRepo.GetByID(id)
}

func (s *EventService) GetEvents() ([]*models.Event, error) {
	return s.eventRepo.GetAll()
}

func (s *EventService) UpdateEvent(id uuid.UUID, req *models.UpdateEventRequest) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.DateTime != nil {
		event.DateTime = *req.DateTime
	}
	if req.TotalTickets != nil {
		event.TotalTickets = *req.TotalTickets
	}
	if req.TicketPrice != nil {
		event.TicketPrice = *req.TicketPrice
	}

	err = s.eventRepo.Update(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) DeleteEvent(id uuid.UUID) error {
	return s.eventRepo.Delete(id)
}

func (s *EventService) GetEventStatistics(id uuid.UUID) (*models.EventStatistics, error) {
	return s.eventRepo.GetStatistics(id)
}
