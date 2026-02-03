package repository

import (
	"agenda-api/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AssignmentRepository struct {
	db *sqlx.DB
}

func NewAssignmentRepository(db *sqlx.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Create(assignment *models.EventAssignment) error {
	query := `
		INSERT INTO event_assignments (id, event_id, user_id, status, assigned_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, assigned_at`

	return r.db.QueryRowx(
		query,
		assignment.ID, assignment.EventID, assignment.UserID, assignment.Status, assignment.AssignedAt,
	).Scan(&assignment.ID, &assignment.AssignedAt)
}

func (r *AssignmentRepository) CreateBatch(assignments []models.EventAssignment) error {
	query := `
		INSERT INTO event_assignments (id, event_id, user_id, status, assigned_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id, user_id) DO NOTHING`

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	for _, a := range assignments {
		_, err := tx.Exec(query, a.ID, a.EventID, a.UserID, a.Status, a.AssignedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *AssignmentRepository) GetByID(id uuid.UUID) (*models.EventAssignment, error) {
	var assignment models.EventAssignment
	query := `SELECT * FROM event_assignments WHERE id = $1`
	err := r.db.Get(&assignment, query, id)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *AssignmentRepository) GetByEventAndUser(eventID, userID uuid.UUID) (*models.EventAssignment, error) {
	var assignment models.EventAssignment
	query := `SELECT * FROM event_assignments WHERE event_id = $1 AND user_id = $2`
	err := r.db.Get(&assignment, query, eventID, userID)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *AssignmentRepository) GetByUserID(userID uuid.UUID, status *models.AssignmentStatus) ([]models.EventAssignmentWithDetails, error) {
	var assignments []models.EventAssignmentWithDetails
	var query string
	var args []interface{}

	baseQuery := `
		SELECT ea.*, u.name as user_name, u.email as user_email,
		       e.title as event_title, e.date::text as event_date
		FROM event_assignments ea
		INNER JOIN users u ON ea.user_id = u.id
		INNER JOIN events e ON ea.event_id = e.id
		WHERE ea.user_id = $1`

	if status != nil {
		query = baseQuery + ` AND ea.status = $2 ORDER BY e.date, e.start_time`
		args = append(args, userID, *status)
	} else {
		query = baseQuery + ` ORDER BY e.date, e.start_time`
		args = append(args, userID)
	}

	err := r.db.Select(&assignments, query, args...)
	return assignments, err
}

func (r *AssignmentRepository) GetByEventID(eventID uuid.UUID) ([]models.EventAssignmentWithDetails, error) {
	var assignments []models.EventAssignmentWithDetails
	query := `
		SELECT ea.*, u.name as user_name, u.email as user_email,
		       e.title as event_title, e.date::text as event_date
		FROM event_assignments ea
		INNER JOIN users u ON ea.user_id = u.id
		INNER JOIN events e ON ea.event_id = e.id
		WHERE ea.event_id = $1
		ORDER BY u.name`
	err := r.db.Select(&assignments, query, eventID)
	return assignments, err
}

func (r *AssignmentRepository) UpdateStatus(id uuid.UUID, status models.AssignmentStatus) error {
	query := `
		UPDATE event_assignments
		SET status = $1, responded_at = $2
		WHERE id = $3`
	now := time.Now()
	_, err := r.db.Exec(query, status, now, id)
	return err
}

func (r *AssignmentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM event_assignments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *AssignmentRepository) DeleteByEventID(eventID uuid.UUID) error {
	query := `DELETE FROM event_assignments WHERE event_id = $1`
	_, err := r.db.Exec(query, eventID)
	return err
}

func (r *AssignmentRepository) GetPendingCountByUserID(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM event_assignments WHERE user_id = $1 AND status = 'pending'`
	err := r.db.Get(&count, query, userID)
	return count, err
}
