package repository

import (
	"agenda-api/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	query := `
		INSERT INTO events (id, title, description, date, start_time, end_time, location, capacity, status, created_by, speaker_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowx(
		query,
		event.ID, event.Title, event.Description, event.Date, event.StartTime, event.EndTime,
		event.Location, event.Capacity, event.Status, event.CreatedBy, event.SpeakerID,
		event.CreatedAt, event.UpdatedAt,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	var event models.Event
	query := `SELECT * FROM events WHERE id = $1`
	err := r.db.Get(&event, query, id)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) GetAll(status *models.EventStatus) ([]models.EventWithAttendeeCount, error) {
	var events []models.EventWithAttendeeCount
	var query string
	var args []interface{}

	baseQuery := `
		SELECT e.*, COALESCE(COUNT(a.id) FILTER (WHERE a.status = 'registered'), 0) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id`

	if status != nil {
		query = baseQuery + ` WHERE e.status = $1 GROUP BY e.id ORDER BY e.date, e.start_time`
		args = append(args, *status)
	} else {
		query = baseQuery + ` GROUP BY e.id ORDER BY e.date, e.start_time`
	}

	err := r.db.Select(&events, query, args...)
	return events, err
}

func (r *EventRepository) GetByDateRange(start, end time.Time) ([]models.EventWithAttendeeCount, error) {
	var events []models.EventWithAttendeeCount
	query := `
		SELECT e.*, COALESCE(COUNT(a.id) FILTER (WHERE a.status = 'registered'), 0) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.date >= $1 AND e.date <= $2 AND e.status = 'published'
		GROUP BY e.id
		ORDER BY e.date, e.start_time`

	err := r.db.Select(&events, query, start, end)
	return events, err
}

func (r *EventRepository) Update(event *models.Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, date = $3, start_time = $4, end_time = $5,
		    location = $6, capacity = $7, status = $8, speaker_id = $9, updated_at = $10
		WHERE id = $11`

	event.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		query,
		event.Title, event.Description, event.Date, event.StartTime, event.EndTime,
		event.Location, event.Capacity, event.Status, event.SpeakerID, event.UpdatedAt, event.ID,
	)
	return err
}

func (r *EventRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *EventRepository) GetBySpeakerID(speakerID uuid.UUID) ([]models.Event, error) {
	var events []models.Event
	query := `SELECT * FROM events WHERE speaker_id = $1 ORDER BY date, start_time`
	err := r.db.Select(&events, query, speakerID)
	return events, err
}
