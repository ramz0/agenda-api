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
		INSERT INTO events (id, title, description, date, start_time, end_time, location, capacity, status, type, team_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowx(
		query,
		event.ID, event.Title, event.Description, event.Date, event.StartTime, event.EndTime,
		event.Location, event.Capacity, event.Status, event.Type, event.TeamID, event.CreatedBy,
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
		SELECT e.*, t.name as team_name,
		       COALESCE(COUNT(a.id) FILTER (WHERE a.status = 'registered'), 0) as attendee_count
		FROM events e
		LEFT JOIN teams t ON e.team_id = t.id
		LEFT JOIN attendance a ON e.id = a.event_id`

	if status != nil {
		query = baseQuery + ` WHERE e.status = $1 GROUP BY e.id, t.name ORDER BY e.date, e.start_time`
		args = append(args, *status)
	} else {
		query = baseQuery + ` GROUP BY e.id, t.name ORDER BY e.date, e.start_time`
	}

	err := r.db.Select(&events, query, args...)
	return events, err
}

func (r *EventRepository) GetByDateRange(start, end time.Time) ([]models.EventWithAttendeeCount, error) {
	var events []models.EventWithAttendeeCount
	query := `
		SELECT e.*, t.name as team_name,
		       COALESCE(COUNT(a.id) FILTER (WHERE a.status = 'registered'), 0) as attendee_count
		FROM events e
		LEFT JOIN teams t ON e.team_id = t.id
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.date >= $1 AND e.date <= $2 AND e.status = 'published'
		GROUP BY e.id, t.name
		ORDER BY e.date, e.start_time`

	err := r.db.Select(&events, query, start, end)
	return events, err
}

func (r *EventRepository) Update(event *models.Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, date = $3, start_time = $4, end_time = $5,
		    location = $6, capacity = $7, status = $8, type = $9, team_id = $10, updated_at = $11
		WHERE id = $12`

	event.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		query,
		event.Title, event.Description, event.Date, event.StartTime, event.EndTime,
		event.Location, event.Capacity, event.Status, event.Type, event.TeamID, event.UpdatedAt, event.ID,
	)
	return err
}

func (r *EventRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *EventRepository) GetByCreatedBy(userID uuid.UUID) ([]models.Event, error) {
	var events []models.Event
	query := `SELECT * FROM events WHERE created_by = $1 ORDER BY date, start_time`
	err := r.db.Select(&events, query, userID)
	return events, err
}

func (r *EventRepository) GetPersonalByUserID(userID uuid.UUID) ([]models.Event, error) {
	var events []models.Event
	query := `SELECT * FROM events WHERE type = 'personal' AND created_by = $1 ORDER BY date, start_time`
	err := r.db.Select(&events, query, userID)
	return events, err
}

func (r *EventRepository) GetTeamEventsByUserID(userID uuid.UUID) ([]models.EventWithAssignment, error) {
	var events []models.EventWithAssignment
	query := `
		SELECT e.*, ea.status as assignment_status, t.name as team_name
		FROM events e
		INNER JOIN event_assignments ea ON e.id = ea.event_id
		LEFT JOIN teams t ON e.team_id = t.id
		WHERE ea.user_id = $1 AND e.type = 'team'
		ORDER BY e.date, e.start_time`
	err := r.db.Select(&events, query, userID)
	return events, err
}

func (r *EventRepository) GetByTeamID(teamID uuid.UUID) ([]models.Event, error) {
	var events []models.Event
	query := `SELECT * FROM events WHERE team_id = $1 ORDER BY date, start_time`
	err := r.db.Select(&events, query, teamID)
	return events, err
}

func (r *EventRepository) GetCalendarByUserID(userID uuid.UUID, start, end time.Time) ([]models.EventWithAssignment, error) {
	var events []models.EventWithAssignment
	query := `
		SELECT e.*, ea.status as assignment_status, t.name as team_name
		FROM events e
		LEFT JOIN event_assignments ea ON e.id = ea.event_id AND ea.user_id = $1
		LEFT JOIN teams t ON e.team_id = t.id
		WHERE e.date >= $2 AND e.date <= $3
		  AND e.status = 'published'
		  AND (
		    (e.type = 'personal' AND e.created_by = $1)
		    OR (e.type = 'team' AND ea.user_id = $1)
		    OR (ea.user_id = $1 AND ea.status = 'approved')
		  )
		ORDER BY e.date, e.start_time`
	err := r.db.Select(&events, query, userID, start, end)
	return events, err
}

func (r *EventRepository) GetByIDWithParticipants(id uuid.UUID) (*models.EventWithParticipants, error) {
	var event models.EventWithParticipants
	query := `
		SELECT e.*, t.name as team_name,
		       COALESCE(COUNT(a.id) FILTER (WHERE a.status = 'registered'), 0) as attendee_count
		FROM events e
		LEFT JOIN teams t ON e.team_id = t.id
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.id = $1
		GROUP BY e.id, t.name`
	err := r.db.Get(&event, query, id)
	if err != nil {
		return nil, err
	}

	// Get participants
	participants, err := r.GetParticipants(id)
	if err == nil {
		event.Participants = participants
	}

	return &event, nil
}

func (r *EventRepository) GetParticipants(eventID uuid.UUID) ([]models.EventParticipant, error) {
	var participants []models.EventParticipant
	query := `
		SELECT ea.user_id, u.name as user_name, u.email as user_email
		FROM event_assignments ea
		INNER JOIN users u ON ea.user_id = u.id
		WHERE ea.event_id = $1`
	err := r.db.Select(&participants, query, eventID)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (r *EventRepository) SetParticipants(eventID uuid.UUID, participantIds []uuid.UUID) error {
	// Delete existing participants for this event
	_, err := r.db.Exec(`DELETE FROM event_assignments WHERE event_id = $1`, eventID)
	if err != nil {
		return err
	}

	// Insert new participants
	if len(participantIds) > 0 {
		query := `
			INSERT INTO event_assignments (id, event_id, user_id, status, assigned_at)
			VALUES ($1, $2, $3, 'approved', $4)`
		for _, userID := range participantIds {
			_, err := r.db.Exec(query, uuid.New(), eventID, userID, time.Now())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
