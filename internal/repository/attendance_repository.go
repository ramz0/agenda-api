package repository

import (
	"agenda-api/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AttendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Create(attendance *models.Attendance) error {
	query := `
		INSERT INTO attendance (id, event_id, user_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.db.QueryRowx(
		query,
		attendance.ID, attendance.EventID, attendance.UserID, attendance.Status, attendance.CreatedAt,
	).Scan(&attendance.ID, &attendance.CreatedAt)
}

func (r *AttendanceRepository) GetByEventAndUser(eventID, userID uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	query := `SELECT * FROM attendance WHERE event_id = $1 AND user_id = $2`
	err := r.db.Get(&attendance, query, eventID, userID)
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (r *AttendanceRepository) GetByEventID(eventID uuid.UUID) ([]models.AttendanceWithUser, error) {
	var attendances []models.AttendanceWithUser
	query := `
		SELECT a.*, u.name as user_name, u.email as user_email
		FROM attendance a
		JOIN users u ON a.user_id = u.id
		WHERE a.event_id = $1 AND a.status = 'registered'
		ORDER BY a.created_at`

	err := r.db.Select(&attendances, query, eventID)
	return attendances, err
}

func (r *AttendanceRepository) GetByUserID(userID uuid.UUID) ([]models.Attendance, error) {
	var attendances []models.Attendance
	query := `SELECT * FROM attendance WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&attendances, query, userID)
	return attendances, err
}

func (r *AttendanceRepository) UpdateStatus(id uuid.UUID, status models.AttendanceStatus) error {
	query := `UPDATE attendance SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *AttendanceRepository) Delete(eventID, userID uuid.UUID) error {
	query := `DELETE FROM attendance WHERE event_id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, eventID, userID)
	return err
}

func (r *AttendanceRepository) CountByEventID(eventID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM attendance WHERE event_id = $1 AND status = 'registered'`
	err := r.db.Get(&count, query, eventID)
	return count, err
}
